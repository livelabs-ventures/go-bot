package git

import (
	"fmt"
	"strings"
	"time"
)

// CommitDetails contains detailed information about a commit
type CommitDetails struct {
	SHA       string
	Author    string
	Date      time.Time
	Message   string
	Subject   string
	Body      string
}

// GetRecentCommits retrieves commits from the last working day
func (c *Client) GetRecentCommits(repoPath string, since time.Time) ([]CommitDetails, error) {
	// Format date for git log
	sinceStr := since.Format("2006-01-02")
	
	// Get commit log with format: SHA|Author|Date|Subject|Body
	output, err := c.runner.RunInDir(repoPath, "git", "log", 
		"--since="+sinceStr,
		"--pretty=format:%H|%an|%aI|%s|%b",
		"--no-merges")
	
	// Check output string regardless of error
	outputStr := string(output)
	
	if err != nil {
		// If no commits found, return empty slice
		lowerOutput := strings.ToLower(outputStr)
		if strings.Contains(lowerOutput, "does not have any commits") || 
		   strings.Contains(lowerOutput, "bad revision") ||
		   strings.Contains(lowerOutput, "unknown revision") ||
		   strings.Contains(lowerOutput, "fatal:") ||
		   strings.Contains(lowerOutput, "not a git repository") {
			return []CommitDetails{}, nil
		}
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	// Parse commits
	commits := []CommitDetails{}
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 4 {
			continue
		}
		
		// Parse date
		date, err := time.Parse(time.RFC3339, parts[2])
		if err != nil {
			continue
		}
		
		// Extract body if available
		body := ""
		if len(parts) >= 5 {
			body = strings.TrimSpace(parts[4])
		}
		
		commit := CommitDetails{
			SHA:     parts[0],
			Author:  parts[1],
			Date:    date,
			Subject: parts[3],
			Body:    body,
			Message: parts[3],
		}
		
		if body != "" {
			commit.Message = commit.Subject + "\n\n" + body
		}
		
		commits = append(commits, commit)
	}
	
	return commits, nil
}

// GetLastWorkingDay calculates the last working day (excludes weekends)
func GetLastWorkingDay(from time.Time) time.Time {
	// Get current day of week
	weekday := from.Weekday()
	
	switch weekday {
	case time.Monday:
		// Last working day was Friday (3 days ago)
		return from.AddDate(0, 0, -3)
	case time.Sunday:
		// Last working day was Friday (2 days ago)
		return from.AddDate(0, 0, -2)
	default:
		// Last working day was yesterday
		return from.AddDate(0, 0, -1)
	}
}

// ExtractWorkItems extracts meaningful work items from commits
func ExtractWorkItems(commits []CommitDetails) []string {
	items := []string{}
	
	for _, commit := range commits {
		// Skip commits with certain patterns
		if isSkippableCommit(commit.Subject) {
			continue
		}
		
		// Extract meaningful item from commit
		item := formatCommitAsWorkItem(commit)
		if item != "" {
			items = append(items, item)
		}
	}
	
	return deduplicateItems(items)
}

// isSkippableCommit checks if a commit should be skipped
func isSkippableCommit(subject string) bool {
	lowered := strings.ToLower(subject)
	skipPatterns := []string{
		"merge",
		"wip",
		"tmp",
		"temp",
		"fix typo",
		"update readme",
		"initial commit",
	}
	
	for _, pattern := range skipPatterns {
		if strings.Contains(lowered, pattern) {
			return true
		}
	}
	
	return false
}

// formatCommitAsWorkItem formats a commit into a work item
func formatCommitAsWorkItem(commit CommitDetails) string {
	subject := commit.Subject
	
	// Remove common prefixes
	prefixes := []string{
		"feat:",
		"fix:",
		"docs:",
		"style:",
		"refactor:",
		"test:",
		"chore:",
		"perf:",
		"ci:",
		"build:",
		"revert:",
	}
	
	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(subject), prefix) {
			subject = strings.TrimSpace(subject[len(prefix):])
			break
		}
	}
	
	// Capitalize first letter
	if len(subject) > 0 {
		subject = strings.ToUpper(subject[:1]) + subject[1:]
	}
	
	// Remove trailing period if present
	subject = strings.TrimSuffix(subject, ".")
	
	return subject
}

// deduplicateItems removes duplicate items while preserving order
func deduplicateItems(items []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	
	for _, item := range items {
		normalized := strings.ToLower(strings.TrimSpace(item))
		if !seen[normalized] {
			seen[normalized] = true
			result = append(result, item)
		}
	}
	
	return result
}