package collector_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/daemon/internal/collector"
)

func TestHashParts(t *testing.T) {
	t.Parallel()

	id, err := collector.HashParts("foo", "bar", "baz")
	assert.NoError(t, err)
	assert.Equal(t, "97df3588b5a3f24b", id)

	id, err = collector.HashParts("foo", "bar")
	assert.NoError(t, err)
	assert.Equal(t, "c3ab8ff13720e8ad", id)

	id, err = collector.HashParts("", "foo", "", "bar", "")
	assert.NoError(t, err)
	assert.Equal(t, "c3ab8ff13720e8ad", id)

	id, err = collector.HashParts("foo", "bar")
	assert.NoError(t, err)
	assert.Equal(t, "c3ab8ff13720e8ad", id) // same as above

	id, err = collector.HashParts("", "", "", "", "")
	assert.ErrorContains(t, err, "empty parts")
	assert.Empty(t, id)
}
