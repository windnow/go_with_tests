package main

import (
	"log"
	"net/http"

	"github.com/windnow/edusrv/internal/gameserver"
	"github.com/windnow/edusrv/internal/inmemstore"
)

func main() {
	server := gameserver.NewServer(inmemstore.NewInMemoryStore())

	if err := http.ListenAndServe(":5000", server); err != nil {
		log.Fatalf("could not listen on port 5000 %v", err)
	}
}
