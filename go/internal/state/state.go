package state

import (
	"sync"

	"rofi-chrome-tab/internal/types"
)

type State struct {
	mu   sync.RWMutex
	tabs []types.Tab
}

func (s *State) Tabs() []types.Tab {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.tabs) == 0 {
		return nil
	}

	cpy := make([]types.Tab, len(s.tabs))
	copy(cpy, s.tabs)
	return cpy
}

func (s *State) SetTabs(tabs []types.Tab) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(tabs) == 0 {
		s.tabs = nil
		return
	}

	cpy := make([]types.Tab, len(tabs))
	copy(cpy, tabs)
	s.tabs = cpy
}
