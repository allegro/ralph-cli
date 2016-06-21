package main

import (
	"bytes"
	"fmt"
)

// ValidationError is the type for aggregating errors that should be presented
// to the user all at once (instead of exiting the program after the first one)
// - useful for parsing config or manifests files.
// In case that this wouldn't be enough, github.com/hashicorp/go-multierror may
// be a pretty good choice.
type ValidationError struct {
	length  int
	path    string
	errMsgs []*string
}

// NewValidationError creates ValidationError and returns it as error type. Path
// should point to the file being validated, and errMsgs should contain detailed
// error messages, but without prefixing them with strings like "config error:"
// or "manifest error:" etc.
func NewValidationError(path string, errMsgs []*string) error {
	return &ValidationError{
		length:  len(errMsgs),
		path:    path,
		errMsgs: errMsgs,
	}
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	var buf bytes.Buffer
	switch e.length {
	case 0:
		// This shouldn't really happen, but never say never...
		return "no errors"
	case 1:
		return fmt.Sprintf("validation error in %s: %s", e.path, *e.errMsgs[0])
	}
	fmt.Fprintf(&buf, "validation errors in %s (%d in total): ", e.path, len(e.errMsgs))
	for i, err := range e.errMsgs {
		fmt.Fprintf(&buf, "(%d) %s", i+1, *err)
		if i+1 < e.length {
			buf.WriteString("; ")
		}
	}
	return buf.String()
}
