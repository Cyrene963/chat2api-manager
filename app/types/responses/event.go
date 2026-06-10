package responses

import "encoding/json"

type Event struct {
	Type           string        `json:"type"`
	ResponseID     string        `json:"response_id,omitempty"`
	Response       *Response     `json:"response,omitempty"`
	OutputIndex    *int          `json:"output_index,omitempty"`
	ContentIndex   *int          `json:"content_index,omitempty"`
	ItemID         string        `json:"item_id,omitempty"`
	Item           *OutputItem   `json:"item,omitempty"`
	Part           *ContentPart  `json:"part,omitempty"`
	Delta          string        `json:"delta,omitempty"`
	Logprobs       []interface{} `json:"logprobs,omitempty"`
	Obfuscation    string        `json:"obfuscation,omitempty"`
	Text           string        `json:"text,omitempty"`
	Name           string        `json:"name,omitempty"`
	CallID         string        `json:"call_id,omitempty"`
	Arguments      string        `json:"arguments,omitempty"`
	SequenceNumber int64         `json:"sequence_number,omitempty"`
}

type EventStream struct {
	sequenceNumber int64
}

func NewEventStream() *EventStream {
	return &EventStream{}
}

func (stream *EventStream) SSE(event Event) string {
	stream.sequenceNumber++
	event.SequenceNumber = stream.sequenceNumber
	return SSE(event)
}

func Int(value int) *int {
	return &value
}

func CreatedEvent(responseID string, model string, created int64, options ...ResponseOptions) Event {
	response := NewResponse(responseID, model, created, "in_progress", []OutputItem{}, options...)
	return Event{Type: "response.created", Response: &response}
}

func InProgressEvent(responseID string, model string, created int64, options ...ResponseOptions) Event {
	response := NewResponse(responseID, model, created, "in_progress", []OutputItem{}, options...)
	return Event{Type: "response.in_progress", Response: &response}
}

func CompletedEvent(responseID string, model string, created int64, output []OutputItem, options ...ResponseOptions) Event {
	response := NewResponse(responseID, model, created, "completed", output, options...)
	return Event{Type: "response.completed", Response: &response}
}

func SSE(event Event) string {
	data, _ := json.Marshal(event)
	if event.Type == "" {
		return "data: " + string(data) + "\n\n"
	}
	return "event: " + event.Type + "\ndata: " + string(data) + "\n\n"
}

func RawSSE(eventType string, payload string) string {
	if eventType == "" {
		return "data: " + payload + "\n\n"
	}
	return "event: " + eventType + "\ndata: " + payload + "\n\n"
}
