package config

import (
	"github.com/spf13/cobra"
)

func (c *Config) RegisterLogLevel(cmd *cobra.Command) {
	key := "log-level"
	cmd.PersistentFlags().String(key, Default.LogLevel, "Log level (debug, info, warn, error)")
	if err := c.viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(key, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"debug", "info", "warn", "error"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}); err != nil {
		panic(err)
	}
}
