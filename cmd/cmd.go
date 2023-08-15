package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gabe565/sponsorblockcast-go/internal/config"
	"github.com/gabe565/sponsorblockcast-go/internal/device"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sponsorblockcast-go",
		Short: "Skip sponsored YouTube segments on local Cast devices",
		RunE:  run,
	}

	config.Interface(cmd)
	config.PausedInterval(cmd)
	config.PlayingInterval(cmd)
	config.Categories(cmd)
	config.InitViper()

	return cmd
}

func run(cmd *cobra.Command, args []string) (err error) {
	config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	entries, err := device.DiscoverCastDNSEntries(ctx)
	if err != nil {
		return err
	}

	var group sync.WaitGroup

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case entry := <-entries:
				if entry.Device == "Google Cast Group" {
					continue
				} else if entry.Device == "" && entry.DeviceName == "" && entry.UUID == "" {
					continue
				}

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
