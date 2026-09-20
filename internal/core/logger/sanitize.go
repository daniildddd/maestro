package logger

import "strings"

// Sanitize strips line breaks so user input cannot forge log entries (CWE-117).
func Sanitize(value string) string {
	value = strings.ReplaceAll(value, "\r", "")

	return strings.ReplaceAll(value, "\n", "")
}
