package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func (c *Config) RegisterYouTubeAPIKey(cmd *cobra.Command) {
	key := "youtube-api-key"
	cmd.PersistentFlags().String(key, Default.YouTubeAPIKey, "YouTube API key for fallback video identification (required on some Chromecast devices).")
	if err := c.viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}

	if env := os.Getenv("SBCYOUTUBEAPIKEY"); env != "" {
		slog.Warn(fmt.Sprintf(`SBCYOUTUBEAPIKEY is deprecated. Please set %q instead.`, "CSS_YOUTUBE_API_KEY="+env))
		c.viper.SetDefault(key, env)
	}
}

func (c *Config) RegisterMuteAds(cmd *cobra.Command) {
	key := "mute-ads"
	cmd.PersistentFlags().Bool(key, Default.MuteAds, "Mutes the device while an ad is playing")
	if err := c.viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}
}
