package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestExecute(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test help command
	os.Args = []string{"standup-bot", "--help"}
	
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Execute()
	
	w.Close()
	os.Stdout = oldStdout
	
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("Execute() with --help error = %v", err)
	}

	// Check help output contains expected text
	if !strings.Contains(output, "Standup Bot facilitates daily standup updates") {
		t.Error("Help output missing description")
	}
}

func TestRunStandup(t *testing.T) {
	// Create temp directory for config
	tempDir, err := os.MkdirTemp("", "standup-bot-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Note: Full integration testing of runStandup would require
	// mocking stdin/stdout and the git client, which is complex.
	// For unit tests, we test the individual components separately.
	
	// Here we just verify the function exists and can be called
	_ = runStandup
}

func TestFlags(t *testing.T) {
	// Test that all flags are properly initialized
	tests := []struct {
		name     string
		flagName string
		expected interface{}
	}{
		{"config flag", "config", false},
		{"direct flag", "direct", false},
		{"merge flag", "merge", false},
		{"name flag", "name", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := rootCmd.Flag(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.flagName)
			}
		})
	}
}

func TestRootCommand(t *testing.T) {
	// Test root command properties
	if rootCmd.Use != "standup-bot" {
		t.Errorf("Root command Use = %v, want standup-bot", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Root command should have a short description")
	}

	if rootCmd.Long == "" {
		t.Error("Root command should have a long description")
	}

	if rootCmd.RunE == nil {
		t.Error("Root command should have a RunE function")
	}
}

func TestConfigFlag(t *testing.T) {
	// Reset flag for testing
	configFlag = false
	
	// Create a test command
	cmd := &cobra.Command{}
	cmd.Flags().BoolVar(&configFlag, "config", false, "Run configuration setup")

	// Parse with config flag
	err := cmd.ParseFlags([]string{"--config"})
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	if !configFlag {
		t.Error("Config flag should be true after parsing --config")
	}

	// Reset and parse without flag
	configFlag = false
	err = cmd.ParseFlags([]string{})
	if err != nil {
		t.Fatalf("Failed to parse empty flags: %v", err)
	}

	if configFlag {
		t.Error("Config flag should be false when not specified")
	}
}