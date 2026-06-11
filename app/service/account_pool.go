package service

import (
	"context"
	"errors"
	"strings"

	"chat2api/app/token_pool"
)

func recordAccessTokenFailure(accessToken string, reason string) {
	token_pool.GetAccessTokenPool().RecordFailure(accessToken, reason)
}

func recordAccessTokenRequestError(accessToken string, err error) {
	if err == nil || token_pool.ShouldIgnoreError(err) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return
	}
	recordAccessTokenFailure(accessToken, err.Error())
}

func recordAccessTokenSuccess(accessToken string) {
	token_pool.GetAccessTokenPool().RecordSuccess(accessToken)
}

func recordAccessTokenOutcome(accessToken string, result *chatResult) {
	if hasChatResultOutput(result) {
		recordAccessTokenSuccess(accessToken)
		return
	}
	if token_pool.GetAccessTokenPool().RecordEmptyResult(accessToken, "empty upstream response") {
		return
	}
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
