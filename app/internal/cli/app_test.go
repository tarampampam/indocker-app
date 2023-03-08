package cli_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/app/internal/cli"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	app := cli.NewApp()

	assert.NotEmpty(t, app.Flags)
}
