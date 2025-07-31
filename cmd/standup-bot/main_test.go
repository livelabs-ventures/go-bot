package main

import (
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	// Save original args and stderr
	oldArgs := os.Args
	oldStderr := os.Stderr
	defer func() {
		os.Args = oldArgs
		os.Stderr = oldStderr
	}()

	// Test with help flag (should exit 0)
	os.Args = []string{"standup-bot", "--help"}
	
	// Redirect stderr to capture output
	r, w, _ := os.Pipe()
	os.Stderr = w

	// We can't easily test main() because it calls os.Exit
	// but we can verify it exists and compiles
	_ = main
	
	w.Close()
	os.Stderr = oldStderr
	r.Close()
}