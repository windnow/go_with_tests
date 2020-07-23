package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMyData(t *testing.T) {
	var d = Data{"Hello, World"}
	want := "Hello, World"
	got := d.Fetch()

	if want != got {
		t.Errorf("Want %q, got %q", want, got)
	}
}

type StubStore struct {
	response string
}

func (s *StubStore) Fetch() string {
	return s.response
}

func TestHandler(t *testing.T) {
	data := "hello, world"
	srv := Server(&StubStore{data})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	srv.ServeHTTP(response, request)

	if response.Body.String() != data {
		t.Errorf(`got "%s", want "%s"`, response.Body.String(), data)
	}
}
