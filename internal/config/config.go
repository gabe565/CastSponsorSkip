package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	castdns "github.com/vishen/go-chromecast/dns"
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
		SkipDelay:        0,

		NetworkInterface: nil,

		SkipSponsors: true,
		Categories:   []string{"sponsor"},
		ActionTypes:  []string{"skip", "mute"},

		YouTubeAPIKey: "",
		MuteAds:       true,
	}
}

type Config struct {
	viper *viper.Viper `mapstructure:"-"`

	LogLevel string `mapstructure:"log-level"`

	DeviceAddrStrs   []string            `mapstructure:"devices"`
	DeviceAddrs      []castdns.CastEntry `mapstructure:"-"`
	DiscoverInterval time.Duration       `mapstructure:"discover-interval"`
	PausedInterval   time.Duration       `mapstructure:"paused-interval"`
	PlayingInterval  time.Duration       `mapstructure:"playing-interval"`
	SkipDelay        time.Duration       `mapstructure:"skip-delay"`

	NetworkInterfaceName string `mapstructure:"network-interface"`
	NetworkInterface     *net.Interface

	SkipSponsors bool `mapstructure:"skip-sponsors"`
	Categories   []string
	ActionTypes  []string `mapstructure:"action-types"`

	YouTubeAPIKey string `mapstructure:"youtube-api-key"`
	MuteAds       bool   `mapstructure:"mute-ads"`
}

func (c *Config) RegisterFlags(cmd *cobra.Command) {
	c.viper = viper.New()
	c.RegisterDevices(cmd)
	c.RegisterLogLevel(cmd)
	c.RegisterNetworkInterface(cmd)
	c.RegisterDiscoverInterval(cmd)
	c.RegisterPausedInterval(cmd)
	c.RegisterPlayingInterval(cmd)
	c.RegisterSkipDelay(cmd)
	c.RegisterSkipSponsors(cmd)
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

	if err := c.viper.Unmarshal(c); err != nil {
		return err
	}

	if c.NetworkInterfaceName != "" {
		var err error
		if c.NetworkInterface, err = net.InterfaceByName(c.NetworkInterfaceName); err != nil {
			return err
		}
	}

	for i, category := range c.Categories {
		c.Categories[i] = strings.TrimSpace(category)
	}

	for i, actionType := range c.ActionTypes {
		c.ActionTypes[i] = strings.TrimSpace(actionType)
	}

	if len(c.DeviceAddrStrs) != 0 {
		c.DeviceAddrs = make([]castdns.CastEntry, 0, len(c.DeviceAddrStrs))
		for _, device := range c.DeviceAddrStrs {
			u := url.URL{Host: device}

			castEntry := castdns.CastEntry{
				DeviceName: device,
				UUID:       device,
			}

			if port := u.Port(); port == "" {
				castEntry.Port = 8009
			} else {
				port, err := strconv.ParseUint(port, 10, 16)
				if err != nil {
					return err
				}

				castEntry.Port = int(port)
			}

			if ip := net.ParseIP(u.Hostname()); ip == nil {
				return fmt.Errorf("failed to parse IP %q", device)
			} else if ip.To4() != nil {
				castEntry.AddrV4 = ip
			} else {
				castEntry.AddrV6 = ip
			}

			c.DeviceAddrs = append(c.DeviceAddrs, castEntry)
		}
	}

	return nil
}
