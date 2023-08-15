package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	InterfaceKey   = "interface"
	InterfaceValue string
)

func Interface(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(InterfaceKey, "i", InterfaceValue, "Network interface to use for multicast dns discovery")
	if err := viper.BindPFlag(InterfaceKey, cmd.PersistentFlags().Lookup(InterfaceKey)); err != nil {
		panic(err)
	}
}
