package device

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/buger/jsonparser"
	"github.com/gabe565/castsponsorskip/internal/config"
	"github.com/gabe565/castsponsorskip/internal/sponsorblock"
	"github.com/gabe565/castsponsorskip/internal/util"
	"github.com/gabe565/castsponsorskip/internal/youtube"
	"github.com/vishen/go-chromecast/application"
	"github.com/vishen/go-chromecast/cast/proto"
	castdns "github.com/vishen/go-chromecast/dns"
)

var (
	listeners  = make(map[string]struct{})
	listenerMu sync.Mutex
)

func Watch(ctx context.Context, entry castdns.CastEntry) {
	if entry.Device == "Google Cast Group" {
		return
	} else if entry.Device == "" && entry.DeviceName == "" && entry.UUID == "" {
		return
	}

	var logger *slog.Logger
	if entry.DeviceName != "" {
		logger = slog.With("device", entry.DeviceName)
	} else {
		logger = slog.With("device", entry.Device)
	}

	if hasVideoOut, err := HasVideoOut(entry); err == nil && !hasVideoOut {
		logger.Debug("Ignoring device.", "reason", "Does not support video")
		return
	}

	listenerMu.Lock()
	if _, ok := listeners[entry.UUID]; ok {
		listenerMu.Unlock()
		logger.Debug("Ignoring device.", "reason", "Already connected")
		return
	}
	listeners[entry.UUID] = struct{}{}
	listenerMu.Unlock()
	defer func() {
		listenerMu.Lock()
		delete(listeners, entry.UUID)
		listenerMu.Unlock()
	}()

	ticker := time.NewTicker(config.Default.PlayingInterval)
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
		if ctx.Err() == nil {
			logger.Error("Failed to connect to device.", "error", err.Error())
		}
		return
	}
	defer func() {
		_ = app.Close(false)
	}()
	if ctx.Err() != nil {
		return
	}

	logger.Info("Connected to cast device.")

	var prevVideoId, prevArtist, prevTitle string
	var mediaSessionId int
	var segments []sponsorblock.Segment

	app.AddMessageFunc(func(msg *api.CastMessage) {
		payload := []byte(msg.GetPayloadUtf8())
		msgType, _ := jsonparser.GetString(payload, "type")
		switch msgType {
		case "RECEIVER_STATUS":
			appId, _ := jsonparser.GetString(payload, "status", "applications", "[0]", "displayName")
			if appId == "YouTube" {
				ticker.Reset(config.Default.PlayingInterval)
			}
		case "MEDIA_STATUS":
			currMediaSessionId, err := jsonparser.GetInt(payload, "status", "[0]", "mediaSessionId")
			if err != nil {
				return
			}

			playerState, _ := jsonparser.GetString(payload, "status", "[0]", "playerState")
			switch playerState {
			case "PLAYING", "BUFFERING":
				if int(currMediaSessionId) == mediaSessionId {
					ticker.Reset(config.Default.PlayingInterval)
				}
			}
		}
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
				if ctx.Err() == nil {
					logger.Error("Lost connection to device.", "error", err.Error())
				}
				return
			}

			castApp, castMedia, _ := app.Status()

			if castApp == nil || castApp.DisplayName != "YouTube" || castMedia == nil {
				mediaSessionId = 0
				ticker.Reset(config.Default.PausedInterval)
				continue
			}

			mediaSessionId = castMedia.MediaSessionId
			if castMedia.PlayerState != "PLAYING" && castMedia.PlayerState != "BUFFERING" {
				ticker.Reset(config.Default.PausedInterval)
				continue
			}

			if castMedia.Media.ContentId == "" {
				var currArtist string
				if castMedia.Media.Metadata.Artist != "" {
					currArtist = castMedia.Media.Metadata.Artist
				} else {
					currArtist = castMedia.Media.Metadata.Subtitle
				}
				currTitle := castMedia.Media.Metadata.Title

				if currArtist == prevArtist && currTitle == prevTitle {
					castMedia.Media.ContentId = prevVideoId
				} else {
					if config.Default.YouTubeAPIKey == "" {
						slog.Warn("Video ID not found. Please set a YouTube API key.")
					} else {
						logger.Info("Video ID not found. Searching for video on YouTube...")
						var err error
						castMedia.Media.ContentId, err = youtube.QueryVideoId(ctx, currArtist, currTitle)
						if err != nil {
							logger.Error("Failed to find video on YouTube.", "error", err.Error())
						}
					}
					prevArtist = currArtist
					prevTitle = currTitle
				}
			}

			if castMedia.Media.ContentId == "" {
				ticker.Reset(config.Default.PausedInterval)
				continue
			}

			if castMedia.Media.ContentId != prevVideoId {
				logger.Info("Detected video stream.", "video_id", castMedia.Media.ContentId)
				prevVideoId = castMedia.Media.ContentId

				var err error
				segments, err = sponsorblock.QuerySegments(ctx, castMedia.Media.ContentId)
				if err == nil {
					if len(segments) == 0 {
						logger.Info("No segments found for video.", "video_id", castMedia.Media.ContentId)
					} else {
						logger.Info("Found segments for video.", "segments", len(segments))
					}
				} else {
					logger.Error("Failed to query segments. Retrying...", "error", err.Error())
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

			ticker.Reset(config.Default.PlayingInterval)
		}
	}
}
