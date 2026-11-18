package alice_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/capcom6/alice-companion/internal/alice"
	"github.com/stretchr/testify/assert"
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
					Payload: strings.Repeat("p", 5000),
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
	assert.Len(t, resp.Response.Buttons[0].Payload, alice.MaxButtonPayloadLength)
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
