package infsstore

import (
	"encoding/json"
	"io"
	"os"

	gs "github.com/windnow/edusrv/internal/gameserver"
	"github.com/windnow/edusrv/internal/tape"
)

// FileSystemPlayerStore ...
type FileSystemPlayerStore struct {
	database io.Writer
	league   gs.League
}

// GetLeague ...
func (f *FileSystemPlayerStore) GetLeague() gs.League {
	return f.league
}

// GetPlayerScore ...
func (f *FileSystemPlayerStore) GetPlayerScore(name string) int {

	player := f.league.Find(name)

	if player != nil {
		return player.Wins
	}

	return 0
}

// RecordWin ...
func (f *FileSystemPlayerStore) RecordWin(name string) {
	player := f.league.Find(name)

	if player != nil {
		player.Wins++
	} else {
		f.league = append(f.league, gs.Player{Name: name, Wins: 1})
	}

	json.NewEncoder(f.database).Encode(f.league)
}

// NewFileSystemPlayerStore ...
func NewFileSystemPlayerStore(database io.ReadWriteSeeker) *FileSystemPlayerStore {
	database.Seek(0, 0)
	league, _ := gs.NewLeague(database)

	return &FileSystemPlayerStore{
		database: &tape.Tape{database.(*os.File)},
		league:   league,
	}
}
