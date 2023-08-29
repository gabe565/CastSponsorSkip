package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Default = &Config{}

type Config struct {
	LogLevel string `mapstructure:"log-level"`

	DiscoverInterval time.Duration `mapstructure:"discover-interval"`
	PausedInterval   time.Duration `mapstructure:"paused-interval"`
	PlayingInterval  time.Duration `mapstructure:"playing-interval"`

	NetworkInterface string `mapstructure:"network-interface"`

	Categories  []string
	ActionTypes []string `mapstructure:"action-types"`

	YouTubeAPIKey string `mapstructure:"youtube-api-key"`
	MuteAds       bool   `mapstructure:"mute-ads"`
}

func (c *Config) RegisterFlags(cmd *cobra.Command) {
	c.RegisterLogLevel(cmd)
	c.RegisterNetworkInterface(cmd)
	c.RegisterDiscoverInterval(cmd)
	c.RegisterPausedInterval(cmd)
	c.RegisterPlayingInterval(cmd)
	c.RegisterCategories(cmd)
	c.RegisterActionTypes(cmd)
	c.RegisterYouTubeAPIKey(cmd)
	c.RegisterMuteAds(cmd)
}

func (c *Config) Load() error {
	viper.SetConfigName("castsponsorskip")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/")
	viper.AddConfigPath("$HOME/")
	viper.AddConfigPath("/etc/castsponsorskip/")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("CSS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
		} else {
			// Config file was found but another error was produced
			panic(fmt.Errorf("Fatal error reading config file: %w \n", err))
		}
	}

	return viper.Unmarshal(c)
}
