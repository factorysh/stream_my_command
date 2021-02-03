package rfc7233

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRange(t *testing.T) {
	r, err := Parse("bytes=500-600,601-999,1001-")
	assert.NoError(t, err)
	fmt.Println(r)
	assert.Equal(t, 3, r.Len())
}

func TestNext(t *testing.T) {
	r, err := Parse("bytes=1-3,5-")
	assert.NoError(t, err)
	values := make([]int, 10)
	for i := 0; i < len(values); i++ {
		values[i] = i
	}
	stack := make([]int, 0)
	for true {
		start, end, infinite, err := r.Next()
		if err != nil {
			if err != io.EOF {
				assert.NoError(t, err)
			}
			break
		}
		var sub []int
		if infinite {
			sub = values[start:]
		} else {
			sub = values[start : end+1]
		}
		for _, v := range sub {
			stack = append(stack, v)
		}
	}
	assert.Equal(t, []int{1, 2, 3, 5, 6, 7, 8, 9}, stack)

}
