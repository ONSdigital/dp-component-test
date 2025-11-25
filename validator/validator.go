package validator

import (
	"net/url"
	"time"

	"github.com/google/uuid"
)

func ValidateTimestamp(value string) bool {
	_, err := time.Parse(time.RFC3339, value)
	return err == nil
}

func ValidateRecentTimestamp(value string) bool {
	parsedTime, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return false
	}

	// Ensure the timestamp is within 10 seconds of now
	timestampAge := time.Since(parsedTime)
	if timestampAge < 0 || timestampAge > 10*time.Second {
		return false
	}

	return true
}

func ValidateUUID(value string) bool {
	err := uuid.Validate(value)
	return err == nil
}

func ValidateURL(value string) bool {
	parsedURL, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}

	if parsedURL.Scheme == "" {
		return false
	}

	if parsedURL.Host == "" {
		return false
	}
	return true
}

func ValidateURIPath(value string) bool {
	parsedURL, err := url.Parse(value)
	if err != nil {
		return false
	}

	if parsedURL.Path == "" {
		return false
	}

	return true
}
