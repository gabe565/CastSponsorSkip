package device

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gabe565/castsponsorskip/internal/config"
	"github.com/gabe565/castsponsorskip/internal/sponsorblock"
	"github.com/gabe565/castsponsorskip/internal/util"
	"github.com/vishen/go-chromecast/application"
	"github.com/vishen/go-chromecast/cast/proto"
	castdns "github.com/vishen/go-chromecast/dns"
)

func Watch(ctx context.Context, entry castdns.CastEntry) {
	var logger *slog.Logger
	if entry.DeviceName != "" {
		logger = slog.With("device", entry.DeviceName)
	} else {
		logger = slog.With("device", entry.Device)
	}

	ticker := time.NewTicker(config.PlayingIntervalValue)
	defer func() {
		ticker.Stop()
	}()

	app := application.NewApplication()

	if err := util.Retry(ctx, 6, 500*time.Millisecond, func(try uint) error {
		if err := app.Start(entry.GetAddr(), entry.GetPort()); err != nil {
			logger.Debug("Failed to connect to device. Retrying...", "try", try, "error", err.Error())

			var subErr error
			if entry, subErr = DiscoverCastDNSEntryByUuid(ctx, entry.UUID); subErr != nil {
				return subErr
			}

			return err
		}
		return nil
	}); err != nil {
		logger.Error("Failed to connect to device.", "error", err.Error())
		return
	}
	defer func() {
		_ = app.Close(false)
	}()
	if ctx.Err() != nil {
		return
	}

	logger.Info("Connected to cast device.")

	var prevVideoId string
	var segments []sponsorblock.Segment

	app.AddMessageFunc(func(msg *api.CastMessage) {
		ticker.Reset(config.PlayingIntervalValue)
	})

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			logger.Debug("Update")

			if err := util.Retry(ctx, 6, 500*time.Millisecond, func(try uint) error {
				if err := app.Update(); err != nil {
					logger.Debug("Failed to update device. Retrying...", "try", try, "error", err.Error())
					return err
				}
				return nil
			}); err != nil {
				logger.Error("Lost connection to device.", "error", err.Error())
				return
			}

			castApp, castMedia, _ := app.Status()

			if castApp == nil || castApp.DisplayName != "YouTube" || castMedia == nil || castMedia.PlayerState != "PLAYING" || castMedia.Media.ContentId == "" {
				segments = nil
				ticker.Reset(config.PausedIntervalValue)
				continue
			}

			if castMedia.Media.ContentId != prevVideoId {
				logger.Info("Detected video stream.", "video_id", castMedia.Media.ContentId)
				segments = nil
				prevVideoId = castMedia.Media.ContentId
			}

			if len(segments) == 0 {
				var err error
				segments, err = sponsorblock.QuerySegments(ctx, castMedia.Media.ContentId)
				if err == nil {
					if len(segments) == 0 {
						logger.Info("No segments found for video.", "video_id", castMedia.Media.ContentId)
					} else {
						logger.Info("Found segments for video.", "segments", len(segments))
					}
				} else {
					logger.Error("Failed to query segments", "error", err.Error())
				}
			}

			for _, segment := range segments {
				if castMedia.CurrentTime > segment.Segment[0] && castMedia.CurrentTime < segment.Segment[1]-1 {
					from := time.Duration(castMedia.CurrentTime) * time.Second
					to := time.Duration(segment.Segment[1]) * time.Second
					logger.Info("Skipping to timestamp.", "category", segment.Category, "from", from, "to", to)
					if err := app.SeekToTime(segment.Segment[1]); err != nil {
						logger.Warn("Failed to seek to timestamp.", "to", to, "error", err.Error())
					}
					break
				}
			}

			if err := app.Skipad(); err == nil {
				logger.Info("Skipped ad.")
			} else if !errors.Is(err, application.ErrNoMediaSkipad) {
				logger.Warn("Failed to skip ad.", "error", err.Error())
			}

			ticker.Reset(config.PlayingIntervalValue)
		}
	}
}
