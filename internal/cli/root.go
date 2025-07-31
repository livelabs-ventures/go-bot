package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/standup-bot/standup-bot/pkg/config"
	"github.com/standup-bot/standup-bot/pkg/git"
	"github.com/standup-bot/standup-bot/pkg/standup"
)

var (
	configFlag bool
	directFlag bool
	mergeFlag  bool
	rootCmd    = &cobra.Command{
		Use:   "standup-bot",
		Short: "A simple CLI tool for daily standup updates via GitHub",
		Long: `Standup Bot facilitates daily standup updates via GitHub.
It collects standup information and commits it to a shared repository,
where GitHub-Slack integration broadcasts updates to the team channel.`,
		RunE: runStandup,
	}
)

func init() {
	rootCmd.Flags().BoolVar(&configFlag, "config", false, "Run configuration setup")
	rootCmd.Flags().BoolVar(&directFlag, "direct", false, "Use direct commit workflow (multi-line commit message)")
	rootCmd.Flags().BoolVar(&mergeFlag, "merge", false, "Merge today's standup pull request")
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func runStandup(cmd *cobra.Command, args []string) error {
	// Create configuration manager
	cfgManager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize config manager: %w", err)
	}

	// Check if we need to run configuration
	if configFlag || !cfgManager.Exists() {
		return runConfiguration(cfgManager)
	}

	// Load configuration
	cfg, err := cfgManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Handle merge command
	if mergeFlag {
		return runMergeDailyStandup(cfg)
	}

	// Run the standup workflow
	if directFlag {
		return runStandupWorkflow(cfg)
	}
	return runStandupWorkflowPR(cfg)
}

func runConfiguration(cfgManager *config.Manager) error {
	fmt.Println("Welcome to Standup Bot!")
	
	// Get repository
	fmt.Print("GitHub Repository (e.g., org/standup-repo): ")
	var repo string
	if _, err := fmt.Scanln(&repo); err != nil {
		return fmt.Errorf("failed to read repository: %w", err)
	}

	// Get user name
	fmt.Print("Your Name: ")
	var name string
	if _, err := fmt.Scanln(&name); err != nil {
		return fmt.Errorf("failed to read name: %w", err)
	}

	// Create configuration
	cfg := config.DefaultConfig()
	cfg.Repository = repo
	cfg.Name = name

	// Save configuration
	if err := cfgManager.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("Configuration saved!")

	// Clone repository if it doesn't exist
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
	expandedPath := cfg.LocalRepoPath
	if expandedPath[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		expandedPath = filepath.Join(homeDir, expandedPath[1:])
	}

	// Clone repository if needed
	if !gitClient.RepositoryExists(expandedPath) {
		fmt.Println("Cloning repository...")
		if err := gitClient.CloneRepository(cfg.Repository, expandedPath); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
		fmt.Println("Repository cloned successfully!")
	}

	fmt.Println("\nSetup complete! Run 'standup-bot' to record your standup.")
	return nil
}

func runStandupWorkflow(cfg *config.Config) error {
	// Create git client
	gitClient := git.NewClient()

	// Check GitHub CLI is installed
	if err := gitClient.CheckGHInstalled(); err != nil {
		return err
	}

	// Check authentication
	if err := gitClient.CheckAuthenticated(); err != nil {
		return err
	}

	// Check if repository exists locally
	if !gitClient.RepositoryExists(cfg.LocalRepoPath) {
		return fmt.Errorf("repository not found at %s. Please run 'standup-bot --config' to set up", cfg.LocalRepoPath)
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
		// If push fails, we should save the standup to a temp file
		tempFile := saveTempStandup(entry, cfg.Name)
		return fmt.Errorf("failed to push changes: %w\nYour standup has been saved to: %s", err, tempFile)
	}

	fmt.Println(" Standup recorded successfully!")
	return nil
}

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

func runStandupWorkflowPR(cfg *config.Config) error {
	// Create git client
	gitClient := git.NewClient()

	// Check GitHub CLI is installed
	if err := gitClient.CheckGHInstalled(); err != nil {
		return err
	}

	// Check authentication
	if err := gitClient.CheckAuthenticated(); err != nil {
		return err
	}

	// Check if repository exists locally
	if !gitClient.RepositoryExists(cfg.LocalRepoPath) {
		return fmt.Errorf("repository not found at %s. Please run 'standup-bot --config' to set up", cfg.LocalRepoPath)
	}

	// Sync repository
	fmt.Println("Syncing repository...")
	if err := gitClient.SyncRepository(cfg.LocalRepoPath); err != nil {
		return fmt.Errorf("failed to sync repository: %w", err)
	}

	// Switch to main branch first
	if err := gitClient.SwitchToMainBranch(cfg.LocalRepoPath); err != nil {
		// If main doesn't exist, we might be in an empty repo
		// Continue anyway as the branch will be created
	}

	// Collect standup entry
	standupManager := standup.NewManager(cfg.LocalRepoPath)
	entry, err := standupManager.CollectEntry(os.Stdin, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to collect standup: %w", err)
	}

	// Create shared daily branch name
	branchName := fmt.Sprintf("standup/%s", entry.Date.Format("2006-01-02"))
	
	// Check if branch already exists
	branchExists := gitClient.BranchExists(cfg.LocalRepoPath, branchName)
	
	if branchExists {
		// Switch to existing branch
		fmt.Println("Switching to today's standup branch...")
		if err := gitClient.SwitchToBranch(cfg.LocalRepoPath, branchName); err != nil {
			return fmt.Errorf("failed to switch to branch: %w", err)
		}
		
		// Pull latest changes from the branch
		fmt.Println("Pulling latest updates...")
		if err := gitClient.PullBranch(cfg.LocalRepoPath, branchName); err != nil {
			// If pull fails (e.g., branch doesn't exist on remote yet), continue
			fmt.Println("Note: Could not pull from remote branch (this is normal for new branches)")
		}
	} else {
		// Create and switch to new branch
		fmt.Println("Creating today's standup branch...")
		if err := gitClient.CreateBranch(cfg.LocalRepoPath, branchName); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}
	}

	// Save entry to file
	fmt.Println("Recording standup...")
	if err := standupManager.SaveEntry(entry, cfg.Name); err != nil {
		return fmt.Errorf("failed to save standup: %w", err)
	}

	// Add and commit changes (without push)
	_, err = gitClient.AddAll(cfg.LocalRepoPath)
	if err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}
	
	commitMessage := fmt.Sprintf("[Standup] %s - %s", cfg.Name, entry.Date.Format("2006-01-02"))
	_, err = gitClient.Commit(cfg.LocalRepoPath, commitMessage)
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	
	// Push the branch
	fmt.Println("Pushing branch...")
	if err := gitClient.PushBranch(cfg.LocalRepoPath, branchName); err != nil {
		tempFile := saveTempStandup(entry, cfg.Name)
		return fmt.Errorf("failed to push changes: %w\nYour standup has been saved to: %s", err, tempFile)
	}

	// Check if PR already exists for today
	prExists, prNumber := gitClient.PRExistsForBranch(cfg.LocalRepoPath, branchName)
	
	if prExists {
		// Update existing PR
		fmt.Printf("Updating existing pull request #%s...\n", prNumber)
		prBody := formatDailyPRBody(cfg.LocalRepoPath, entry.Date)
		if err := gitClient.UpdatePullRequest(cfg.LocalRepoPath, prNumber, prBody); err != nil {
			fmt.Printf("Warning: Could not update PR body: %v\n", err)
		}
	} else {
		// Create new PR
		fmt.Println("Creating pull request...")
		prTitle := fmt.Sprintf("[Standup] %s", entry.Date.Format("2006-01-02"))
		prBody := formatDailyPRBody(cfg.LocalRepoPath, entry.Date)
		
		if err := gitClient.CreatePullRequest(cfg.LocalRepoPath, prTitle, prBody); err != nil {
			return fmt.Errorf("failed to create pull request: %w", err)
		}
	}

	fmt.Println("✅ Standup recorded successfully!")
	fmt.Println("💡 To merge today's standups, run: standup-bot --merge")
	return nil
}

func formatPRBody(entry *standup.Entry, userName string) string {
	body := fmt.Sprintf("## Daily Standup - %s\n\n", userName)
	body += fmt.Sprintf("**Date**: %s\n\n", entry.Date.Format("2006-01-02"))
	
	body += "### Yesterday\n"
	for _, item := range entry.Yesterday {
		body += fmt.Sprintf("- %s\n", item)
	}
	
	body += "\n### Today\n"
	for _, item := range entry.Today {
		body += fmt.Sprintf("- %s\n", item)
	}
	
	body += "\n### Blockers\n"
	if entry.Blockers == "" {
		body += "None\n"
	} else {
		body += entry.Blockers + "\n"
	}
	
	return body
}

func formatDailyPRBody(repoPath string, date time.Time) string {
	body := fmt.Sprintf("# Daily Standups - %s\n\n", date.Format("2006-01-02"))
	
	// Read all standup files for today
	standupDir := filepath.Join(repoPath, "stand-ups")
	files, err := os.ReadDir(standupDir)
	if err != nil {
		return body + "Error reading standup files\n"
	}
	
	// Collect all standups for today
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		
		// Extract user name from filename
		userName := strings.TrimSuffix(file.Name(), ".md")
		
		// Read the file to get today's entry
		content, err := os.ReadFile(filepath.Join(standupDir, file.Name()))
		if err != nil {
			continue
		}
		
		// Parse today's standup from the content
		if todayStandup := extractTodayStandup(string(content), date); todayStandup != "" {
			body += fmt.Sprintf("## %s\n\n%s\n\n---\n\n", userName, todayStandup)
		}
	}
	
	body += "\n💡 To merge this PR, run: `standup-bot --merge`\n"
	
	return body
}

func extractTodayStandup(content string, date time.Time) string {
	dateStr := date.Format("2006-01-02")
	lines := strings.Split(content, "\n")
	
	inTodaySection := false
	var todayContent []string
	sectionCount := 0
	
	for _, line := range lines {
		// Check if we found a date header matching today
		if strings.HasPrefix(line, "## ") && strings.Contains(line, dateStr) {
			inTodaySection = true
			sectionCount++
			// Only process the first matching section
			if sectionCount > 1 {
				break
			}
			continue
		}
		
		// Check if we hit the next date section or separator (stop collecting)
		if inTodaySection {
			if strings.HasPrefix(line, "## ") || line == "---" {
				break
			}
		}
		
		// Collect content for today
		if inTodaySection {
			todayContent = append(todayContent, line)
		}
	}
	
	return strings.TrimSpace(strings.Join(todayContent, "\n"))
}

func runMergeDailyStandup(cfg *config.Config) error {
	// Create git client
	gitClient := git.NewClient()

	// Check GitHub CLI is installed
	if err := gitClient.CheckGHInstalled(); err != nil {
		return err
	}

	// Check authentication
	if err := gitClient.CheckAuthenticated(); err != nil {
		return err
	}

	// Check if repository exists locally
	if !gitClient.RepositoryExists(cfg.LocalRepoPath) {
		return fmt.Errorf("repository not found at %s. Please run 'standup-bot --config' to set up", cfg.LocalRepoPath)
	}

	// Get today's branch name
	today := time.Now()
	branchName := fmt.Sprintf("standup/%s", today.Format("2006-01-02"))
	
	// Check if PR exists for today
	prExists, prNumber := gitClient.PRExistsForBranch(cfg.LocalRepoPath, branchName)
	
	if !prExists {
		return fmt.Errorf("no pull request found for today's standups")
	}
	
	// Merge the PR
	fmt.Printf("Merging pull request #%s...\n", prNumber)
	if err := gitClient.MergePullRequestByNumber(cfg.LocalRepoPath, prNumber); err != nil {
		return fmt.Errorf("failed to merge pull request: %w", err)
	}
	
	fmt.Println("✅ Today's standups have been merged successfully!")
	
	// Switch back to main branch locally
	fmt.Println("Switching back to main branch...")
	if err := gitClient.SwitchToMainBranch(cfg.LocalRepoPath); err != nil {
		fmt.Printf("Warning: Could not switch to main branch: %v\n", err)
	}
	
	// Pull latest changes
	if err := gitClient.SyncRepository(cfg.LocalRepoPath); err != nil {
		fmt.Printf("Warning: Could not sync repository: %v\n", err)
	}
	
	return nil
}