package sponsorblock

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"gabe565.com/castsponsorskip/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuerySegmentsRequest(t *testing.T) {
	type args struct {
		id          string
		categories  []string
		actionTypes []string
	}
	tests := []struct {
		name     string
		args     args
		wantPath string
		wantErr  require.ErrorAssertionFunc
	}{
		{
			"1",
			args{"dQw4w9WgXcQ", []string{"sponsor"}, []string{"skip", "mute"}},
			"/api/skipSegments/5f6b",
			require.NoError,
		},
		{
			"2",
			args{"y8Kyi0WNg40", []string{"sponsor", "selfpromo"}, []string{"skip", "mute"}},
			"/api/skipSegments/30cc",
			require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := config.New()
			conf.Categories = tt.args.categories
			conf.ActionTypes = tt.args.actionTypes

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.wantPath, r.URL.Path)
				assert.Equal(t, tt.args.categories, r.URL.Query()["category"])
				assert.Equal(t, tt.args.actionTypes, r.URL.Query()["actionType"])
				_, _ = w.Write([]byte("[]"))
			}))
			t.Cleanup(server.Close)

			tempURL, err := url.Parse(server.URL)
			require.NoError(t, err)
			defaultURL := baseURL
			t.Cleanup(func() {
				baseURL = defaultURL
			})
			baseURL = *tempURL

			_, err = QuerySegments(t.Context(), conf, tt.args.id)
			tt.wantErr(t, err)
		})
	}
}

func TestQuerySegmentsResponse(t *testing.T) {
	testResponse := []byte(`[{"videoID":"fy9jO8JHaPo","segments":[{"category":"sponsor","actionType":"skip","segment":[53.433,57.705],"UUID":"e992e1c6dcebe5f21fc5dc68cfec12bc58cf7a68161983e74b56c89f1ac1d2c87","videoDuration":1399.461,"locked":0,"votes":0,"description":""},{"category":"sponsor","actionType":"skip","segment":[388.815,421.601],"UUID":"6c1c415479595a922bb1c67e4091bd804329be60a2013e17a421bc43137ae47b7","videoDuration":1399.461,"locked":0,"votes":-1,"description":""},{"category":"sponsor","actionType":"skip","segment":[927.803,997.758],"UUID":"3230cb569b64c51076d47c8000f7671e6231691667f0b722790d19be1ea578387","videoDuration":1399.461,"locked":0,"votes":0,"description":""}]}]`)

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		handler http.HandlerFunc
		want    []Segment
		wantErr require.ErrorAssertionFunc
	}{
		{
			"200 OK video in response",
			args{"fy9jO8JHaPo"},
			func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write(testResponse)
			},
			[]Segment{
				{Segment: [2]float32{53.433, 57.705}, UUID: "e992e1c6dcebe5f21fc5dc68cfec12bc58cf7a68161983e74b56c89f1ac1d2c87", Category: "sponsor", VideoDuration: 1399.461, ActionType: "skip", Locked: 0, Votes: 0, Description: ""},
				{Segment: [2]float32{388.815, 421.601}, UUID: "6c1c415479595a922bb1c67e4091bd804329be60a2013e17a421bc43137ae47b7", Category: "sponsor", VideoDuration: 1399.461, ActionType: "skip", Locked: 0, Votes: -1, Description: ""},
				{Segment: [2]float32{927.803, 997.758}, UUID: "3230cb569b64c51076d47c8000f7671e6231691667f0b722790d19be1ea578387", Category: "sponsor", VideoDuration: 1399.461, ActionType: "skip", Locked: 0, Votes: 0, Description: ""},
			},
			require.NoError,
		},
		{
			"200 OK video not in response",
			args{"dQw4w9WgXcQ"},
			func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write(testResponse)
			},
			nil,
			require.NoError,
		},
		{
			"400 Bad Request",
			args{"dQw4w9WgXcQ"},
			func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`["No valid categories provided."]`))
			},
			nil,
			require.Error,
		},
		{
			"404 Not Found",
			args{"y8Kyi0WNg40"},
			func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("Not Found"))
			},
			nil,
			require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			t.Cleanup(server.Close)

			tempURL, err := url.Parse(server.URL)
			require.NoError(t, err)
			defaultURL := baseURL
			t.Cleanup(func() {
				baseURL = defaultURL
			})
			baseURL = *tempURL

			got, err := QuerySegments(t.Context(), config.New(), tt.args.id)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
