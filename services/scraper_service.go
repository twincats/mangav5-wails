package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

const DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"

// ScraperService handles rule-based scraping
type ScraperService struct {
	browserService *BrowserService
	client         *resty.Client
}

// NewScraperService creates a new instance
func NewScraperService(bs *BrowserService) *ScraperService {
	client := resty.New()
	client.SetCookieJar(nil) // Enable CookieJar
	client.SetHeader("User-Agent", DefaultUserAgent)
	client.SetTimeout(30 * time.Second)

	return &ScraperService{
		browserService: bs,
		client:         client,
	}
}

// Scrape executes the scraping rule
func (s *ScraperService) Scrape(rule SiteRule, overrideURL string) (map[string]interface{}, error) {
	targetURL := overrideURL
	params := make(map[string]interface{})

	// Handle Override URL (User Input)
	if targetURL != "" {
		// Parse URL to handle query parameters (like offset, limit)
		// We do this before regex extraction so that ?offset=100 doesn't pollute the ID extraction
		u, err := url.Parse(targetURL)
		if err == nil {
			// Extract query parameters
			query := u.Query()
			for k, v := range query {
				if len(v) > 0 {
					params[k] = v[0]
				}
			}

			// Use the path part for ID extraction logic
			// Reconstruct URL without query params for the pattern matching
			// Note: We keep the scheme and host if present
			targetURLWithoutQuery := targetURL
			if strings.Contains(targetURL, "?") {
				targetURLWithoutQuery = strings.Split(targetURL, "?")[0]
			}

			// Use the cleaner URL for subsequent ID extraction
			targetURL = targetURLWithoutQuery
		}

		// Case 1: Input is a Full URL (e.g. https://site.com/manga/id/)
		if strings.HasPrefix(targetURL, "http") {
			// Try to extract params using Entry URL as a template
			if rule.Entry != nil && rule.Entry.URL != "" {
				extracted := s.extractParamsFromURLTemplate(rule.Entry.URL, targetURL)
				for k, v := range extracted {
					params[k] = v
				}
			}

			// If rule has an entry regex, try to extract ID from the full URL (Override/Supplement)
			if rule.Entry != nil && rule.Entry.Regex != "" {
				re, err := regexp.Compile(rule.Entry.Regex)
				if err == nil {
					matches := re.FindStringSubmatch(targetURL)
					names := re.SubexpNames()
					if len(matches) > 0 {
						for i, name := range names {
							if i > 0 && i < len(matches) && name != "" {
								params[name] = matches[i]
							}
						}
					}
				}
			}

			// If "id" is still missing, fallback to legacy suffix matching logic?
			// The extractParamsFromURLTemplate should cover it, but let's keep the legacy logic as a safe fallback
			// if the template matching failed (e.g. slightly different URL structure).
			if _, ok := params["id"]; !ok && rule.Entry != nil && strings.Contains(rule.Entry.URL, "{id}") {
				// Simple suffix match
				prefix := strings.Split(rule.Entry.URL, "{id}")[0]
				if strings.HasPrefix(targetURL, prefix) {
					idPart := strings.TrimPrefix(targetURL, prefix)
					idPart = strings.TrimSuffix(idPart, "/")
					if idPart != "" {
						params["id"] = idPart
					}
				}
			}

			params["url"] = targetURL
		} else {
			// Case 2: Input is just an ID (e.g. my-manga-id)
			params["id"] = targetURL

			// Construct full URL if Entry URL template exists
			if rule.Entry != nil && strings.Contains(rule.Entry.URL, "{id}") {
				targetURL = strings.ReplaceAll(rule.Entry.URL, "{id}", targetURL)
				params["url"] = targetURL
			}
		}
	} else if rule.Entry != nil {
		// Case 3: No override, use default Entry URL
		targetURL = rule.Entry.URL
		params["url"] = targetURL
	}

	switch rule.Strategy {
	case "static":
		return s.scrapeStatic(targetURL, rule, params)
	case "browser":
		return s.scrapeBrowser(targetURL, rule, params)
	case "api":
		return s.scrapeAPI(targetURL, rule, params)
	case "auto":
		// Default to static, fallback logic could be added here
		return s.scrapeStatic(targetURL, rule, params)
	default:
		return nil, fmt.Errorf("unknown strategy: %s", rule.Strategy)
	}
}

func (s *ScraperService) extractParamsFromURLTemplate(templateStr, urlStr string) map[string]interface{} {
	params := make(map[string]interface{})

	// Normalize: remove trailing slash from template for robust matching
	// We will allow optional slash in the regex
	templateStr = strings.TrimSuffix(templateStr, "/")

	// 1. Escape the template to make it safe for regex
	regexStr := regexp.QuoteMeta(templateStr)

	// 2. Replace {id} placeholder with named capture group
	// \{id\} is the escaped version of {id}
	// We use [^/]+ to match the segment (stops at next slash)
	regexStr = strings.ReplaceAll(regexStr, "\\{id\\}", "(?P<id>[^/]+)")

	// 3. Anchor and allow optional trailing slash and extra segments
	// Was: regexStr = "^" + regexStr + "/?$"
	// New: Allow anything after the match if it starts with /
	regexStr = "^" + regexStr + "(?:/.*)?$"

	re, err := regexp.Compile(regexStr)
	if err != nil {
		return params
	}

	matches := re.FindStringSubmatch(urlStr)
	if len(matches) == 0 {
		return params
	}

	names := re.SubexpNames()
	for i, name := range names {
		if i > 0 && i < len(matches) && name != "" {
			params[name] = matches[i]
		}
	}

	return params
}

func (s *ScraperService) scrapeStatic(url string, rule SiteRule, params map[string]interface{}) (map[string]interface{}, error) {
	if url == "" {
		return nil, fmt.Errorf("url is required for static strategy")
	}

	// Initialize Context
	ctx := make(map[string]interface{})
	for k, v := range params {
		ctx[k] = v
	}
	ctx["url"] = url
	if _, ok := ctx["id"]; !ok {
		ctx["id"] = url // Default ID to url if not provided
	}

	// Extract params from URL if regex provided (Legacy support or secondary extraction)
	if rule.Entry != nil && rule.Entry.Regex != "" {
		re, err := regexp.Compile(rule.Entry.Regex)
		if err == nil {
			matches := re.FindStringSubmatch(url)
			names := re.SubexpNames()
			if len(matches) > 0 {
				for i, name := range names {
					if i > 0 && i < len(matches) && name != "" {
						// Only set if not already set by upstream
						if _, exists := ctx[name]; !exists {
							ctx[name] = matches[i]
						}
					}
				}
			}
		}
	}

	req := s.client.R()
	req.SetHeader("User-Agent", DefaultUserAgent)
	if rule.Entry != nil && rule.Entry.Headers != nil {
		req.SetHeaders(rule.Entry.Headers)
	}

	resp, err := req.Get(url)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		return nil, err
	}

	// Store default selection
	ctx["__default_selection__"] = doc.Selection

	// Execute API Steps if present
	if rule.API != nil && len(rule.API.Steps) > 0 {
		if err := s.executeAPISteps(ctx, rule.API.Steps); err != nil {
			return nil, err
		}
	}

	return s.extractFromContext(ctx, rule.Extract), nil
}

func (s *ScraperService) scrapeBrowser(url string, rule SiteRule, params map[string]interface{}) (map[string]interface{}, error) {
	if url == "" {
		return nil, fmt.Errorf("url is required for browser strategy")
	}

	// Initialize Context
	ctx := make(map[string]interface{})
	for k, v := range params {
		ctx[k] = v
	}
	ctx["url"] = url
	if _, ok := ctx["id"]; !ok {
		ctx["id"] = url
	}

	if err := s.browserService.initBrowser(); err != nil {
		return nil, err
	}

	page := s.browserService.browser.MustPage(url)
	defer page.Close()

	// Apply Wait Config
	if rule.WaitConfig != nil {
		wc := rule.WaitConfig
		if wc.Timeout > 0 {
			// Note: Rod context timeout logic would go here if needed
		}

		if !wc.SkipNavigationWait {
			page.MustWaitLoad()
		}

		if !wc.SkipRenderStable {
			page.MustWaitStable()
		}

		// Wait for specific selectors
		if len(wc.ContainerSelectors) > 0 {
			for _, sel := range wc.ContainerSelectors {
				if err := page.WaitElementsMoreThan(sel, 0); err != nil {
					// Log or ignore?
				}
			}
		}
	} else {
		// Default wait
		page.MustWaitStable()
	}

	htmlStr, err := page.HTML()
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		return nil, err
	}

	return s.extractFields(doc.Selection, rule.Extract), nil
}

func (s *ScraperService) scrapeAPI(url string, rule SiteRule, params map[string]interface{}) (map[string]interface{}, error) {
	if rule.API == nil || len(rule.API.Steps) == 0 {
		return nil, fmt.Errorf("api strategy requires api steps")
	}

	// Initialize Context
	ctx := make(map[string]interface{})
	for k, v := range params {
		ctx[k] = v
	}

	// Ensure default pagination parameters if missing
	if _, ok := ctx["offset"]; !ok {
		ctx["offset"] = "0"
	}
	if _, ok := ctx["limit"]; !ok {
		ctx["limit"] = "100"
	}

	ctx["url"] = url
	// Default 'id' to url for convenience when passing IDs directly
	if _, ok := ctx["id"]; !ok {
		ctx["id"] = url
	}

	// Extract params from URL if regex provided
	if rule.Entry != nil && rule.Entry.Regex != "" && url != "" {
		re, err := regexp.Compile(rule.Entry.Regex)
		if err == nil {
			matches := re.FindStringSubmatch(url)
			names := re.SubexpNames()
			if len(matches) > 0 {
				for i, name := range names {
					if i > 0 && i < len(matches) && name != "" {
						ctx[name] = matches[i]
					}
				}
			}
		}
	}

	if err := s.executeAPISteps(ctx, rule.API.Steps); err != nil {
		return nil, err
	}

	return s.extractFromContext(ctx, rule.Extract), nil
}

func (s *ScraperService) executeAPISteps(ctx map[string]interface{}, steps []APIStep) error {
	for _, step := range steps {
		// Process URL template if needed
		stepURLRaw := s.processTemplate(step.Request.URL, ctx)
		stepURL := ""
		if str, ok := stepURLRaw.(string); ok {
			stepURL = strings.TrimSpace(str)
		} else {
			// Should not happen for API steps URL template with ctx map
			stepURL = fmt.Sprintf("%v", stepURLRaw)
		}

		// Check for unreplaced placeholders
		if strings.Contains(stepURL, "{") && strings.Contains(stepURL, "}") {
			return fmt.Errorf("step %s failed: URL %s contains unreplaced placeholders. Available keys: %v", step.ID, stepURL, getKeys(ctx))
		}

		req := s.client.R()
		req.SetHeader("User-Agent", DefaultUserAgent)
		if step.Request.Headers != nil {
			req.SetHeaders(step.Request.Headers)
		}

		var resp *resty.Response
		var err error

		method := strings.ToUpper(step.Request.Method)
		if method == "" {
			method = "GET"
		}

		switch method {
		case "POST":
			resp, err = req.Post(stepURL)
		default:
			resp, err = req.Get(stepURL)
		}

		if err != nil {
			return fmt.Errorf("step %s failed: %w", step.ID, err)
		}

		body := resp.String()

		// Parse response based on type
		if step.Response == "html" {
			ctx[step.ID] = body
			ctx[step.ID+"_raw"] = body
		} else {
			ctx[step.ID] = gjson.Parse(body).Value()
			ctx[step.ID+"_raw"] = body
		}
	}
	return nil
}

// extractFromContext handles extraction when source is a map of step results
func (s *ScraperService) extractFromContext(ctx map[string]interface{}, rules []FieldRule) map[string]interface{} {
	result := make(map[string]interface{})

	var defaultSource interface{} = ctx
	if sel, ok := ctx["__default_selection__"]; ok {
		defaultSource = sel
	}

	for _, field := range rules {
		if field.Type == "template" {
			// Template type always uses the full context (ctx) as source
			result[field.Name] = s.extractFieldGeneric(ctx, field)
		} else if field.From != "" {
			// Try _raw first to avoid marshaling overhead/issues
			if val, ok := ctx[field.From+"_raw"]; ok {
				result[field.Name] = s.extractFieldGeneric(val, field)
			} else if val, ok := ctx[field.From]; ok {
				// If source is a string (HTML/JSON string)
				// If source is interface{} (parsed JSON)
				result[field.Name] = s.extractFieldGeneric(val, field)
			}
		} else {
			// Use default source (ctx for API, Selection for Static)
			result[field.Name] = s.extractFieldGeneric(defaultSource, field)
		}
	}
	return result
}

func (s *ScraperService) extractFieldGeneric(source interface{}, field FieldRule) interface{} {
	// If type is template, it uses the source as context
	if field.Type == "template" {
		// If children are defined, we must extract them first to build the context
		if len(field.Children) > 0 {
			childData := make(map[string]interface{})
			var multiKeys []string
			var maxLen int

			ctx, isCtx := source.(map[string]interface{})
			sel, isSel := source.(*goquery.Selection)

			for _, child := range field.Children {
				var val interface{}
				childSource := source

				// If we are in a context map and child specifies 'From'
				if isCtx && child.From != "" {
					if v, ok := ctx[child.From+"_raw"]; ok {
						childSource = v
					} else if v, ok := ctx[child.From]; ok {
						childSource = v
					} else {
						// If 'From' is specified but not found in context, check if we are in 'template' extraction mode
						// where 'source' might be just one step's data (if passed from parent),
						// OR 'source' is the big context map.
						// If childSource is nil, we can't extract.
						childSource = nil
					}
				}

				// If we are in a context map, and child type is CSS, and no From was specified
				// We should try to use the default selection from context
				if isCtx && child.From == "" && child.Type == "css" {
					if sel, ok := ctx["__default_selection__"]; ok {
						childSource = sel
					}
				}

				if childSource != nil {
					// Recursively extract
					// Note: extractFieldGeneric expects 'source' to be what the type needs.
					// If type is json, source should be string or json-compatible.

					// If source is Selection and child is CSS, pass Selection
					// If source is Selection and child is Template, pass Selection
					// If source is Selection and child is JSON, extract text? (Handled in extractFields but here we call extractFieldGeneric directly)

					if isSel {
						// Special handling for Selection source
						switch child.Type {
						case "css", "template":
							val = s.extractFieldGeneric(sel, child)
						case "json":
							// For JSON child of a template in CSS context (unlikely but possible?)
							// Treat text as JSON?
							// Let's defer to extractFieldGeneric which handles json type if source is string
							// But source is Selection here.
							// extractFieldGeneric needs to handle Selection for json type?
							// Currently extractFieldGeneric handles json type by expecting string or object.

							// Let's make extractFieldGeneric smart enough to handle Selection for JSON type too?
							// OR extract the text here.
							// Reuse extractFields logic?
							// Let's assume child of template in CSS context is usually CSS.
							val = s.extractFieldGeneric(sel, child)
						default:
							val = s.extractFieldGeneric(sel, child)
						}
					} else {
						val = s.extractFieldGeneric(childSource, child)
					}
				}

				childData[child.Name] = val

				// Check for Multiple
				// If child is multiple, val should be []interface{}
				if child.Multiple {
					if arr, ok := val.([]interface{}); ok {
						multiKeys = append(multiKeys, child.Name)
						if len(arr) > maxLen {
							maxLen = len(arr)
						}
					}
				}
			}

			// If we have multiple values, we need to generate a list of results
			if len(multiKeys) > 0 {
				var results []interface{}
				for i := 0; i < maxLen; i++ {
					itemData := make(map[string]interface{})
					for k, v := range childData {
						isMulti := false
						for _, mk := range multiKeys {
							if mk == k {
								isMulti = true
								break
							}
						}

						if isMulti {
							arr := v.([]interface{})
							if i < len(arr) {
								itemData[k] = arr[i]
							} else {
								itemData[k] = ""
							}
						} else {
							itemData[k] = v
						}
					}
					// If parent field (template) is marked as multiple, or if children are multiple and parent is template
					// For a template type, if it produces multiple values, it returns []string (or []interface{})
					// The previous logic returns ONE string if not explicit loop?
					// Wait, if children are multiple, the template should probably generate multiple strings.

					// processTemplate now returns interface{} (string or []string)
					// Here we expect single string because we are iterating manually
					res := s.processTemplate(field.Template, itemData)
					if str, ok := res.(string); ok {
						results = append(results, str)
					} else {
						// Should not happen if itemData is map
						results = append(results, res)
					}
				}
				return results
			}

			// Single result
			return s.processTemplate(field.Template, childData)
		}

		return s.processTemplate(field.Template, source)
	}

	// If type is text (fixed value)
	if field.Type == "text" {
		return field.Text
	}

	// If type is css, source must be HTML string OR Selection (if passed from template parent)
	if field.Type == "css" {
		if sel, ok := source.(*goquery.Selection); ok {
			return s.extractCSS(sel, field)
		}
		htmlStr, ok := source.(string)
		if !ok {
			return nil
		}
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
		return s.extractCSS(doc.Selection, field)
	}

	// If type is json, source can be JSON string or object
	if field.Type == "json" {
		// If it's already an object, marshal it back? or use reflection?
		// gjson works on strings.
		jsonStr, ok := source.(string)
		if !ok {
			// Marshall back to JSON string so gjson can work
			b, err := json.Marshal(source)
			if err == nil {
				jsonStr = string(b)
			} else {
				jsonStr = fmt.Sprintf("%v", source)
			}
		}
		return s.extractJSON(jsonStr, field)
	}

	return nil
}

func (s *ScraperService) extractFields(selection *goquery.Selection, rules []FieldRule) map[string]interface{} {
	result := make(map[string]interface{})
	for _, rule := range rules {
		switch rule.Type {
		case "json":
			// Special case: Extract text from CSS selection and treat as JSON
			// 1. Get the element(s) using selector
			sel := selection
			if rule.Selector != "" {
				sel = selection.Find(rule.Selector)
			}

			// 2. Get text
			jsonStr := sel.Text()
			if rule.Trim {
				jsonStr = strings.TrimSpace(jsonStr)
			}

			// 3. Extract using JSON logic
			result[rule.Name] = s.extractJSON(jsonStr, rule)
		case "template":
			// Template field at the top level of extractFields (CSS/Static context)
			result[rule.Name] = s.extractFieldGeneric(selection, rule)
		case "text":
			result[rule.Name] = rule.Text
		default:
			// Default to CSS extraction
			result[rule.Name] = s.extractCSS(selection, rule)
		}
	}
	return result
}

func (s *ScraperService) extractCSS(sel *goquery.Selection, rule FieldRule) interface{} {
	// Select
	current := sel
	if rule.Selector != "" {
		current = sel.Find(rule.Selector)
	}

	// Handle Multiple
	if rule.Multiple {
		var items []interface{}
		current.Each(func(i int, selection *goquery.Selection) {
			// Filter Logic
			if rule.Filter != "" {
				match := selection.Is(rule.Filter) || selection.Find(rule.Filter).Length() > 0

				if rule.FilterMode == "not" {
					if match {
						return // Skip this item
					}
				} else {
					// Default: Keep only if it matches (FilterMode="has" or empty default implies filtering FOR something?)
					// Usually if 'filter' is set, we want items that match.
					if !match {
						return // Skip this item
					}
				}
			}

			if len(rule.Children) > 0 {
				items = append(items, s.extractFields(selection, rule.Children))
			} else {
				val := s.extractValue(selection, rule)
				items = append(items, val)
			}
		})
		return items
	}

	// Single
	if len(rule.Children) > 0 {
		return s.extractFields(current.First(), rule.Children)
	}

	return s.extractValue(current.First(), rule)
}

func (s *ScraperService) extractValue(sel *goquery.Selection, rule FieldRule) interface{} {
	var val string

	// Attribute or Text
	if len(rule.Attr) > 0 {
		// Prefer first found attribute
		for _, attr := range rule.Attr {
			if v, exists := sel.Attr(attr); exists {
				val = v
				break
			}
		}
	} else {
		val = sel.Text()
	}

	// Trim
	if rule.Trim {
		val = strings.TrimSpace(val)
	}

	// Regex
	if rule.Regex != "" {
		re, err := regexp.Compile(rule.Regex)
		if err == nil {
			matches := re.FindStringSubmatch(val)
			if len(matches) > 1 {
				val = matches[1] // Return first capture group
			} else if len(matches) > 0 {
				val = matches[0] // Return full match
			}
		}
	}

	return val
}

func (s *ScraperService) extractJSON(jsonStr string, rule FieldRule) interface{} {
	if rule.Path == "" {
		return nil
	}

	// Normalize Path: GJSON requires double quotes for string literals in queries.
	// Users might use single quotes in the JSON rule to avoid escaping hell.
	// We replace ' with " inside the path, but we need to be careful not to break other things.
	// Simple approach: Replace ' with " if it looks like a query value.
	// Or just replace all single quotes with double quotes? GJSON paths usually don't contain single quotes as syntax.
	path := strings.ReplaceAll(rule.Path, "'", "\"")

	res := gjson.Get(jsonStr, path)

	if rule.Multiple {
		if !res.IsArray() {
			return []interface{}{}
		}
		var items []interface{}
		res.ForEach(func(key, value gjson.Result) bool {
			if len(rule.Children) > 0 {
				// Recursively extract children on this object
				childMap := make(map[string]interface{})
				for _, child := range rule.Children {
					// IMPORTANT: For children of a JSON array item, we pass the ITEM's JSON string
					// We use extractFieldGeneric to support mixed types (json, text, etc.)
					childMap[child.Name] = s.extractFieldGeneric(value.String(), child)
				}
				items = append(items, childMap)
			} else {
				items = append(items, value.Value())
			}
			return true
		})
		return items
	}

	// Single
	if len(rule.Children) > 0 {
		childMap := make(map[string]interface{})
		for _, child := range rule.Children {
			// Same logic: operate on the current object (res)
			childMap[child.Name] = s.extractFieldGeneric(res.String(), child)
		}
		return childMap
	}

	val := res.String()
	// Regex on string value
	if rule.Regex != "" {
		re, err := regexp.Compile(rule.Regex)
		if err == nil {
			matches := re.FindStringSubmatch(val)
			if len(matches) > 1 {
				val = matches[1]
			}
		}
	}

	return val
}

func (s *ScraperService) renderTemplate(tmplStr string, data interface{}) string {
	t, err := template.New("tmpl").Parse(tmplStr)
	if err != nil {
		return tmplStr
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return tmplStr
	}
	return buf.String()
}

func (s *ScraperService) processTemplate(tmplStr string, data interface{}) interface{} {
	// If data is array/slice, process template for each item
	if arr, ok := data.([]interface{}); ok {
		var results []string
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				results = append(results, s.simpleRender(tmplStr, m))
			} else {
				// Fallback for non-map items in array?
				results = append(results, s.renderTemplate(tmplStr, item))
			}
		}
		return results
	}

	// Single item processing
	if m, ok := data.(map[string]interface{}); ok {
		return s.simpleRender(tmplStr, m)
	}
	return s.renderTemplate(tmplStr, data)
}

func (s *ScraperService) simpleRender(tmpl string, data map[string]interface{}) string {
	result := tmpl
	for k, v := range data {
		placeholder := "{" + k + "}"
		valStr := fmt.Sprintf("%v", v)
		result = strings.ReplaceAll(result, placeholder, valStr)
	}
	return result
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
