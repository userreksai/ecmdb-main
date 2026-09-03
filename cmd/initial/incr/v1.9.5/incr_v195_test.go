package v195

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	incr := &incrV195{}
	assert.Equal(t, "v1.9.5", incr.Version())
}
