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
		ctx := r.Context()

		data := make(chan string, 1)

		go func() {
			data <- store.Fetch()
		}()
		select {
		case d := <-data:
			fmt.Fprint(w, d)
		case <-ctx.Done():
			store.Cancel()
		}
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
