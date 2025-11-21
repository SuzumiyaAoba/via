package cli

import (
	"net/url"
	"os"
)

// File and URL detection helpers

// isURL checks if the given string is a valid URL with a scheme
func isURL(filename string) bool {
	u, err := url.Parse(filename)
	return err == nil && u.Scheme != ""
}

// fileExists checks if a file exists at the given path
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// isFileOrURL checks if the filename is either a valid URL or an existing file
func isFileOrURL(filename string) bool {
	return isURL(filename) || fileExists(filename)
}
