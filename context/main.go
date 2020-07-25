package main

import (
	"context"
	"fmt"
	"net/http"
)

// Store ...
type Store interface {
	Fetch(ctx context.Context) (string, error)
}

// Server ...
func Server(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := store.Fetch(r.Context())
		if err != nil {
			return
		}
		fmt.Fprint(w, data)
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
