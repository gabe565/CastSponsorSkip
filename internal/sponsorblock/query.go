package sponsorblock

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

	query := make(url.Values, len(config.Default.Categories))
	for _, category := range config.Default.Categories {
		query.Add("category", category)
	}

	u := baseUrl
	u.Path = path.Join("api", "skipSegments", checksum[:4])
	u.RawQuery = query.Encode()

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
			return nil, fmt.Errorf("%w: %s", ErrStatusCode, resp.Status)
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
