package collector

import (
	"errors"
	"net"
)

type UIDResolver interface {
	// Resolve returns a unique ID for the host.
	Resolve() (string, error)
}

type HardwareMACResolver struct{}

var _ UIDResolver = (*HardwareMACResolver)(nil) // ensure interface is implemented

// Resolve returns the MAC address of the first non-loopback interface that is up.
// If no such interface is found, an error is returned. This is used to generate a unique ID for the host.
func (HardwareMACResolver) Resolve() (string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, ifa := range ifas {
		if ifa.Flags&net.FlagLoopback != 0 {
			continue
		}

		if ifa.Flags&net.FlagUp == 0 {
			continue
		}

		if ifa.HardwareAddr == nil {
			continue
		}

		return ifa.HardwareAddr.String(), nil
	}

	return "", errors.New("no interfaces found")
}
