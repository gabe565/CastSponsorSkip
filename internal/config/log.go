package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (c *Config) RegisterLogLevel(cmd *cobra.Command) {
	key := "log-level"
	cmd.PersistentFlags().String(key, "info", "Log level (debug, info, warn, error)")
	if err := viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(key, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"debug", "info", "warn", "error"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}); err != nil {
		panic(err)
	}
}
