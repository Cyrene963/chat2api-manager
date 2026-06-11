package service

import (
	"strings"

	"chat2api/app/token_pool"
)

func recordAccessTokenFailure(accessToken string, reason string) {
	token_pool.GetAccessTokenPool().RecordFailure(accessToken, reason)
}

func recordAccessTokenSuccess(accessToken string) {
	token_pool.GetAccessTokenPool().RecordSuccess(accessToken)
}

func recordAccessTokenOutcome(accessToken string, result *chatResult) {
	if hasChatResultOutput(result) {
		recordAccessTokenSuccess(accessToken)
		return
	}
	recordAccessTokenFailure(accessToken, "empty upstream response")
}

func hasChatResultOutput(result *chatResult) bool {
	if result == nil {
		return false
	}
	if strings.TrimSpace(result.Content) != "" {
		return true
	}
	return len(result.ToolCalls) > 0
}
