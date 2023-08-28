package cmd

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/gabe565/castsponsorskip/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func randDuration() time.Duration {
	highest := int((10 * time.Minute).Seconds())
	randSecs := rand.Intn(highest)
	return time.Duration(randSecs) * time.Second
}

func TestFlags(t *testing.T) {
	viper.Reset()

	discoverInterval := randDuration()
	pausedInterval := randDuration()
	playingInterval := randDuration()

	var cmd *cobra.Command
	if !assert.NotPanics(t, func() {
		cmd = NewCommand("", "")
	}) {
		return
	}
	cmd.SetArgs([]string{
		"--log-level=debug",
		"--network-interface=eno1",
		"--discover-interval=" + discoverInterval.String(),
		"--paused-interval=" + pausedInterval.String(),
		"--playing-interval=" + playingInterval.String(),
		"--categories=a,b,c",
		"--youtube-api-key=AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe",
	})
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}

	if err := cmd.Execute(); !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "debug", config.Default.LogLevel)
	assert.Equal(t, "eno1", config.Default.NetworkInterface)
	assert.Equal(t, discoverInterval, config.Default.DiscoverInterval)
	assert.Equal(t, pausedInterval, config.Default.PausedInterval)
	assert.Equal(t, playingInterval, config.Default.PlayingInterval)
	assert.Equal(t, []string{"a", "b", "c"}, config.Default.Categories)
	assert.Equal(t, "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe", config.Default.YouTubeAPIKey)
}

func TestEnvs(t *testing.T) {
	viper.Reset()

	discoverInterval := randDuration()
	pausedInterval := randDuration()
	playingInterval := randDuration()

	defer func() {
		_ = os.Unsetenv("CSS_LOG_LEVEL")
		_ = os.Unsetenv("CSS_NETWORK_INTERFACE")
		_ = os.Unsetenv("CSS_DISCOVER_INTERVAL")
		_ = os.Unsetenv("CSS_PAUSED_INTERVAL")
		_ = os.Unsetenv("CSS_PLAYING_INTERVAL")
		_ = os.Unsetenv("CSS_CATEGORIES")
		_ = os.Unsetenv("CSS_YOUTUBE_API_KEY")
	}()
	_ = os.Setenv("CSS_LOG_LEVEL", "warn")
	_ = os.Setenv("CSS_NETWORK_INTERFACE", "eno1")
	_ = os.Setenv("CSS_DISCOVER_INTERVAL", discoverInterval.String())
	_ = os.Setenv("CSS_PAUSED_INTERVAL", pausedInterval.String())
	_ = os.Setenv("CSS_PLAYING_INTERVAL", playingInterval.String())
	_ = os.Setenv("CSS_CATEGORIES", "a,b,c")
	_ = os.Setenv("CSS_YOUTUBE_API_KEY", "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe")

	var cmd *cobra.Command
	if !assert.NotPanics(t, func() {
		cmd = NewCommand("", "")
	}) {
		return
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}

	if err := cmd.Execute(); !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "warn", config.Default.LogLevel)
	assert.Equal(t, "eno1", config.Default.NetworkInterface)
	assert.Equal(t, discoverInterval, config.Default.DiscoverInterval)
	assert.Equal(t, pausedInterval, config.Default.PausedInterval)
	assert.Equal(t, playingInterval, config.Default.PlayingInterval)
	assert.Equal(t, []string{"a", "b", "c"}, config.Default.Categories)
	assert.Equal(t, "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe", config.Default.YouTubeAPIKey)
}
