package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gabe565/castsponsorskip/internal/config"
	"github.com/gabe565/castsponsorskip/internal/device"
	"github.com/spf13/cobra"
)

var long = `Skip sponsored YouTube segments on local Cast devices.

When run, this program will watch all Google Cast devices on the LAN.
If a Cast device begins playing a YouTube video, sponsored segments are fetched from the SponsorBlock API.
When the device reaches a sponsored segment, the CastSponsorSkip will quickly seek to the end of the segment.

Additionally, CastSponsorSkip will look for skippable YouTube ads, and automatically hit the skip button when it becomes available.`

func NewCommand(version, commit string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "castsponsorskip",
		Short:   "Skip sponsored YouTube segments on local Cast devices",
		Long:    long,
		PreRunE: preRun,
		RunE:    run,
		Version: buildVersion(version, commit),

		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	CompletionFlag(cmd)
	config.Interface(cmd)
	config.PausedInterval(cmd)
	config.PlayingInterval(cmd)
	config.Categories(cmd)
	cmd.InitDefaultVersionFlag()

	return cmd
}

func preRun(cmd *cobra.Command, args []string) error {
	config.Load()
	return nil
}

func run(cmd *cobra.Command, args []string) (err error) {
	if completionFlag != "" {
		return completion(cmd)
	}

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

func buildVersion(version, commit string) string {
	if commit != "" {
		version += " (" + commit + ")"
	}
	return version
}
