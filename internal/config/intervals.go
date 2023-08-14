package config

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	PausedIntervalKey   = "paused-interval"
	PausedIntervalValue = time.Minute
)

func PausedInterval(cmd *cobra.Command) {
	cmd.PersistentFlags().Duration(PausedIntervalKey, PausedIntervalValue, "Interval to scan paused devices")
	if err := viper.BindPFlag(PausedIntervalKey, cmd.PersistentFlags().Lookup(PausedIntervalKey)); err != nil {
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
}
