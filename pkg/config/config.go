package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/standup-bot/standup-bot/pkg/types"
)

// Config holds the application configuration
type Config struct {
	Repository    string `json:"repository"`
	Name          string `json:"name"`
	LocalRepoPath string `json:"localRepoPath"`
}

// GetRepository returns the repository as a typed value
func (c *Config) GetRepository() (types.Repository, error) {
	return types.NewRepository(c.Repository)
}

// GetUserName returns the user name as a typed value
func (c *Config) GetUserName() (types.UserName, error) {
	return types.NewUserName(c.Name)
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Repository == "" {
		return fmt.Errorf("repository cannot be empty")
	}
	
	if c.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	
	if c.LocalRepoPath == "" {
		return fmt.Errorf("local repository path cannot be empty")
	}
	
	// Validate repository format
	if _, err := c.GetRepository(); err != nil {
		return fmt.Errorf("invalid repository: %w", err)
	}
	
	// Validate user name
	if _, err := c.GetUserName(); err != nil {
		return fmt.Errorf("invalid user name: %w", err)
	}
	
	return nil
}

// Manager handles configuration operations
type Manager struct {
	configDir  string
	configFile string
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".standup-bot")
	configFile := filepath.Join(configDir, "config.json")

	return &Manager{
		configDir:  configDir,
		configFile: configFile,
	}, nil
}

// ConfigDir returns the configuration directory path
func (m *Manager) ConfigDir() string {
	return m.configDir
}

// Load reads the configuration from disk
func (m *Manager) Load() (*Config, error) {
	if !m.Exists() {
		return nil, ErrConfigNotFound
	}

	data, err := os.ReadFile(m.configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file at %s: %w", m.configFile, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w (file: %s)", err, m.configFile)
	}

	// Expand tilde in local repo path
	if err := m.expandPath(&cfg); err != nil {
		return nil, fmt.Errorf("failed to expand path: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// ErrConfigNotFound indicates the configuration file doesn't exist
var ErrConfigNotFound = fmt.Errorf("configuration file not found")

// expandPath expands tilde in the local repo path
func (m *Manager) expandPath(cfg *Config) error {
	if cfg.LocalRepoPath != "" && cfg.LocalRepoPath[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		cfg.LocalRepoPath = filepath.Join(homeDir, cfg.LocalRepoPath[1:])
	}
	return nil
}

// Save writes the configuration to disk
func (m *Manager) Save(cfg *Config) error {
	// Validate configuration before saving
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid configuration: %w", err)
	}

	// Ensure config directory exists
	if err := m.ensureConfigDir(); err != nil {
		return err
	}

	// Create a copy for saving with tilde notation
	saveCfg := m.prepareForSave(cfg)

	// Marshal configuration
	data, err := json.MarshalIndent(saveCfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(m.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file to %s: %w", m.configFile, err)
	}

	return nil
}

// ensureConfigDir creates the configuration directory if it doesn't exist
func (m *Manager) ensureConfigDir() error {
	if err := os.MkdirAll(m.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory at %s: %w", m.configDir, err)
	}
	return nil
}

// prepareForSave creates a copy of the config with tilde notation for paths
func (m *Manager) prepareForSave(cfg *Config) Config {
	saveCfg := *cfg
	homeDir, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(saveCfg.LocalRepoPath, homeDir) {
		saveCfg.LocalRepoPath = "~" + saveCfg.LocalRepoPath[len(homeDir):]
	}
	return saveCfg
}

// Exists checks if configuration file exists
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.configFile)
	return err == nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		LocalRepoPath: "~/.standup-bot/repo",
	}
}