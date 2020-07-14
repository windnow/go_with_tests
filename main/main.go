package main

import (
	"fmt"
	"net/http"

	"github.com/windnow/go_with_tests/dependinj"
)

type my_str string

func (m my_str) Write(b []byte) (int, error) {
	fmt.Println("=====>", string(b))
	return 0, nil
}

func main() {
	var m my_str
	dependinj.Greet(m, "Yermek")
	myLst()
}

func myGreeterHandler(w http.ResponseWriter, r *http.Request) {
	dependinj.Greet(w, "Guest")
}

func myLst() {
	http.ListenAndServe(":8000", http.HandlerFunc(myGreeterHandler))
}
