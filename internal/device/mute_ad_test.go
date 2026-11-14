package device

import (
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"gabe565.com/castsponsorskip/internal/config"
	"github.com/buger/jsonparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vishen/go-chromecast/application"
	"github.com/vishen/go-chromecast/cast"
	mockCast "github.com/vishen/go-chromecast/cast/mocks"
	pb "github.com/vishen/go-chromecast/cast/proto"
)

func castMessage(t *testing.T, payload any) *pb.CastMessage {
	t.Helper()
	b, err := json.Marshal(payload)
	require.NoError(t, err)
	s := string(b)
	protocolVersion := pb.CastMessage_CASTV2_1_0
	payloadType := pb.CastMessage_STRING
	return &pb.CastMessage{ProtocolVersion: &protocolVersion, PayloadType: &payloadType, PayloadUtf8: &s}
}

// TestMuteAdUnmutesWhenSessionCloses covers the YouTube receiver closing the
// session while an ad is being skipped. Depending on the go-chromecast
// version, Skipad either panics or returns an error; either way the device
// must not be left muted.
func TestMuteAdUnmutesWhenSessionCloses(t *testing.T) {
	recvChan := make(chan *pb.CastMessage, 5)
	closeHandled := make(chan struct{})
	var mutedValues []bool

	conn := &mockCast.Conn{}
	conn.On("MsgChan").Return(recvChan)
	conn.On("Start", mock.Anything, mock.Anything).Return(nil)
	conn.On("Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			requestID := args.Int(0)
			payload, err := json.Marshal(args.Get(1))
			require.NoError(t, err)
			payloadType, _ := jsonparser.GetString(payload, "type")
			namespace := args.String(4)

			switch {
			case payloadType == "GET_STATUS" && namespace == "urn:x-cast:com.google.cast.receiver":
				resp := cast.ReceiverStatusResponse{
					PayloadHeader: cast.PayloadHeader{Type: "RECEIVER_STATUS", RequestId: requestID},
				}
				resp.Status.Applications = []cast.Application{
					{AppId: "233637DE", DisplayName: "YouTube", TransportId: "transport-1"},
				}
				recvChan <- castMessage(t, resp)
			case payloadType == "GET_STATUS" && namespace == "urn:x-cast:com.google.cast.media":
				recvChan <- castMessage(t, cast.MediaStatusResponse{
					PayloadHeader: cast.PayloadHeader{Type: "MEDIA_STATUS", RequestId: requestID},
					Status: []cast.Media{{
						MediaSessionId: 1,
						PlayerState:    StatePlaying,
						CustomData:     cast.CustomData{PlayerState: StateAd},
					}},
				})
			case payloadType == "SET_VOLUME":
				muted, err := jsonparser.GetBoolean(payload, "volume", "muted")
				require.NoError(t, err)
				mutedValues = append(mutedValues, muted)
			case payloadType == "SKIP_AD":
				// The receiver goes away instead of skipping the ad. Wait until
				// the CLOSE has been processed so the test is deterministic.
				recvChan <- castMessage(t, cast.PayloadHeader{Type: "CLOSE"})
				select {
				case <-closeHandled:
				case <-time.After(5 * time.Second):
					t.Error("timed out waiting for CLOSE to be handled")
				}
			}
		}).Return(nil)

	app := application.NewApplication(
		application.WithConnection(conn),
		application.WithCacheDisabled(true),
		application.WithSkipadSleep(time.Millisecond),
		application.WithSkipadRetries(5),
	)
	app.AddMessageFunc(func(msg *pb.CastMessage) {
		if msgType, _ := jsonparser.GetString([]byte(msg.GetPayloadUtf8()), "type"); msgType == "CLOSE" {
			close(closeHandled)
		}
	})
	require.NoError(t, app.Start("127.0.0.1", 8009))

	_, _, castVol := app.Status()
	require.NotNil(t, castVol)

	d := &Device{
		config: &config.Config{MuteAds: true},
		app:    app,
		logger: slog.New(slog.DiscardHandler),
	}

	func() {
		// A panic from Skipad is still recovered by BeginTick, so swallow it
		// here and only check the mute state.
		defer func() { _ = recover() }()
		d.muteAd(castVol)
	}()

	assert.Equal(t, []bool{true, false}, mutedValues, "device should be muted, then unmuted")
}
