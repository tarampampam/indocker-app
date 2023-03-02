package healthcheck_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"gh.tarampamp.am/indocker-app/daemon/internal/cli/start/healthcheck"
	"gh.tarampamp.am/indocker-app/daemon/internal/env"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := healthcheck.NewCommand()

	assert.Equal(t, "healthcheck", cmd.Name)
	assert.GreaterOrEqual(t, 2, len(cmd.Flags))

	assert.Equal(t, env.HTTPPort.String(), cmd.Flags[0].(*cli.UintFlag).EnvVars[0])
	assert.Equal(t, env.HTTPSPort.String(), cmd.Flags[1].(*cli.UintFlag).EnvVars[0])
}
