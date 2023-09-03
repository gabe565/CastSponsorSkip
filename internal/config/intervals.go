package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

func (c *Config) RegisterDiscoverInterval(cmd *cobra.Command) {
	key := "discover-interval"
	cmd.PersistentFlags().Duration(key, Default.DiscoverInterval, "Interval to restart the DNS discovery client")
	if err := c.viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(key, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"5m", "10m", "15m"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}); err != nil {
		panic(err)
	}

	if env := os.Getenv("SBCSCANINTERVAL"); env != "" {
		parsed, err := strconv.Atoi(env)
		if err == nil {
			val := (time.Duration(parsed) * time.Second).String()
			slog.Warn(fmt.Sprintf(`SBCSCANINTERVAL is deprecated. Please set %q instead.`, "CSS_DISCOVER_INTERVAL="+val))
			c.viper.SetDefault(key, val)
		}
	}
}

func (c *Config) RegisterPausedInterval(cmd *cobra.Command) {
	key := "paused-interval"
	cmd.PersistentFlags().Duration(key, Default.PausedInterval, "Interval to scan paused devices")
	if err := c.viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(key, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"1m", "2m", "5m", "10m", "30m", "1h"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}); err != nil {
		panic(err)
	}
}

func (c *Config) RegisterPlayingInterval(cmd *cobra.Command) {
	key := "playing-interval"
	cmd.PersistentFlags().Duration(key, Default.PlayingInterval, "Interval to scan playing devices")
	if err := c.viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(key, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"1s", "2s"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}); err != nil {
		panic(err)
	}

	if env := os.Getenv("SBCPOLLINTERVAL"); env != "" {
		parsed, err := strconv.Atoi(env)
		if err == nil {
			val := (time.Duration(parsed) * time.Second).String()
			slog.Warn(fmt.Sprintf(`SBCPOLLINTERVAL is deprecated. Please set %q instead.`, "CSS_PLAYING_INTERVAL="+val))
			c.viper.SetDefault(key, val)
		}
	}
}
