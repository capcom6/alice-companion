package alice

// Config holds the Alice webhook configuration.
type Config struct {
	// SharedSecret is the secret required in the webhook URL's key query
	// parameter. It is sourced from ALICE__SHARED_SECRET (process environment
	// or .env) or alice.shared_secret in CONFIG_PATH YAML; an empty value
	// rejects every request.
	SharedSecret string
}
