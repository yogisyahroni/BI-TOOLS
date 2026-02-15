package services

import (
	"fmt"
	"runtime/debug"
)

// ErrorTracker handles centralized error reporting and stack trace capture
type ErrorTracker struct{}

var DefaultErrorTracker = &ErrorTracker{}

// TrackError logs an error with stack trace and metadata
func (et *ErrorTracker) TrackError(context string, err error, metadata map[string]interface{}) {
	if err == nil {
		return
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	// Capture stack trace
	stack := string(debug.Stack())
	metadata["stack_trace"] = stack
	metadata["error"] = err.Error()

	LogError(context, "Error tracked", metadata)
}

// TrackPanic handles panic recovery logging
func (et *ErrorTracker) TrackPanic(context string, r interface{}, metadata map[string]interface{}) {
	err, ok := r.(error)
	if !ok {
		err = fmt.Errorf("%v", r)
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["panic"] = true

	et.TrackError(context, err, metadata)
}

// Global helper
func TrackError(context string, err error, metadata map[string]interface{}) {
	DefaultErrorTracker.TrackError(context, err, metadata)
}
