package stream

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLongBuffer(t *testing.T) {
	l, err := NewLongBuffer(os.TempDir())
	assert.NoError(t, err)
	r := rand.New(rand.NewSource(time.Now().Unix()))
	buff := make([]byte, 3*1024*1024)
	for i := 0; i < 10; i++ {
		_, _ = r.Read(buff)
		n, err := l.Write(buff)
		assert.NoError(t, err)
		assert.Equal(t, 3*1024*1024, n)
	}
	assert.Equal(t, 30*1024*1024, l.Len())
}
