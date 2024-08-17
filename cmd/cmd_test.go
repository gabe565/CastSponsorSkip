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
	discoverInterval := randDuration()
	pausedInterval := randDuration()
	playingInterval := randDuration()
	skipDelay := randDuration()
	ignoreSegmentDuration := randDuration()
	networkInterface := getNetworkInterfaceName(t)

	cmd := NewCommand("", "")
	cmd.SetArgs([]string{
		"--log-level=debug",
		"--devices=192.168.1.1,192.168.1.2",
		"--discover-interval=" + discoverInterval.String(),
		"--paused-interval=" + pausedInterval.String(),
		"--playing-interval=" + playingInterval.String(),
		"--skip-delay=" + skipDelay.String(),
		"--ignore-segment-duration=" + ignoreSegmentDuration.String(),
		"--network-interface=" + networkInterface,
		"--categories=a,b,c",
		"--action-types=d,e,f",
		"--youtube-api-key=AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe",
		"--mute-ads=false",
	})
	cmd.RunE = func(_ *cobra.Command, _ []string) error { return nil }

	require.NoError(t, cmd.Execute())

	conf := config.FromContext(cmd.Context())
	assert.Equal(t, "debug", conf.LogLevel)
	assert.Equal(t, []string{"192.168.1.1", "192.168.1.2"}, conf.DeviceAddrStrs)
	assert.Equal(t, discoverInterval, conf.DiscoverInterval)
	assert.Equal(t, pausedInterval, conf.PausedInterval)
	assert.Equal(t, playingInterval, conf.PlayingInterval)
	assert.Equal(t, skipDelay, conf.SkipDelay)
	assert.Equal(t, ignoreSegmentDuration, conf.IgnoreSegmentDuration)
	assert.Equal(t, networkInterface, conf.NetworkInterfaceName)
	assert.Equal(t, []string{"a", "b", "c"}, conf.Categories)
	assert.Equal(t, []string{"d", "e", "f"}, conf.ActionTypes)
	assert.Equal(t, "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe", conf.YouTubeAPIKey)
	assert.False(t, conf.MuteAds)
	assert.Len(t, conf.DeviceAddrs, 2)
	assert.NotNil(t, conf.NetworkInterface)
}

func TestEnvs(t *testing.T) {
	discoverInterval := randDuration()
	pausedInterval := randDuration()
	playingInterval := randDuration()
	skipDelay := randDuration()
	ignoreSegmentDuration := randDuration()
	networkInterface := getNetworkInterfaceName(t)

	t.Setenv("CSS_LOG_LEVEL", "warn")
	t.Setenv("CSS_DEVICES", "192.168.1.1,192.168.1.2")
	t.Setenv("CSS_DISCOVER_INTERVAL", discoverInterval.String())
	t.Setenv("CSS_PAUSED_INTERVAL", pausedInterval.String())
	t.Setenv("CSS_PLAYING_INTERVAL", playingInterval.String())
	t.Setenv("CSS_SKIP_DELAY", skipDelay.String())
	t.Setenv("CSS_IGNORE_SEGMENT_DURATION", ignoreSegmentDuration.String())
	t.Setenv("CSS_NETWORK_INTERFACE", networkInterface)
	t.Setenv("CSS_CATEGORIES", "a,b,c")
	t.Setenv("CSS_ACTION_TYPES", "d,e,f")
	t.Setenv("CSS_YOUTUBE_API_KEY", "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe")
	t.Setenv("CSS_MUTE_ADS", "false")

	cmd := NewCommand("", "")
	cmd.RunE = func(_ *cobra.Command, _ []string) error { return nil }

	require.NoError(t, cmd.Execute())

	conf := config.FromContext(cmd.Context())
	assert.Equal(t, "warn", conf.LogLevel)
	assert.Equal(t, []string{"192.168.1.1", "192.168.1.2"}, conf.DeviceAddrStrs)
	assert.Equal(t, discoverInterval, conf.DiscoverInterval)
	assert.Equal(t, pausedInterval, conf.PausedInterval)
	assert.Equal(t, playingInterval, conf.PlayingInterval)
	assert.Equal(t, skipDelay, conf.SkipDelay)
	assert.Equal(t, ignoreSegmentDuration, conf.IgnoreSegmentDuration)
	assert.Equal(t, networkInterface, conf.NetworkInterfaceName)
	assert.Equal(t, []string{"a", "b", "c"}, conf.Categories)
	assert.Equal(t, []string{"d", "e", "f"}, conf.ActionTypes)
	assert.Equal(t, "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe", conf.YouTubeAPIKey)
	assert.False(t, conf.MuteAds)
	assert.Len(t, conf.DeviceAddrs, 2)
	assert.NotNil(t, conf.NetworkInterface)
}

func TestSBCEnvs(t *testing.T) {
	discoverInterval := randDuration()
	playingInterval := randDuration()

	t.Setenv("SBCSCANINTERVAL", strconv.Itoa(int(discoverInterval.Seconds())))
	t.Setenv("SBCPOLLINTERVAL", strconv.Itoa(int(playingInterval.Seconds())))
	t.Setenv("SBCCATEGORIES", "a b c")
	t.Setenv("SBCYOUTUBEAPIKEY", "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe")

	cmd := NewCommand("", "")
	cmd.RunE = func(_ *cobra.Command, _ []string) error { return nil }

	require.NoError(t, cmd.Execute())

	conf := config.FromContext(cmd.Context())
	assert.Equal(t, discoverInterval, conf.DiscoverInterval)
	assert.Equal(t, playingInterval, conf.PlayingInterval)
	assert.Equal(t, []string{"a", "b", "c"}, conf.Categories)
	assert.Equal(t, "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe", conf.YouTubeAPIKey)
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
