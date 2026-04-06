package errors

// Domain-level error codes used by the service layer. The API layer maps
// them to HTTP status codes via StatusFromError. These constants have the
// same numeric values as their HTTP counterparts but avoid importing net/http
// so the service package stays transport-agnostic.
const (
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeConflict     = 409
	CodeGone         = 410
	CodeInternal     = 500
)

// ServiceError is a domain error that carries a status code for the API layer.
type ServiceError struct {
	Code    int
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

func ErrBadRequest(msg string) *ServiceError {
	return &ServiceError{Code: CodeBadRequest, Message: msg}
}

func ErrUnauthorized(msg string) *ServiceError {
	return &ServiceError{Code: CodeUnauthorized, Message: msg}
}

func ErrForbidden(msg string) *ServiceError {
	return &ServiceError{Code: CodeForbidden, Message: msg}
}

func ErrNotFound(msg string) *ServiceError {
	return &ServiceError{Code: CodeNotFound, Message: msg}
}

func ErrConflict(msg string) *ServiceError {
	return &ServiceError{Code: CodeConflict, Message: msg}
}

func ErrGone(msg string) *ServiceError {
	return &ServiceError{Code: CodeGone, Message: msg}
}

func ErrInternal(msg string) *ServiceError {
	return &ServiceError{Code: CodeInternal, Message: msg}
}

// StatusFromError extracts the HTTP status code from a ServiceError,
// defaulting to 500 for unknown error types.
func StatusFromError(err error) int {
	if se, ok := err.(*ServiceError); ok {
		return se.Code
	}
	return CodeInternal
}

// ErrorCode derives a stable snake_case API error code from the error message.
// For example "invalid credentials" → "invalid_credentials".
// Falls back to the generic HTTP status code name if the message is not a
// recognized pattern.
func ErrorCode(err error) string {
	se, ok := err.(*ServiceError)
	if !ok {
		return "internal_error"
	}
	// Convert common messages to stable codes
	msg := se.Message
	switch {
	case msg == "invalid credentials":
		return "invalid_credentials"
	case msg == "account banned":
		return "account_banned"
	case msg == "email not verified":
		return "email_not_verified"
	default:
		return statusToCode(se.Code)
	}
}

func statusToCode(code int) string {
	switch code {
	case CodeBadRequest:
		return "bad_request"
	case CodeUnauthorized:
		return "unauthorized"
	case CodeForbidden:
		return "forbidden"
	case CodeNotFound:
		return "not_found"
	case CodeConflict:
		return "conflict"
	case CodeGone:
		return "gone"
	default:
		return "internal_error"
	}
}
