package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenizh/go-capturer"
)

func Test_MainHelp(t *testing.T) {
	os.Args = []string{"", "--help"}
	exitFn = func(code int) { require.Equal(t, 0, code) }

	output := capturer.CaptureStdout(main)

	assert.Contains(t, output, "USAGE:")
	assert.Contains(t, output, "COMMANDS:")
	assert.Contains(t, output, "GLOBAL OPTIONS:")
}
