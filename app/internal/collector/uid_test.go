package collector_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/app/internal/collector"
)

func TestHardwareMACResolver_Resolve(t *testing.T) {
	t.Parallel()

	mac, err := collector.HardwareMACResolver{}.Resolve()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(mac), 17)
}
