package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gabe565/castsponsorskip/internal/config"
	"github.com/gabe565/castsponsorskip/internal/device"
	"github.com/gabe565/castsponsorskip/internal/youtube"
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
	config.Default.RegisterFlags(cmd)
	cmd.InitDefaultVersionFlag()

	return cmd
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := config.Default.Load(); err != nil {
		return err
	}

	if config.Default.LogLevel != "info" {
		var level slog.Level
		switch config.Default.LogLevel {
		case "debug":
			level = slog.LevelDebug
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		default:
			slog.Warn("Invalid log level. Defaulting to info.")
		}
		if level != slog.LevelInfo {
			slog.SetDefault(slog.New(slog.NewTextHandler(cmd.ErrOrStderr(), &slog.HandlerOptions{
				Level: level,
			})))
		}
	}

	return nil
}

func run(cmd *cobra.Command, args []string) (err error) {
	if completionFlag != "" {
		return completion(cmd)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	if config.Default.YouTubeAPIKey != "" {
		if err := youtube.CreateService(ctx); err != nil {
			return err
		}
	}

	entries, err := device.BeginDiscover(ctx)
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
				group.Add(1)
				go func() {
					device.Watch(ctx, entry)
					group.Done()
				}()
			}
		}
	}()

	<-ctx.Done()
	slog.Info("Gracefully closing connections... Press Ctrl+C again to force exit.")

	forceCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	go func() {
		group.Wait()
		cancel()
	}()
	forceCtx, cancel = signal.NotifyContext(forceCtx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()
	<-forceCtx.Done()
	slog.Info("Exiting.")
	return nil
}

func buildVersion(version, commit string) string {
	if commit != "" {
		version += " (" + commit + ")"
	}
	return version
}
