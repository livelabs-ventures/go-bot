package git

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// MockCommandRunner for testing commit functionality
type MockCommitRunner struct {
	outputs map[string]string
	errors  map[string]error
}

func (m *MockCommitRunner) Run(name string, args ...string) ([]byte, error) {
	return m.RunInDir("", name, args...)
}

func (m *MockCommitRunner) RunInDir(dir, name string, args ...string) ([]byte, error) {
	key := fmt.Sprintf("%s:%s:%s", dir, name, strings.Join(args, " "))
	if err, ok := m.errors[key]; ok {
		return nil, err
	}
	if output, ok := m.outputs[key]; ok {
		return []byte(output), nil
	}
	return nil, fmt.Errorf("unexpected command: %s", key)
}

func TestGetRecentCommits(t *testing.T) {
	tests := []struct {
		name        string
		since       time.Time
		gitOutput   string
		gitError    error
		wantCommits int
		wantErr     bool
	}{
		{
			name:  "successful commit retrieval",
			since: time.Now().AddDate(0, 0, -1),
			gitOutput: `abc1234|John Doe|2025-07-30T10:00:00Z|feat: Add new feature|This adds a cool new feature
def5678|Jane Smith|2025-07-30T14:30:00Z|fix: Fix critical bug|
ghi9012|John Doe|2025-07-30T16:00:00Z|docs: Update README|Updated installation instructions`,
			wantCommits: 3,
			wantErr:     false,
		},
		{
			name:        "no commits found",
			since:       time.Now().AddDate(0, 0, -1),
			gitOutput:   "",
			wantCommits: 0,
			wantErr:     false,
		},
		{
			name:        "repository has no commits",
			since:       time.Now().AddDate(0, 0, -1),
			gitOutput:   "fatal: your current branch 'main' does not have any commits yet",
			gitError:    fmt.Errorf("exit status 128"),
			wantCommits: 0,
			wantErr:     false,
		},
		{
			name:        "git command error",
			since:       time.Now().AddDate(0, 0, -1),
			gitOutput:   "",
			gitError:    fmt.Errorf("git not found"),
			wantCommits: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &MockCommitRunner{
				outputs: make(map[string]string),
				errors:  make(map[string]error),
			}
			
			sinceStr := tt.since.Format("2006-01-02")
			key := fmt.Sprintf("test-repo:git:log --since=%s --pretty=format:%%H|%%an|%%aI|%%s|%%b --no-merges", sinceStr)
			
			if tt.gitError != nil {
				runner.errors[key] = tt.gitError
			}
			runner.outputs[key] = tt.gitOutput
			
			client := NewClientWithRunner(runner)
			commits, err := client.GetRecentCommits("test-repo", tt.since)
			
			// Debug output for failing test
			if tt.name == "repository has no commits" && err != nil {
				t.Logf("Debug: output = %q, error = %v", tt.gitOutput, err)
			}
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if len(commits) != tt.wantCommits {
				t.Errorf("expected %d commits, got %d", tt.wantCommits, len(commits))
			}
			
			// Verify commit details if we have commits
			if tt.wantCommits > 0 && len(commits) > 0 {
				// Check first commit
				if commits[0].SHA != "abc1234" {
					t.Errorf("expected first commit SHA 'abc1234', got '%s'", commits[0].SHA)
				}
				if commits[0].Author != "John Doe" {
					t.Errorf("expected first commit author 'John Doe', got '%s'", commits[0].Author)
				}
				if commits[0].Subject != "feat: Add new feature" {
					t.Errorf("expected first commit subject 'feat: Add new feature', got '%s'", commits[0].Subject)
				}
				if commits[0].Body != "This adds a cool new feature" {
					t.Errorf("expected first commit body 'This adds a cool new feature', got '%s'", commits[0].Body)
				}
			}
		})
	}
}

func TestGetLastWorkingDay(t *testing.T) {
	tests := []struct {
		name     string
		from     time.Time
		expected time.Time
	}{
		{
			name:     "Monday returns Friday",
			from:     time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC),  // Monday
			expected: time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),  // Friday
		},
		{
			name:     "Sunday returns Friday",
			from:     time.Date(2024, 1, 7, 10, 0, 0, 0, time.UTC),  // Sunday
			expected: time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),  // Friday
		},
		{
			name:     "Tuesday returns Monday",
			from:     time.Date(2024, 1, 9, 10, 0, 0, 0, time.UTC),  // Tuesday
			expected: time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC),  // Monday
		},
		{
			name:     "Friday returns Thursday",
			from:     time.Date(2024, 1, 12, 10, 0, 0, 0, time.UTC), // Friday
			expected: time.Date(2024, 1, 11, 10, 0, 0, 0, time.UTC), // Thursday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLastWorkingDay(tt.from)
			if !result.Equal(tt.expected) {
				t.Errorf("expected %s, got %s", tt.expected.Format("2006-01-02"), result.Format("2006-01-02"))
			}
		})
	}
}

func TestExtractWorkItems(t *testing.T) {
	tests := []struct {
		name     string
		commits  []CommitDetails
		expected []string
	}{
		{
			name: "extract from conventional commits",
			commits: []CommitDetails{
				{Subject: "feat: Add user authentication"},
				{Subject: "fix: Resolve login timeout issue"},
				{Subject: "docs: Update API documentation"},
				{Subject: "refactor: Simplify database queries"},
			},
			expected: []string{
				"Add user authentication",
				"Resolve login timeout issue",
				"Update API documentation",
				"Simplify database queries",
			},
		},
		{
			name: "skip merge commits and trivial changes",
			commits: []CommitDetails{
				{Subject: "Merge pull request #123"},
				{Subject: "fix typo"},
				{Subject: "WIP: incomplete feature"},
				{Subject: "temp: debugging code"},
				{Subject: "feat: Implement search functionality"},
			},
			expected: []string{
				"Implement search functionality",
			},
		},
		{
			name: "handle non-conventional commits",
			commits: []CommitDetails{
				{Subject: "Added new dashboard widgets"},
				{Subject: "Fixed critical security vulnerability"},
				{Subject: "Implemented OAuth2 integration"},
			},
			expected: []string{
				"Added new dashboard widgets",
				"Fixed critical security vulnerability",
				"Implemented OAuth2 integration",
			},
		},
		{
			name: "deduplicate similar items",
			commits: []CommitDetails{
				{Subject: "feat: add user profile"},
				{Subject: "feat: Add user profile"},
				{Subject: "fix: fix user profile bug"},
			},
			expected: []string{
				"Add user profile",
				"Fix user profile bug",
			},
		},
		{
			name:     "empty commits",
			commits:  []CommitDetails{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractWorkItems(tt.commits)
			
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d items, got %d", len(tt.expected), len(result))
				t.Errorf("expected: %v", tt.expected)
				t.Errorf("got: %v", result)
				return
			}
			
			for i, item := range result {
				if item != tt.expected[i] {
					t.Errorf("item %d: expected '%s', got '%s'", i, tt.expected[i], item)
				}
			}
		})
	}
}