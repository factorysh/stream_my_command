package stream

import (
	"bytes"
	"os"
	"testing"

	"io/ioutil"

	"github.com/stretchr/testify/assert"
)

func TestBucket(t *testing.T) {
	home, err := ioutil.TempDir(os.TempDir(), "buck_")
	assert.NoError(t, err)
	bucketSize := 6
	b, err := NewBucket(home, bucketSize)
	assert.NoError(t, err)
	txt := []byte("Je mange des carottes")
	n, err := b.Write(txt)
	assert.Equal(t, len(txt), n)
	assert.NoError(t, err)
	nBuckets := 4
	assert.Equal(t, nBuckets, b.n)
	assert.Equal(t, b.Cache(), txt[(nBuckets-1)*bucketSize:])
	err = b.Close()
	assert.NoError(t, err)
	buff := bytes.NewBuffer(nil)
	err = b.Copy(0, buff)
	assert.NoError(t, err)
	assert.Equal(t, txt, buff.Bytes())
}
