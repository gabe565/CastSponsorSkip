package sponsorblock

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gabe565/castsponsorskip/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestQuerySegmentsRequest(t *testing.T) {
	type args struct {
		id         string
		categories []string
	}
	tests := []struct {
		name         string
		args         args
		wantPath     string
		errAssertion assert.ErrorAssertionFunc
	}{
		{
			"1",
			args{"dQw4w9WgXcQ", []string{"sponsor"}},
			"/api/skipSegments/5f6b",
			assert.NoError,
		},
		{
			"2",
			args{"y8Kyi0WNg40", []string{"sponsor", "selfpromo"}},
			"/api/skipSegments/30cc",
			assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func(c []string) {
				config.Default.Categories = c
			}(config.Default.Categories)
			config.Default.Categories = tt.args.categories

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.wantPath, r.URL.Path)
				assert.Equal(t, tt.args.categories, r.URL.Query()["category"])
				_, _ = w.Write([]byte("[]"))
			}))
			defer server.Close()

			tempUrl, err := url.Parse(server.URL)
			if !assert.NoError(t, err) {
				return
			}
			defer func(u url.URL) {
				baseUrl = u
			}(baseUrl)
			baseUrl = *tempUrl

			if _, err := QuerySegments(context.Background(), tt.args.id); !tt.errAssertion(t, err) {
				return
			}
		})
	}
}

func TestQuerySegmentsResponse(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name         string
		args         args
		handler      http.HandlerFunc
		want         []Segment
		errAssertion assert.ErrorAssertionFunc
	}{
		{
			"200 OK",
			args{"dQw4w9WgXcQ"},
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`[{"videoID": "dQw4w9WgXcQ", "segments": [{}]}, {"videoID": "y8Kyi0WNg40", "segments": [{}, {}]}]`))
			},
			[]Segment{{}},
			assert.NoError,
		},
		{
			"400 Bad Request",
			args{"dQw4w9WgXcQ"},
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			nil,
			assert.Error,
		},
		{
			"404 Not Found",
			args{"y8Kyi0WNg40"},
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			nil,
			assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			tempUrl, err := url.Parse(server.URL)
			if !assert.NoError(t, err) {
				return
			}
			defer func(u url.URL) {
				baseUrl = u
			}(baseUrl)
			baseUrl = *tempUrl

			got, err := QuerySegments(context.Background(), tt.args.id)
			if !tt.errAssertion(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
