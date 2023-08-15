package cmd

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gabe565/sponsorblockcast/internal/config"
	"github.com/gabe565/sponsorblockcast/internal/device"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	castdns "github.com/vishen/go-chromecast/dns"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sponsorblockcast",
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

	var iface *net.Interface
	interfaceName := viper.GetString(config.InterfaceKey)
	if interfaceName != "" {
		iface, err = net.InterfaceByName(interfaceName)
		if err != nil {
			return err
		}
		slog.Info("Searching for devices...", "interface", interfaceName)
	} else {
		slog.Info("Searching for devices...")
	}

	entries, err := castdns.DiscoverCastDNSEntries(ctx, iface)
	if err != nil {
		slog.Error("Failed to fetch devices", "error", err.Error())
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
