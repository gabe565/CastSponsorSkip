package device

import (
	"errors"
	"strconv"

	"github.com/vishen/go-chromecast/dns"
)

type Capability uint8

func (c *Capability) UnmarshalString(s string) error {
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	*c = Capability(i & 31)
	return nil
}

const (
	VideoOut Capability = 1 << iota
	VideoIn
	AudioOut
	AudioIn
	DevMode
)

var ErrMissingCapabilities = errors.New("capabilities not found")

func HasVideoOut(entry dns.CastEntry) (bool, error) {
	var capability Capability
	capStr, ok := entry.InfoFields["ca"]
	if !ok {
		return false, ErrMissingCapabilities
	}

	if err := capability.UnmarshalString(capStr); err != nil {
		return false, err
	}

	return capability&VideoOut != 0, nil
}
