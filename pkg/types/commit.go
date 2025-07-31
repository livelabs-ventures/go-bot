package types

import (
	"fmt"
	"strings"
)

// CommitMessage represents a git commit message with validation
type CommitMessage struct {
	Title string
	Body  string
}

// NewCommitMessage creates a new commit message with validation
func NewCommitMessage(title, body string) (CommitMessage, error) {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)

	if title == "" {
		return CommitMessage{}, fmt.Errorf("commit title cannot be empty")
	}

	if len(title) > 72 {
		return CommitMessage{}, fmt.Errorf("commit title too long (max 72 characters, got %d)", len(title))
	}

	// Check for conventional commit format if it looks like one
	if strings.Contains(title, ":") {
		parts := strings.SplitN(title, ":", 2)
		if len(parts[0]) == 0 {
			return CommitMessage{}, fmt.Errorf("invalid commit format: missing type before colon")
		}
		if len(parts) == 2 && len(strings.TrimSpace(parts[1])) == 0 {
			return CommitMessage{}, fmt.Errorf("invalid commit format: missing description after colon")
		}
	}

	return CommitMessage{
		Title: title,
		Body:  body,
	}, nil
}

// String returns the full commit message
func (c CommitMessage) String() string {
	if c.Body == "" {
		return c.Title
	}
	return fmt.Sprintf("%s\n\n%s", c.Title, c.Body)
}

// StandupCommitMessage creates a commit message for a standup entry
func StandupCommitMessage(userName UserName, dateStr string, summary string) CommitMessage {
	title := fmt.Sprintf("[Standup] %s - %s", userName, dateStr)
	return CommitMessage{
		Title: title,
		Body:  summary,
	}
}