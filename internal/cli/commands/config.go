package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/standup-bot/standup-bot/pkg/config"
	"github.com/standup-bot/standup-bot/pkg/git"
)

// RunConfiguration handles the configuration setup workflow
func RunConfiguration(cfgManager *config.Manager) error {
	fmt.Println("Welcome to Standup Bot!")
	
	cfg, err := collectConfigurationInput()
	if err != nil {
		return err
	}

	if err := cfgManager.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("Configuration saved!")

	if err := setupRepository(cfg); err != nil {
		return err
	}

	fmt.Println("\nSetup complete! Run 'standup-bot' to record your standup.")
	return nil
}

// collectConfigurationInput prompts the user for configuration values
func collectConfigurationInput() (*config.Config, error) {
	// Get repository
	fmt.Print("GitHub Repository (e.g., org/standup-repo): ")
	var repo string
	if _, err := fmt.Scanln(&repo); err != nil {
		return nil, fmt.Errorf("failed to read repository: %w", err)
	}

	// Get user name
	fmt.Print("Your Name: ")
	var name string
	if _, err := fmt.Scanln(&name); err != nil {
		return nil, fmt.Errorf("failed to read name: %w", err)
	}

	cfg := config.DefaultConfig()
	cfg.Repository = repo
	cfg.Name = name

	return cfg, nil
}

// setupRepository clones the repository if it doesn't exist
func setupRepository(cfg *config.Config) error {
	gitClient := git.NewClient()
	
	// Check GitHub CLI is installed
	if err := gitClient.CheckGHInstalled(); err != nil {
		return err
	}

	// Check authentication
	if err := gitClient.CheckAuthenticated(); err != nil {
		return err
	}

	// Expand tilde in local repo path for actual operations
	expandedPath := expandPath(cfg.LocalRepoPath)

	// Clone repository if needed
	if !gitClient.RepositoryExists(expandedPath) {
		fmt.Println("Cloning repository...")
		if err := gitClient.CloneRepository(cfg.Repository, expandedPath); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
		fmt.Println("Repository cloned successfully!")
	}

	return nil
}

// expandPath expands tilde in paths
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, path[1:])
	}
	return path
}