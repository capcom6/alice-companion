package alice

import "strings"

const (
	greetingText       = "Hello! I am your Alice companion. How can I help you?"
	echoPrefix         = "You said: "
	emptyUtteranceText = "I did not catch that. Please say it again."
)

// Reply builds the protocol response for a webhook request. The session user is
// deliberately not inspected: anonymous turns get only a greeting or an echo,
// never memory or IoT data.
func Reply(req *WebhookRequest) Response {
	text := greetingText
	if !req.Session.New {
		text = echoText(req.Request.Command, req.Request.OriginalUtterance)
	}

	resp := Response{
		Version: ProtocolVersion,
		Session: ResponseSession{
			MessageID: req.Session.MessageID,
			SessionID: req.Session.SessionID,
		},
		Response: ResponseBody{
			Text:       text,
			TTS:        ttsSafe(text),
			EndSession: false,
			Buttons:    nil,
		},
	}
	resp.Normalize()

	return resp
}

func echoText(command, original string) string {
	phrase := command
	if phrase == "" {
		phrase = original
	}
	if phrase == "" {
		return emptyUtteranceText
	}

	return echoPrefix + phrase
}

// ttsSafe removes characters that would open TTS markup while keeping the
// echoed user phrase otherwise intact.
func ttsSafe(text string) string {
	text = strings.ReplaceAll(text, "<", "")
	return strings.ReplaceAll(text, ">", "")
}
