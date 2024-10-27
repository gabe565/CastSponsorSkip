package device

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"gabe565.com/castsponsorskip/internal/config"
	"gabe565.com/castsponsorskip/internal/sponsorblock"
	"gabe565.com/castsponsorskip/internal/util"
	"gabe565.com/castsponsorskip/internal/youtube"
	"github.com/buger/jsonparser"
	"github.com/vishen/go-chromecast/application"
	"github.com/vishen/go-chromecast/cast"
	api "github.com/vishen/go-chromecast/cast/proto"
	castdns "github.com/vishen/go-chromecast/dns"
)

const (
	StatePlaying   = "PLAYING"
	StateBuffering = "BUFFERING"
	StateIdle      = "IDLE"
	StateAd        = 1081

	NoMutedSegment   = -1
	NoSkippedSegment = -1
)

//nolint:gochecknoglobals
var (
	listeners  = make(map[string]struct{})
	listenerMu sync.Mutex
)

type Device struct {
	config *config.Config
	ctx    context.Context
	cancel context.CancelFunc

	mu     sync.Mutex
	entry  castdns.CastEntry
	opts   []application.ApplicationOption
	app    *application.Application
	logger *slog.Logger

	tickInterval time.Duration
	ticker       *time.Ticker

	state             string
	meta              VideoMeta
	segments          []sponsorblock.Segment
	prevSegmentIdx    int
	prevSegmentIgnore time.Time
	mutedSegmentID    int
}

func NewDevice(config *config.Config, entry castdns.CastEntry, opts ...Option) *Device {
	if entry.Device == "Google Cast Group" {
		return nil
	} else if entry.Device == "" && entry.DeviceName == "" && entry.UUID == "" {
		return nil
	}

	var logger *slog.Logger
	if entry.DeviceName != "" {
		logger = slog.With("device", entry.DeviceName)
	} else {
		logger = slog.With("device", entry.Device)
	}

	if hasVideoOut, err := HasVideoOut(entry); err == nil && !hasVideoOut {
		logger.Debug("Ignoring device.", "reason", "Does not support video")
		return nil
	}

	listenerMu.Lock()
	if _, ok := listeners[entry.UUID]; ok {
		listenerMu.Unlock()
		logger.Debug("Ignoring device.", "reason", "Already connected")
		return nil
	}
	listeners[entry.UUID] = struct{}{}
	listenerMu.Unlock()

	device := &Device{
		config:         config,
		entry:          entry,
		logger:         logger,
		mutedSegmentID: NoMutedSegment,
		prevSegmentIdx: NoSkippedSegment,
	}

	for _, opt := range opts {
		opt(device)
	}

	return device
}

func (d *Device) Close() error {
	defer func() {
		listenerMu.Lock()
		delete(listeners, d.entry.UUID)
		listenerMu.Unlock()
	}()

	d.unmuteSegment()
	if d.cancel != nil {
		d.cancel()
	}

	if d.ticker != nil {
		d.ticker.Stop()
	}

	return d.closeApp()
}

func (d *Device) closeApp() error {
	if d.app != nil {
		defer func() {
			if r := recover(); r != nil {
				d.logger.Debug("Panic during close", "error", r, "stack", string(debug.Stack()))
			}
		}()

		return d.app.Close(false)
	}
	return nil
}

func (d *Device) BeginTick(opts ...application.ApplicationOption) error {
	defer func() {
		if r := recover(); r != nil {
			d.logger.Error("Recovered from panic.", "error", r)
			_, _ = os.Stderr.Write(debug.Stack())
		}
	}()

	if err := d.connect(opts...); err != nil {
		d.logger.Error("Failed to connect to device.", "error", err.Error())
		return err
	}

	if d.ticker != nil {
		d.ticker.Stop()
	}
	d.ticker = time.NewTicker(d.config.PlayingInterval)

	for {
		select {
		case <-d.ctx.Done():
			return d.ctx.Err()
		case <-d.ticker.C:
			if err := d.tick(); err != nil {
				d.logger.Error("Lost connection to device.", "error", err.Error())
				return err
			}
		}
	}
}

func (d *Device) tick() error {
	if err := d.update(); err != nil {
		return err
	}

	castApp, castMedia, castVol := d.app.Status()

	if castApp == nil || castApp.DisplayName != "YouTube" || castMedia == nil {
		d.changeTickInterval(d.config.PausedInterval)
		return nil
	}

	if castMedia.PlayerState != StatePlaying && castMedia.PlayerState != StateBuffering {
		d.changeTickInterval(d.config.PausedInterval)
		return nil
	}

	switch castMedia.CustomData.PlayerState {
	case StateAd:
		d.muteAd(castVol)
	default:
		if castMedia.Media.Metadata.Artist != "" {
			d.meta.CurrArtist = castMedia.Media.Metadata.Artist
		} else {
			d.meta.CurrArtist = castMedia.Media.Metadata.Subtitle
		}
		d.meta.CurrTitle = castMedia.Media.Metadata.Title

		switch {
		case !d.config.SkipSponsors:
			if d.meta.CurrVideoID != d.meta.PrevVideoID || !d.meta.SameVideo() {
				d.meta.PrevVideoID = d.meta.CurrVideoID
				d.meta.PrevArtist = d.meta.CurrArtist
				d.meta.PrevTitle = d.meta.CurrTitle
				d.logger.Info("Detected video stream.", "video_id", castMedia.Media.ContentId)
			}
		case castMedia.Media.ContentId != "":
			d.meta.CurrVideoID = castMedia.Media.ContentId
		default:
			d.queryVideoID()
		}

		if d.meta.CurrVideoID != d.meta.PrevVideoID {
			d.segments = nil
			d.prevSegmentIdx = NoSkippedSegment
			if d.meta.CurrVideoID != "" {
				d.logger.Info("Detected video stream.", "video_id", d.meta.CurrVideoID)
				d.meta.PrevVideoID = d.meta.CurrVideoID
				go d.querySegments()
			}
			d.unmuteSegment()
			break
		}

		for i, segment := range d.segments {
			if (segment.Segment[0]+float32(d.config.SkipDelay.Seconds())) <= castMedia.CurrentTime && castMedia.CurrentTime < segment.Segment[1]-1 {
				d.handleSegment(castMedia, castVol, segment, i)
			}
		}

		if d.mutedSegmentID != NoMutedSegment {
			segment := d.segments[d.mutedSegmentID]
			if castMedia.CurrentTime < segment.Segment[0]-1 || segment.Segment[1] <= castMedia.CurrentTime {
				from := time.Duration(castMedia.CurrentTime) * time.Second
				to := time.Duration(segment.Segment[1]) * time.Second
				d.logger.Info("Unmute segment.", "category", segment.Category, "from", from, "to", to)
				d.unmuteSegment()
			}
		}
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if d.state != StateIdle {
		d.changeTickInterval(d.config.PlayingInterval)
	}
	return nil
}

func (d *Device) connect(opts ...application.ApplicationOption) error {
	logLevel := slog.LevelInfo
	if d.app != nil {
		logLevel = slog.LevelDebug
		_ = d.closeApp()
	}
	d.opts = opts
	opts = append(
		opts,
		application.WithSkipadSleep(d.config.PlayingInterval),
		application.WithSkipadRetries(int(time.Minute/d.config.PlayingInterval)),
		application.WithIface(d.config.NetworkInterface),
	)
	d.app = application.NewApplication(opts...)
	d.app.AddMessageFunc(d.onMessage)

	if err := util.Retry(d.ctx, 6, 500*time.Millisecond, func(try uint) error {
		if err := d.app.Start(d.entry.GetAddr(), d.entry.GetPort()); err != nil {
			d.logger.Debug("Failed to connect to device. Retrying...", "try", try, "error", err.Error())

			newEntry, subErr := DiscoverCastDNSEntryByUUID(d.ctx, d.config, d.entry.UUID)
			if subErr != nil {
				return subErr
			}
			d.entry = newEntry

			return err
		}
		return nil
	}); err != nil {
		return err
	}
	if d.ctx.Err() == nil {
		d.logger.Log(d.ctx, logLevel, "Connected to cast device.")
	}

	return nil
}

func (d *Device) onMessage(msg *api.CastMessage) {
	payload := []byte(msg.GetPayloadUtf8())
	msgType, _ := jsonparser.GetString(payload, "type")
	switch msgType {
	case "RECEIVER_STATUS":
		appID, _ := jsonparser.GetString(payload, "status", "applications", "[0]", "displayName")
		d.mu.Lock()
		defer d.mu.Unlock()
		if appID == "YouTube" && d.state != StateIdle {
			d.changeTickInterval(d.config.PlayingInterval)
		}
	case "MEDIA_STATUS":
		var playerState string
		customPlayerState, _ := jsonparser.GetInt(payload, "status", "[0]", "customData", "playerState")
		if customPlayerState == 5 {
			playerState = StateIdle
		} else {
			playerState, _ = jsonparser.GetString(payload, "status", "[0]", "playerState")
		}
		d.state = playerState
		switch playerState {
		case StatePlaying, StateBuffering:
			d.changeTickInterval(d.config.PlayingInterval)
		case StateIdle:
			d.changeTickInterval(d.config.PausedInterval)
			d.unmuteSegment()
		}
	case "CLOSE":
		d.unmuteSegment()
		d.segments = nil
		d.prevSegmentIdx = NoSkippedSegment
		d.meta.Clear()
	}
}

func (d *Device) update() error {
	d.logger.Debug("Requesting update.")

	err := d.app.Update()
	if err != nil {
		d.logger.Debug("Failed to update device. Reconnecting...", "error", err.Error())

		if subErr := d.connect(d.opts...); subErr != nil {
			d.logger.Debug("Failed to reconnect.", "error", subErr.Error())
			return err
		}

		if subErr := d.app.Update(); subErr == nil {
			return nil
		}
	}

	return err
}

func (d *Device) queryVideoID() {
	if d.meta.Empty() || d.meta.SameVideo() {
		return
	}

	d.meta.PrevArtist = d.meta.CurrArtist
	d.meta.PrevTitle = d.meta.CurrTitle
	d.unmuteSegment()
	d.segments = nil
	d.prevSegmentIdx = NoSkippedSegment

	if d.config.YouTubeAPIKey == "" {
		d.logger.Error("Video ID not set. Please configure a YouTube API key.")
	} else {
		d.logger.Info("Video ID not set. Searching YouTube for video ID...")
		go func() {
			err := util.Retry(d.ctx, 3, time.Second, func(_ uint) error {
				contentID, err := youtube.QueryVideoID(d.ctx, d.meta.CurrArtist, d.meta.CurrTitle)
				if err != nil {
					d.logger.Error("YouTube search failed.", "error", err.Error())
					return err
				}

				d.meta.CurrVideoID = contentID
				return nil
			})
			if err == nil {
				d.logger.Debug("YouTube search returned video ID.", "video_id", d.meta.CurrVideoID)
			} else {
				d.logger.Debug("Halting YouTube search retries.")
			}
		}()
	}
}

func (d *Device) muteAd(castVol *cast.Volume) {
	var shouldUnmute bool
	if d.config.MuteAds && !castVol.Muted {
		d.logger.Info("Detected ad. Muting and attempting to skip...")
		if err := d.app.SetMuted(true); err == nil {
			shouldUnmute = true
		} else {
			d.logger.Warn("Failed to mute ad.", "error", err.Error())
		}
	} else {
		d.logger.Info("Detected ad. Attempting to skip...")
	}

	if err := d.app.Skipad(); err == nil {
		d.logger.Info("Skipped ad.")
	} else if !errors.Is(err, application.ErrNoMediaSkipad) {
		d.logger.Warn("Failed to skip ad.", "error", err.Error())
	}

	if shouldUnmute {
		if err := d.app.SetMuted(false); err != nil {
			d.logger.Warn("Failed to unmute ad.", "error", err.Error())
		}
	}
}

func (d *Device) handleSegment(castMedia *cast.Media, castVol *cast.Volume, segment sponsorblock.Segment, i int) {
	from := time.Duration(castMedia.CurrentTime) * time.Second
	to := time.Duration(segment.Segment[1]) * time.Second
	switch segment.ActionType {
	case sponsorblock.ActionTypeSkip:
		if i == d.prevSegmentIdx {
			if now := time.Now(); now.Before(d.prevSegmentIgnore) {
				d.logger.Debug("Ignoring segment.", "category", segment.Category, "from", from, "to", to, "until", d.prevSegmentIgnore.Truncate(time.Second).String())
				d.prevSegmentIgnore = now.Add(d.config.IgnoreSegmentDuration)
				return
			}
		}

		d.logger.Info("Skipping to timestamp.", "category", segment.Category, "from", from, "to", to)
		// Cast API seems to ignore decimals, so add 100ms to seek time in case sponsorship ends at 0.9 seconds.
		if err := d.app.SeekToTime(segment.Segment[1] + 0.1); err != nil {
			d.logger.Warn("Failed to seek to timestamp.", "to", segment.Segment[1], "error", err.Error())
		}
		castMedia.CurrentTime = segment.Segment[1]
		d.prevSegmentIdx = i
		d.prevSegmentIgnore = time.Now().Add(d.config.IgnoreSegmentDuration)
	case sponsorblock.ActionTypeMute:
		if !castVol.Muted || i != d.mutedSegmentID {
			d.logger.Info("Mute segment.", "category", segment.Category, "from", from, "to", to)
			if err := d.app.SetMuted(true); err == nil {
				d.mutedSegmentID = i
			} else {
				d.logger.Warn("Failed to mute "+segment.Category+".", "error", err.Error())
			}
		}
	}
}

func (d *Device) unmuteSegment() {
	if d.mutedSegmentID != NoMutedSegment {
		if err := d.app.SetMuted(false); err == nil {
			d.mutedSegmentID = NoMutedSegment
		} else {
			d.logger.Warn("Failed to unmute after video change.", "error", err.Error())
		}
	}
}

func (d *Device) querySegments() {
	if d.meta.CurrVideoID == "" {
		return
	}

	if err := util.Retry(d.ctx, 10, 500*time.Millisecond, func(_ uint) error {
		var err error
		d.segments, err = sponsorblock.QuerySegments(d.ctx, d.config, d.meta.CurrVideoID)
		return err
	}); err == nil {
		if len(d.segments) == 0 {
			d.logger.Info("No segments found for video.", "video_id", d.meta.CurrVideoID)
		} else {
			d.logger.Info("Found segments for video.", "segments", len(d.segments))
		}
	} else {
		d.logger.Error("Failed to query segments. Retrying...", "error", err.Error())
	}
}

func (d *Device) changeTickInterval(interval time.Duration) {
	if d.ticker != nil && interval != d.tickInterval {
		d.ticker.Reset(interval)
		d.tickInterval = interval
	}
}
