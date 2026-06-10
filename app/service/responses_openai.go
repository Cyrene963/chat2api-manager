package service

import (
	"bytes"
	"chat2api/app/common"
	"chat2api/app/conf"
	"chat2api/app/types/responses"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func shouldUseOpenAIResponses(apiReq *responses.ApiReq) bool {
	if apiReq == nil {
		return false
	}
	if apiReq.Background || strings.TrimSpace(apiReq.PreviousResponseID) != "" || len(apiReq.Include) > 0 {
		return true
	}
	if isDeepResearchModel(apiReq.Model) {
		return true
	}
	for _, tool := range apiReq.Tools {
		switch strings.TrimSpace(tool.Type) {
		case "", "function", "image_generation":
			continue
		default:
			return true
		}
	}
	return false
}

func isDeepResearchModel(model string) bool {
	switch strings.TrimSpace(model) {
	case "o3-deep-research", "o4-mini-deep-research":
		return true
	default:
		return false
	}
}

func runOpenAIResponses(c *gin.Context, apiReq *responses.ApiReq) error {
	appConf := conf.GetApp()
	apiKey := strings.TrimSpace(appConf.OpenAIApiKey)
	if apiKey == "" {
		return fmt.Errorf("openai_api_key is required for deep research / OpenAI Responses mode")
	}
	payload := normalizeOpenAIResponsesRequest(apiReq)
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	baseURL := strings.TrimSpace(appConf.OpenAIBaseUrl)
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	resp, err := doOpenAIResponsesRequest(c, http.MethodPost, strings.TrimRight(baseURL, "/")+"/responses", apiKey, body, payload.Stream)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return proxyOpenAIResponsesResponse(c, resp, payload.Stream)
}

func OpenAIResponseByID(c *gin.Context) {
	if err := openAIResponseByID(c); err != nil {
		common.ErrorResponse(c, http.StatusBadGateway, "openai responses request failed", err.Error())
	}
}

func openAIResponseByID(c *gin.Context) error {
	appConf := conf.GetApp()
	apiKey := strings.TrimSpace(appConf.OpenAIApiKey)
	if apiKey == "" {
		return fmt.Errorf("openai_api_key is required for OpenAI Responses polling")
	}
	responseID := strings.TrimSpace(c.Param("id"))
	if responseID == "" {
		common.ErrorResponse(c, http.StatusBadRequest, "response id is required", nil)
		return nil
	}
	baseURL := strings.TrimSpace(appConf.OpenAIBaseUrl)
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	resp, err := doOpenAIResponsesRequest(c, http.MethodGet, strings.TrimRight(baseURL, "/")+"/responses/"+responseID, apiKey, nil, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return proxyOpenAIResponsesResponse(c, resp, false)
}

func normalizeOpenAIResponsesRequest(apiReq *responses.ApiReq) responses.ApiReq {
	payload := *apiReq
	payload.Model = strings.TrimSpace(payload.Model)
	payload.Tools = normalizeOpenAITools(payload.Tools)
	if payload.Model == "" && (payload.Background || len(payload.Include) > 0 || len(payload.Tools) > 0 || strings.TrimSpace(payload.PreviousResponseID) != "") {
		payload.Model = "o3-deep-research"
	}
	if payload.Background || isDeepResearchModel(payload.Model) {
		payload.Background = true
		if payload.Reasoning == nil {
			payload.Reasoning = &responses.Reasoning{}
		}
		if payload.Reasoning.Context == "" {
			payload.Reasoning.Context = "current_turn"
		}
		if payload.Reasoning.Summary == nil {
			payload.Reasoning.Summary = "auto"
		}
		if len(payload.Tools) == 0 {
			payload.Tools = []responses.Tool{{Type: "web_search_preview"}}
		}
		if payload.ToolChoice == nil {
			payload.ToolChoice = "required"
		}
	}
	if payload.Reasoning != nil && !containsString(payload.Include, "reasoning.encrypted_content") {
		payload.Include = append(payload.Include, "reasoning.encrypted_content")
	}
	return payload
}

func normalizeOpenAITools(tools []responses.Tool) []responses.Tool {
	normalized := make([]responses.Tool, 0, len(tools))
	for _, tool := range tools {
		switch strings.TrimSpace(tool.Type) {
		case "web_search":
			tool.Type = "web_search_preview"
		}
		normalized = append(normalized, tool)
	}
	return normalized
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == target {
			return true
		}
	}
	return false
}

func doOpenAIResponsesRequest(c *gin.Context, method string, url string, apiKey string, body []byte, stream bool) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(c.Request.Context(), method, url, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	} else {
		req.Header.Set("Accept", "application/json")
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func proxyOpenAIResponsesResponse(c *gin.Context, resp *http.Response, stream bool) error {
	for key, values := range resp.Header {
		if strings.EqualFold(key, "content-length") || strings.EqualFold(key, "transfer-encoding") {
			continue
		}
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}
	c.Status(resp.StatusCode)
	if !stream {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		_, err = c.Writer.Write(data)
		return err
	}
	buffer := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := c.Writer.Write(buffer[:n]); writeErr != nil {
				return writeErr
			}
			c.Writer.Flush()
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	c.Writer.Flush()
	return nil
}
