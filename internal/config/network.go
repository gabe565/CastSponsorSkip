package config

import (
	"net"
	"strings"

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
	if err := cmd.RegisterFlagCompletionFunc(InterfaceKey, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		interfaces, err := net.Interfaces()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(interfaces))
		for _, iface := range interfaces {
			addrs, _ := iface.Addrs()
			addrStr := make([]string, 0, len(addrs))
			for _, addr := range addrs {
				addrStr = append(addrStr, addr.String())
			}
			completions = append(completions, iface.Name+"\t"+strings.Join(addrStr, ", "))
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}
}
