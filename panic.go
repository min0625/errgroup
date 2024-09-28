package errgroup

import (
	"fmt"
	"runtime/debug"
)

// A PanicError wraps an error recovered from an unhandled panic
// when calling a function passed to Go or TryGo.
type PanicError struct {
	Recovered error
	Stack     []byte
}

func (p PanicError) Error() string {
	// A Go Error method conventionally does not include a stack dump, so omit it
	// here. (Callers who care can extract it from the Stack field.)
	return fmt.Sprintf("recovered from errgroup.Group: %v", p.Recovered)
}

func (p PanicError) Unwrap() error { return p.Recovered }

// A PanicValue wraps a value that does not implement the error interface,
// recovered from an unhandled panic when calling a function passed to Go or
// TryGo.
type PanicValue struct {
	Recovered any
	Stack     []byte
}

func (p PanicValue) String() string {
	if len(p.Stack) > 0 {
		return fmt.Sprintf("recovered from errgroup.Group: %v\n%s", p.Recovered, p.Stack)
	}

	return fmt.Sprintf("recovered from errgroup.Group: %v", p.Recovered)
}

func exception(recovered any) any {
	if recovered == nil {
		return nil
	}

	if p, ok := recovered.(PanicError); ok {
		return p
	}

	if p, ok := recovered.(PanicValue); ok {
		return p
	}

	if err, ok := recovered.(error); ok {
		return PanicError{
			Recovered: err,
			Stack:     debug.Stack(),
		}
	}

	return PanicValue{
		Recovered: recovered,
		Stack:     debug.Stack(),
	}
}
