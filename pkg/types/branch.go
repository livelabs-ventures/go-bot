package types

import (
	"fmt"
	"regexp"
)

// BranchName represents a git branch name with validation
type BranchName string

// branchNameRegex defines valid branch name pattern
var branchNameRegex = regexp.MustCompile(`^[a-zA-Z0-9/_-]+$`)

// NewBranchName creates a new validated branch name
func NewBranchName(name string) (BranchName, error) {
	if name == "" {
		return "", fmt.Errorf("branch name cannot be empty")
	}

	if !branchNameRegex.MatchString(name) {
		return "", fmt.Errorf("invalid branch name: %s (must contain only alphanumeric, /, _, - characters)", name)
	}

	// Check for consecutive slashes
	if regexp.MustCompile(`//+`).MatchString(name) {
		return "", fmt.Errorf("branch name cannot contain consecutive slashes")
	}

	// Check for leading/trailing slashes
	if name[0] == '/' || name[len(name)-1] == '/' {
		return "", fmt.Errorf("branch name cannot start or end with a slash")
	}

	return BranchName(name), nil
}

// String returns the branch name as a string
func (b BranchName) String() string {
	return string(b)
}

// StandupBranchName creates a branch name for a standup on a specific date
func StandupBranchName(dateStr string) BranchName {
	// This is a safe operation since we control the format
	return BranchName(fmt.Sprintf("standup/%s", dateStr))
}