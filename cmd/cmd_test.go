package cmd

import (
	"bytes"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gabe565/castsponsorskip/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func randDuration() time.Duration {
	highest := int((10 * time.Minute).Seconds())
	randSecs := rand.Intn(highest)
	return time.Duration(randSecs) * time.Second
}

func TestFlags(t *testing.T) {
	defer func() {
		config.Reset()
	}()

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
		"--action-types=d,e,f",
		"--youtube-api-key=AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe",
		"--mute-ads=false",
	})
	cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }

	if err := cmd.Execute(); !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "debug", config.Default.LogLevel)
	assert.Equal(t, "eno1", config.Default.NetworkInterface)
	assert.Equal(t, discoverInterval, config.Default.DiscoverInterval)
	assert.Equal(t, pausedInterval, config.Default.PausedInterval)
	assert.Equal(t, playingInterval, config.Default.PlayingInterval)
	assert.Equal(t, []string{"a", "b", "c"}, config.Default.Categories)
	assert.Equal(t, []string{"d", "e", "f"}, config.Default.ActionTypes)
	assert.Equal(t, "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe", config.Default.YouTubeAPIKey)
	assert.Equal(t, false, config.Default.MuteAds)
}

func TestEnvs(t *testing.T) {
	defer func() {
		config.Reset()
	}()

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
		_ = os.Unsetenv("CSS_MUTE_ADS")
	}()
	_ = os.Setenv("CSS_LOG_LEVEL", "warn")
	_ = os.Setenv("CSS_NETWORK_INTERFACE", "eno1")
	_ = os.Setenv("CSS_DISCOVER_INTERVAL", discoverInterval.String())
	_ = os.Setenv("CSS_PAUSED_INTERVAL", pausedInterval.String())
	_ = os.Setenv("CSS_PLAYING_INTERVAL", playingInterval.String())
	_ = os.Setenv("CSS_CATEGORIES", "a,b,c")
	_ = os.Setenv("CSS_ACTION_TYPES", "d,e,f")
	_ = os.Setenv("CSS_YOUTUBE_API_KEY", "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe")
	_ = os.Setenv("CSS_MUTE_ADS", "false")

	var cmd *cobra.Command
	if !assert.NotPanics(t, func() {
		cmd = NewCommand("", "")
	}) {
		return
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }

	if err := cmd.Execute(); !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "warn", config.Default.LogLevel)
	assert.Equal(t, "eno1", config.Default.NetworkInterface)
	assert.Equal(t, discoverInterval, config.Default.DiscoverInterval)
	assert.Equal(t, pausedInterval, config.Default.PausedInterval)
	assert.Equal(t, playingInterval, config.Default.PlayingInterval)
	assert.Equal(t, []string{"a", "b", "c"}, config.Default.Categories)
	assert.Equal(t, []string{"d", "e", "f"}, config.Default.ActionTypes)
	assert.Equal(t, "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe", config.Default.YouTubeAPIKey)
	assert.Equal(t, false, config.Default.MuteAds)
}

func TestSBCEnvs(t *testing.T) {
	defer func() {
		config.Reset()
	}()

	discoverInterval := randDuration()
	playingInterval := randDuration()

	defer func() {
		_ = os.Unsetenv("SBCSCANINTERVAL")
		_ = os.Unsetenv("SBCPOLLINTERVAL")
		_ = os.Unsetenv("SBCCATEGORIES")
		_ = os.Unsetenv("SBCYOUTUBEAPIKEY")
	}()
	_ = os.Setenv("SBCSCANINTERVAL", strconv.Itoa(int(discoverInterval.Seconds())))
	_ = os.Setenv("SBCPOLLINTERVAL", strconv.Itoa(int(playingInterval.Seconds())))
	_ = os.Setenv("SBCCATEGORIES", "a b c")
	_ = os.Setenv("SBCYOUTUBEAPIKEY", "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe")

	var cmd *cobra.Command
	if !assert.NotPanics(t, func() {
		cmd = NewCommand("", "")
	}) {
		return
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }

	if err := cmd.Execute(); !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, discoverInterval, config.Default.DiscoverInterval)
	assert.Equal(t, playingInterval, config.Default.PlayingInterval)
	assert.Equal(t, []string{"a", "b", "c"}, config.Default.Categories)
	assert.Equal(t, "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe", config.Default.YouTubeAPIKey)
}

func TestCompletionFlag(t *testing.T) {
	tests := []struct {
		shell        string
		errAssertion assert.ErrorAssertionFunc
	}{
		{"bash", assert.NoError},
		{"zsh", assert.NoError},
		{"fish", assert.NoError},
		{"powershell", assert.NoError},
		{"invalid", assert.Error},
	}
	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			cmd := NewCommand("", "")
			cmd.SetArgs([]string{"--completion", tt.shell})

			var buf bytes.Buffer
			cmd.SetOut(&buf)

			if err := cmd.Execute(); !tt.errAssertion(t, err) {
				return
			}

			assert.NotZero(t, buf.Bytes())
		})
	}
}
