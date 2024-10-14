package cert_test

import (
	"context"
	"testing"

	"gh.tarampamp.am/indocker-app/app/internal/cert"
)

func TestFoo(t *testing.T) {
	var r = cert.NewResolver()

	t.Log(r.Resolve(context.Background()))
}
