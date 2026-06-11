package token_pool

import (
	"chat2api/app/common"
	"strings"
	"sync"
)

var (
	instance *AccessTokenPool
	once     sync.Once

	nowUnix = func() int64 {
		return common.GetTimestampSecond(0)
	}

	settingsMu sync.RWMutex
	settings   = Settings{BadThreshold: 3}
)

type Settings struct {
	RotationEnabled       bool
	RotationWindowSeconds int64
	RotationMaxUses       int
	BadThreshold          int
}

type AccessTokenPool struct {
	mu           sync.Mutex
	AccessTokens []*AccessToken
	index        int
}

type AccessToken struct {
	Token     string `yaml:"token,omitempty"`
	ExpiresAt int64  `yaml:"expires_at,omitempty"`
	Proxy     string `yaml:"proxy,omitempty"`

	CanUseAt              int64  `yaml:"-"`
	RotationCanUseAt      int64  `yaml:"-"`
	RotationWindowStartAt int64  `yaml:"-"`
	RotationUseCount      int    `yaml:"-"`
	FailureCount          int    `yaml:"-"`
	MarkedAt              int64  `yaml:"-"`
	MarkedReason          string `yaml:"-"`
	LastFailureAt         int64  `yaml:"-"`
	LastFailureReason     string `yaml:"-"`
}

func Configure(next Settings) {
	settingsMu.Lock()
	defer settingsMu.Unlock()
	settings = normalizeSettings(next)
}

func currentSettings() Settings {
	settingsMu.RLock()
	defer settingsMu.RUnlock()
	return settings
}

func normalizeSettings(next Settings) Settings {
	if next.BadThreshold <= 0 {
		next.BadThreshold = 3
	}
	if next.RotationWindowSeconds <= 0 || next.RotationMaxUses <= 0 {
		next.RotationEnabled = false
	}
	return next
}

func newAccessTokenPool() *AccessTokenPool {
	return &AccessTokenPool{
		AccessTokens: make([]*AccessToken, 0),
		index:        -1,
	}
}

func GetAccessTokenPool() *AccessTokenPool {
	once.Do(func() {
		instance = newAccessTokenPool()
	})
	return instance
}

func (a *AccessTokenPool) AddAccessToken(accessToken *AccessToken) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.AccessTokens = append(a.AccessTokens, accessToken)
}

func (a *AccessTokenPool) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.AccessTokens = make([]*AccessToken, 0)
	a.index = -1
}

func (a *AccessTokenPool) AppendAccessTokens(accessTokens []*AccessToken) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.AccessTokens = append(a.AccessTokens, accessTokens...)
}

func (a *AccessTokenPool) Size() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.AccessTokens)
}

func (a *AccessTokenPool) MarkedSize() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	count := 0
	for _, v := range a.AccessTokens {
		if v != nil && v.MarkedAt > 0 {
			count++
		}
	}
	return count
}

func (a *AccessTokenPool) IsEmpty() bool {
	return a.CanUseSize() == 0
}

func (a *AccessTokenPool) CanUseSize() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := nowUnix()
	cfg := currentSettings()
	count := 0
	for _, v := range a.AccessTokens {
		if a.isAvailableLocked(v, now, cfg) {
			count++
		}
	}
	return count
}

func (a *AccessTokenPool) GetToken() string {
	accessToken := a.GetAccessToken()
	if accessToken == nil {
		return ""
	}
	return accessToken.Token
}

func (a *AccessTokenPool) GetAccessToken() *AccessToken {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.AccessTokens) == 0 {
		return nil
	}

	now := nowUnix()
	cfg := currentSettings()
	total := len(a.AccessTokens)

	for i := 0; i < total; i++ {
		a.index = (a.index + 1) % total
		token := a.AccessTokens[a.index]
		if !a.isAvailableLocked(token, now, cfg) {
			continue
		}
		a.reserveLocked(token, now, cfg)
		return token
	}

	return nil
}

func (a *AccessTokenPool) SetCanUseAt(token string, canUseAt int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	token = normalizeTokenString(token)
	for _, v := range a.AccessTokens {
		if normalizeTokenString(v.Token) == token {
			v.CanUseAt = canUseAt
			break
		}
	}
}

func (a *AccessTokenPool) RecordSuccess(token string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	token = normalizeTokenString(token)
	if token == "" {
		return
	}
	for _, v := range a.AccessTokens {
		if normalizeTokenString(v.Token) != token || v.MarkedAt > 0 {
			continue
		}
		v.FailureCount = 0
		v.LastFailureReason = ""
		return
	}
}

func (a *AccessTokenPool) RecordFailure(token string, reason string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	token = normalizeTokenString(token)
	if token == "" {
		return false
	}
	cfg := currentSettings()
	now := nowUnix()
	for _, v := range a.AccessTokens {
		if normalizeTokenString(v.Token) != token || v.MarkedAt > 0 {
			continue
		}
		v.FailureCount++
		v.LastFailureAt = now
		v.LastFailureReason = strings.TrimSpace(reason)
		if v.FailureCount >= cfg.BadThreshold {
			v.MarkedAt = now
			v.MarkedReason = v.LastFailureReason
			if v.MarkedReason == "" {
				v.MarkedReason = "marked after repeated failures"
			}
			v.CanUseAt = now
			v.RotationCanUseAt = now
			return true
		}
		return false
	}
	return false
}

func (a *AccessTokenPool) isAvailableLocked(token *AccessToken, now int64, cfg Settings) bool {
	if token == nil {
		return false
	}
	if token.MarkedAt > 0 {
		return false
	}
	if token.ExpiresAt > 0 && token.ExpiresAt <= now {
		return false
	}
	if token.CanUseAt > now {
		return false
	}
	if cfg.RotationEnabled && token.RotationCanUseAt > now {
		return false
	}
	return true
}

func (a *AccessTokenPool) reserveLocked(token *AccessToken, now int64, cfg Settings) {
	if !cfg.RotationEnabled {
		return
	}
	if token.RotationWindowStartAt == 0 || now-token.RotationWindowStartAt >= cfg.RotationWindowSeconds {
		token.RotationWindowStartAt = now
		token.RotationUseCount = 0
		token.RotationCanUseAt = 0
	}
	token.RotationUseCount++
	if token.RotationUseCount >= cfg.RotationMaxUses {
		token.RotationCanUseAt = token.RotationWindowStartAt + cfg.RotationWindowSeconds
	}
}

func normalizeTokenString(token string) string {
	token = strings.TrimSpace(token)
	token = strings.TrimPrefix(token, "Bearer ")
	return strings.TrimSpace(token)
}
