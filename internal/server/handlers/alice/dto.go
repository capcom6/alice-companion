package alice

// ErrorResponse is the body returned when a request is rejected.
type ErrorResponse struct {
	Error string `json:"error"`
}
