package things

import (
	"fmt"
	"io"
	"sync"
)

// GetErrors combines multiple errors into one.
func GetErrors(errors ...error) error {
	return errorList(errors).Get()
}

// Errors is an error tracker.
// It is thread safe. Unfinished
type Errors struct {
	errors   errorList
	errChans []chan error
	mutex    sync.Mutex
}

// GetErrChan returns a channel where new errors are written to.
// Every call creates a new channel, and channels are expected to be read from.
func (e *Errors) GetErrChan() <-chan error {
	c := make(chan error)
	e.mutex.Lock()
	e.errChans = append(e.errChans, c)
	e.mutex.Unlock()
	return c
}

// Add adds errors to the system, writing non-nil errors to channels returned from GetErrChan
// and adding them to the error list.
func (e *Errors) Add(errors ...error) {
	e.mutex.Lock()
	for _, err := range errors {
		if err == nil {
			continue
		}

		e.errors = append(e.errors, err)
		for _, c := range e.errChans {
			c <- err
		}
	}
	e.mutex.Unlock()
}

// Errored returns true if an error has been recorded.
func (e *Errors) Errored() bool {
	e.mutex.Lock()
	if len(e.errors) > 0 {
		e.mutex.Unlock()
		return true
	}
	e.mutex.Unlock()
	return false
}

// CloseOnErr will close the closer when an error is encountered,
// adding the error returned by Close if non-nil.
func (e *Errors) CloseOnErr(closer io.Closer) {
	c := e.GetErrChan()
	// do this in a new routine because we don't want to block if nobody is listening
	go func() {
		_, ok := <-c
		if !ok {
			return
		}
		e.Add(closer.Close())
	}()
}

// Close closes Errors.
// All error channels are closed, and the errors, if any,
// are returned and removed from the system.
// It is safe to re-use Errors after calling Close.
func (e *Errors) Close() error {
	e.mutex.Lock()
	for _, c := range e.errChans {
		close(c)
	}
	e.errChans = nil
	err := e.errors.Get()
	e.errors = nil
	e.mutex.Unlock()
	return err
}

type errorList []error

func (e errorList) Get() error {
	switch len(e) {
	case 0:
		return nil
	case 1:
		return e[0]
	default:
		return e
	}
}

func (e errorList) Error() (msg string) {
	for i, err := range e {
		if i > 0 {
			msg += "\n"
		}
		msg += fmt.Sprintf("error %v: %v", i+1, err.Error())
	}
	return
}
