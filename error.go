package goa

import (
	"fmt"
	"runtime"
)

// Error with stack trace
type Error interface {
	Error() string // Error message
	Stack() string // Error stack trace
}

// Argument error
type ArgumentError interface {
	Error() string         // Error message
	Stack() string         // Error stack trace
	ArgName() string       // Name of invalid argument
	ArgValue() interface{} // Value of invalid argument
}

// Max size of stack trace string in bytes
var maxStackBytes = 4096

// Build error with stack trace information
func NewError(msg string) Error {
	stack := make([]byte, maxStackBytes)
	runtime.Stack(stack, false)
	return &errorInfo{msg, string(stack)}
}

// Helper method with fmt.Errorf like behavior
func NewErrorf(format string, a ...interface{}) Error {
	stack := make([]byte, maxStackBytes)
	runtime.Stack(stack, false)
	return &errorInfo{fmt.Sprintf(format, a...), string(stack)}
}

// Build argument error from message, argument name and value
func NewArgumentError(msg string, argName string, argValue interface{}) ArgumentError {
	stack := make([]byte, maxStackBytes)
	runtime.Stack(stack, false)
	return &argumentErrorInfo{&errorInfo{msg, string(stack)}, argName, argValue}
}

// Error implementation
type errorInfo struct {
	message string
	stack   string
}

func (err *errorInfo) Error() string { return err.message }
func (err *errorInfo) Stack() string { return err.stack }

// ArgumentError implementation
type argumentErrorInfo struct {
	*errorInfo
	argName  string
	argValue interface{}
}

func (err *argumentErrorInfo) ArgName() string       { return err.argName }
func (err *argumentErrorInfo) ArgValue() interface{} { return err.argValue }
