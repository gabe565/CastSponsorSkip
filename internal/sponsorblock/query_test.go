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
		id          string
		categories  []string
		actionTypes []string
	}
	tests := []struct {
		name         string
		args         args
		wantPath     string
		errAssertion assert.ErrorAssertionFunc
	}{
		{
			"1",
			args{"dQw4w9WgXcQ", []string{"sponsor"}, []string{"skip", "mute"}},
			"/api/skipSegments/5f6b",
			assert.NoError,
		},
		{
			"2",
			args{"y8Kyi0WNg40", []string{"sponsor", "selfpromo"}, []string{"skip", "mute"}},
			"/api/skipSegments/30cc",
			assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func(categories []string, actionTypes []string) {
				config.Default.Categories = categories
				config.Default.ActionTypes = actionTypes
			}(config.Default.Categories, config.Default.ActionTypes)
			config.Default.Categories = tt.args.categories
			config.Default.ActionTypes = tt.args.actionTypes

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.wantPath, r.URL.Path)
				assert.Equal(t, tt.args.categories, r.URL.Query()["category"])
				assert.Equal(t, tt.args.actionTypes, r.URL.Query()["actionType"])
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
	testResponse := []byte(`[{"videoID":"fy9jO8JHaPo","segments":[{"category":"sponsor","actionType":"skip","segment":[53.433,57.705],"UUID":"e992e1c6dcebe5f21fc5dc68cfec12bc58cf7a68161983e74b56c89f1ac1d2c87","videoDuration":1399.461,"locked":0,"votes":0,"description":""},{"category":"sponsor","actionType":"skip","segment":[388.815,421.601],"UUID":"6c1c415479595a922bb1c67e4091bd804329be60a2013e17a421bc43137ae47b7","videoDuration":1399.461,"locked":0,"votes":-1,"description":""},{"category":"sponsor","actionType":"skip","segment":[927.803,997.758],"UUID":"3230cb569b64c51076d47c8000f7671e6231691667f0b722790d19be1ea578387","videoDuration":1399.461,"locked":0,"votes":0,"description":""}]}]`)

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
			"200 OK video in response",
			args{"fy9jO8JHaPo"},
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(testResponse)
			},
			[]Segment{
				{Segment: [2]float32{53.433, 57.705}, UUID: "e992e1c6dcebe5f21fc5dc68cfec12bc58cf7a68161983e74b56c89f1ac1d2c87", Category: "sponsor", VideoDuration: 1399.461, ActionType: "skip", Locked: 0, Votes: 0, Description: ""},
				{Segment: [2]float32{388.815, 421.601}, UUID: "6c1c415479595a922bb1c67e4091bd804329be60a2013e17a421bc43137ae47b7", Category: "sponsor", VideoDuration: 1399.461, ActionType: "skip", Locked: 0, Votes: -1, Description: ""},
				{Segment: [2]float32{927.803, 997.758}, UUID: "3230cb569b64c51076d47c8000f7671e6231691667f0b722790d19be1ea578387", Category: "sponsor", VideoDuration: 1399.461, ActionType: "skip", Locked: 0, Votes: 0, Description: ""},
			},
			assert.NoError,
		},
		{
			"200 OK video not in response",
			args{"dQw4w9WgXcQ"},
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(testResponse)
			},
			nil,
			assert.NoError,
		},
		{
			"400 Bad Request",
			args{"dQw4w9WgXcQ"},
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`["No valid categories provided."]`))
			},
			nil,
			assert.Error,
		},
		{
			"404 Not Found",
			args{"y8Kyi0WNg40"},
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("Not Found"))
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
