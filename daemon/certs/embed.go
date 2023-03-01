package certs

import _ "embed"

//go:embed fullchain.pem
var fullchain []byte

//go:embed privkey.pem
var privkey []byte

// FullChain returns the fullchain.pem file content.
func FullChain() []byte { return fullchain }

// PrivateKey returns the privkey.pem file content.
func PrivateKey() []byte { return privkey }
