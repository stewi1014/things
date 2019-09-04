package things

import (
	"fmt"
	"strconv"
)

// Err* functions are methods for creating more detailed errors.
//
// They support two main types,
// The error slice Errors,
// and key:value storage wrapper WrappedError.
//
// They attempt to maintain a structure of a WrappedError that nests Errors,
// ommiting one or both if appropriate. Circular references are not allowed.

// ErrAdd combines multiple errors into one, wrapping with Errors if more than one error is non-nil.
func ErrAdd(err *error, errs ...error) {
	if we, ok := (*err).(WrappedError); ok {
		err = &we.Err
	}

	el, ok := (*err).(Errors)
	if !ok {
		el = Errors{
			*err,
		}
	}
	el = append(el, errs...)
	*err = el.Get()
}

// ErrSet sets the key 'key' to the value 'value', wrapping the error with WrappedError if needed.
// It can be retreived at a later point by using ErrGet.
func ErrSet(err *error, key interface{}, value interface{}) {
	we, ok := (*err).(WrappedError)
	if !ok {
		we.Err = *err
		we.Values = make(map[interface{}]interface{})
		*err = we
	}

	we.Values[key] = value
}

// ErrGet returns the value for the key 'key', returning nil and false if the value is not found.
// It recursively traverses the error tree, returning the first instance of the key.
func ErrGet(err error, key interface{}) (interface{}, bool) {
	if we, ok := err.(WrappedError); ok {
		if val, ok := we.Values[key]; ok {
			return val, true
		}
		return ErrGet(we.Err, key)
	}

	if el, ok := err.(Errors); ok {
		for i := range el {
			if val, ok := ErrGet(el[i], key); ok {
				return val, true
			}
		}
	}

	return nil, false
}

// ErrIs recursively traverses the error tree, returning true if an underlying error is 'check'.
func ErrIs(err error, check error) bool {
	return ErrCheck(err, func(err error) bool {
		return err == check
	})
}

// ErrCheck recursively traverses the error tree,
// returning true if the check function is true for any of the underlying errors.
func ErrCheck(err error, check func(error) bool) bool {
	if we, ok := err.(WrappedError); ok {
		return ErrCheck(we.Err, check)
	}

	if el, ok := err.(Errors); ok {
		for _, elerr := range el {
			if ErrCheck(elerr, check) {
				return true
			}
		}
	}

	return check(err)
}

// WrappedError wraps an error with key:value storage for information supplimentary to the error.
type WrappedError struct {
	Err    error
	Values map[interface{}]interface{}
}

// Error implements error
func (we WrappedError) Error() string {
	return we.Err.Error()
}

// String formats the key:value map along with the error message.
func (we WrappedError) String() string {
	return we.Err.Error() + "|" + fmt.Sprint(we.Values)
}

// Errors is a slice of errors.
type Errors []error

// Error implements error
func (el Errors) Error() (str string) {
	for i, err := range el {
		if i > 0 {
			str += ", "
		}
		var errMsg string
		if err == nil {
			errMsg = "<nil>" // This should only happend if Get hasn't been called.
		} else {
			errMsg = err.Error()
		}
		str += strconv.Itoa(i) + ":\"" + errMsg + "\""
	}
	return str
}

// Get returns nil, the error, or the ErrorList for 0, 1 and >1 non-nil errors in ErrorList.
// Returning Errors with no actual errors and no Get call will result in false positives.
func (el Errors) Get() error {
	end := len(el)
	for i := 0; i < end; i++ {
		if el[i] == nil {
			copy(el[i:], el[i+1:])
			end--
			i--
		}
	}
	switch end {
	case 0:
		return nil
	case 1:
		return el[0]
	default:
		return el[:end]
	}
}
