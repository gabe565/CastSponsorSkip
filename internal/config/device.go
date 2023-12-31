package config

import (
	"github.com/spf13/cobra"
)

func (c *Config) RegisterDevices(cmd *cobra.Command) {
	key := "devices"
	cmd.PersistentFlags().StringSlice(key, Default.DeviceAddrStrs, "Comma-separated list of device addresses. This will disable discovery and is not recommended unless discovery fails")
	if err := c.viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}
}
