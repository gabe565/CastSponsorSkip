package device

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gabe565/sponsorblockcast/internal/config"
	"github.com/gabe565/sponsorblockcast/internal/sponsorblock"
	"github.com/vishen/go-chromecast/application"
	"github.com/vishen/go-chromecast/cast/proto"
	castdns "github.com/vishen/go-chromecast/dns"
)

func Watch(ctx context.Context, entry castdns.CastEntry) {
	logGroup := slog.Group("device", "type", entry.Device, "name", entry.DeviceName)

	slog.With(logGroup).Info("Found cast device")

	ticker := time.NewTicker(config.PlayingIntervalValue)
	defer func() {
		ticker.Stop()
	}()

	app := application.NewApplication()

	if err := app.Start(entry.GetAddr(), entry.GetPort()); err != nil {
		slog.With(logGroup).Warn("Failed to start application", "error", err.Error())
		return
	}
	defer func() {
		_ = app.Close(false)
	}()

	var prevVideoId string
	var segments []sponsorblock.Segment

	if err := app.Update(); err != nil {
		slog.With(logGroup).Warn("Failed to update application")
		return
	}

	app.AddMessageFunc(func(msg *api.CastMessage) {
		ticker.Reset(config.PlayingIntervalValue)
	})

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			slog.With(logGroup).Debug("Update")

			if err := app.Update(); err != nil {
				slog.With(logGroup).Warn("Failed to update application", "error", err.Error())
				continue
			}

			castApp, castMedia, _ := app.Status()

			if castApp == nil || castApp.DisplayName != "YouTube" || castMedia == nil || castMedia.PlayerState != "PLAYING" || castMedia.Media.ContentId == "" {
				segments = nil
				ticker.Reset(config.PausedIntervalValue)
				continue
			}

			if castMedia.Media.ContentId != prevVideoId {
				slog.With(logGroup).Info("Watching stream", "app_name", castApp.DisplayName, "content_id", castMedia.Media.ContentId)
				segments = nil
				prevVideoId = castMedia.Media.ContentId
			}

			if len(segments) == 0 {
				var err error
				segments, err = sponsorblock.QuerySegments(castMedia.Media.ContentId)
				if err == nil {
					if len(segments) == 0 {
						slog.With(logGroup).Info("No segments found for video", "video_id", castMedia.Media.ContentId)
					} else {
						slog.With(logGroup).Info("Found segments for video", "segments", len(segments))
					}
				} else {
					slog.With(logGroup).Error("Failed to query segments", "error", err.Error())
				}
			}

			for _, segment := range segments {
				if castMedia.CurrentTime > segment.Segment[0] && castMedia.CurrentTime < segment.Segment[1]-1 {
					slog.With(logGroup).Info("Skipping to timestamp", "category", segment.Category, "timestamp", castMedia.CurrentTime, "segment", segment.Segment)
					if err := app.SeekToTime(segment.Segment[1]); err != nil {
						slog.With(logGroup).Warn("Failed to seek to timestamp", "to", segment.Segment[1], "error", err.Error())
					}
					break
				}
			}

			if err := app.Skipad(); err == nil {
				slog.With(logGroup).Info("Skipped ad")
			} else if !errors.Is(err, application.ErrNoMediaSkipad) {
				slog.With(logGroup).Warn("Failed to skip ad", "error", err.Error())
			}

			ticker.Reset(config.PlayingIntervalValue)
		}
	}
}
