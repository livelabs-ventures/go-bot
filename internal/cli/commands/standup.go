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
func RunStandupDirect(cfg *config.Config, jsonInput, outputFormat string) error {
	gitClient := git.NewClient()

	if err := validateEnvironment(gitClient, cfg); err != nil {
		return handleError(err, outputFormat)
	}

	// Sync repository
	if outputFormat != "json" {
		fmt.Println("Syncing repository...")
	}
	if err := gitClient.SyncRepository(cfg.LocalRepoPath); err != nil {
		return handleError(fmt.Errorf("failed to sync repository: %w", err), outputFormat)
	}

	// Collect standup entry
	standupManager := standup.NewManager(cfg.LocalRepoPath)
	var entry *standup.Entry
	var err error

	if jsonInput != "" {
		// Parse JSON input
		entry, err = standup.ParseJSONInput(jsonInput)
		if err != nil {
			return handleError(fmt.Errorf("failed to parse JSON input: %w", err), outputFormat)
		}
	} else {
		// Interactive mode
		entry, err = standupManager.CollectEntry(os.Stdin, os.Stdout)
		if err != nil {
			return handleError(fmt.Errorf("failed to collect standup: %w", err), outputFormat)
		}
	}

	// Save entry to file
	if outputFormat != "json" {
		fmt.Println("\nRecording standup...")
	}
	filePath, err := standupManager.GetStandupFilePath(cfg.Name)
	if err != nil {
		return handleError(fmt.Errorf("failed to get standup file path: %w", err), outputFormat)
	}
	
	if err := standupManager.SaveEntry(entry, cfg.Name); err != nil {
		return handleError(fmt.Errorf("failed to save standup: %w", err), outputFormat)
	}

	// Commit and push
	if outputFormat != "json" {
		fmt.Println("Pushing changes...")
	}
	commitMessage := standupManager.FormatCommitMessage(entry, cfg.Name)
	if err := gitClient.CommitAndPush(cfg.LocalRepoPath, commitMessage); err != nil {
		// If push fails, save to temp file
		tempFile := saveTempStandup(entry, cfg.Name)
		errMsg := fmt.Errorf("failed to push changes: %w\nYour standup has been saved to: %s", err, tempFile)
		return handleError(errMsg, outputFormat)
	}

	// Handle output
	if outputFormat == "json" {
		output := standup.JSONOutput{
			Success:  true,
			Message:  "Standup recorded successfully",
			Date:     entry.Date.Format("2006-01-02"),
			User:     cfg.Name,
			Yesterday: entry.Yesterday,
			Today:    entry.Today,
			Blockers: entry.Blockers,
			FilePath: filePath,
		}
		jsonStr, err := standup.FormatJSONOutput(output)
		if err != nil {
			return err
		}
		fmt.Println(jsonStr)
	} else {
		fmt.Println("✅ Standup recorded successfully!")
	}
	
	return nil
}

// RunStandupPR runs the pull request workflow
func RunStandupPR(cfg *config.Config, jsonInput, outputFormat string) error {
	gitClient := git.NewClient()

	if err := validateEnvironment(gitClient, cfg); err != nil {
		return handleError(err, outputFormat)
	}

	// Sync repository
	if outputFormat != "json" {
		fmt.Println("Syncing repository...")
	}
	if err := gitClient.SyncRepository(cfg.LocalRepoPath); err != nil {
		return handleError(fmt.Errorf("failed to sync repository: %w", err), outputFormat)
	}

	// Ensure main branch exists
	if err := ensureMainBranch(cfg.LocalRepoPath, gitClient); err != nil {
		return handleError(err, outputFormat)
	}

	// Collect standup entry
	standupManager := standup.NewManager(cfg.LocalRepoPath)
	var entry *standup.Entry
	var err error

	if jsonInput != "" {
		// Parse JSON input
		entry, err = standup.ParseJSONInput(jsonInput)
		if err != nil {
			return handleError(fmt.Errorf("failed to parse JSON input: %w", err), outputFormat)
		}
	} else {
		// Interactive mode
		entry, err = standupManager.CollectEntry(os.Stdin, os.Stdout)
		if err != nil {
			return handleError(fmt.Errorf("failed to collect standup: %w", err), outputFormat)
		}
	}

	// Handle branch and PR creation
	prInfo, err := createOrUpdateStandupPR(cfg, gitClient, standupManager, entry, outputFormat)
	if err != nil {
		return handleError(err, outputFormat)
	}

	// Handle output
	if outputFormat == "json" {
		filePath, _ := standupManager.GetStandupFilePath(cfg.Name)
		output := standup.JSONOutput{
			Success:   true,
			Message:   "Standup recorded and PR created/updated successfully",
			Date:      entry.Date.Format("2006-01-02"),
			User:      cfg.Name,
			Yesterday: entry.Yesterday,
			Today:     entry.Today,
			Blockers:  entry.Blockers,
			FilePath:  filePath,
			PRNumber:  prInfo.Number,
			PRUrl:     prInfo.URL,
		}
		jsonStr, err := standup.FormatJSONOutput(output)
		if err != nil {
			return err
		}
		fmt.Println(jsonStr)
	} else {
		fmt.Println("✅ Standup recorded successfully!")
		fmt.Println("💡 To merge today's standups, run: standup-bot --merge")
	}
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
func createOrUpdateStandupPR(cfg *config.Config, gitClient *git.Client, standupManager *standup.Manager, entry *standup.Entry, outputFormat string) (*PRInfo, error) {
	branchName := fmt.Sprintf("standup/%s", entry.Date.Format("2006-01-02"))
	
	// Handle branch creation or switching
	if err := handleBranch(cfg.LocalRepoPath, gitClient, branchName); err != nil {
		return nil, err
	}

	// Save entry to file
	if outputFormat != "json" {
		fmt.Println("Recording standup...")
	}
	if err := standupManager.SaveEntry(entry, cfg.Name); err != nil {
		return nil, fmt.Errorf("failed to save standup: %w", err)
	}

	// Commit changes
	if err := commitStandupChanges(cfg, gitClient, entry); err != nil {
		return nil, err
	}

	// Push the branch
	if outputFormat != "json" {
		fmt.Println("Pushing branch...")
	}
	if err := gitClient.PushBranch(cfg.LocalRepoPath, branchName); err != nil {
		tempFile := saveTempStandup(entry, cfg.Name)
		return nil, fmt.Errorf("failed to push changes: %w\nYour standup has been saved to: %s", err, tempFile)
	}

	// Create or update PR
	return handlePullRequest(cfg, gitClient, branchName, entry.Date, outputFormat)
}

// handleBranch creates or switches to the standup branch
func handleBranch(repoPath string, gitClient *git.Client, branchName string) error {
	return handleBranchWithOutput(repoPath, gitClient, branchName, "")
}

// handleBranchWithOutput creates or switches to the standup branch with optional output format
func handleBranchWithOutput(repoPath string, gitClient *git.Client, branchName string, outputFormat string) error {
	if gitClient.BranchExists(repoPath, branchName) {
		if outputFormat != "json" {
			fmt.Println("Switching to today's standup branch...")
		}
		if err := gitClient.SwitchToBranch(repoPath, branchName); err != nil {
			return fmt.Errorf("failed to switch to branch: %w", err)
		}
		
		if outputFormat != "json" {
			fmt.Println("Pulling latest updates...")
		}
		if err := gitClient.PullBranch(repoPath, branchName); err != nil {
			if outputFormat != "json" {
				fmt.Println("Note: Could not pull from remote branch (this is normal for new branches)")
			}
		}
	} else {
		if outputFormat != "json" {
			fmt.Println("Creating today's standup branch...")
		}
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
func handlePullRequest(cfg *config.Config, gitClient *git.Client, branchName string, date time.Time, outputFormat string) (*PRInfo, error) {
	prExists, prNumber := gitClient.PRExistsForBranch(cfg.LocalRepoPath, branchName)
	
	if prExists {
		if outputFormat != "json" {
			fmt.Printf("Updating existing pull request #%s...\n", prNumber)
		}
		prBody := FormatDailyPRBody(cfg.LocalRepoPath, date)
		if err := gitClient.UpdatePullRequest(cfg.LocalRepoPath, prNumber, prBody); err != nil {
			if outputFormat != "json" {
				fmt.Printf("Warning: Could not update PR body: %v\n", err)
			}
		}
		return &PRInfo{
			Number: prNumber,
			URL:    fmt.Sprintf("https://github.com/%s/pull/%s", getRepoName(cfg.LocalRepoPath), prNumber),
		}, nil
	} else {
		if outputFormat != "json" {
			fmt.Println("Creating pull request...")
		}
		prTitle := fmt.Sprintf("[Standup] %s", date.Format("2006-01-02"))
		prBody := FormatDailyPRBody(cfg.LocalRepoPath, date)
		
		if err := gitClient.CreatePullRequest(cfg.LocalRepoPath, prTitle, prBody); err != nil {
			return nil, fmt.Errorf("failed to create pull request: %w", err)
		}
		
		// Get the PR number of the newly created PR
		_, newPRNumber := gitClient.PRExistsForBranch(cfg.LocalRepoPath, branchName)
		return &PRInfo{
			Number: newPRNumber,
			URL:    fmt.Sprintf("https://github.com/%s/pull/%s", getRepoName(cfg.LocalRepoPath), newPRNumber),
		}, nil
	}
}

// getRepoName extracts the repository name from the local path
func getRepoName(repoPath string) string {
	// This is a simplified version - in production you'd want to parse the git remote URL
	return filepath.Base(repoPath)
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

// PRInfo holds information about a pull request
type PRInfo struct {
	Number string
	URL    string
}

// handleError formats errors based on output format
func handleError(err error, outputFormat string) error {
	if outputFormat == "json" {
		output := standup.JSONOutput{
			Success: false,
			Error:   err.Error(),
			Date:    time.Now().Format("2006-01-02"),
		}
		jsonStr, _ := standup.FormatJSONOutput(output)
		fmt.Println(jsonStr)
		return nil // Don't return error so JSON is printed
	}
	return err
}

// RunStandupSuggest analyzes recent commits and suggests standup content
func RunStandupSuggest(cfg *config.Config, outputFormat string) error {
	gitClient := git.NewClient()

	if err := validateEnvironment(gitClient, cfg); err != nil {
		return handleError(err, outputFormat)
	}

	// Get last working day
	lastWorkingDay := git.GetLastWorkingDay(time.Now())
	
	// Get recent commits
	commits, err := gitClient.GetRecentCommits(cfg.LocalRepoPath, lastWorkingDay)
	if err != nil {
		return handleError(fmt.Errorf("failed to get recent commits: %w", err), outputFormat)
	}

	// Extract work items from commits
	workItems := git.ExtractWorkItems(commits)

	// Create suggestion
	suggestion := standup.JSONSuggestion{
		Date:      time.Now().Format("2006-01-02"),
		Yesterday: workItems,
		Today:     []string{"Continue work from yesterday", "Review pull requests"},
		Blockers:  "None",
	}

	// Add commit info
	for _, commit := range commits {
		suggestion.Based_on.Commits = append(suggestion.Based_on.Commits, standup.CommitInfo{
			SHA:     commit.SHA[:7],
			Date:    commit.Date.Format("2006-01-02 15:04:05"),
			Author:  commit.Author,
			Message: commit.Subject,
		})
	}

	// Output suggestion
	if outputFormat == "json" {
		jsonStr, err := standup.FormatJSONSuggestion(suggestion)
		if err != nil {
			return handleError(err, outputFormat)
		}
		fmt.Println(jsonStr)
	} else {
		// Human-readable format
		fmt.Println("Suggested standup content based on recent commits:")
		fmt.Println("\nYesterday:")
		if len(suggestion.Yesterday) == 0 {
			fmt.Println("- No commits found")
		} else {
			for _, item := range suggestion.Yesterday {
				fmt.Printf("- %s\n", item)
			}
		}
		fmt.Println("\nToday:")
		for _, item := range suggestion.Today {
			fmt.Printf("- %s\n", item)
		}
		fmt.Printf("\nBlockers: %s\n", suggestion.Blockers)
		
		if len(suggestion.Based_on.Commits) > 0 {
			fmt.Printf("\nBased on %d commits since %s\n", len(suggestion.Based_on.Commits), lastWorkingDay.Format("2006-01-02"))
		}
	}

	return nil
}