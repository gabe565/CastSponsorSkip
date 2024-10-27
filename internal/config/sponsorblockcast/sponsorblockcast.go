package sponsorblockcast

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"gabe565.com/castsponsorskip/internal/config/names"
)

func Provider() SponsorBlockCast {
	return SponsorBlockCast{}
}

type SponsorBlockCast struct{}

var ErrUnsupported = errors.New("sponsorblockcast provider does not support this method")

func (s SponsorBlockCast) ReadBytes() ([]byte, error) {
	return nil, ErrUnsupported
}

func (s SponsorBlockCast) Read() (map[string]any, error) {
	result := make(map[string]any, 4)

	if env := os.Getenv("SBCSCANINTERVAL"); env != "" {
		parsed, err := strconv.Atoi(env)
		if err == nil {
			val := (time.Duration(parsed) * time.Second).String()
			slog.Warn(fmt.Sprintf(`SBCSCANINTERVAL is deprecated. Please set %q instead.`, "CSS_DISCOVER_INTERVAL="+val))
			result[names.FlagDiscoverInterval] = val
		}
	}

	if env := os.Getenv("SBCPOLLINTERVAL"); env != "" {
		parsed, err := strconv.Atoi(env)
		if err == nil {
			val := (time.Duration(parsed) * time.Second).String()
			slog.Warn(fmt.Sprintf(`SBCPOLLINTERVAL is deprecated. Please set %q instead.`, "CSS_PLAYING_INTERVAL="+val))
			result[names.FlagPlayingInterval] = val
		}
	}

	if env := os.Getenv("SBCCATEGORIES"); env != "" {
		val := strings.Split(env, " ")
		slog.Warn(fmt.Sprintf(`SBCCATEGORIES is deprecated. Please set %q instead.`, "CSS_CATEGORIES="+strings.Join(val, ",")))
		result[names.FlagCategories] = val
	}

	if env := os.Getenv("SBCYOUTUBEAPIKEY"); env != "" {
		slog.Warn(fmt.Sprintf(`SBCYOUTUBEAPIKEY is deprecated. Please set %q instead.`, "CSS_YOUTUBE_API_KEY="+env))
		result[names.FlagYouTubeAPIKey] = env
	}

	return result, nil
}
