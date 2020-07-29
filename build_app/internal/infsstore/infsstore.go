package infsstore

import (
	"encoding/json"
	"io"

	gs "github.com/windnow/edusrv/internal/gameserver"
)

// FileSystemPlayerStore ...
type FileSystemPlayerStore struct {
	database io.ReadWriteSeeker
}

// GetLeague ...
func (f *FileSystemPlayerStore) GetLeague() gs.League {
	f.database.Seek(0, 0)
	league, _ := gs.NewLeague(f.database)
	return league
}

// GetPlayerScore ...
func (f *FileSystemPlayerStore) GetPlayerScore(name string) int {

	player := f.GetLeague().Find(name)

	if player != nil {
		return player.Wins
	}

	return 0
}

// RecordWin ...
func (f *FileSystemPlayerStore) RecordWin(name string) {
	league := f.GetLeague()
	player := league.Find(name)

	if player != nil {
		player.Wins++
	} else {
		league = append(league, gs.Player{Name: name, Wins: 1})
	}

	f.database.Seek(0, 0)
	json.NewEncoder(f.database).Encode(league)
}

// NewFileSystemPlayerStore ...
func NewFileSystemPlayerStore(database io.ReadWriteSeeker) *FileSystemPlayerStore {
	return &FileSystemPlayerStore{database}
}
