package cmd

import (
	"context"
	_ "embed"
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

//go:embed description.md
var long string

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
	cmd.SetVersionTemplate("CastSponsorSkip {{ .Version }}\n")

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

	slog.Info("CastSponsorSkip " + cmd.Version)

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
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
					if d := device.NewDevice(entry, device.WithContext(ctx)); d != nil {
						_ = d.BeginTick()
						_ = d.Close()
					}
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
