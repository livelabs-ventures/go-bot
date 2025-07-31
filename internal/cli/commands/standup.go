package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/standup-bot/standup-bot/pkg/config"
	"github.com/standup-bot/standup-bot/pkg/git"
	"github.com/standup-bot/standup-bot/pkg/standup"
)

// RunStandupDirect runs the direct commit workflow (no PR)
func RunStandupDirect(cfg *config.Config) error {
	gitClient := git.NewClient()

	if err := validateEnvironment(gitClient, cfg); err != nil {
		return err
	}

	// Sync repository
	fmt.Println("Syncing repository...")
	if err := gitClient.SyncRepository(cfg.LocalRepoPath); err != nil {
		return fmt.Errorf("failed to sync repository: %w", err)
	}

	// Collect standup entry
	standupManager := standup.NewManager(cfg.LocalRepoPath)
	entry, err := standupManager.CollectEntry(os.Stdin, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to collect standup: %w", err)
	}

	// Save entry to file
	fmt.Println("\nRecording standup...")
	if err := standupManager.SaveEntry(entry, cfg.Name); err != nil {
		return fmt.Errorf("failed to save standup: %w", err)
	}

	// Commit and push
	fmt.Println("Pushing changes...")
	commitMessage := standupManager.FormatCommitMessage(entry, cfg.Name)
	if err := gitClient.CommitAndPush(cfg.LocalRepoPath, commitMessage); err != nil {
		// If push fails, save to temp file
		tempFile := saveTempStandup(entry, cfg.Name)
		return fmt.Errorf("failed to push changes: %w\nYour standup has been saved to: %s", err, tempFile)
	}

	fmt.Println(" Standup recorded successfully!")
	return nil
}

// RunStandupPR runs the pull request workflow
func RunStandupPR(cfg *config.Config) error {
	gitClient := git.NewClient()

	if err := validateEnvironment(gitClient, cfg); err != nil {
		return err
	}

	// Sync repository
	fmt.Println("Syncing repository...")
	if err := gitClient.SyncRepository(cfg.LocalRepoPath); err != nil {
		return fmt.Errorf("failed to sync repository: %w", err)
	}

	// Ensure main branch exists
	if err := ensureMainBranch(cfg.LocalRepoPath, gitClient); err != nil {
		return err
	}

	// Collect standup entry
	standupManager := standup.NewManager(cfg.LocalRepoPath)
	entry, err := standupManager.CollectEntry(os.Stdin, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to collect standup: %w", err)
	}

	// Handle branch and PR creation
	if err := createOrUpdateStandupPR(cfg, gitClient, standupManager, entry); err != nil {
		return err
	}

	fmt.Println("✅ Standup recorded successfully!")
	fmt.Println("💡 To merge today's standups, run: standup-bot --merge")
	return nil
}

// validateEnvironment checks if GitHub CLI is installed and authenticated
func validateEnvironment(gitClient *git.Client, cfg *config.Config) error {
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

// ensureMainBranch ensures the main branch exists
func ensureMainBranch(repoPath string, gitClient *git.Client) error {
	mainExistsLocal := gitClient.BranchExists(repoPath, "main")
	mainExistsRemote := gitClient.RemoteBranchExists(repoPath, "main")
	
	if !mainExistsLocal && !mainExistsRemote {
		fmt.Println("Creating initial main branch...")
		if err := createInitialMainBranch(repoPath, gitClient); err != nil {
			return fmt.Errorf("failed to create initial main branch: %w", err)
		}
	} else if mainExistsLocal {
		if err := gitClient.SwitchToMainBranch(repoPath); err != nil {
			return fmt.Errorf("failed to switch to main branch: %w", err)
		}
	}

	return nil
}

// createOrUpdateStandupPR handles the PR workflow for a standup entry
func createOrUpdateStandupPR(cfg *config.Config, gitClient *git.Client, standupManager *standup.Manager, entry *standup.Entry) error {
	branchName := fmt.Sprintf("standup/%s", entry.Date.Format("2006-01-02"))
	
	// Handle branch creation or switching
	if err := handleBranch(cfg.LocalRepoPath, gitClient, branchName); err != nil {
		return err
	}

	// Save entry to file
	fmt.Println("Recording standup...")
	if err := standupManager.SaveEntry(entry, cfg.Name); err != nil {
		return fmt.Errorf("failed to save standup: %w", err)
	}

	// Commit changes
	if err := commitStandupChanges(cfg, gitClient, entry); err != nil {
		return err
	}

	// Push the branch
	fmt.Println("Pushing branch...")
	if err := gitClient.PushBranch(cfg.LocalRepoPath, branchName); err != nil {
		tempFile := saveTempStandup(entry, cfg.Name)
		return fmt.Errorf("failed to push changes: %w\nYour standup has been saved to: %s", err, tempFile)
	}

	// Create or update PR
	return handlePullRequest(cfg, gitClient, branchName, entry.Date)
}

// handleBranch creates or switches to the standup branch
func handleBranch(repoPath string, gitClient *git.Client, branchName string) error {
	if gitClient.BranchExists(repoPath, branchName) {
		fmt.Println("Switching to today's standup branch...")
		if err := gitClient.SwitchToBranch(repoPath, branchName); err != nil {
			return fmt.Errorf("failed to switch to branch: %w", err)
		}
		
		fmt.Println("Pulling latest updates...")
		if err := gitClient.PullBranch(repoPath, branchName); err != nil {
			fmt.Println("Note: Could not pull from remote branch (this is normal for new branches)")
		}
	} else {
		fmt.Println("Creating today's standup branch...")
		if err := gitClient.CreateBranch(repoPath, branchName); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}
	}
	return nil
}

// commitStandupChanges adds and commits the standup changes
func commitStandupChanges(cfg *config.Config, gitClient *git.Client, entry *standup.Entry) error {
	_, err := gitClient.AddAll(cfg.LocalRepoPath)
	if err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}
	
	commitMessage := fmt.Sprintf("[Standup] %s - %s", cfg.Name, entry.Date.Format("2006-01-02"))
	_, err = gitClient.Commit(cfg.LocalRepoPath, commitMessage)
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	
	return nil
}

// handlePullRequest creates or updates the PR
func handlePullRequest(cfg *config.Config, gitClient *git.Client, branchName string, date time.Time) error {
	prExists, prNumber := gitClient.PRExistsForBranch(cfg.LocalRepoPath, branchName)
	
	if prExists {
		fmt.Printf("Updating existing pull request #%s...\n", prNumber)
		prBody := FormatDailyPRBody(cfg.LocalRepoPath, date)
		if err := gitClient.UpdatePullRequest(cfg.LocalRepoPath, prNumber, prBody); err != nil {
			fmt.Printf("Warning: Could not update PR body: %v\n", err)
		}
	} else {
		fmt.Println("Creating pull request...")
		prTitle := fmt.Sprintf("[Standup] %s", date.Format("2006-01-02"))
		prBody := FormatDailyPRBody(cfg.LocalRepoPath, date)
		
		if err := gitClient.CreatePullRequest(cfg.LocalRepoPath, prTitle, prBody); err != nil {
			return fmt.Errorf("failed to create pull request: %w", err)
		}
	}
	
	return nil
}

// saveTempStandup saves a standup to a temporary file in case of errors
func saveTempStandup(entry *standup.Entry, userName string) string {
	tempFile := fmt.Sprintf("/tmp/standup-%s-%s.txt", userName, entry.Date.Format("2006-01-02"))
	
	content := fmt.Sprintf("Standup for %s on %s\n\nYesterday:\n", userName, entry.Date.Format("2006-01-02"))
	for _, item := range entry.Yesterday {
		content += fmt.Sprintf("- %s\n", item)
	}
	content += "\nToday:\n"
	for _, item := range entry.Today {
		content += fmt.Sprintf("- %s\n", item)
	}
	content += fmt.Sprintf("\nBlockers:\n%s\n", entry.Blockers)

	// Ignore error for temp file save
	_ = os.WriteFile(tempFile, []byte(content), 0644)
	
	return tempFile
}

// createInitialMainBranch creates the initial main branch with a README
func createInitialMainBranch(repoPath string, gitClient *git.Client) error {
	// Create a README file
	readmePath := filepath.Join(repoPath, "README.md")
	readmeContent := `# Team Standups

This repository contains daily standup updates from the team.

## Structure

Each team member has their own markdown file in the ` + "`stand-ups/`" + ` directory.
`
	
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}
	
	// Create stand-ups directory
	standupDir := filepath.Join(repoPath, "stand-ups")
	if err := os.MkdirAll(standupDir, 0755); err != nil {
		return fmt.Errorf("failed to create stand-ups directory: %w", err)
	}
	
	// Add, commit, and push
	if _, err := gitClient.AddAll(repoPath); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}
	
	if _, err := gitClient.Commit(repoPath, "Initial repository setup"); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	
	// Push to create main branch on remote
	if err := gitClient.PushBranch(repoPath, "main"); err != nil {
		return fmt.Errorf("failed to push main branch: %w", err)
	}
	
	return nil
}