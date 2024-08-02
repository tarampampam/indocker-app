package collector

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// HashParts returns a hash of the given parts. The hash is a hex-encoded string of the first 16 bytes of the SHA-256
// hash of the parts. Empty parts are ignored.
func HashParts(parts ...string) (string, error) {
	if len(parts) == 0 {
		return "", errors.New("empty parts")
	}

	var (
		h     = sha256.New()
		wrote int
	)

	for _, part := range parts {
		if w, err := h.Write([]byte(part)); err != nil {
			return "", err
		} else {
			wrote += w
		}
	}

	if wrote == 0 {
		return "", errors.New("empty parts")
	}

	var hash = hex.EncodeToString(h.Sum(nil))

	if len(hash) > 16 { //nolint:mnd
		return hash[:16], nil
	}

	return "", errors.New("too short hash") // never happens
}
