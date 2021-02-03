package rfc7233

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRange(t *testing.T) {
	r, err := Parse("bytes=500-600,601-999,1001-")
	assert.NoError(t, err)
	fmt.Println(r)
	assert.Len(t, r, 3)
}
