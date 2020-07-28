package inmemstore

import (
	gs "github.com/windnow/edusrv/internal/gameserver"
)

// InMemoryPlayerStore ...
type InMemoryPlayerStore struct {
	store map[string]int
}

// RecordWin ...
func (i *InMemoryPlayerStore) RecordWin(name string) {
	i.store[name]++
}

// GetPlayerScore ...
func (i *InMemoryPlayerStore) GetPlayerScore(name string) int {
	return i.store[name]
}

// GetLeague ...
func (i *InMemoryPlayerStore) GetLeague() []gs.Player {
	var league []gs.Player

	for name, wins := range i.store {
		league = append(league, gs.Player{name, wins})
	}
	return league
}

// NewInMemoryStore ...
func NewInMemoryStore() *InMemoryPlayerStore {
	return &InMemoryPlayerStore{
		map[string]int{},
	}
}
