package service

import (
	"chat2api/app/types/completions"
	"chat2api/app/types/responses"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func streamResponsesTextEvents(c *gin.Context, model string, resp *http.Response, responseOptions responses.ResponseOptions) (*chatResult, error) {
	c.Header("Content-Type", "text/event-stream")
	responseID := responses.ResponseID()
	itemID := responses.MessageID()
	created := time.Now().Unix()
	stream := responses.NewEventStream()
	if err := writeResponsesLifecycleStart(c, stream, responseID, model, created, responseOptions); err != nil {
		return nil, err
	}
	if err := startResponsesTextItem(c, stream, itemID, 0); err != nil {
		return nil, err
	}
	c.Writer.Flush()
	result, err := handleChatStream(resp, func(event chatStreamEvent) error {
		if event.Delta == "" {
			return nil
		}
		if _, err := c.Writer.WriteString(stream.SSE(responses.Event{Type: "response.output_text.delta", ItemID: itemID, OutputIndex: responses.Int(0), ContentIndex: responses.Int(0), Delta: event.Delta})); err != nil {
			return err
		}
		c.Writer.Flush()
		return nil
	})
	if err != nil {
		return nil, err
	}
	completedItem, err := finishResponsesTextItem(c, stream, itemID, 0, result.Content)
	if err != nil {
		return nil, err
	}
	if _, err := c.Writer.WriteString(stream.SSE(responses.CompletedEvent(responseID, model, created, []responses.OutputItem{completedItem}, responseOptions))); err != nil {
		return nil, err
	}
	_, _ = c.Writer.WriteString("data: [DONE]\n\n")
	c.Writer.Flush()
	return result, nil
}

func streamResponsesFunctionCallingEvents(c *gin.Context, apiReq *completions.ApiReq, resp *http.Response, responseOptions responses.ResponseOptions) (*chatResult, error) {
	c.Header("Content-Type", "text/event-stream")
	responseID := responses.ResponseID()
	itemID := responses.MessageID()
	created := time.Now().Unix()
	stream := responses.NewEventStream()
	if err := writeResponsesLifecycleStart(c, stream, responseID, apiReq.Model, created, responseOptions); err != nil {
		return nil, err
	}
	c.Writer.Flush()

	detector := completions.NewStreamToolDetector(completions.ToolifyTriggerSignal)
	textItemStarted := false
	toolStreamFinished := false

	result, err := handleChatStream(resp, func(event chatStreamEvent) error {
		if detector.State() == "tool_parsing" {
			detector.AppendParsing(event.Delta)
			if !detector.HasCompleteToolBlock() {
				return nil
			}
			calls := detector.Finalize()
			if len(calls) == 0 || completions.ValidateParsedToolCalls(calls, apiReq.Tools) != nil {
				toolStreamFinished = true
				return errToolCallsStreamFinished
			}
			event.Result.ToolCalls = completions.ToolCallsFromParsed(calls, false)
			event.Result.ToolContent = completions.ToolCallPrefixText(event.Text)
			event.Result.FinishReason = "tool_calls"
			prefixItems := []responses.OutputItem(nil)
			startIndex := 0
			if textItemStarted {
				textItem, err := finishResponsesTextItem(c, stream, itemID, 0, event.Result.ToolContent)
				if err != nil {
					return err
				}
				prefixItems = append(prefixItems, textItem)
				startIndex = 1
			}
			if err := writeResponsesToolCallEvents(c, stream, responseID, apiReq.Model, created, event.Result.ToolCalls, prefixItems, startIndex, responseOptions); err != nil {
				return err
			}
			toolStreamFinished = true
			return errToolCallsStreamFinished
		}

		detected, content := detector.ProcessChunk(event.Delta)
		if content != "" {
			if !textItemStarted {
				if err := startResponsesTextItem(c, stream, itemID, 0); err != nil {
					return err
				}
				textItemStarted = true
			}
			if _, err := c.Writer.WriteString(stream.SSE(responses.Event{Type: "response.output_text.delta", ItemID: itemID, OutputIndex: responses.Int(0), ContentIndex: responses.Int(0), Delta: content})); err != nil {
				return err
			}
			c.Writer.Flush()
		}
		if detected {
			return nil
		}
		return nil
	})
	if err != nil && err != errToolCallsStreamFinished {
		return nil, err
	}
	if result == nil {
		result = &chatResult{}
	}
	if toolStreamFinished {
		_, _ = c.Writer.WriteString("data: [DONE]\n\n")
		c.Writer.Flush()
		return result, nil
	}
	if detector.State() == "tool_parsing" {
		if calls := detector.Finalize(); len(calls) > 0 && completions.ValidateParsedToolCalls(calls, apiReq.Tools) == nil {
			result.ToolCalls = completions.ToolCallsFromParsed(calls, false)
			result.ToolContent = completions.ToolCallPrefixText(result.Content)
			result.FinishReason = "tool_calls"
			prefixItems := []responses.OutputItem(nil)
			startIndex := 0
			if textItemStarted {
				textItem, err := finishResponsesTextItem(c, stream, itemID, 0, result.ToolContent)
				if err != nil {
					return nil, err
				}
				prefixItems = append(prefixItems, textItem)
				startIndex = 1
			}
			if err := writeResponsesToolCallEvents(c, stream, responseID, apiReq.Model, created, result.ToolCalls, prefixItems, startIndex, responseOptions); err != nil {
				return nil, err
			}
			_, _ = c.Writer.WriteString("data: [DONE]\n\n")
			c.Writer.Flush()
			return result, nil
		}
	} else if text := detector.FlushText(); text != "" {
		if !textItemStarted {
			if err := startResponsesTextItem(c, stream, itemID, 0); err != nil {
				return nil, err
			}
			textItemStarted = true
		}
		if _, err := c.Writer.WriteString(stream.SSE(responses.Event{Type: "response.output_text.delta", ItemID: itemID, OutputIndex: responses.Int(0), ContentIndex: responses.Int(0), Delta: text})); err != nil {
			return nil, err
		}
	}
	if textItemStarted {
		completedItem, err := finishResponsesTextItem(c, stream, itemID, 0, result.Content)
		if err != nil {
			return nil, err
		}
		if _, err := c.Writer.WriteString(stream.SSE(responses.CompletedEvent(responseID, apiReq.Model, created, []responses.OutputItem{completedItem}, responseOptions))); err != nil {
			return nil, err
		}
	} else {
		if _, err := c.Writer.WriteString(stream.SSE(responses.CompletedEvent(responseID, apiReq.Model, created, []responses.OutputItem{}, responseOptions))); err != nil {
			return nil, err
		}
	}
	_, _ = c.Writer.WriteString("data: [DONE]\n\n")
	c.Writer.Flush()
	return result, nil
}

func writeResponsesLifecycleStart(c *gin.Context, stream *responses.EventStream, responseID string, model string, created int64, responseOptions responses.ResponseOptions) error {
	if _, err := c.Writer.WriteString(stream.SSE(responses.CreatedEvent(responseID, model, created, responseOptions))); err != nil {
		return err
	}
	if _, err := c.Writer.WriteString(stream.SSE(responses.InProgressEvent(responseID, model, created, responseOptions))); err != nil {
		return err
	}
	return nil
}

func startResponsesTextItem(c *gin.Context, stream *responses.EventStream, itemID string, outputIndex int) error {
	item := responses.TextOutputItem(itemID, "", "in_progress")
	if _, err := c.Writer.WriteString(stream.SSE(responses.Event{Type: "response.output_item.added", OutputIndex: responses.Int(outputIndex), Item: &item})); err != nil {
		return err
	}
	part := responses.OutputTextPart("")
	if _, err := c.Writer.WriteString(stream.SSE(responses.Event{Type: "response.content_part.added", ItemID: itemID, OutputIndex: responses.Int(outputIndex), ContentIndex: responses.Int(0), Part: &part})); err != nil {
		return err
	}
	return nil
}

func finishResponsesTextItem(c *gin.Context, stream *responses.EventStream, itemID string, outputIndex int, text string) (responses.OutputItem, error) {
	if _, err := c.Writer.WriteString(stream.SSE(responses.Event{Type: "response.output_text.done", ItemID: itemID, OutputIndex: responses.Int(outputIndex), ContentIndex: responses.Int(0), Text: text})); err != nil {
		return responses.OutputItem{}, err
	}
	part := responses.OutputTextPart(text)
	if _, err := c.Writer.WriteString(stream.SSE(responses.Event{Type: "response.content_part.done", ItemID: itemID, OutputIndex: responses.Int(outputIndex), ContentIndex: responses.Int(0), Part: &part})); err != nil {
		return responses.OutputItem{}, err
	}
	completedItem := responses.TextOutputItem(itemID, text, "completed")
	if _, err := c.Writer.WriteString(stream.SSE(responses.Event{Type: "response.output_item.done", OutputIndex: responses.Int(outputIndex), Item: &completedItem})); err != nil {
		return responses.OutputItem{}, err
	}
	return completedItem, nil
}

func writeResponsesToolCallEvents(c *gin.Context, stream *responses.EventStream, responseID string, model string, created int64, toolCalls []completions.ToolCall, prefixItems []responses.OutputItem, startIndex int, responseOptions responses.ResponseOptions) error {
	output := make([]responses.OutputItem, 0, len(prefixItems)+len(toolCalls))
	output = append(output, prefixItems...)
	for i, toolCall := range toolCalls {
		outputIndex := startIndex + i
		itemID := responses.MessageID()
		item := responses.FunctionCallOutputItem(itemID, toolCall.ID, toolCall.Function.Name, "", "in_progress")
		if _, err := c.Writer.WriteString(stream.SSE(responses.Event{Type: "response.output_item.added", OutputIndex: responses.Int(outputIndex), Item: &item})); err != nil {
			return err
		}
		if _, err := c.Writer.WriteString(stream.SSE(responses.Event{Type: "response.function_call_arguments.delta", ItemID: itemID, OutputIndex: responses.Int(outputIndex), Delta: toolCall.Function.Arguments})); err != nil {
			return err
		}
		if _, err := c.Writer.WriteString(stream.SSE(responses.Event{Type: "response.function_call_arguments.done", ItemID: itemID, OutputIndex: responses.Int(outputIndex), Arguments: toolCall.Function.Arguments})); err != nil {
			return err
		}
		completed := responses.FunctionCallOutputItem(itemID, toolCall.ID, toolCall.Function.Name, toolCall.Function.Arguments, "completed")
		if _, err := c.Writer.WriteString(stream.SSE(responses.Event{Type: "response.output_item.done", OutputIndex: responses.Int(outputIndex), Item: &completed})); err != nil {
			return err
		}
		output = append(output, completed)
	}
	if _, err := c.Writer.WriteString(stream.SSE(responses.CompletedEvent(responseID, model, created, output, responseOptions))); err != nil {
		return err
	}
	c.Writer.Flush()
	return nil
}
