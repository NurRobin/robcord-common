package httputil

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// DecodeJSON limits the request body to maxSize bytes, then JSON-decodes
// it into v. On failure it writes a 400 error response and returns false.
func DecodeJSON(w http.ResponseWriter, r *http.Request, maxSize int64, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

// DecodeJSONOptional is like DecodeJSON but tolerates an empty body
// (io.EOF). Returns true when v was populated or the body was empty,
// false (with a 400 response written) on a real parse error.
func DecodeJSONOptional(w http.ResponseWriter, r *http.Request, maxSize int64, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		if errors.Is(err, io.EOF) {
			return true
		}
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

// DecodeJSONStrict is like DecodeJSON but rejects unknown fields.
func DecodeJSONStrict(w http.ResponseWriter, r *http.Request, maxSize int64, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

// DecodeJSONStrictOptional is like DecodeJSONStrict but tolerates an
// empty body (io.EOF).
func DecodeJSONStrictOptional(w http.ResponseWriter, r *http.Request, maxSize int64, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		if errors.Is(err, io.EOF) {
			return true
		}
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}
