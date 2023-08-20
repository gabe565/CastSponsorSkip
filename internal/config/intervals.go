package config

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	DiscoverIntervalKey   = "discover-interval"
	DiscoverIntervalValue = 5 * time.Minute
)

func DiscoverInterval(cmd *cobra.Command) {
	cmd.PersistentFlags().Duration(DiscoverIntervalKey, DiscoverIntervalValue, "Interval to restart the DNS discovery client")
	if err := viper.BindPFlag(DiscoverIntervalKey, cmd.PersistentFlags().Lookup(DiscoverIntervalKey)); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(DiscoverIntervalKey, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"5m", "10m", "15m"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}); err != nil {
		panic(err)
	}
}

var (
	PausedIntervalKey   = "paused-interval"
	PausedIntervalValue = time.Minute
)

func PausedInterval(cmd *cobra.Command) {
	cmd.PersistentFlags().Duration(PausedIntervalKey, PausedIntervalValue, "Interval to scan paused devices")
	if err := viper.BindPFlag(PausedIntervalKey, cmd.PersistentFlags().Lookup(PausedIntervalKey)); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(PausedIntervalKey, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"1m", "2m", "5m", "10m", "30m", "1h"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}); err != nil {
		panic(err)
	}
}

var (
	PlayingIntervalKey   = "playing-interval"
	PlayingIntervalValue = time.Second
)

func PlayingInterval(cmd *cobra.Command) {
	cmd.PersistentFlags().Duration(PlayingIntervalKey, PlayingIntervalValue, "Interval to scan playing devices")
	if err := viper.BindPFlag(PlayingIntervalKey, cmd.PersistentFlags().Lookup(PlayingIntervalKey)); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(PlayingIntervalKey, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"1s", "2s"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}); err != nil {
		panic(err)
	}
}
