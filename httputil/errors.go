package httputil

import (
	"log/slog"
	"net/http"
)

// ErrorResponse is the standard error format used across all robcord services.
// Both Zentrale and Workspace return this structure for all error responses.
type ErrorResponse struct {
	Error   string `json:"error"`   // stable snake_case code (e.g. "not_found")
	Message string `json:"message"` // human-readable description
}

// WriteErrorResponse writes a structured JSON error response with a machine-readable
// code and a human-readable message.
func WriteErrorResponse(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, ErrorResponse{Error: code, Message: message})
}

// WriteInternalError logs an internal error with context and writes a generic
// 500 response. This replaces the repeated pattern of slog.Error + writeError
// that appears ~100 times across the codebase.
func WriteInternalError(w http.ResponseWriter, context string, err error) {
	slog.Error("internal error", "context", context, "error", err)
	WriteErrorResponse(w, http.StatusInternalServerError, "internal_error", "internal error")
}

// IgnoreError logs an intentionally ignored error at debug level. Use this
// instead of silently discarding errors with _ = operation() to make the
// intent explicit and keep a trace in debug logs.
func IgnoreError(err error, reason string) {
	if err != nil {
		slog.Debug("ignored error", "reason", reason, "error", err)
	}
}
