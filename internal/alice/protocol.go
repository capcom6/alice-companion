// Package alice contains the Alice Dialogs protocol DTOs and the greeting and
// echo logic served by the HTTP webhook.
package alice

import "unicode/utf8"

const (
	// ProtocolVersion is the only supported Dialogs protocol version.
	ProtocolVersion = "1.0"

	// MaxTextLength caps response.text in characters.
	MaxTextLength = 1024
	// MaxTTSLength caps response.tts in characters.
	MaxTTSLength = 1024
	// MaxButtonTitleLength caps buttons[].title in characters.
	MaxButtonTitleLength = 64
	// MaxButtonURLLength caps buttons[].url in bytes.
	MaxButtonURLLength = 1024
	// MaxButtonPayloadLength caps buttons[].payload in bytes.
	MaxButtonPayloadLength = 4096
	// MaxSessionIDLength caps session.session_id in characters.
	MaxSessionIDLength = 64

	maxMessageID int64 = 99_999_999
)

// WebhookRequest is the Alice Dialogs request body. A top-level user_id is
// deliberately absent: the only supported identity is Session.User.UserID.
type WebhookRequest struct {
	Meta    Meta      `json:"meta"`
	Session Session   `json:"session"`
	Request Utterance `json:"request"`
	Version string    `json:"version"`
}

// Meta carries request metadata used for correlation and diagnostics.
type Meta struct {
	Locale   string `json:"locale"`
	Timezone string `json:"timezone"`
	ClientID string `json:"client_id"`
}

// Session is the Dialogs session block. New marks the first turn of a session.
type Session struct {
	MessageID   int64        `json:"message_id"`
	SessionID   string       `json:"session_id"`
	SkillID     string       `json:"skill_id"`
	User        *User        `json:"user,omitempty"`
	Application *Application `json:"application,omitempty"`
	New         bool         `json:"new"`
}

// User is the authenticated Dialogs user. An absent User means the turn is
// anonymous.
type User struct {
	UserID string `json:"user_id"`
}

// Application is the client application that sent the request. It is suitable
// for correlation only and is never used as an identity key.
type Application struct {
	ApplicationID string `json:"application_id"`
}

// Utterance is the user's request text.
type Utterance struct {
	Command           string `json:"command"`
	OriginalUtterance string `json:"original_utterance"`
	Type              string `json:"type"`
}

// Response is the Alice Dialogs response body.
type Response struct {
	Version  string          `json:"version"`
	Session  ResponseSession `json:"session"`
	Response ResponseBody    `json:"response"`
}

// ResponseSession echoes the request session identifiers.
type ResponseSession struct {
	MessageID int64  `json:"message_id"`
	SessionID string `json:"session_id"`
}

// ResponseBody is the spoken answer.
type ResponseBody struct {
	Text       string   `json:"text,omitempty"`
	TTS        string   `json:"tts,omitempty"`
	EndSession bool     `json:"end_session"`
	Buttons    []Button `json:"buttons,omitempty"`
}

// Button is a response button.
type Button struct {
	Title   string `json:"title"`
	URL     string `json:"url,omitempty"`
	Payload string `json:"payload,omitempty"`
}

// Normalize enforces the protocol version and every documented size cap.
func (r *Response) Normalize() {
	r.Version = ProtocolVersion
	r.Session.SessionID = truncateRunes(r.Session.SessionID, MaxSessionIDLength)
	r.Session.MessageID = clampMessageID(r.Session.MessageID)
	r.Response.Text = truncateRunes(r.Response.Text, MaxTextLength)
	r.Response.TTS = truncateRunes(r.Response.TTS, MaxTTSLength)

	for i := range r.Response.Buttons {
		button := &r.Response.Buttons[i]
		button.Title = truncateRunes(button.Title, MaxButtonTitleLength)
		button.URL = truncateBytes(button.URL, MaxButtonURLLength)
		button.Payload = truncateBytes(button.Payload, MaxButtonPayloadLength)
	}
}

func clampMessageID(id int64) int64 {
	if id < 0 {
		return 0
	}
	if id > maxMessageID {
		return maxMessageID
	}

	return id
}

func truncateRunes(s string, limit int) string {
	if utf8.RuneCountInString(s) <= limit {
		return s
	}

	return string([]rune(s)[:limit])
}

func truncateBytes(s string, limit int) string {
	if len(s) <= limit {
		return s
	}

	cut := limit
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}

	return s[:cut]
}
