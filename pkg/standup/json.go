package standup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// JSONInput represents the structure for JSON input
type JSONInput struct {
	Yesterday []string `json:"yesterday"`
	Today     []string `json:"today"`
	Blockers  string   `json:"blockers"`
}

// JSONOutput represents the structure for JSON output
type JSONOutput struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
	Error     string    `json:"error,omitempty"`
	Date      string    `json:"date"`
	User      string    `json:"user,omitempty"`
	Yesterday []string  `json:"yesterday,omitempty"`
	Today     []string  `json:"today,omitempty"`
	Blockers  string    `json:"blockers,omitempty"`
	FilePath  string    `json:"file_path,omitempty"`
	CommitSHA string    `json:"commit_sha,omitempty"`
	PRNumber  string    `json:"pr_number,omitempty"`
	PRUrl     string    `json:"pr_url,omitempty"`
}


// CommitInfo represents information about a commit
type CommitInfo struct {
	SHA     string `json:"sha"`
	Date    string `json:"date"`
	Author  string `json:"author"`
	Message string `json:"message"`
}

// ParseJSONInput parses JSON input into a standup entry
// The jsonStr can be:
// - Direct JSON string
// - File path to a JSON file
// - "-" to read from stdin
func ParseJSONInput(jsonStr string) (*Entry, error) {
	var jsonData []byte
	var err error

	if jsonStr == "-" {
		// Read from stdin
		jsonData, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read from stdin: %w", err)
		}
	} else if strings.HasPrefix(jsonStr, "{") || strings.HasPrefix(jsonStr, "[") {
		// Direct JSON string
		jsonData = []byte(jsonStr)
	} else {
		// Assume it's a file path
		jsonData, err = os.ReadFile(jsonStr)
		if err != nil {
			// If file read fails, try to parse as direct JSON anyway
			jsonData = []byte(jsonStr)
		}
	}

	var input JSONInput
	if err := json.Unmarshal(jsonData, &input); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}

	// Validate input
	if len(input.Yesterday) == 0 && len(input.Today) == 0 {
		return nil, fmt.Errorf("at least one of 'yesterday' or 'today' must have entries")
	}

	// Set default blockers if empty
	if input.Blockers == "" {
		input.Blockers = "None"
	}

	return &Entry{
		Date:      time.Now(),
		Yesterday: input.Yesterday,
		Today:     input.Today,
		Blockers:  input.Blockers,
	}, nil
}

// FormatJSONOutput formats the output as JSON
func FormatJSONOutput(output JSONOutput) (string, error) {
	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON output: %w", err)
	}
	return string(jsonBytes), nil
}

