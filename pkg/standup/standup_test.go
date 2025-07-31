package standup

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCollectEntry(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedYest   []string
		expectedToday  []string
		expectedBlock  string
	}{
		{
			name: "normal input",
			input: `Finished API endpoints
Fixed bug in auth
` + "\n" + `Work on frontend
Write tests
` + "\n" + `None
`,
			expectedYest:  []string{"Finished API endpoints", "Fixed bug in auth"},
			expectedToday: []string{"Work on frontend", "Write tests"},
			expectedBlock: "None",
		},
		{
			name: "single line entries",
			input: `Worked on stuff
` + "\n" + `More work
` + "\n" + `Waiting for review
`,
			expectedYest:  []string{"Worked on stuff"},
			expectedToday: []string{"More work"},
			expectedBlock: "Waiting for review",
		},
		{
			name: "empty blockers",
			input: `Did things
` + "\n" + `Will do things
` + "\n" + `
`,
			expectedYest:  []string{"Did things"},
			expectedToday: []string{"Will do things"},
			expectedBlock: "None",
		},
		{
			name:          "all empty",
			input:         "\n\n\n",
			expectedYest:  []string{},
			expectedToday: []string{},
			expectedBlock: "None",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			writer := &bytes.Buffer{}
			
			manager := NewManager("/test/repo")
			entry, err := manager.CollectEntry(reader, writer)
			
			if err != nil {
				t.Fatalf("CollectEntry() error = %v", err)
			}

			// Check yesterday
			if !slicesEqual(entry.Yesterday, tt.expectedYest) {
				t.Errorf("Yesterday = %v, want %v", entry.Yesterday, tt.expectedYest)
			}

			// Check today
			if !slicesEqual(entry.Today, tt.expectedToday) {
				t.Errorf("Today = %v, want %v", entry.Today, tt.expectedToday)
			}

			// Check blockers
			if entry.Blockers != tt.expectedBlock {
				t.Errorf("Blockers = %v, want %v", entry.Blockers, tt.expectedBlock)
			}

			// Verify prompts were displayed
			output := writer.String()
			if !strings.Contains(output, "What did you do yesterday?") {
				t.Error("Missing yesterday prompt")
			}
			if !strings.Contains(output, "What will you do today?") {
				t.Error("Missing today prompt")
			}
			if !strings.Contains(output, "Any blockers?") {
				t.Error("Missing blockers prompt")
			}
		})
	}
}

func TestSaveEntry(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "standup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewManager(tempDir)
	
	// Test entry
	entry := &Entry{
		Date:      time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		Yesterday: []string{"Completed API endpoints", "Fixed authentication bug"},
		Today:     []string{"Work on frontend", "Write unit tests"},
		Blockers:  "None",
	}

	// Save entry
	err = manager.SaveEntry(entry, "Alice")
	if err != nil {
		t.Fatalf("SaveEntry() error = %v", err)
	}

	// Check file was created
	filePath := filepath.Join(tempDir, "stand-ups", "alice.md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	contentStr := string(content)

	// Check header
	if !strings.Contains(contentStr, "# Alice's Standups") {
		t.Error("Missing header in standup file")
	}

	// Check date
	if !strings.Contains(contentStr, "## 2024-01-31") {
		t.Error("Missing or incorrect date")
	}

	// Check yesterday section
	if !strings.Contains(contentStr, "**Yesterday:**") {
		t.Error("Missing yesterday section")
	}
	if !strings.Contains(contentStr, "- Completed API endpoints") {
		t.Error("Missing yesterday item 1")
	}
	if !strings.Contains(contentStr, "- Fixed authentication bug") {
		t.Error("Missing yesterday item 2")
	}

	// Check today section
	if !strings.Contains(contentStr, "**Today:**") {
		t.Error("Missing today section")
	}
	if !strings.Contains(contentStr, "- Work on frontend") {
		t.Error("Missing today item 1")
	}
	if !strings.Contains(contentStr, "- Write unit tests") {
		t.Error("Missing today item 2")
	}

	// Check blockers
	if !strings.Contains(contentStr, "**Blockers:**\nNone") {
		t.Error("Missing or incorrect blockers section")
	}

	// Test appending another entry
	entry2 := &Entry{
		Date:      time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		Yesterday: []string{"Finished frontend work"},
		Today:     []string{"Code review"},
		Blockers:  "Waiting for API deployment",
	}

	err = manager.SaveEntry(entry2, "Alice")
	if err != nil {
		t.Fatalf("SaveEntry() second entry error = %v", err)
	}

	// Read updated content
	content2, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	contentStr2 := string(content2)

	// Check both entries exist
	if !strings.Contains(contentStr2, "## 2024-01-31") {
		t.Error("First entry missing after append")
	}
	if !strings.Contains(contentStr2, "## 2024-02-01") {
		t.Error("Second entry missing")
	}
	if !strings.Contains(contentStr2, "Waiting for API deployment") {
		t.Error("Second entry blockers missing")
	}

	// Header should only appear once
	headerCount := strings.Count(contentStr2, "# Alice's Standups")
	if headerCount != 1 {
		t.Errorf("Header appears %d times, want 1", headerCount)
	}
}

func TestSaveEntryEmptyItems(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "standup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewManager(tempDir)
	
	// Test entry with empty items
	entry := &Entry{
		Date:      time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		Yesterday: []string{},
		Today:     []string{},
		Blockers:  "Blocked by everything",
	}

	err = manager.SaveEntry(entry, "Bob")
	if err != nil {
		t.Fatalf("SaveEntry() error = %v", err)
	}

	// Check file content
	filePath := filepath.Join(tempDir, "stand-ups", "bob.md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	contentStr := string(content)

	// Check default messages for empty items
	if !strings.Contains(contentStr, "- Nothing to report") {
		t.Error("Missing default yesterday message")
	}
	if !strings.Contains(contentStr, "- Nothing planned") {
		t.Error("Missing default today message")
	}
}

func TestFormatCommitMessage(t *testing.T) {
	manager := NewManager("/test/repo")

	tests := []struct {
		name     string
		entry    *Entry
		userName string
		expected []string // Expected strings that should be in the message
	}{
		{
			name: "normal entry",
			entry: &Entry{
				Date:      time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
				Yesterday: []string{"Completed API", "Fixed bugs"},
				Today:     []string{"Frontend work", "Tests"},
				Blockers:  "None",
			},
			userName: "Alice",
			expected: []string{
				"[Standup] Alice - 2024-01-31",
				"Yesterday: Completed API; Fixed bugs",
				"Today: Frontend work; Tests",
				"Blockers: None",
			},
		},
		{
			name: "empty items",
			entry: &Entry{
				Date:      time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
				Yesterday: []string{},
				Today:     []string{},
				Blockers:  "Everything",
			},
			userName: "Bob",
			expected: []string{
				"[Standup] Bob - 2024-01-31",
				"Yesterday: Nothing to report",
				"Today: Nothing planned",
				"Blockers: Everything",
			},
		},
		{
			name: "single items",
			entry: &Entry{
				Date:      time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
				Yesterday: []string{"Worked"},
				Today:     []string{"More work"},
				Blockers:  "Waiting for review",
			},
			userName: "Charlie",
			expected: []string{
				"[Standup] Charlie - 2024-01-31",
				"Yesterday: Worked",
				"Today: More work",
				"Blockers: Waiting for review",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := manager.FormatCommitMessage(tt.entry, tt.userName)

			for _, expected := range tt.expected {
				if !strings.Contains(message, expected) {
					t.Errorf("Commit message missing expected string: %q\nGot message:\n%s", expected, message)
				}
			}

			// Check format has proper newlines
			lines := strings.Split(message, "\n")
			if len(lines) < 5 {
				t.Errorf("Commit message should have at least 5 lines, got %d", len(lines))
			}

			// Check title format
			if !strings.HasPrefix(lines[0], "[Standup]") {
				t.Error("Commit message should start with [Standup]")
			}

			// Check empty line after title
			if lines[1] != "" {
				t.Error("Should have empty line after title")
			}
		})
	}
}

// Helper function
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}