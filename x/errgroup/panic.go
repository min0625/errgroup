// Package errgroup provides a more robust error group implementation
// that extends golang.org/x/sync/errgroup with panic recovery.
package errgroup

import (
	"github.com/min0625/errgroup"
)

// A PanicError wraps an error recovered from an unhandled panic
// when calling a function passed to Go or TryGo.
type PanicError = errgroup.PanicError

// A PanicValue wraps a value that does not implement the error interface,
// recovered from an unhandled panic when calling a function passed to Go or
// TryGo.
type PanicValue = errgroup.PanicValue
