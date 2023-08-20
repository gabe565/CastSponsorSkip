package config

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	YouTubeAPIKeyKey   = "youtube-api-key"
	YouTubeAPIKeyValue string
)

func YouTubeAPIKey(cmd *cobra.Command) {
	cmd.PersistentFlags().String(YouTubeAPIKeyKey, YouTubeAPIKeyValue, "YouTube API key for fallback video identification (required on some Chromecast devices).")
	if err := viper.BindPFlag(YouTubeAPIKeyKey, cmd.PersistentFlags().Lookup(YouTubeAPIKeyKey)); err != nil {
		panic(err)
	}

	if env := os.Getenv("SBCYOUTUBEAPIKEY"); env != "" {
		viper.SetDefault(YouTubeAPIKeyKey, env)
	}
}
