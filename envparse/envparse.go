package envparse

import (
	"os"
	"strconv"
	"time"
)

// String returns the value of the environment variable named by key,
// or fallback if the variable is empty or unset.
func String(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Int returns the integer value of the environment variable named by key,
// or fallback if the variable is empty, unset, or not a valid integer.
func Int(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// Bool returns the boolean value of the environment variable named by key,
// or fallback if the variable is empty, unset, or not a valid boolean.
func Bool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

// Duration returns the duration value of the environment variable named by key,
// or fallback if the variable is empty, unset, or not a valid duration string.
func Duration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
