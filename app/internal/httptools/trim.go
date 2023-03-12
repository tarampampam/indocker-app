package httptools

import (
	"net"
	"strings"
)

// TrimHostPortSuffix trims the given suffixes from the host. If no suffixes are given, ".indocker.app" is used.
// The port will be removed as well.
//
// Examples:
//
//	TrimHostPortSuffix("foo") -> "foo"
//	TrimHostPortSuffix("foo.indocker.app") -> "foo"
//	TrimHostPortSuffix("foo.indocker.app:8080") -> "foo"
func TrimHostPortSuffix(host string, suffixes ...string) string {
	if len(suffixes) == 0 {
		suffixes = []string{".indocker.app"}
	}

	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	for _, suffix := range suffixes {
		if strings.HasSuffix(strings.ToLower(host), strings.ToLower(suffix)) {
			host = host[:len(host)-len(suffix)]

			break
		}
	}

	return host
}
