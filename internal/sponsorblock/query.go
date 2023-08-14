package sponsorblock

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var Categories = []string{"sponsor"}

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

func QuerySegments(id string) ([]Segment, error) {
	u := url.URL{
		Scheme: "https",
		Host:   "sponsor.ajay.app",
		Path:   "/api/skipSegments",
	}

	query := make(url.Values)
	query.Set("videoID", id)
	for _, category := range Categories {
		query.Add("category", category)
	}
	u.RawQuery = query.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrStatusCode, resp.Status)
	}

	var segments []Segment
	if err := json.NewDecoder(resp.Body).Decode(&segments); err != nil {
		return nil, err
	}

	return segments, nil
}
