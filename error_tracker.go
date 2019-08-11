package things

import (
	"fmt"
)

// NewErrorTracker creates a new error tracker
func NewErrorTracker() *ErrorTracker {
	return new(ErrorTracker)
}

// Errors combines multiple errors into one.
func Errors(errors ...error) *ErrorTracker {
	et := NewErrorTracker()
	et.Add(errors...)
	return et
}

// ErrorTracker keeps track of errors that occur, allowing them to be collected and handled as one
type ErrorTracker struct {
	Errors   []error
	callback func(error)
}

// Add adds the given errors to the error set if they are non-nil.
// ErrorTrackers passed to this function are merged
func (t *ErrorTracker) Add(err ...error) {
	for _, e := range err {
		if errTracker, ok := e.(*ErrorTracker); ok {
			t.Add(errTracker.Errors...)
			continue
		}
		if e != nil {
			t.Errors = append(t.Errors, e)

			if t.callback != nil {
				t.callback(e)
			}
		}
	}
}

// Get returns nil, the collected error, or itself for 0, 1 or >1 caught errors.
func (t *ErrorTracker) Get() error {
	switch len(t.Errors) {
	case 0:
		return nil

	case 1:
		return t.Errors[0]

	default:
		return t
	}
}

// HasError returns true if any errors has been caught
func (t *ErrorTracker) HasError() bool {
	if len(t.Errors) > 0 {
		return true
	}

	return false
}

// ErrorTracker implements error
func (t *ErrorTracker) Error() string {
	switch len(t.Errors) {
	case 0:
		return "no errors! remember to use ErrorTracker.Get()"
	case 1:
		return t.Errors[0].Error()
	default:
		var str string
		for i := range t.Errors {
			if i > 0 {
				str += "\n"
			}
			str += fmt.Sprintf("%v: %v", i, t.Errors[i].Error())
		}
		return str
	}
}

// SetErrorCallback calls the given function when a non nil error is added
func (t *ErrorTracker) SetErrorCallback(callback func(error)) {
	t.callback = callback
}
