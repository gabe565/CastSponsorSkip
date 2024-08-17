package config

import (
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/gabe565/castsponsorskip/internal/config/names"
	"github.com/spf13/cobra"
	castdns "github.com/vishen/go-chromecast/dns"
)

type Config struct {
	LogLevel  string `yaml:"log-level"`
	LogFormat string `yaml:"log-format"`

	DeviceAddrStrs        []string            `yaml:"devices"`
	DeviceAddrs           []castdns.CastEntry `yaml:"-"`
	DiscoverInterval      time.Duration       `yaml:"discover-interval"`
	PausedInterval        time.Duration       `yaml:"paused-interval"`
	PlayingInterval       time.Duration       `yaml:"playing-interval"`
	SkipDelay             time.Duration       `yaml:"skip-delay"`
	IgnoreSegmentDuration time.Duration       `yaml:"ignore-segment-duration"`

	NetworkInterfaceName string         `yaml:"network-interface"`
	NetworkInterface     *net.Interface `yaml:"-"`

	SkipSponsors bool     `yaml:"skip-sponsors"`
	Categories   []string `yaml:"categories"`
	ActionTypes  []string `yaml:"action-types"`

	YouTubeAPIKey string `yaml:"youtube-api-key"`
	MuteAds       bool   `yaml:"mute-ads"`
}

func New() *Config {
	return &Config{
		LogLevel:  strings.ToLower(slog.LevelInfo.String()),
		LogFormat: FormatAuto.String(),

		DiscoverInterval:      5 * time.Minute,
		PausedInterval:        time.Minute,
		PlayingInterval:       500 * time.Millisecond,
		IgnoreSegmentDuration: time.Minute,

		SkipSponsors: true,
		Categories:   []string{"sponsor"},
		ActionTypes:  []string{"skip", "mute"},

		MuteAds: true,
	}
}

func RegisterFlags(cmd *cobra.Command) {
	fs := cmd.Flags()
	c := New()

	fs.String(names.FlagConfig, "", "Config file path")
	fs.String(names.FlagLogLevel, c.LogLevel, "Log level (one of: debug, info, warn, error)")
	fs.String(names.FlagLogFormat, c.LogFormat, "Log format (one of: "+strings.Join(LogFormatStrings(), ", ")+")")

	fs.StringSlice(names.FlagDevices, c.DeviceAddrStrs, "Comma-separated list of device addresses. This will disable discovery and is not recommended unless discovery fails")
	fs.Duration(names.FlagDiscoverInterval, c.DiscoverInterval, "Interval to restart the DNS discovery client")
	fs.Duration(names.FlagPausedInterval, c.PausedInterval, "Interval to scan paused devices")
	fs.Duration(names.FlagPlayingInterval, c.PlayingInterval, "Interval to scan playing devices")
	fs.Duration(names.FlagSkipDelay, c.SkipDelay, "Delay skipping the start of a segment")
	fs.Duration(names.FlagIgnoreSegmentDuration, c.IgnoreSegmentDuration, "Ignores the previous sponsored segment for a set amount of time. Useful if you want to to go back and watch a segment.")

	fs.StringP(names.FlagNetworkInterface, "i", c.NetworkInterfaceName, "Network interface to use for multicast dns discovery. (default all interfaces)")

	fs.Bool(names.FlagSkipSponsors, c.SkipSponsors, "Skip sponsored segments with SponsorBlock")
	fs.StringSliceP(names.FlagCategories, "c", c.Categories, "Comma-separated list of SponsorBlock categories to skip")
	fs.StringSlice(names.FlagActionTypes, c.ActionTypes, "SponsorBlock action types to handle. Shorter segments that overlap with content can be muted instead of skipped.")

	fs.String(names.FlagYouTubeAPIKey, c.YouTubeAPIKey, "YouTube API key for fallback video identification (required on some Chromecast devices).")
	fs.Bool(names.FlagMuteAds, c.MuteAds, "Mutes the device while an ad is playing")
}
