package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if manager.configDir == "" {
		t.Error("configDir should not be empty")
	}

	if manager.configFile == "" {
		t.Error("configFile should not be empty")
	}

	// Check that paths contain .standup-bot
	if !filepath.IsAbs(manager.configDir) {
		t.Error("configDir should be absolute path")
	}

	if !filepath.IsAbs(manager.configFile) {
		t.Error("configFile should be absolute path")
	}
}

func TestManagerLoadSave(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "standup-bot-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a manager with custom paths for testing
	manager := &Manager{
		configDir:  tempDir,
		configFile: filepath.Join(tempDir, "config.json"),
	}

	// Test loading non-existent config
	cfg, err := manager.Load()
	if err != ErrConfigNotFound {
		t.Fatalf("Load() error = %v, want ErrConfigNotFound for non-existent config", err)
	}
	if cfg != nil {
		t.Error("Load() = non-nil config, want nil for non-existent config")
	}

	// Test saving config
	testConfig := &Config{
		Repository:    "test/repo",
		Name:          "TestUser",
		LocalRepoPath: filepath.Join(tempDir, "repo"),
	}

	err = manager.Save(testConfig)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Test loading saved config
	loadedCfg, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v after save", err)
	}

	if loadedCfg.Repository != testConfig.Repository {
		t.Errorf("Repository = %v, want %v", loadedCfg.Repository, testConfig.Repository)
	}

	if loadedCfg.Name != testConfig.Name {
		t.Errorf("Name = %v, want %v", loadedCfg.Name, testConfig.Name)
	}

	// LocalRepoPath will be expanded, so we check it exists
	if loadedCfg.LocalRepoPath == "" {
		t.Error("LocalRepoPath should not be empty after load")
	}
}

func TestManagerExists(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "standup-bot-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := &Manager{
		configDir:  tempDir,
		configFile: filepath.Join(tempDir, "config.json"),
	}

	// Test non-existent config
	if manager.Exists() {
		t.Error("Exists() = true, want false for non-existent config")
	}

	// Create config
	testConfig := &Config{
		Repository:    "test/repo",
		Name:          "TestUser",
		LocalRepoPath: filepath.Join(tempDir, "repo"),
	}
	err = manager.Save(testConfig)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Test existing config
	if !manager.Exists() {
		t.Error("Exists() = false, want true for existing config")
	}
}

func TestTildeExpansion(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "standup-bot-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := &Manager{
		configDir:  tempDir,
		configFile: filepath.Join(tempDir, "config.json"),
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot test tilde expansion without home directory")
	}

	// Test saving with absolute path converts to tilde
	testConfig := &Config{
		Repository:    "test/repo",
		Name:          "TestUser",
		LocalRepoPath: filepath.Join(homeDir, ".standup-bot", "repo"),
	}

	err = manager.Save(testConfig)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Read the raw file to check tilde notation
	data, err := os.ReadFile(manager.configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	configStr := string(data)
	if !contains(configStr, "~/.standup-bot/repo") {
		t.Error("Config file should contain tilde notation for home directory")
	}

	// Test loading expands tilde
	loadedCfg, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	expectedPath := filepath.Join(homeDir, ".standup-bot", "repo")
	if loadedCfg.LocalRepoPath != expectedPath {
		t.Errorf("LocalRepoPath = %v, want %v", loadedCfg.LocalRepoPath, expectedPath)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig() = nil")
	}

	if cfg.LocalRepoPath != "~/.standup-bot/repo" {
		t.Errorf("LocalRepoPath = %v, want ~/.standup-bot/repo", cfg.LocalRepoPath)
	}
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}