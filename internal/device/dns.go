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
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	entries, err := castdns.DiscoverCastDNSEntries(ctx, config.Default.NetworkInterface)
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
	ch := make(chan castdns.CastEntry)
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		defer close(ch)
		defer cancel()

		if len(config.Default.DeviceAddrs) == 0 {
			if config.Default.NetworkInterface != nil {
				slog.Info("Searching for devices...", "interface", config.Default.NetworkInterfaceName)
			} else {
				slog.Info("Searching for devices...")
			}

			for {
				select {
				case <-ctx.Done():
					return
				default:
					if err := DiscoverCastDNSEntries(ctx, config.Default.NetworkInterface, ch); err != nil {
						slog.Error("Failed to discover devices.", "error", err.Error())
						continue
					}
				}
			}
		} else {
			if config.Default.NetworkInterface != nil {
				slog.Info("Connecting to configured devices...", "interface", config.Default.NetworkInterfaceName)
			} else {
				slog.Info("Connecting to configured devices...")
			}

			timer := time.NewTimer(0)
			defer timer.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					for _, castEntry := range config.Default.DeviceAddrs {
						ch <- castEntry
					}
					timer.Reset(config.Default.DiscoverInterval)
				}
			}
		}
	}()

	return ch, nil
}
