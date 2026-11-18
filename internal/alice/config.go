package alice

// Config holds the Alice webhook configuration.
type Config struct {
	// SharedSecret is the secret required in the X-Alice-Secret header. It is
	// sourced from the ALICE__SHARED_SECRET environment variable only, and an
	// empty value rejects every request.
	SharedSecret string
}
