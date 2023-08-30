package sponsorblock

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"

	"github.com/gabe565/castsponsorskip/internal/config"
)

type Video struct {
	VideoID  string    `json:"videoID"`
	Segments []Segment `json:"segments"`
}

type Segment struct {
	Segment       [2]float32
	UUID          string
	Category      string
	VideoDuration float32
	ActionType    string
	Locked        int
	Votes         int
	Description   string
}

var ErrStatusCode = errors.New("invalid response status")

var baseUrl = url.URL{
	Scheme: "https",
	Host:   "sponsor.ajay.app",
}

func QuerySegments(ctx context.Context, id string) ([]Segment, error) {
	checksumBytes := sha256.Sum256([]byte(id))
	checksum := hex.EncodeToString(checksumBytes[:])

	query := make(url.Values, len(config.Default.Categories)+len(config.Default.ActionTypes))
	for _, category := range config.Default.Categories {
		query.Add("category", category)
	}
	for _, actionType := range config.Default.ActionTypes {
		query.Add("actionType", actionType)
	}

	u := baseUrl
	u.Path = path.Join("api", "skipSegments", checksum[:4])
	u.RawQuery = query.Encode()

	slog.Debug("Request segments", "url", u.String())
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, nil
		} else {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("%w: %s %s", ErrStatusCode, resp.Status, body)
		}
	}

	var videos []Video
	if err := json.NewDecoder(resp.Body).Decode(&videos); err != nil {
		return nil, err
	}

	for _, video := range videos {
		if video.VideoID == id {
			return video.Segments, nil
		}
	}

	return nil, nil
}
