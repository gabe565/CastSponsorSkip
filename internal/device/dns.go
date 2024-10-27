package device

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gabe565.com/castsponsorskip/internal/config"
	castdns "github.com/vishen/go-chromecast/dns"
)

var ErrDeviceNotFound = errors.New("device not found")

func DiscoverCastDNSEntryByUUID(ctx context.Context, conf *config.Config, uuid string) (castdns.CastEntry, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	entries, err := castdns.DiscoverCastDNSEntries(ctx, conf.NetworkInterface)
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

func DiscoverCastDNSEntries(ctx context.Context, conf *config.Config, ch chan castdns.CastEntry) error {
	subCtx, cancel := context.WithTimeout(ctx, conf.DiscoverInterval)
	defer cancel()

	entries, err := castdns.DiscoverCastDNSEntries(subCtx, conf.NetworkInterface)
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

func BeginDiscover(ctx context.Context, conf *config.Config) (<-chan castdns.CastEntry, error) {
	ch := make(chan castdns.CastEntry)
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		defer close(ch)
		defer cancel()

		if len(conf.DeviceAddrs) == 0 {
			if conf.NetworkInterface != nil {
				slog.Info("Searching for devices...", "interface", conf.NetworkInterfaceName)
			} else {
				slog.Info("Searching for devices...")
			}

			for {
				select {
				case <-ctx.Done():
					return
				default:
					if err := DiscoverCastDNSEntries(ctx, conf, ch); err != nil {
						slog.Error("Failed to discover devices.", "error", err.Error())
						continue
					}
				}
			}
		} else {
			if conf.NetworkInterface != nil {
				slog.Info("Connecting to configured devices...", "interface", conf.NetworkInterfaceName)
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
					for _, castEntry := range conf.DeviceAddrs {
						ch <- castEntry
					}
					timer.Reset(conf.DiscoverInterval)
				}
			}
		}
	}()

	return ch, nil
}
