package device

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/gabe565/sponsorblockcast/internal/config"
	castdns "github.com/vishen/go-chromecast/dns"
)

var ErrDeviceNotFound = errors.New("device not found")

func DiscoverCastDNSEntryByUuid(ctx context.Context, uuid string) (castdns.CastEntry, error) {
	var iface *net.Interface
	if config.InterfaceValue != "" {
		var err error
		iface, err = net.InterfaceByName(config.InterfaceValue)
		if err != nil {
			return castdns.CastEntry{}, err
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	entries, err := castdns.DiscoverCastDNSEntries(ctx, iface)
	if err != nil {
		return castdns.CastEntry{}, err
	}

	for {
		select {
		case <-ctx.Done():
			return castdns.CastEntry{}, fmt.Errorf("%w: %s", ErrDeviceNotFound, uuid)
		case entry := <-entries:
			if entry.UUID == uuid {
				return entry, nil
			}
		}
	}
}

func DiscoverCastDNSEntries(ctx context.Context) (<-chan castdns.CastEntry, error) {
	var iface *net.Interface
	if config.InterfaceValue != "" {
		var err error
		iface, err = net.InterfaceByName(config.InterfaceValue)
		if err != nil {
			return nil, err
		}
		slog.Info("Searching for devices...", "interface", config.InterfaceValue)
	} else {
		slog.Info("Searching for devices...")
	}

	entries, err := castdns.DiscoverCastDNSEntries(ctx, iface)
	if err != nil {
		return nil, err
	}

	return entries, nil
}
