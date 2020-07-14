package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Sleeper mock `Sleep`
type Sleeper interface {
	Sleep()
}

// ConfigurableSleeper ...
type ConfigurableSleeper struct {
	duration time.Duration
	sleep    func(time.Duration)
}

// Sleep ...
func (c *ConfigurableSleeper) Sleep() {
	c.sleep(c.duration)
}

// DefaultSleeper implement `Sleep` method of `time`
type DefaultSleeper struct{}

// Sleep ... implement
func (d *DefaultSleeper) Sleep() {
	time.Sleep(1 * time.Second)
}
func main() {
	//sleeper := &DefaultSleeper{}
	sleeper := &ConfigurableSleeper{10 * time.Second, time.Sleep}
	Countdown(os.Stdout, sleeper)
}

const finalWord = "Go!"
const countdownStart = 3

// Countdown ..
func Countdown(out io.Writer, sleeper Sleeper) {

	for i := countdownStart; i > 0; i-- {
		sleeper.Sleep()
		fmt.Fprintln(out, i)
	}

	sleeper.Sleep()
	fmt.Fprint(out, finalWord)
}
