package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gabe565/sponsorblockcast/internal/device"
	"github.com/spf13/cobra"
	castdns "github.com/vishen/go-chromecast/dns"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sponsorblockcast",
		Short: "Skip sponsored YouTube segments on local Cast devices",
		RunE:  run,
	}

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	slog.Info("Searching for devices...")

	entries, err := castdns.DiscoverCastDNSEntries(ctx, nil)
	if err != nil {
		slog.Error("Failed to fetch devices", "error", err)
	}

	var group sync.WaitGroup

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case entry := <-entries:
				group.Add(1)
				go func() {
					defer func() {
						group.Done()
					}()
					device.Watch(ctx, entry)
				}()
			}
		}
	}()

	<-ctx.Done()
	slog.Info("Gracefully closing connections...")
	group.Wait()
	slog.Info("Exiting")
	return nil
}
