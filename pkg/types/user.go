package types

import (
	"fmt"
	"regexp"
	"strings"
)

// UserName represents a validated user name
type UserName string

// userNameRegex defines valid user name pattern
var userNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-'.]+$`)

// NewUserName creates a new validated user name
func NewUserName(name string) (UserName, error) {
	name = strings.TrimSpace(name)
	
	if name == "" {
		return "", fmt.Errorf("user name cannot be empty")
	}

	if len(name) > 100 {
		return "", fmt.Errorf("user name too long (max 100 characters)")
	}

	if !userNameRegex.MatchString(name) {
		return "", fmt.Errorf("user name contains invalid characters (only letters, numbers, spaces, hyphens, apostrophes, and periods allowed)")
	}

	return UserName(name), nil
}

// String returns the user name as a string
func (u UserName) String() string {
	return string(u)
}

// FileName returns a sanitized version suitable for file names
func (u UserName) FileName() string {
	// Convert to lowercase and replace spaces with hyphens
	fileName := strings.ToLower(string(u))
	fileName = strings.ReplaceAll(fileName, " ", "-")
	fileName = strings.ReplaceAll(fileName, "'", "")
	fileName = strings.ReplaceAll(fileName, ".", "")
	
	// Remove any double hyphens that might have been created
	for strings.Contains(fileName, "--") {
		fileName = strings.ReplaceAll(fileName, "--", "-")
	}
	
	return fileName
}