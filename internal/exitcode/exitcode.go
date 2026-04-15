package exitcode

const (
	Success      = 0
	GeneralError = 1
	UsageError   = 2
	AuthError    = 3
	APIError     = 4
	ConfigError  = 5
)

// FromHTTPStatus maps an HTTP status code to a CLI exit code.
// 4xx errors map to 10 + (status - 400), e.g. 401 -> 11, 404 -> 14, 429 -> 39.
// 5xx errors map to 60 + (status - 500), e.g. 500 -> 60, 503 -> 63.
// Other errors return APIError (4).
func FromHTTPStatus(status int) int {
	switch {
	case status >= 400 && status < 500:
		return 10 + (status - 400)
	case status >= 500 && status < 600:
		return 60 + (status - 500)
	default:
		return APIError
	}
}
