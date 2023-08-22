package config

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (c *Config) RegisterDiscoverInterval(cmd *cobra.Command) {
	key := "discover-interval"
	cmd.PersistentFlags().Duration(key, 5*time.Minute, "Interval to restart the DNS discovery client")
	if err := viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(key, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"5m", "10m", "15m"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}); err != nil {
		panic(err)
	}
}

func (c *Config) RegisterPausedInterval(cmd *cobra.Command) {
	key := "paused-interval"
	cmd.PersistentFlags().Duration(key, time.Minute, "Interval to scan paused devices")
	if err := viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
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
	cmd.PersistentFlags().Duration(key, time.Second, "Interval to scan playing devices")
	if err := viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(key, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"1s", "2s"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}); err != nil {
		panic(err)
	}
}
