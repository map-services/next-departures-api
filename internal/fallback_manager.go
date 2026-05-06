package internal

import (
	"sync/atomic"
)

type FallbackManager interface {
	IsSiriRateLimited() bool
	SetSiriRateLimited(limited bool)
}

type fallbackManager struct {
	siriRateLimited atomic.Bool
}

func NewFallbackManager() FallbackManager {
	return &fallbackManager{}
}

func (m *fallbackManager) IsSiriRateLimited() bool {
	return m.siriRateLimited.Load()
}

func (m *fallbackManager) SetSiriRateLimited(limited bool) {
	m.siriRateLimited.Store(limited)
}
