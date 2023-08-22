package youtube

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabe565/castsponsorskip/internal/config"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	ErrNoVideos = errors.New("search returned no videos")
	ErrNoId     = errors.New("search result does not have a valid video ID")
)

func QueryVideoId(ctx context.Context, artist, title string) (string, error) {
	service, err := youtube.NewService(ctx, option.WithAPIKey(config.Default.YouTubeAPIKey))
	if err != nil {
		return "", err
	}

	response, err := service.Search.List([]string{"id"}).
		Q(fmt.Sprintf(`%q+intitle:%q`, artist, title)).
		MaxResults(1).
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
