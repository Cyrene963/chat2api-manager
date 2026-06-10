package responses

type OutputItem struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"`
	Status           string                 `json:"status,omitempty"`
	Role             string                 `json:"role,omitempty"`
	Content          *[]ContentPart         `json:"content,omitempty"`
	EncryptedContent string                 `json:"encrypted_content,omitempty"`
	Summary          *[]interface{}         `json:"summary,omitempty"`
	Phase            string                 `json:"phase,omitempty"`
	CallID           string                 `json:"call_id,omitempty"`
	Name             string                 `json:"name,omitempty"`
	Arguments        string                 `json:"arguments,omitempty"`
	Result           string                 `json:"result,omitempty"`
	RevisedPrompt    string                 `json:"revised_prompt,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type ContentPart struct {
	Type        string        `json:"type"`
	Annotations []interface{} `json:"annotations"`
	Logprobs    []interface{} `json:"logprobs"`
	Text        string        `json:"text"`
}

func ImageOutputItem(id string, result string, prompt string) OutputItem {
	return OutputItem{
		ID:            id,
		Type:          "image_generation_call",
		Status:        "completed",
		Result:        result,
		RevisedPrompt: prompt,
	}
}

func TextOutputItem(id string, text string, status string) OutputItem {
	content := []ContentPart{}
	if status != "in_progress" {
		content = []ContentPart{OutputTextPart(text)}
	}
	return OutputItem{
		ID:      id,
		Type:    "message",
		Status:  status,
		Role:    "assistant",
		Phase:   "final_answer",
		Content: &content,
	}
}

func FunctionCallOutputItem(id string, callID string, name string, arguments string, status string) OutputItem {
	return OutputItem{
		ID:        id,
		Type:      "function_call",
		Status:    status,
		CallID:    callID,
		Name:      name,
		Arguments: arguments,
	}
}

func ReasoningOutputItem(id string, encryptedContent string) OutputItem {
	content := []ContentPart{}
	summary := []interface{}{}
	return OutputItem{
		ID:               id,
		Type:             "reasoning",
		Content:          &content,
		EncryptedContent: encryptedContent,
		Summary:          &summary,
	}
}

func OutputTextPart(text string) ContentPart {
	return ContentPart{
		Type:        "output_text",
		Annotations: []interface{}{},
		Logprobs:    []interface{}{},
		Text:        text,
	}
}
