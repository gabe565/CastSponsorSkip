package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func InitViper() func() {
	return func() {
		viper.SetConfigName("sponsorblockcast-go")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("$HOME/.config/")
		viper.AddConfigPath("$HOME/")
		viper.AddConfigPath("/etc/sponsorblockcast-go/")

		viper.AutomaticEnv()
		viper.SetEnvPrefix("SBC")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config file not found; ignore error
			} else {
				// Config file was found but another error was produced
				panic(fmt.Errorf("Fatal error reading config file: %w \n", err))
			}
		}
	}
}
