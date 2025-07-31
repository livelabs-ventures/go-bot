package standup

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestParseJSONInput(t *testing.T) {
	// Create a temp JSON file for file input tests
	tempFile, err := os.CreateTemp("", "standup-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	
	validJSON := `{"yesterday": ["File task"], "today": ["Another task"], "blockers": "None"}`
	if _, err := tempFile.Write([]byte(validJSON)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()

	tests := []struct {
		name      string
		jsonStr   string
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *Entry)
	}{
		{
			name: "valid input with all fields",
			jsonStr: `{
				"yesterday": ["Completed task A", "Fixed bug B"],
				"today": ["Start task C", "Review PRs"],
				"blockers": "Waiting for API access"
			}`,
			wantErr: false,
			validate: func(t *testing.T, e *Entry) {
				if len(e.Yesterday) != 2 {
					t.Errorf("expected 2 yesterday items, got %d", len(e.Yesterday))
				}
				if len(e.Today) != 2 {
					t.Errorf("expected 2 today items, got %d", len(e.Today))
				}
				if e.Blockers != "Waiting for API access" {
					t.Errorf("expected blockers 'Waiting for API access', got %s", e.Blockers)
				}
			},
		},
		{
			name: "valid input with empty blockers",
			jsonStr: `{
				"yesterday": ["Task A"],
				"today": ["Task B"],
				"blockers": ""
			}`,
			wantErr: false,
			validate: func(t *testing.T, e *Entry) {
				if e.Blockers != "None" {
					t.Errorf("expected blockers 'None', got %s", e.Blockers)
				}
			},
		},
		{
			name: "valid input without blockers field",
			jsonStr: `{
				"yesterday": ["Task A"],
				"today": ["Task B"]
			}`,
			wantErr: false,
			validate: func(t *testing.T, e *Entry) {
				if e.Blockers != "None" {
					t.Errorf("expected blockers 'None', got %s", e.Blockers)
				}
			},
		},
		{
			name:    "invalid JSON",
			jsonStr: `{invalid json}`,
			wantErr: true,
			errMsg:  "invalid JSON input",
		},
		{
			name: "empty yesterday and today",
			jsonStr: `{
				"yesterday": [],
				"today": [],
				"blockers": "None"
			}`,
			wantErr: true,
			errMsg:  "at least one of 'yesterday' or 'today' must have entries",
		},
		{
			name: "only yesterday entries",
			jsonStr: `{
				"yesterday": ["Task A"],
				"today": []
			}`,
			wantErr: false,
			validate: func(t *testing.T, e *Entry) {
				if len(e.Yesterday) != 1 {
					t.Errorf("expected 1 yesterday item, got %d", len(e.Yesterday))
				}
				if len(e.Today) != 0 {
					t.Errorf("expected 0 today items, got %d", len(e.Today))
				}
			},
		},
		{
			name:    "read from file",
			jsonStr: tempFile.Name(),
			wantErr: false,
			validate: func(t *testing.T, e *Entry) {
				if len(e.Yesterday) != 1 || e.Yesterday[0] != "File task" {
					t.Errorf("expected yesterday to contain 'File task', got %v", e.Yesterday)
				}
				if len(e.Today) != 1 || e.Today[0] != "Another task" {
					t.Errorf("expected today to contain 'Another task', got %v", e.Today)
				}
			},
		},
		{
			name:    "non-existent file falls back to direct parse",
			jsonStr: "/non/existent/file.json",
			wantErr: true,
			errMsg:  "invalid JSON input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := ParseJSONInput(tt.jsonStr)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if entry == nil {
				t.Errorf("expected non-nil entry")
				return
			}
			
			// Check date is set to today
			today := time.Now().Format("2006-01-02")
			if entry.Date.Format("2006-01-02") != today {
				t.Errorf("expected date %s, got %s", today, entry.Date.Format("2006-01-02"))
			}
			
			if tt.validate != nil {
				tt.validate(t, entry)
			}
		})
	}
}

func TestFormatJSONOutput(t *testing.T) {
	output := JSONOutput{
		Success:   true,
		Message:   "Test successful",
		Date:      "2025-07-31",
		User:      "testuser",
		Yesterday: []string{"Task A", "Task B"},
		Today:     []string{"Task C"},
		Blockers:  "None",
		FilePath:  "/path/to/file.md",
		PRNumber:  "42",
		PRUrl:     "https://github.com/org/repo/pull/42",
	}
	
	jsonStr, err := FormatJSONOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Parse back to verify
	var parsed JSONOutput
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}
	
	if parsed.Success != output.Success {
		t.Errorf("expected success %v, got %v", output.Success, parsed.Success)
	}
	if parsed.Message != output.Message {
		t.Errorf("expected message '%s', got '%s'", output.Message, parsed.Message)
	}
	if parsed.PRNumber != output.PRNumber {
		t.Errorf("expected PR number '%s', got '%s'", output.PRNumber, parsed.PRNumber)
	}
}


// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}