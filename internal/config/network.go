package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var InterfaceKey = "interface"

func Interface(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(InterfaceKey, "i", "", "Network interface to use for multicast dns discovery")
	if err := viper.BindPFlag(InterfaceKey, cmd.PersistentFlags().Lookup(InterfaceKey)); err != nil {
		panic(err)
	}
}
