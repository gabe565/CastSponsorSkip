package youtube

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gabe565/castsponsorskip/internal/config"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	ErrNotConnected = errors.New("not connected to YouTube")
	ErrNoVideos     = errors.New("search returned no videos")
	ErrNoId         = errors.New("search result does not have a valid video ID")
)

var service *youtube.Service

func CreateService(ctx context.Context) error {
	var err error
	service, err = youtube.NewService(
		ctx,
		option.WithAPIKey(config.Default.YouTubeAPIKey),
		option.WithTelemetryDisabled(),
	)
	return err
}

func QueryVideoId(ctx context.Context, artist, title string) (string, error) {
	if service == nil {
		return "", ErrNotConnected
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
		return "", ErrNoVideos
	}

	item := response.Items[0]
	if item.Id == nil || item.Id.VideoId == "" {
		return "", ErrNoId
	}

	return item.Id.VideoId, nil
}
