package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

// ScraperService handles rule-based scraping
type ScraperService struct {
	browserService *BrowserService
	client         *resty.Client
}

// NewScraperService creates a new instance
func NewScraperService(bs *BrowserService) *ScraperService {
	return &ScraperService{
		browserService: bs,
		client:         resty.New(),
	}
}

// Scrape executes the scraping rule
func (s *ScraperService) Scrape(rule SiteRule, overrideURL string) (map[string]interface{}, error) {
	targetURL := overrideURL
	if targetURL == "" && rule.Entry != nil {
		targetURL = rule.Entry.URL
	}

	switch rule.Strategy {
	case "static":
		return s.scrapeStatic(targetURL, rule)
	case "browser":
		return s.scrapeBrowser(targetURL, rule)
	case "api":
		return s.scrapeAPI(targetURL, rule)
	case "auto":
		// Default to static, fallback logic could be added here
		return s.scrapeStatic(targetURL, rule)
	default:
		return nil, fmt.Errorf("unknown strategy: %s", rule.Strategy)
	}
}

func (s *ScraperService) scrapeStatic(url string, rule SiteRule) (map[string]interface{}, error) {
	if url == "" {
		return nil, fmt.Errorf("url is required for static strategy")
	}

	req := s.client.R()
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

	return s.extractFields(doc.Selection, rule.Extract), nil
}

func (s *ScraperService) scrapeBrowser(url string, rule SiteRule) (map[string]interface{}, error) {
	if url == "" {
		return nil, fmt.Errorf("url is required for browser strategy")
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

func (s *ScraperService) scrapeAPI(url string, rule SiteRule) (map[string]interface{}, error) {
	if rule.API == nil || len(rule.API.Steps) == 0 {
		return nil, fmt.Errorf("api strategy requires api steps")
	}

	// Context to store results from steps
	ctx := make(map[string]interface{})
	ctx["url"] = url
	// Default 'id' to url for convenience when passing IDs directly
	ctx["id"] = url

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

	for _, step := range rule.API.Steps {
		// Process URL template if needed
		stepURL := s.processTemplate(step.Request.URL, ctx)

		// Check for unreplaced placeholders
		if strings.Contains(stepURL, "{") && strings.Contains(stepURL, "}") {
			return nil, fmt.Errorf("step %s failed: URL %s contains unreplaced placeholders. Available keys: %v", step.ID, stepURL, getKeys(ctx))
		}

		req := s.client.R()
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
			return nil, fmt.Errorf("step %s failed: %w", step.ID, err)
		}

		body := resp.String()
		
		// Parse response based on type
		if step.Response == "html" {
			// If it's HTML, we might need to extract data immediately?
			// The schema doesn't specify extraction per step, but usually API steps 
			// build up context or the final step provides data for 'Extract'.
			// Assuming the last step's response is used for main extraction if not specified otherwise.
			// But for now, let's store the raw body in context.
			ctx[step.ID] = body
			ctx[step.ID+"_raw"] = body
		} else {
			// Default JSON
			// Store as parsed JSON object or raw string?
			// gjson.Parse(body).Value() gives interface{}
			ctx[step.ID] = gjson.Parse(body).Value()
			ctx[step.ID+"_raw"] = body
		}
	}

	// For extraction, we need a source. 
	// The rule.Extract fields might refer to specific steps via 'from'.
	// If 'from' is missing, what is the default source? probably the last step.
	
	// We need a flexible extractor that can handle the map context.
	// For simplicity, we will assume 'from' is mandatory for API strategy OR we use the last response.
	
	return s.extractFromContext(ctx, rule.Extract), nil
}

// extractFromContext handles extraction when source is a map of step results
func (s *ScraperService) extractFromContext(ctx map[string]interface{}, rules []FieldRule) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Find default source (last step?)
	// For now, let's require 'from' or use the first available key if single
	
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
			// Try to find a default source or use ctx directly (for template)
			// For non-template, if From is missing, we might default to one of the steps?
			// But since we don't know which one, let's skip or try the last one added.
			// Currently, just skipping if not template.
			result[field.Name] = s.extractFieldGeneric(ctx, field)
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

				if childSource != nil {
					// Recursively extract
					// Note: extractFieldGeneric expects 'source' to be what the type needs.
					// If type is json, source should be string or json-compatible.
					val = s.extractFieldGeneric(childSource, child)
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
					results = append(results, s.processTemplate(field.Template, itemData))
				}
				return results
			}

			// Single result
			return s.processTemplate(field.Template, childData)
		}

		return s.processTemplate(field.Template, source)
	}

	// If type is css, source must be HTML string
	if field.Type == "css" {
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
		if rule.Type == "json" {
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
		} else {
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
					// However, if the child specifies 'From', it might be trying to access a different step?
					// Usually, children of a 'json' multiple extraction are relative to the item.
					// BUT, the schema allows 'From' on any rule.
					// If 'From' is present, we should probably ignore the current item context?
					// BUT extractJSON signature takes jsonStr (the current item).
					// If we want to support 'From' inside a nested JSON loop, we'd need the global context, which we don't have here.
					// assumption: Children of a JSON extraction operate on the extracted item.
					
					childMap[child.Name] = s.extractJSON(value.String(), child)
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
			childMap[child.Name] = s.extractJSON(res.String(), child)
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

func (s *ScraperService) processTemplate(tmplStr string, data interface{}) string {
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
