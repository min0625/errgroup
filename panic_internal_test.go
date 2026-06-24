package errgroup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// White-box test: the public callers always guard recover() != nil, so the
// nil branch of exception is only reachable from here.
func Test_exception_Nil(t *testing.T) {
	t.Parallel()

	assert.Nil(t, exception(nil))
}
