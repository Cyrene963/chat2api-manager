package responses

type Response struct {
	ID                   string                 `json:"id"`
	Object               string                 `json:"object"`
	CreatedAt            int64                  `json:"created_at"`
	Status               string                 `json:"status"`
	Background           bool                   `json:"background"`
	CompletedAt          *int64                 `json:"completed_at"`
	Error                interface{}            `json:"error"`
	FrequencyPenalty     float64                `json:"frequency_penalty"`
	IncompleteDetails    interface{}            `json:"incomplete_details"`
	Instructions         interface{}            `json:"instructions"`
	MaxOutputTokens      interface{}            `json:"max_output_tokens"`
	MaxToolCalls         interface{}            `json:"max_tool_calls"`
	Model                string                 `json:"model"`
	Moderation           interface{}            `json:"moderation"`
	Output               []OutputItem           `json:"output"`
	ParallelToolCalls    bool                   `json:"parallel_tool_calls"`
	PresencePenalty      float64                `json:"presence_penalty"`
	PreviousResponseID   interface{}            `json:"previous_response_id"`
	PromptCacheKey       interface{}            `json:"prompt_cache_key"`
	PromptCacheRetention interface{}            `json:"prompt_cache_retention"`
	Reasoning            *Reasoning             `json:"reasoning"`
	SafetyIdentifier     interface{}            `json:"safety_identifier"`
	ServiceTier          string                 `json:"service_tier"`
	Store                bool                   `json:"store"`
	Temperature          interface{}            `json:"temperature"`
	Text                 interface{}            `json:"text"`
	ToolChoice           interface{}            `json:"tool_choice"`
	Tools                []Tool                 `json:"tools"`
	TopLogprobs          int                    `json:"top_logprobs"`
	TopP                 interface{}            `json:"top_p"`
	Truncation           string                 `json:"truncation"`
	Usage                interface{}            `json:"usage"`
	User                 interface{}            `json:"user"`
	Metadata             map[string]interface{} `json:"metadata"`
}

type Reasoning struct {
	Context string      `json:"context,omitempty"`
	Effort  string      `json:"effort,omitempty"`
	Summary interface{} `json:"summary"`
}

type ResponseOptions struct {
	Instructions      interface{}
	MaxOutputTokens   interface{}
	ParallelToolCalls bool
	Reasoning         *Reasoning
	Store             bool
	Temperature       interface{}
	Text              interface{}
	ToolChoice        interface{}
	Tools             []Tool
	TopP              interface{}
	Truncation        string
	Usage             interface{}
	Metadata          map[string]interface{}
}

func OptionsFromRequest(req *ApiReq) ResponseOptions {
	if req == nil {
		return defaultResponseOptions()
	}
	options := defaultResponseOptions()
	if req.Instructions != "" {
		options.Instructions = req.Instructions
	}
	options.MaxOutputTokens = req.MaxOutputTokens
	if req.ParallelToolCalls != nil {
		options.ParallelToolCalls = *req.ParallelToolCalls
	}
	options.Reasoning = normalizeReasoning(req.Reasoning)
	if req.Store != nil {
		options.Store = *req.Store
	}
	if req.Temperature != nil {
		options.Temperature = req.Temperature
	}
	if req.Text != nil {
		options.Text = req.Text
	}
	if req.ToolChoice != nil {
		options.ToolChoice = req.ToolChoice
	}
	if req.Tools != nil {
		options.Tools = req.Tools
	}
	if req.TopP != nil {
		options.TopP = req.TopP
	}
	if req.Truncation != "" {
		options.Truncation = req.Truncation
	}
	if req.Metadata != nil {
		options.Metadata = req.Metadata
	}
	return options
}

func defaultResponseOptions() ResponseOptions {
	return ResponseOptions{
		Temperature: 1.0,
		Text: map[string]interface{}{
			"format": map[string]string{"type": "text"},
		},
		ToolChoice: "auto",
		Tools:      []Tool{},
		TopP:       1.0,
		Truncation: "disabled",
		Metadata:   map[string]interface{}{},
	}
}

func normalizeReasoning(reasoning *Reasoning) *Reasoning {
	if reasoning == nil {
		return nil
	}
	normalized := *reasoning
	if normalized.Effort != "" && normalized.Context == "" {
		normalized.Context = "current_turn"
	}
	return &normalized
}

func NewResponse(responseID string, model string, created int64, status string, output []OutputItem, options ...ResponseOptions) Response {
	resolved := defaultResponseOptions()
	if len(options) > 0 {
		resolved = mergeResponseOptions(resolved, options[0])
	}
	var completedAt *int64
	if status == "completed" {
		completed := created
		completedAt = &completed
	}
	if output == nil {
		output = []OutputItem{}
	}
	if resolved.Tools == nil {
		resolved.Tools = []Tool{}
	}
	if resolved.Metadata == nil {
		resolved.Metadata = map[string]interface{}{}
	}
	return Response{
		ID:                   responseID,
		Object:               "response",
		CreatedAt:            created,
		Status:               status,
		Background:           false,
		CompletedAt:          completedAt,
		Error:                nil,
		FrequencyPenalty:     0,
		IncompleteDetails:    nil,
		Instructions:         resolved.Instructions,
		MaxOutputTokens:      resolved.MaxOutputTokens,
		MaxToolCalls:         nil,
		Model:                model,
		Moderation:           nil,
		Output:               output,
		ParallelToolCalls:    resolved.ParallelToolCalls,
		PresencePenalty:      0,
		PreviousResponseID:   nil,
		PromptCacheKey:       nil,
		PromptCacheRetention: nil,
		Reasoning:            resolved.Reasoning,
		SafetyIdentifier:     nil,
		ServiceTier:          "auto",
		Store:                resolved.Store,
		Temperature:          resolved.Temperature,
		Text:                 resolved.Text,
		ToolChoice:           resolved.ToolChoice,
		Tools:                resolved.Tools,
		TopLogprobs:          0,
		TopP:                 resolved.TopP,
		Truncation:           resolved.Truncation,
		Usage:                resolved.Usage,
		User:                 nil,
		Metadata:             resolved.Metadata,
	}
}

func mergeResponseOptions(base ResponseOptions, override ResponseOptions) ResponseOptions {
	base.Instructions = override.Instructions
	base.MaxOutputTokens = override.MaxOutputTokens
	base.ParallelToolCalls = override.ParallelToolCalls
	base.Reasoning = override.Reasoning
	base.Store = override.Store
	if override.Temperature != nil {
		base.Temperature = override.Temperature
	}
	if override.Text != nil {
		base.Text = override.Text
	}
	if override.ToolChoice != nil {
		base.ToolChoice = override.ToolChoice
	}
	if override.Tools != nil {
		base.Tools = override.Tools
	}
	if override.TopP != nil {
		base.TopP = override.TopP
	}
	if override.Truncation != "" {
		base.Truncation = override.Truncation
	}
	base.Usage = override.Usage
	if override.Metadata != nil {
		base.Metadata = override.Metadata
	}
	return base
}
