package infsstore

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	gs "github.com/windnow/edusrv/internal/gameserver"
	"github.com/windnow/edusrv/internal/tape"
)

// FileSystemPlayerStore ...
type FileSystemPlayerStore struct {
	database *json.Encoder
	league   gs.League
}

// GetLeague ...
func (f *FileSystemPlayerStore) GetLeague() gs.League {
	sort.Slice(f.league, func(i, j int) bool {
		return f.league[i].Wins > f.league[j].Wins
	})
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

	f.database.Encode(f.league)
}

// NewFileSystemPlayerStore ...
func NewFileSystemPlayerStore(file *os.File) (*FileSystemPlayerStore, error) {

	err := initialisePlayerDBFile(file)
	if err != nil {
		return nil, fmt.Errorf("problem initialising player db file, %v", err)
	}

	league, err := gs.NewLeague(file)

	if err != nil {
		return nil, fmt.Errorf("problem loading player store from file %s, %v", file.Name(), err)
	}

	return &FileSystemPlayerStore{
		database: json.NewEncoder(&tape.Tape{
			File: file,
		}),
		league: league,
	}, nil
}

func initialisePlayerDBFile(file *os.File) error {
	file.Seek(0, 0)
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("problem getting file info from file %v, %s", file.Name(), err)
	}

	if info.Size() == 0 {
		file.Write([]byte("[]"))
		file.Seek(0, 0)
	}
	return nil
}
