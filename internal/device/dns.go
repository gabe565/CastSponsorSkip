package device

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/gabe565/castsponsorskip/internal/config"
	castdns "github.com/vishen/go-chromecast/dns"
)

var ErrDeviceNotFound = errors.New("device not found")

func DiscoverCastDNSEntryByUuid(ctx context.Context, uuid string) (castdns.CastEntry, error) {
	var iface *net.Interface
	if config.Default.NetworkInterface != "" {
		var err error
		iface, err = net.InterfaceByName(config.Default.NetworkInterface)
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

func DiscoverCastDNSEntries(ctx context.Context, iface *net.Interface, ch chan castdns.CastEntry) error {
	subCtx, cancel := context.WithTimeout(ctx, config.Default.DiscoverInterval)
	defer cancel()

	entries, err := castdns.DiscoverCastDNSEntries(subCtx, iface)
	if err != nil {
		return err
	}

	for {
		select {
		case <-subCtx.Done():
			return nil
		case entry := <-entries:
			ch <- entry
		}
	}
}

func BeginDiscover(ctx context.Context) (<-chan castdns.CastEntry, error) {
	var iface *net.Interface
	if config.Default.NetworkInterface != "" {
		var err error
		iface, err = net.InterfaceByName(config.Default.NetworkInterface)
		if err != nil {
			return nil, err
		}
		slog.Info("Searching for devices...", "interface", config.Default.NetworkInterface)
	} else {
		slog.Info("Searching for devices...")
	}

	ch := make(chan castdns.CastEntry)
	go func() {
		defer func() {
			close(ch)
		}()

		for {
			if ctx.Err() != nil {
				return
			}

			if err := DiscoverCastDNSEntries(ctx, iface, ch); err != nil {
				slog.Error("Failed to discover devices.", "error", err.Error())
				continue
			}
		}
	}()

	return ch, nil
}
