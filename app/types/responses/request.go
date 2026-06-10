package responses

type ApiReq struct {
	Model              string                 `json:"model"`
	Input              interface{}            `json:"input"`
	Instructions       string                 `json:"instructions"`
	Background         bool                   `json:"background"`
	Stream             bool                   `json:"stream"`
	PreviousResponseID string                 `json:"previous_response_id"`
	Tools              []Tool                 `json:"tools"`
	ToolChoice         interface{}            `json:"tool_choice"`
	MaxOutputTokens    interface{}            `json:"max_output_tokens"`
	MaxToolCalls       interface{}            `json:"max_tool_calls"`
	ParallelToolCalls  *bool                  `json:"parallel_tool_calls"`
	Reasoning          *Reasoning             `json:"reasoning"`
	Store              *bool                  `json:"store"`
	Temperature        interface{}            `json:"temperature"`
	Text               interface{}            `json:"text"`
	TopP               interface{}            `json:"top_p"`
	Truncation         string                 `json:"truncation"`
	Include            []string               `json:"include"`
	Metadata           map[string]interface{} `json:"metadata"`
	User               interface{}            `json:"user"`
}

type Tool struct {
	Type               string                 `json:"type"`
	Name               string                 `json:"name,omitempty"`
	Description        string                 `json:"description,omitempty"`
	Parameters         map[string]interface{} `json:"parameters,omitempty"`
	Strict             *bool                  `json:"strict,omitempty"`
	Model              string                 `json:"model,omitempty"`
	Action             string                 `json:"action,omitempty"`
	Size               string                 `json:"size,omitempty"`
	Quality            string                 `json:"quality,omitempty"`
	OutputFormat       string                 `json:"output_format,omitempty"`
	SearchContextSize  interface{}            `json:"search_context_size,omitempty"`
	Filters            map[string]interface{} `json:"filters,omitempty"`
	UserLocation       map[string]interface{} `json:"user_location,omitempty"`
	SearchContentTypes []string               `json:"search_content_types,omitempty"`
	VectorStoreIDs     []string               `json:"vector_store_ids,omitempty"`
	ServerLabel        string                 `json:"server_label,omitempty"`
	ServerURL          string                 `json:"server_url,omitempty"`
	RequireApproval    string                 `json:"require_approval,omitempty"`
	AllowedTools       []string               `json:"allowed_tools,omitempty"`
	Container          map[string]interface{} `json:"container,omitempty"`
}
