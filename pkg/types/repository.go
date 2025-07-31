package types

import (
	"fmt"
	"strings"
)

// Repository represents a GitHub repository
type Repository struct {
	Owner string
	Name  string
}

// NewRepository creates a new Repository from a string in "owner/name" format
func NewRepository(fullName string) (Repository, error) {
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 {
		return Repository{}, fmt.Errorf("invalid repository format: expected 'owner/name', got '%s'", fullName)
	}

	if parts[0] == "" || parts[1] == "" {
		return Repository{}, fmt.Errorf("repository owner and name cannot be empty")
	}

	return Repository{
		Owner: parts[0],
		Name:  parts[1],
	}, nil
}

// String returns the repository in "owner/name" format
func (r Repository) String() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

// Validate checks if the repository is valid
func (r Repository) Validate() error {
	if r.Owner == "" {
		return fmt.Errorf("repository owner cannot be empty")
	}
	if r.Name == "" {
		return fmt.Errorf("repository name cannot be empty")
	}
	return nil
}