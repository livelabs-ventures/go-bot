package commands

import (
	"testing"
)

func TestMCPServerTypes(t *testing.T) {
	// Test that our argument types are properly defined
	t.Run("SubmitStandupArgs", func(t *testing.T) {
		args := SubmitStandupArgs{
			Yesterday: []string{"Task 1", "Task 2"},
			Today:     []string{"Task 3", "Task 4"},
			Blockers:  "None",
			Direct:    false,
		}
		
		if len(args.Yesterday) != 2 {
			t.Errorf("Expected 2 yesterday items, got %d", len(args.Yesterday))
		}
		
		if len(args.Today) != 2 {
			t.Errorf("Expected 2 today items, got %d", len(args.Today))
		}
		
		if args.Blockers != "None" {
			t.Errorf("Expected blockers to be 'None', got %s", args.Blockers)
		}
		
		if args.Direct != false {
			t.Errorf("Expected direct to be false, got %v", args.Direct)
		}
	})
	
	t.Run("CreateStandupPRArgs", func(t *testing.T) {
		args := CreateStandupPRArgs{
			Merge: true,
		}
		
		if args.Merge != true {
			t.Errorf("Expected merge to be true, got %v", args.Merge)
		}
	})
	
	t.Run("GetStandupStatusArgs", func(t *testing.T) {
		// This type has no fields, just ensure it can be instantiated
		_ = GetStandupStatusArgs{}
	})
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "contains substring",
			s:        "Hello, World!",
			substr:   "World",
			expected: true,
		},
		{
			name:     "does not contain substring",
			s:        "Hello, World!",
			substr:   "Goodbye",
			expected: false,
		},
		{
			name:     "empty substring",
			s:        "Hello, World!",
			substr:   "",
			expected: true,
		},
		{
			name:     "substring longer than string",
			s:        "Hi",
			substr:   "Hello",
			expected: false,
		},
		{
			name:     "exact match",
			s:        "Hello",
			substr:   "Hello",
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsString(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("containsString(%q, %q) = %v; want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}