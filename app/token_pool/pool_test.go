package token_pool

import "testing"

func TestGetAccessTokenHonorsRotationWindow(t *testing.T) {
	oldNow := nowUnix
	oldSettings := currentSettings()
	defer func() {
		nowUnix = oldNow
		Configure(oldSettings)
	}()

	nowUnix = func() int64 { return 1000 }
	Configure(Settings{
		RotationEnabled:       true,
		RotationWindowSeconds: 60,
		RotationMaxUses:       2,
		BadThreshold:          3,
	})

	pool := &AccessTokenPool{
		AccessTokens: []*AccessToken{
			{Token: "Bearer token-a", ExpiresAt: 2000},
		},
		index: -1,
	}

	if got := pool.GetAccessToken(); got == nil || got.Token != "Bearer token-a" {
		t.Fatalf("first use = %+v, want token-a", got)
	}
	if got := pool.GetAccessToken(); got == nil || got.Token != "Bearer token-a" {
		t.Fatalf("second use = %+v, want token-a", got)
	}
	if got := pool.GetAccessToken(); got != nil {
		t.Fatalf("third use = %+v, want nil during cooldown", got)
	}

	nowUnix = func() int64 { return 1061 }
	if got := pool.GetAccessToken(); got == nil || got.Token != "Bearer token-a" {
		t.Fatalf("post-window use = %+v, want token-a", got)
	}
}

func TestRecordFailureMarksBadToken(t *testing.T) {
	oldNow := nowUnix
	oldSettings := currentSettings()
	defer func() {
		nowUnix = oldNow
		Configure(oldSettings)
	}()

	nowUnix = func() int64 { return 2000 }
	Configure(Settings{BadThreshold: 3})

	pool := &AccessTokenPool{
		AccessTokens: []*AccessToken{
			{Token: "Bearer token-b", ExpiresAt: 3000},
		},
		index: -1,
	}

	if marked := pool.RecordFailure("Bearer token-b", "first failure"); marked {
		t.Fatal("first failure should not mark token")
	}
	if marked := pool.RecordFailure("Bearer token-b", "second failure"); marked {
		t.Fatal("second failure should not mark token")
	}
	if marked := pool.RecordFailure("Bearer token-b", "third failure"); !marked {
		t.Fatal("third failure should mark token")
	}
	if got := pool.MarkedSize(); got != 1 {
		t.Fatalf("marked size = %d, want 1", got)
	}
	if got := pool.GetAccessToken(); got != nil {
		t.Fatalf("marked token should not be reusable, got %+v", got)
	}
}
