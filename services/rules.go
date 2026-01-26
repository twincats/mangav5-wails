package services

// SiteRule defines the scraping rules for a specific site
type SiteRule struct {
	Site       string       `json:"site"`
	Domains    []string     `json:"domains"`
	Strategy   string       `json:"strategy"` // static, browser, api, auto
	Entry      *EntryRule   `json:"entry,omitempty"`
	API        *APIWorkflow `json:"api,omitempty"`
	Extract    []FieldRule  `json:"extract"`
	WaitConfig *WaitConfig  `json:"wait_config,omitempty"`
}

type EntryRule struct {
	URL     string            `json:"url"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Regex   string            `json:"regex,omitempty"` // Extract params from URL
}

type APIWorkflow struct {
	Steps []APIStep `json:"steps"`
}

type APIStep struct {
	ID       string      `json:"id"`
	Request  APIRequest  `json:"request"`
	Response string      `json:"response,omitempty"` // json, html. default json
}

type APIRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type FieldRule struct {
	Name string `json:"name"`
	Type string `json:"type"` // css, json, template

	// Common
	Multiple bool        `json:"multiple"`
	Trim     bool        `json:"trim"`
	Regex    string      `json:"regex,omitempty"`
	Children []FieldRule `json:"children,omitempty"`
	From     string      `json:"from,omitempty"`

	// CSS
	Selector   string   `json:"selector,omitempty"`
	Attr       []string `json:"attr,omitempty"`
	Filter     string   `json:"filter,omitempty"`
	FilterMode string   `json:"filter_mode,omitempty"` // has, not

	// JSON
	Path string `json:"path,omitempty"`

	// Template
	Template string `json:"template,omitempty"`

	// Text (Fixed Value)
	Text string `json:"text,omitempty"`
}

type WaitConfig struct {
	ContainerSelectors []string `json:"container_selectors,omitempty"`
	ContentSelectors   []string `json:"content_selectors,omitempty"`
	MinTextLength      int      `json:"min_text_length,omitempty"`
	RequireImageLoaded bool     `json:"require_image_loaded,omitempty"`
	Timeout            int      `json:"timeout_ms,omitempty"`
	PollInterval       int      `json:"poll_ms,omitempty"`
	SkipWaits          bool     `json:"skip_waits,omitempty"`
	SkipRenderStable   bool     `json:"skip_render_stable,omitempty"`
	SkipNavigationWait bool     `json:"skip_navigation_wait,omitempty"`
}
