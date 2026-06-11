package completions

import "testing"

func TestBuildChatRequestPassesModelThrough(t *testing.T) {
	for _, model := range []string{
		"gpt-5-4-thinking",
		"gpt-5-5",
		"gpt-5-5-instant",
		"gpt-5-5-pro",
		"gpt-5-5-thinking",
		"gpt-5.4",
		"gpt-5.4-mini",
		"gpt-5.4-pro",
		"gpt-5.5",
		"gpt-5.5-pro",
		"o3-deep-research",
		"o4-mini-deep-research",
	} {
		t.Run(model, func(t *testing.T) {
			req := BuildChatRequest(&ApiReq{Model: model})
			if req.Model != model {
				t.Fatalf("model = %q, want %q", req.Model, model)
			}
		})
	}
}
