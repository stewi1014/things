package things

import (
	"sync"
	"time"
)

// OnlyEvery returns a function that will only return true once for every time period
func Limit(period time.Duration) func() bool {
	return (&limiter{duration: period, next: time.Now().Add(period)}).ok
}

type limiter struct {
	duration time.Duration
	next     time.Time
	mutex    sync.RWMutex
}

func (l *limiter) ok() bool {
	l.mutex.RLock()
	if l.next.Before(time.Now()) {
		l.mutex.RUnlock()
		l.mutex.Lock()
		l.next = time.Now().Add(l.duration)
		l.mutex.Unlock()
		return true
	}
	l.mutex.RUnlock()
	return false
}

// Counter keeps track of how many times Counter.Count has been called within period
type Counter struct {
	mutex  sync.Mutex
	period time.Duration
	number int

	callback     func()
	callbackTh   int
	callbackFlip bool
}

// NewCounter creates a new counter with the given period
func NewCounter(period time.Duration) *Counter {
	c := new(Counter)
	return c
}

// Count Adds one to the counter and returns the current count
func (c *Counter) Count() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.number++

	if c.callbackTh != 0 {
		if c.callbackFlip && c.number < c.callbackTh {
			c.callbackFlip = false
		} else if !c.callbackFlip && c.number >= c.callbackTh {
			c.callback()
			c.callbackFlip = true
		}
	}

	return c.number
}

// Number returns the current number
func (c *Counter) Number() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.number
}

// SetCallback sets the counter's callback function
// The callback function is called when the callback threshold is reached
func (c *Counter) SetCallback(f func()) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.callback = f
}

// SetThreshold sets the counter's callback threshold
func (c *Counter) SetThreshold(n int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.callbackTh = n
}

// RunFunc is a function that takes a time.Duration indicating the time since the last call
type RunFunc func(time.Duration)

// Run creates a new Runner, calls Runner.Run with the given arguments and returns the runner
func Run(runFunc RunFunc, d time.Duration) *Runner {
	r := NewRunner()
	r.Run(runFunc, d)
	return r
}

// NewRunner creates a new Runner
func NewRunner() *Runner {
	r := Runner(make(chan struct{}))
	return &r
}

// Runner is a simple way to control reoccurring tasks
type Runner chan struct{}

// Run calls the given RunFunc after every duration of d
// Run blocks until it is stopped by Stop or StopOne
// Run can be called multiple times with different function, however no distinction between the functions will be made.
// runFunc will not be called until it's previous call returns.
func (r *Runner) Run(runFunc RunFunc, d time.Duration) {
	ticker := time.NewTicker(d)
	last := time.Now()

	for {
		<-ticker.C

		select {
		case <-*r:
			return
		}

		runFunc(time.Now().Sub(last))
		last = time.Now()
	}
}

// StopOne will stop the ongoing execution of one runFunc.
// The halted routine will be the next to reach it's timeout
// StopOne blocks until the routien is stopped.
func (r *Runner) StopOne() {
	*r <- struct{}{}
}
