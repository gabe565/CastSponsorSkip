package youtube

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gabe565/castsponsorskip/internal/config"
	"github.com/gabe565/castsponsorskip/internal/util"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	ErrNotConnected = errors.New("not connected to YouTube")
	ErrNoVideos     = errors.New("search returned no videos")
	ErrNoId         = errors.New("search result does not have a valid video ID")
)

var service *youtube.Service

func CreateService(ctx context.Context, opts ...option.ClientOption) error {
	var err error
	opts = append(
		opts,
		option.WithAPIKey(config.Default.YouTubeAPIKey),
		option.WithTelemetryDisabled(),
	)
	service, err = youtube.NewService(ctx, opts...)
	return err
}

func QueryVideoId(ctx context.Context, artist, title string) (string, error) {
	if service == nil {
		return "", util.HaltRetries(ErrNotConnected)
	}

	query := fmt.Sprintf(`%q+intitle:%q`, artist, title)
	slog.Debug("Searching for video ID", "query", query)
	response, err := service.Search.List([]string{"id"}).
		Q(query).
		MaxResults(1).
		Context(ctx).
		Do()
	if err != nil {
		return "", err
	}

	if len(response.Items) == 0 || response.Items[0] == nil {
		return "", util.HaltRetries(ErrNoVideos)
	}

	item := response.Items[0]
	if item.Id == nil || item.Id.VideoId == "" {
		return "", util.HaltRetries(ErrNoId)
	}

	return item.Id.VideoId, nil
}
