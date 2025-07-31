package commands

import (
	"fmt"
	"time"

	"github.com/standup-bot/standup-bot/pkg/config"
	"github.com/standup-bot/standup-bot/pkg/git"
)

// RunMergeDailyStandup handles merging the daily standup PR
func RunMergeDailyStandup(cfg *config.Config) error {
	gitClient := git.NewClient()

	// Validate environment
	if err := validateMergeEnvironment(gitClient, cfg); err != nil {
		return err
	}

	// Get today's branch name
	today := time.Now()
	branchName := fmt.Sprintf("standup/%s", today.Format("2006-01-02"))
	
	// Find and merge the PR
	if err := findAndMergePR(gitClient, cfg.LocalRepoPath, branchName); err != nil {
		return err
	}
	
	fmt.Println("✅ Today's standups have been merged successfully!")
	
	// Clean up local repository
	if err := cleanupAfterMerge(gitClient, cfg.LocalRepoPath); err != nil {
		// Non-fatal errors, just warn
		fmt.Printf("Warning during cleanup: %v\n", err)
	}
	
	return nil
}

// validateMergeEnvironment checks prerequisites for merging
func validateMergeEnvironment(gitClient *git.Client, cfg *config.Config) error {
	if err := gitClient.CheckGHInstalled(); err != nil {
		return err
	}

	if err := gitClient.CheckAuthenticated(); err != nil {
		return err
	}

	if !gitClient.RepositoryExists(cfg.LocalRepoPath) {
		return fmt.Errorf("repository not found at %s. Please run 'standup-bot --config' to set up", cfg.LocalRepoPath)
	}

	return nil
}

// findAndMergePR finds and merges the PR for the given branch
func findAndMergePR(gitClient *git.Client, repoPath, branchName string) error {
	prExists, prNumber := gitClient.PRExistsForBranch(repoPath, branchName)
	
	if !prExists {
		return fmt.Errorf("no pull request found for today's standups")
	}
	
	fmt.Printf("Merging pull request #%s...\n", prNumber)
	if err := gitClient.MergePullRequestByNumber(repoPath, prNumber); err != nil {
		return fmt.Errorf("failed to merge pull request: %w", err)
	}
	
	return nil
}

// cleanupAfterMerge switches back to main and syncs the repository
func cleanupAfterMerge(gitClient *git.Client, repoPath string) error {
	fmt.Println("Switching back to main branch...")
	if err := gitClient.SwitchToMainBranch(repoPath); err != nil {
		return fmt.Errorf("could not switch to main branch: %w", err)
	}
	
	// Pull latest changes
	if err := gitClient.SyncRepository(repoPath); err != nil {
		return fmt.Errorf("could not sync repository: %w", err)
	}
	
	return nil
}