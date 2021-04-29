package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArguments(t *testing.T) {
	a, err := NewArguments("pim", "$1", "poum")
	assert.NoError(t, err)
	v, err := a.Values("pam")
	assert.NoError(t, err)
	assert.Equal(t, []string{"pim", "pam", "poum"}, v)
}
