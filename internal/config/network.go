package config

import (
	"net"
	"strings"

	"github.com/spf13/cobra"
)

func (c *Config) RegisterNetworkInterface(cmd *cobra.Command) {
	key := "network-interface"
	cmd.PersistentFlags().StringP(key, "i", Default.NetworkInterfaceName, "Network interface to use for multicast dns discovery. (default all interfaces)")
	if err := c.viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(key, completeNetworkInterface); err != nil {
		panic(err)
	}
}

func completeNetworkInterface(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
}
