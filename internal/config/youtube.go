package config

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (c *Config) RegisterYouTubeAPIKey(cmd *cobra.Command) {
	key := "youtube-api-key"
	cmd.PersistentFlags().String(key, "", "YouTube API key for fallback video identification (required on some Chromecast devices).")
	if err := viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}

	if env := os.Getenv("SBCYOUTUBEAPIKEY"); env != "" {
		viper.SetDefault(key, env)
	}
}

func (c *Config) RegisterMuteAds(cmd *cobra.Command) {
	key := "mute-ads"
	cmd.PersistentFlags().Bool(key, false, "Enables experimental support for muting unskippable ads")
	if err := viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}
}
