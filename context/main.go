package main

import (
	"fmt"
	"net/http"
)

// Store ...
type Store interface {
	Fetch() string
	Cancel()
}

// Server ...
func Server(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, store.Fetch())
	}
}

func main() {

}

// Data ...
type Data struct {
	resp string
}

// Fetch ...
func (d *Data) Fetch() string {
	return d.resp
}
