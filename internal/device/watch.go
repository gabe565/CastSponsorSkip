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

const (
	StatePlaying   = "PLAYING"
	StateBuffering = "BUFFERING"
	StateAd        = 1081

	NoMutedSegment = -1
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

	app := application.NewApplication(
		application.WithSkipadSleep(config.Default.PlayingInterval),
		application.WithSkipadRetries(int(time.Minute/config.Default.PlayingInterval)),
	)

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
	mutedSegmentId := NoMutedSegment

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
			case StatePlaying, StateBuffering:
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

			castApp, castMedia, castVol := app.Status()

			if castApp == nil || castApp.DisplayName != "YouTube" || castMedia == nil {
				mediaSessionId = 0
				ticker.Reset(config.Default.PausedInterval)
				continue
			}

			mediaSessionId = castMedia.MediaSessionId
			if castMedia.PlayerState != StatePlaying && castMedia.PlayerState != StateBuffering {
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
						if err := util.Retry(ctx, 10, 500*time.Millisecond, func(try uint) (err error) {
							castMedia.Media.ContentId, err = youtube.QueryVideoId(ctx, currArtist, currTitle)
							if errors.Is(err, youtube.ErrNoVideos) || errors.Is(err, youtube.ErrNoId) {
								return util.HaltRetries(err)
							}
							return err
						}); err != nil {
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

			switch castMedia.CustomData.PlayerState {
			case StateAd:
				var shouldUnmute bool
				if config.Default.MuteAds && !castVol.Muted {
					logger.Info("Detected ad. Muting and attempting to skip...")
					if err := app.SetMuted(true); err == nil {
						shouldUnmute = true
					} else {
						logger.Warn("Failed to mute ad.", "error", err.Error())
					}
				} else {
					logger.Info("Detected ad. Attempting to skip...")
				}

				if err := app.Skipad(); err == nil {
					logger.Info("Skipped ad.")
				} else if !errors.Is(err, application.ErrNoMediaSkipad) {
					logger.Warn("Failed to skip ad.", "error", err.Error())
				}

				if shouldUnmute {
					if err := app.SetMuted(false); err != nil {
						logger.Warn("Failed to unmute ad.", "error", err.Error())
					}
				}
			default:
				if castMedia.Media.ContentId != prevVideoId {
					logger.Info("Detected video stream.", "video_id", castMedia.Media.ContentId)
					prevVideoId = castMedia.Media.ContentId

					if mutedSegmentId != NoMutedSegment {
						if err := app.SetMuted(false); err == nil {
							mutedSegmentId = NoMutedSegment
						} else {
							logger.Warn("Failed to unmute after video change.", "error", err.Error())
						}
					}

					if err := util.Retry(ctx, 10, 500*time.Millisecond, func(try uint) (err error) {
						segments, err = sponsorblock.QuerySegments(ctx, castMedia.Media.ContentId)
						return err
					}); err == nil {
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
					if segment.Segment[0] <= castMedia.CurrentTime && castMedia.CurrentTime < segment.Segment[1]-1 {
						from := time.Duration(castMedia.CurrentTime) * time.Second
						to := time.Duration(segment.Segment[1]) * time.Second
						switch segment.ActionType {
						case sponsorblock.ActionTypeSkip:
							logger.Info("Skipping to timestamp.", "category", segment.Category, "from", from, "to", to)
							if err := app.SeekToTime(segment.Segment[1]); err != nil {
								logger.Warn("Failed to seek to timestamp.", "to", segment.Segment[1], "error", err.Error())
							}
							castMedia.CurrentTime = segment.Segment[1]
						case sponsorblock.ActionTypeMute:
							if !castVol.Muted || mutedSegmentId != NoMutedSegment {
								logger.Info("Mute segment.", "category", segment.Category, "from", from, "to", to)
								if err := app.SetMuted(true); err == nil {
									mutedSegmentId = i
								} else {
									logger.Warn("Failed to mute "+segment.Category+".", "error", err.Error())
								}
							}
						}
					}
				}

				if mutedSegmentId != NoMutedSegment {
					segment := segments[mutedSegmentId]
					if castMedia.CurrentTime < segment.Segment[0] || segment.Segment[1] <= castMedia.CurrentTime {
						from := time.Duration(castMedia.CurrentTime) * time.Second
						to := time.Duration(segment.Segment[1]) * time.Second
						logger.Info("Unmute segment.", "category", segment.Category, "from", from, "to", to)
						if err := app.SetMuted(false); err == nil {
							mutedSegmentId = NoMutedSegment
						} else {
							logger.Warn("Failed to unmute segment.", "error", err.Error())
						}
					}
				}
			}

			ticker.Reset(config.Default.PlayingInterval)
		}
	}
}
