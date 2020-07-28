package infsstore

import (
	"encoding/json"
	"io"

	gs "github.com/windnow/edusrv/internal/gameserver"
)

// FileSystemPlayerStore ...
type FileSystemPlayerStore struct {
	database io.Reader
}

// GetLeague ...
func (f *FileSystemPlayerStore) GetLeague() []gs.Player {
	var league []gs.Player
	json.NewDecoder(f.database).Decode(&league)
	return league
}

// NewFileSystemPlayerStore ...
func NewFileSystemPlayerStore(database io.Reader) *FileSystemPlayerStore {
	return &FileSystemPlayerStore{database}
}
