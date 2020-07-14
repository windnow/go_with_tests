package main

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

// CountdownOperationSpy ...
type CountdownOperationsSpy struct {
	Calls []string
}

const write = "write"
const sleep = "sleep"

// Sleep spy for sleep method
func (s *CountdownOperationsSpy) Sleep() {
	s.Calls = append(s.Calls, sleep)
}

// Write spy for write method
func (s *CountdownOperationsSpy) Write(w []byte) (n int, err error) {
	s.Calls = append(s.Calls, write)
	return
}

type SpyTime struct {
	durationSlept time.Duration
}

func (s *SpyTime) Sleep(duration time.Duration) {
	s.durationSlept = duration
}

// SpySleeper struct for implement mock behavior
type SpySleeper struct {
	Calls int
}

// Sleep Mock func
func (s *SpySleeper) Sleep() {
	s.Calls++
}
func TestCountdown(t *testing.T) {
	t.Run("prints 3 to Go!", func(t *testing.T) {
		bufer := &bytes.Buffer{}
		Countdown(bufer, &CountdownOperationsSpy{})

		got := bufer.String()
		want := `3
2
1
Go!`

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})
	t.Run("sleep before every print", func(t *testing.T) {
		spySleepPrinter := &CountdownOperationsSpy{}
		Countdown(spySleepPrinter, spySleepPrinter)

		want := []string{
			sleep, write,
			sleep, write,
			sleep, write,
			sleep, write,
		}

		if !reflect.DeepEqual(want, spySleepPrinter.Calls) {
			t.Errorf("wanted calls %v got %v", want, spySleepPrinter.Calls)
		}
	})
}

func TestConfigurableSleeper(t *testing.T) {
	sleepTime := 5 * time.Second

	spyTime := &SpyTime{}
	sleeper := ConfigurableSleeper{sleepTime, spyTime.Sleep}
	sleeper.Sleep()

	if spyTime.durationSlept != sleepTime {
		t.Errorf("should have slept for %v but slept for %v", sleepTime, spyTime.durationSlept)
	}

}
