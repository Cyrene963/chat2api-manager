package responses

type ApiReq struct {
	Model             string                 `json:"model"`
	Input             interface{}            `json:"input"`
	Instructions      string                 `json:"instructions"`
	Stream            bool                   `json:"stream"`
	Tools             []Tool                 `json:"tools"`
	ToolChoice        interface{}            `json:"tool_choice"`
	MaxOutputTokens   interface{}            `json:"max_output_tokens"`
	ParallelToolCalls *bool                  `json:"parallel_tool_calls"`
	Reasoning         *Reasoning             `json:"reasoning"`
	Store             *bool                  `json:"store"`
	Temperature       interface{}            `json:"temperature"`
	Text              interface{}            `json:"text"`
	TopP              interface{}            `json:"top_p"`
	Truncation        string                 `json:"truncation"`
	Metadata          map[string]interface{} `json:"metadata"`
}

type Tool struct {
	Type         string                 `json:"type"`
	Name         string                 `json:"name,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Strict       *bool                  `json:"strict,omitempty"`
	Model        string                 `json:"model,omitempty"`
	Action       string                 `json:"action,omitempty"`
	Size         string                 `json:"size,omitempty"`
	Quality      string                 `json:"quality,omitempty"`
	OutputFormat string                 `json:"output_format,omitempty"`
}
