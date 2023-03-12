package httptools_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/app/internal/httptools"
)

func TestTrimHostPortSuffix(t *testing.T) {
	for _, testCase := range []struct {
		giveHost     string
		giveSuffixes []string
		want         string
	}{
		{"example.com", nil, "example.com"},
		{"example.com:8080", []string{".com"}, "example"},
		{"example.com", []string{".CoM"}, "example"},
		{"foo.indocker.app", nil, "foo"},
		{"[host]:port", nil, "host"},
		{"[host]:321", nil, "host"},
		{"host:321", nil, "host"},
		{"host", nil, "host"},
		{"foo.indocker.app:1234", nil, "foo"},
		{"aAa.ExAmPlE.CoM", []string{".example.com"}, "aAa"},
		{"aAa.ExAmPlE.CoM", []string{".eXaMpLe.cOm"}, "aAa"},
		{".ExAmPlE.CoM", []string{".eXaMpLe.cOm"}, ""},
		{".indocker.app:3322", nil, ""},
		{"..indocker.app:3322", nil, "."},
	} {
		tt := testCase

		t.Run(fmt.Sprintf("%s -> %s", tt.giveHost, tt.want), func(t *testing.T) {
			assert.Equal(t, tt.want, httptools.TrimHostPortSuffix(tt.giveHost, tt.giveSuffixes...))
		})
	}
}
