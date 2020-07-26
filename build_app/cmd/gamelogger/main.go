package main

import (
	"log"
	"net/http"

	"github.com/windnow/edusrv/internal/gameserver"
)

type inMemoryPlayerStore struct{}

func (i *inMemoryPlayerStore) GetPlayerScore(name string) int {
	return 123
}

func main() {
	server := gameserver.NewServer(&inMemoryPlayerStore{})

	if err := http.ListenAndServe(":5000", server); err != nil {
		log.Fatalf("could not listen on port 5000 %v", err)
	}
}
