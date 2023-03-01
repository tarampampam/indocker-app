package certs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/daemon/certs"
)

func TestFullChain(t *testing.T) {
	assert.NotEmpty(t, certs.FullChain())
}

func TestPrivateKey(t *testing.T) {
	assert.NotEmpty(t, certs.PrivateKey())
}
