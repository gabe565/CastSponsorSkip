package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Default Config

func init() {
	Reset()
}

func Reset() {
	Default = Config{
		LogLevel: "info",

		DiscoverInterval: 5 * time.Minute,
		PausedInterval:   time.Minute,
		PlayingInterval:  500 * time.Millisecond,

		NetworkInterface: "",

		Categories:  []string{"sponsor"},
		ActionTypes: []string{"skip", "mute"},

		YouTubeAPIKey: "",
		MuteAds:       true,
	}
}

type Config struct {
	viper *viper.Viper `mapstructure:"-"`

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
	c.viper = viper.New()
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
	c.viper.SetConfigName("castsponsorskip")
	c.viper.SetConfigType("yaml")
	c.viper.AddConfigPath("$HOME/.config/")
	c.viper.AddConfigPath("$HOME/")
	c.viper.AddConfigPath("/etc/castsponsorskip/")

	c.viper.AutomaticEnv()
	c.viper.SetEnvPrefix("CSS")
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := c.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
		} else {
			// Config file was found but another error was produced
			return fmt.Errorf("fatal error reading config file: %w", err)
		}
	}

	return c.viper.Unmarshal(c)
}
