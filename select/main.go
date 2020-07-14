package main

import (
	"fmt"
	"net/http"
	"time"
)

var tenSecondDuration = 10 * time.Second

// Racer ...
func Racer(a, b string) (winner string, e error) {
	return ConfigurableRacer(a, b, tenSecondDuration)
}

// ConfigurableRacer ...
func ConfigurableRacer(a, b string, d time.Duration) (winner string, e error) {
	select {
	case <-ping(a):
		return a, nil
	case <-ping(b):
		return b, nil
	case <-time.After(d):
		//return "", errors.New(fmt.Sprintf("Timeout waiting for %s and %s", a, b))
		return "", fmt.Errorf("timeout waiting for %s and %s", a, b)
	}
}
func ping(url string) chan struct{} {
	ch := make(chan struct{})
	go func() {
		http.Get(url)
		close(ch)
	}()
	return ch
}

func measureResponseTime(url string) time.Duration {
	start := time.Now()
	http.Get(url)
	return time.Since(start)
}

func main() {
	a := "https://mail.ru"
	b := "https://google.com"

	faster, _ := Racer(a, b)
	fmt.Println(faster)
}
