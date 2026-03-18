package httputil

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// writeDecodeError writes a 413 if the body exceeded the size limit,
// or a 400 for any other JSON parse error.
func writeDecodeError(w http.ResponseWriter, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		WriteError(w, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}
	WriteError(w, http.StatusBadRequest, "invalid request body")
}

// DecodeJSON limits the request body to maxSize bytes, then JSON-decodes
// it into v. On failure it writes a 400 (or 413 if oversized) error
// response and returns false.
func DecodeJSON(w http.ResponseWriter, r *http.Request, maxSize int64, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeDecodeError(w, err)
		return false
	}
	return true
}

// DecodeJSONOptional is like DecodeJSON but tolerates an empty body
// (io.EOF). Returns true when v was populated or the body was empty,
// false (with an error response written) on a real parse error.
func DecodeJSONOptional(w http.ResponseWriter, r *http.Request, maxSize int64, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		if errors.Is(err, io.EOF) {
			return true
		}
		writeDecodeError(w, err)
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
		writeDecodeError(w, err)
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
		writeDecodeError(w, err)
		return false
	}
	return true
}
