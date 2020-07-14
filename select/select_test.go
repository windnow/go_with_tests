package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makeDelayedServer(delay time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
				<html>
					<head>
						<title>test</title>
					</head>
					<body>
					<h1>Hello, world</h1>
					<p>This is just test server</p>
				</html>
				`))
	}))

}
func TestRacer(t *testing.T) {
	t.Run("check fast and slow URL's", func(t *testing.T) {
		slowServer := makeDelayedServer(20 * time.Millisecond)
		fastServer := makeDelayedServer(0 * time.Millisecond)

		defer func() {
			fmt.Println("Closing servers")
			slowServer.Close()
			fastServer.Close()
			fmt.Println("Closed")
		}()

		slowURL := slowServer.URL
		fastURL := fastServer.URL

		fmt.Println("Slow URL:", slowURL)
		fmt.Println("Fast URL:", fastURL)

		want := fastURL
		got, _ := Racer(slowURL, fastURL)

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("returns an error if a server doesn't respond within 10s", func(t *testing.T) {
		// server, that wait 100 millesecond, before request
		server := makeDelayedServer(100 * time.Millisecond)

		defer func() {
			server.Close()
		}()

		_, err := ConfigurableRacer(server.URL, server.URL, 200*time.Millisecond)
		if err == nil {
			t.Error("Expected an error but didn't get one")
		}
	})
}
