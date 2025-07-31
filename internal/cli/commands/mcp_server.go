package commands

import (
	"fmt"
	"os"
	"time"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"github.com/standup-bot/standup-bot/pkg/config"
	"github.com/standup-bot/standup-bot/pkg/git"
	"github.com/standup-bot/standup-bot/pkg/standup"
)

// SubmitStandupArgs represents arguments for submit_standup tool
type SubmitStandupArgs struct {
	Yesterday []string `json:"yesterday" jsonschema:"required,description=List of tasks completed yesterday"`
	Today     []string `json:"today" jsonschema:"required,description=List of tasks planned for today"`
	Blockers  string   `json:"blockers" jsonschema:"description=Any blockers or impediments (default: None)"`
	Direct    bool     `json:"direct" jsonschema:"description=Use direct commit workflow instead of PR workflow (default: false)"`
}

// CreateStandupPRArgs represents arguments for create_standup_pr tool
type CreateStandupPRArgs struct {
	Merge bool `json:"merge" jsonschema:"description=Whether to merge the PR after creation (default: false)"`
}

// GetStandupStatusArgs represents arguments for get_standup_status tool
type GetStandupStatusArgs struct{}

// RunMCPServer starts the MCP server
func RunMCPServer() error {
	// Create a channel to keep the server running
	done := make(chan struct{})

	// Create MCP server with stdio transport
	server := mcp.NewServer(
		stdio.NewStdioServerTransport(),
		mcp.WithName("standup-bot-mcp"),
		mcp.WithVersion("1.0.0"),
	)

	// Register submit_standup tool
	err := server.RegisterTool(
		"submit_standup",
		"Submit daily standup with yesterday's accomplishments, today's plans, and blockers",
		handleSubmitStandup,
	)
	if err != nil {
		return fmt.Errorf("failed to register submit_standup tool: %w", err)
	}

	// Register create_standup_pr tool
	err = server.RegisterTool(
		"create_standup_pr",
		"Create a pull request with standup entries for the day",
		handleCreateStandupPR,
	)
	if err != nil {
		return fmt.Errorf("failed to register create_standup_pr tool: %w", err)
	}

	// Register get_standup_status tool
	err = server.RegisterTool(
		"get_standup_status",
		"Check if today's standup has been completed",
		handleGetStandupStatus,
	)
	if err != nil {
		return fmt.Errorf("failed to register get_standup_status tool: %w", err)
	}

	// Log to stderr to avoid interfering with stdio transport
	fmt.Fprintln(os.Stderr, "Starting standup-bot MCP server...")

	// Serve the MCP server
	err = server.Serve()
	if err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	// Block to keep the server running
	<-done
	return nil
}

// handleSubmitStandup handles the submit_standup tool
func handleSubmitStandup(args SubmitStandupArgs) (*mcp.ToolResponse, error) {
	// Set default blockers if empty
	if args.Blockers == "" {
		args.Blockers = "None"
	}

	// Create standup entry
	entry := &standup.Entry{
		Date:      time.Now(),
		Yesterday: args.Yesterday,
		Today:     args.Today,
		Blockers:  args.Blockers,
	}

	// Load configuration
	cfgManager, err := config.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize config manager: %w", err)
	}

	cfg, err := cfgManager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Submit standup
	var result string
	if args.Direct {
		err = submitStandupDirect(cfg, entry)
		if err != nil {
			return nil, err
		}
		result = fmt.Sprintf("Standup submitted successfully via direct commit for %s", entry.Date.Format("2006-01-02"))
	} else {
		prInfo, err := submitStandupPR(cfg, entry)
		if err != nil {
			return nil, err
		}
		result = fmt.Sprintf("Standup submitted successfully via PR #%s for %s", prInfo.Number, entry.Date.Format("2006-01-02"))
	}

	return mcp.NewToolResponse(
		mcp.NewTextContent(result),
	), nil
}

// handleCreateStandupPR handles the create_standup_pr tool
func handleCreateStandupPR(args CreateStandupPRArgs) (*mcp.ToolResponse, error) {
	// Load configuration
	cfgManager, err := config.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize config manager: %w", err)
	}

	cfg, err := cfgManager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	gitClient := git.NewClient()
	
	// Validate environment
	if err := validateEnvironment(gitClient, cfg); err != nil {
		return nil, err
	}

	// Check if there's a PR for today
	date := time.Now()
	branchName := fmt.Sprintf("standup/%s", date.Format("2006-01-02"))
	
	prExists, prNumber := gitClient.PRExistsForBranch(cfg.LocalRepoPath, branchName)
	if !prExists {
		return nil, fmt.Errorf("no standup PR found for today (%s)", date.Format("2006-01-02"))
	}

	result := fmt.Sprintf("Standup PR #%s exists for %s", prNumber, date.Format("2006-01-02"))

	// Merge if requested
	if args.Merge {
		if err := gitClient.MergePullRequestByNumber(cfg.LocalRepoPath, prNumber); err != nil {
			return nil, fmt.Errorf("failed to merge PR: %w", err)
		}
		result += " and has been merged"
	}

	return mcp.NewToolResponse(
		mcp.NewTextContent(result),
	), nil
}

// handleGetStandupStatus handles the get_standup_status tool
func handleGetStandupStatus(args GetStandupStatusArgs) (*mcp.ToolResponse, error) {
	// Load configuration
	cfgManager, err := config.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize config manager: %w", err)
	}

	cfg, err := cfgManager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if standup file exists for today
	standupManager := standup.NewManager(cfg.LocalRepoPath)
	filePath, err := standupManager.GetStandupFilePath(cfg.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get standup file path: %w", err)
	}

	// Read the file and check for today's entry
	hasToday, err := checkTodayStandup(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to check standup status: %w", err)
	}

	status := "incomplete"
	message := fmt.Sprintf("No standup found for today (%s)", time.Now().Format("2006-01-02"))
	
	if hasToday {
		status = "complete"
		message = fmt.Sprintf("Standup completed for today (%s)", time.Now().Format("2006-01-02"))
	}

	// Also check for PR
	gitClient := git.NewClient()
	branchName := fmt.Sprintf("standup/%s", time.Now().Format("2006-01-02"))
	prExists, prNumber := gitClient.PRExistsForBranch(cfg.LocalRepoPath, branchName)
	
	if prExists {
		message += fmt.Sprintf(" - PR #%s exists", prNumber)
	}

	return mcp.NewToolResponse(
		mcp.NewTextContent(fmt.Sprintf("Status: %s\n%s", status, message)),
	), nil
}

// submitStandupDirect handles direct commit workflow
func submitStandupDirect(cfg *config.Config, entry *standup.Entry) error {
	gitClient := git.NewClient()

	if err := validateEnvironment(gitClient, cfg); err != nil {
		return err
	}

	// Sync repository
	if err := gitClient.SyncRepository(cfg.LocalRepoPath); err != nil {
		return fmt.Errorf("failed to sync repository: %w", err)
	}

	// Save entry
	standupManager := standup.NewManager(cfg.LocalRepoPath)
	if err := standupManager.SaveEntry(entry, cfg.Name); err != nil {
		return fmt.Errorf("failed to save standup: %w", err)
	}

	// Commit and push
	commitMessage := standupManager.FormatCommitMessage(entry, cfg.Name)
	if err := gitClient.CommitAndPush(cfg.LocalRepoPath, commitMessage); err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}

	return nil
}

// submitStandupPR handles PR workflow
func submitStandupPR(cfg *config.Config, entry *standup.Entry) (*PRInfo, error) {
	gitClient := git.NewClient()

	if err := validateEnvironment(gitClient, cfg); err != nil {
		return nil, err
	}

	// Sync repository
	if err := gitClient.SyncRepository(cfg.LocalRepoPath); err != nil {
		return nil, fmt.Errorf("failed to sync repository: %w", err)
	}

	// Ensure main branch exists
	if err := ensureMainBranch(cfg.LocalRepoPath, gitClient); err != nil {
		return nil, err
	}

	// Save and create PR
	standupManager := standup.NewManager(cfg.LocalRepoPath)
	return createOrUpdateStandupPR(cfg, gitClient, standupManager, entry, "json")
}

// checkTodayStandup checks if today's standup exists in the file
func checkTodayStandup(filePath string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	todayStr := time.Now().Format("2006-01-02")
	dateHeader := fmt.Sprintf("## %s", todayStr)
	
	// Check if content contains today's date header
	contentStr := string(content)
	return containsString(contentStr, dateHeader), nil
}

// containsString is a simple string contains check
func containsString(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}