package service

import (
	"chat2api/app/types/completions"
	"chat2api/app/types/responses"

	"github.com/gin-gonic/gin"
)

func runResponsesTextChat(c *gin.Context, apiReq *completions.ApiReq, streamResponses bool, responseOptions responses.ResponseOptions) (*chatResult, error) {
	chatReq := completions.BuildChatRequest(apiReq)
	resp, accessToken, err := sendChatRequest(c, chatReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if handleResponseError(c, resp, accessToken) {
		return nil, nil
	}
	if streamResponses {
		if completions.HasTools(apiReq) {
			return streamResponsesFunctionCallingEvents(c, apiReq, resp, responseOptions)
		}
		_, err := streamResponsesTextEvents(c, apiReq.Model, resp, responseOptions)
		return nil, err
	}
	return handlerResponse(c, apiReq, resp)
}
