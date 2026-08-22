package store

import (
	"errors"
	"example.com/bridge-health-monitor-service/domain"
	"sync"
)

var ErrNotFound = errors.New("bridge not found")

type Store struct {
	mu      sync.RWMutex
	items   map[string]domain.Bridge
	history map[string][]string
}

func New() *Store {
	s := &Store{items: map[string]domain.Bridge{}, history: map[string][]string{}}
	for _, b := range []domain.Bridge{
		{ID: "br-201", Name: "东江大桥", River: "东江", Condition: "monitored", LastSurvey: "2026-07-30", RiskScore: 24, Revision: 1},
		{ID: "br-202", Name: "港湾跨线桥", River: "海湾航道", Condition: "watch", LastSurvey: "2026-08-03", RiskScore: 61, Revision: 1},
	} {
		s.items[b.ID] = b
	}
	return s
}

func (s *Store) List() []domain.Bridge {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Bridge, 0, len(s.items))
	for _, b := range s.items {
		out = append(out, b)
	}
	return out
}

func (s *Store) Get(id string) (domain.Bridge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.items[id]
	if !ok {
		return domain.Bridge{}, ErrNotFound
	}
	return b, nil
}

// UpdateCondition applies a new condition when the expected revision matches.
func (s *Store) UpdateCondition(id, value string, expected int) (domain.Bridge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.items[id]
	if !ok {
		return domain.Bridge{}, ErrNotFound
	}
	if expected > 0 && current.Revision != expected {
		return domain.Bridge{}, ErrRevisionConflict
	}
	current.Condition = value
	current.Revision++
	s.items[id] = current
	s.history[id] = append(s.history[id], value)
	return current, nil
}

func (s *Store) History(id string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]string(nil), s.history[id]...)
	return out
}

var ErrRevisionConflict = errors.New("bridge revision conflict")
