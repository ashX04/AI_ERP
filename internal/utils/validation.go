package utils

import (
	"net/url"
	"regexp"
)

var (
	// Regex for validating file IDs
	fileIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// ValidateFileID checks if the file ID is safe
func ValidateFileID(id string) bool {
	return fileIDRegex.MatchString(id) && len(id) <= 100
}

// SanitizeURL ensures the URL is safe
func SanitizeURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}
