package alice_test

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/capcom6/alice-companion/internal/alice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseNormalizeCapsFields(t *testing.T) {
	t.Parallel()

	resp := alice.Response{
		Version: "0.1",
		Session: alice.ResponseSession{
			MessageID: 123456789012,
			SessionID: strings.Repeat("s", 100),
		},
		Response: alice.ResponseBody{
			Text:       strings.Repeat("t", 2000),
			TTS:        strings.Repeat("u", 2000),
			EndSession: false,
			Buttons: []alice.Button{
				{
					Title:   strings.Repeat("b", 100),
					URL:     strings.Repeat("я", 600),
					Payload: json.RawMessage(`{"action":"next"}`),
				},
			},
		},
	}

	resp.Normalize()

	assert.Equal(t, alice.ProtocolVersion, resp.Version)
	assert.Len(t, []rune(resp.Session.SessionID), alice.MaxSessionIDLength)
	assert.LessOrEqual(t, resp.Session.MessageID, int64(99_999_999))
	assert.Len(t, []rune(resp.Response.Text), alice.MaxTextLength)
	assert.Len(t, []rune(resp.Response.TTS), alice.MaxTTSLength)
	assert.Len(t, []rune(resp.Response.Buttons[0].Title), alice.MaxButtonTitleLength)
	assert.Len(t, resp.Response.Buttons[0].URL, alice.MaxButtonURLLength)
	assert.True(t, utf8.ValidString(resp.Response.Buttons[0].URL))
	assert.JSONEq(t, `{"action":"next"}`, string(resp.Response.Buttons[0].Payload))
}

func TestResponseNormalizeKeepsShortValues(t *testing.T) {
	t.Parallel()

	resp := alice.Response{
		Version: alice.ProtocolVersion,
		Session: alice.ResponseSession{
			MessageID: 42,
			SessionID: "session-1",
		},
		Response: alice.ResponseBody{
			Text:       "hello",
			TTS:        "hello",
			EndSession: false,
			Buttons:    nil,
		},
	}

	resp.Normalize()

	assert.Equal(t, int64(42), resp.Session.MessageID)
	assert.Equal(t, "session-1", resp.Session.SessionID)
	assert.Equal(t, "hello", resp.Response.Text)
	assert.Equal(t, "hello", resp.Response.TTS)
}

func TestResponseBodySerializesEmptyText(t *testing.T) {
	t.Parallel()

	resp := alice.Response{
		Version: alice.ProtocolVersion,
		Session: alice.ResponseSession{
			MessageID: 1,
			SessionID: "session-1",
		},
		Response: alice.ResponseBody{
			Text:       "",
			TTS:        "hello",
			EndSession: false,
			Buttons:    nil,
		},
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"text":""`)
}

func TestButtonPayloadSerializesAsJSONObject(t *testing.T) {
	t.Parallel()

	resp := alice.Response{
		Version: alice.ProtocolVersion,
		Session: alice.ResponseSession{
			MessageID: 1,
			SessionID: "session-1",
		},
		Response: alice.ResponseBody{
			Text:       "hello",
			TTS:        "hello",
			EndSession: false,
			Buttons: []alice.Button{
				{
					Title:   "next",
					URL:     "",
					Payload: json.RawMessage(`{"action":"next"}`),
				},
			},
		},
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"payload":{"action":"next"}`)
	assert.NotContains(t, string(data), `"payload":"`)
}
