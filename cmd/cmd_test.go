package cmd

import (
	"bytes"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/gabe565/castsponsorskip/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func randDuration() time.Duration {
	highest := int((10 * time.Minute).Seconds())
	randSecs := rand.Intn(highest) //nolint:gosec
	return time.Duration(randSecs) * time.Second
}

func getNetworkInterfaceName(t *testing.T) string {
	interfaces, err := net.Interfaces()
	require.NoError(t, err)
	return interfaces[0].Name
}

func TestFlags(t *testing.T) {
	t.Cleanup(func() {
		config.Default = config.NewDefault()
	})

	discoverInterval := randDuration()
	pausedInterval := randDuration()
	playingInterval := randDuration()
	networkInterface := getNetworkInterfaceName(t)

	var cmd *cobra.Command
	if !assert.NotPanics(t, func() {
		cmd = NewCommand("", "")
	}) {
		return
	}
	cmd.SetArgs([]string{
		"--log-level=debug",
		"--network-interface=" + networkInterface,
		"--discover-interval=" + discoverInterval.String(),
		"--paused-interval=" + pausedInterval.String(),
		"--playing-interval=" + playingInterval.String(),
		"--categories=a,b,c",
		"--action-types=d,e,f",
		"--youtube-api-key=AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe",
		"--mute-ads=false",
		"--devices=192.168.1.1,192.168.1.2",
	})
	cmd.RunE = func(_ *cobra.Command, _ []string) error { return nil }

	require.NoError(t, cmd.Execute())

	assert.Equal(t, "debug", config.Default.LogLevel)
	assert.Equal(t, networkInterface, config.Default.NetworkInterfaceName)
	assert.Equal(t, discoverInterval, config.Default.DiscoverInterval)
	assert.Equal(t, pausedInterval, config.Default.PausedInterval)
	assert.Equal(t, playingInterval, config.Default.PlayingInterval)
	assert.Equal(t, []string{"a", "b", "c"}, config.Default.Categories)
	assert.Equal(t, []string{"d", "e", "f"}, config.Default.ActionTypes)
	assert.Equal(t, "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe", config.Default.YouTubeAPIKey)
	assert.False(t, config.Default.MuteAds)
	assert.Equal(t, []string{"192.168.1.1", "192.168.1.2"}, config.Default.DeviceAddrStrs)
	assert.Len(t, config.Default.DeviceAddrs, 2)
}

func TestEnvs(t *testing.T) {
	t.Cleanup(func() {
		config.Default = config.NewDefault()
	})

	discoverInterval := randDuration()
	pausedInterval := randDuration()
	playingInterval := randDuration()
	networkInterface := getNetworkInterfaceName(t)

	t.Setenv("CSS_LOG_LEVEL", "warn")
	t.Setenv("CSS_NETWORK_INTERFACE", networkInterface)
	t.Setenv("CSS_DISCOVER_INTERVAL", discoverInterval.String())
	t.Setenv("CSS_PAUSED_INTERVAL", pausedInterval.String())
	t.Setenv("CSS_PLAYING_INTERVAL", playingInterval.String())
	t.Setenv("CSS_CATEGORIES", "a,b,c")
	t.Setenv("CSS_ACTION_TYPES", "d,e,f")
	t.Setenv("CSS_YOUTUBE_API_KEY", "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe")
	t.Setenv("CSS_MUTE_ADS", "false")
	t.Setenv("CSS_DEVICES", "192.168.1.1,192.168.1.2")

	var cmd *cobra.Command
	if !assert.NotPanics(t, func() {
		cmd = NewCommand("", "")
	}) {
		return
	}
	cmd.RunE = func(_ *cobra.Command, _ []string) error { return nil }

	require.NoError(t, cmd.Execute())

	assert.Equal(t, "warn", config.Default.LogLevel)
	assert.Equal(t, networkInterface, config.Default.NetworkInterfaceName)
	assert.Equal(t, discoverInterval, config.Default.DiscoverInterval)
	assert.Equal(t, pausedInterval, config.Default.PausedInterval)
	assert.Equal(t, playingInterval, config.Default.PlayingInterval)
	assert.Equal(t, []string{"a", "b", "c"}, config.Default.Categories)
	assert.Equal(t, []string{"d", "e", "f"}, config.Default.ActionTypes)
	assert.Equal(t, "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe", config.Default.YouTubeAPIKey)
	assert.False(t, config.Default.MuteAds)
	assert.Equal(t, []string{"192.168.1.1", "192.168.1.2"}, config.Default.DeviceAddrStrs)
	assert.Len(t, config.Default.DeviceAddrs, 2)
}

func TestSBCEnvs(t *testing.T) {
	t.Cleanup(func() {
		config.Default = config.NewDefault()
	})

	discoverInterval := randDuration()
	playingInterval := randDuration()

	t.Setenv("SBCSCANINTERVAL", strconv.Itoa(int(discoverInterval.Seconds())))
	t.Setenv("SBCPOLLINTERVAL", strconv.Itoa(int(playingInterval.Seconds())))
	t.Setenv("SBCCATEGORIES", "a b c")
	t.Setenv("SBCYOUTUBEAPIKEY", "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe")

	var cmd *cobra.Command
	if !assert.NotPanics(t, func() {
		cmd = NewCommand("", "")
	}) {
		return
	}
	cmd.RunE = func(_ *cobra.Command, _ []string) error { return nil }

	require.NoError(t, cmd.Execute())

	assert.Equal(t, discoverInterval, config.Default.DiscoverInterval)
	assert.Equal(t, playingInterval, config.Default.PlayingInterval)
	assert.Equal(t, []string{"a", "b", "c"}, config.Default.Categories)
	assert.Equal(t, "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe", config.Default.YouTubeAPIKey)
}

func TestCompletionFlag(t *testing.T) {
	tests := []struct {
		shell   string
		wantErr require.ErrorAssertionFunc
	}{
		{"bash", require.NoError},
		{"zsh", require.NoError},
		{"fish", require.NoError},
		{"powershell", require.NoError},
		{"invalid", require.Error},
	}
	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			cmd := NewCommand("", "")
			cmd.SetArgs([]string{"--completion", tt.shell})

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			tt.wantErr(t, cmd.Execute())
			assert.NotZero(t, buf.Bytes())
		})
	}
}
