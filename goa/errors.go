package main

import (
	"fmt"
)

// Make it possible to report multiple errors at once
type errors []error

// Generate summary error message
func (e errors) Error() string {
	if len(e) == 0 {
		return "No error"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	msg := fmt.Sprintf("%d errors:\n", len(e))
	for i, err := range e {
		msg += fmt.Sprintf("%d %s\n", i, err.Error())
	}
	return msg
}

// Add error to list
func (e *errors) add(err error) {
	if err == nil {
		panic("goa: internal error - trying to record a nil error")
	}
	*e = append(*e, err)
}

// Only add error if not nil
func (e *errors) addIf(err error) {
	if err != nil {
		e.add(err)
	}
}
