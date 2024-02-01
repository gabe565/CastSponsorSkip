package youtube

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabe565/castsponsorskip/internal/config"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func TestQueryVideoId(t *testing.T) {
	type args struct {
		ctx    context.Context
		artist string
		title  string
	}
	tests := []struct {
		name      string
		args      args
		found     bool
		want      string
		wantQuery string
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			"simple",
			args{artist: "Rick Astley", title: "Rick Astley - Never Gonna Give You Up (Official Music Video)"},
			true,
			"dQw4w9WgXcQ",
			`"Rick Astley"+intitle:"Rick Astley - Never Gonna Give You Up (Official Music Video)"`,
			assert.NoError,
		},
		{
			"not found",
			args{artist: "gabe565", title: "Nonexistent video"},
			false,
			"",
			`"gabe565"+intitle:"Nonexistent video"`,
			assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.wantQuery, r.URL.Query().Get("q"))

				response := &youtube.SearchListResponse{}
				if tt.found {
					response.Items = []*youtube.SearchResult{{
						Id: &youtube.ResourceId{VideoId: tt.want},
						Snippet: &youtube.SearchResultSnippet{
							ChannelTitle: tt.args.artist,
							Title:        tt.args.title,
						},
					}}
				}

				b, err := json.Marshal(response)
				if !assert.NoError(t, err) {
					return
				}

				_, _ = w.Write(b)
			}))
			defer server.Close()

			defer func(key string) {
				config.Default.YouTubeAPIKey = key
			}(config.Default.YouTubeAPIKey)
			config.Default.YouTubeAPIKey = "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe"
			if err := CreateService(context.Background(), option.WithEndpoint(server.URL)); !assert.NoError(t, err) {
				return
			}
			defer func() {
				service = nil
			}()

			got, err := QueryVideoId(tt.args.ctx, tt.args.artist, tt.args.title)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
