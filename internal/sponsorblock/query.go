package sponsorblock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gabe565/castsponsorskip/internal/config"
)

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

func QuerySegments(ctx context.Context, id string) ([]Segment, error) {
	u := url.URL{
		Scheme: "https",
		Host:   "sponsor.ajay.app",
		Path:   "/api/skipSegments",
	}

	query := make(url.Values)
	query.Set("videoID", id)
	for _, category := range config.CategoriesValue {
		query.Add("category", category)
	}
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

	var segments []Segment
	if err := json.NewDecoder(resp.Body).Decode(&segments); err != nil {
		return nil, err
	}

	return segments, nil
}
