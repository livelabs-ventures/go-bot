package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CommandRunner interface for executing commands (allows mocking in tests)
type CommandRunner interface {
	Run(name string, args ...string) ([]byte, error)
	RunInDir(dir, name string, args ...string) ([]byte, error)
}

// RealCommandRunner implements CommandRunner using actual system commands
type RealCommandRunner struct{}

// Run executes a command and returns its output
func (r *RealCommandRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

// RunInDir executes a command in a specific directory
func (r *RealCommandRunner) RunInDir(dir, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

// Client handles Git operations via GitHub CLI
type Client struct {
	runner CommandRunner
}

// NewClient creates a new Git client
func NewClient() *Client {
	return &Client{
		runner: &RealCommandRunner{},
	}
}

// NewClientWithRunner creates a new Git client with a custom command runner
func NewClientWithRunner(runner CommandRunner) *Client {
	return &Client{
		runner: runner,
	}
}

// CheckGHInstalled checks if GitHub CLI is installed
func (c *Client) CheckGHInstalled() error {
	output, err := c.runner.Run("gh", "--version")
	if err != nil {
		return fmt.Errorf("GitHub CLI not found. Please install it from https://cli.github.com/")
	}

	// Verify it's actually gh by checking output
	if !strings.Contains(string(output), "gh version") {
		return fmt.Errorf("gh command found but appears to be incorrect")
	}

	return nil
}

// CheckAuthenticated checks if user is authenticated with GitHub
func (c *Client) CheckAuthenticated() error {
	_, err := c.runner.Run("gh", "auth", "status")
	if err != nil {
		return fmt.Errorf("not authenticated with GitHub. Please run 'gh auth login'")
	}
	return nil
}

// CloneRepository clones a repository to the specified path
func (c *Client) CloneRepository(repo, targetPath string) error {
	// Ensure parent directory exists
	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Clone the repository
	output, err := c.runner.Run("gh", "repo", "clone", repo, targetPath)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// SyncRepository syncs the repository with the remote
func (c *Client) SyncRepository(repoPath string) error {
	// First, check if there are any branches (empty repo check)
	branchListOutput, err := c.runner.RunInDir(repoPath, "git", "branch", "-a")
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}
	
	// If no branches, this is an empty repo - nothing to sync
	if strings.TrimSpace(string(branchListOutput)) == "" {
		return nil
	}

	// Fetch all changes
	output, err := c.runner.RunInDir(repoPath, "git", "fetch", "--all")
	if err != nil {
		return fmt.Errorf("failed to fetch: %w\nOutput: %s", err, string(output))
	}

	// Get current branch
	branchOutput, err := c.runner.RunInDir(repoPath, "git", "branch", "--show-current")
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	branch := strings.TrimSpace(string(branchOutput))

	// If no branch (detached HEAD or empty repo), nothing to sync
	if branch == "" {
		return nil
	}

	// Check if origin/branch exists
	_, err = c.runner.RunInDir(repoPath, "git", "rev-parse", fmt.Sprintf("origin/%s", branch))
	if err != nil {
		// Remote branch doesn't exist yet, nothing to sync
		return nil
	}

	// Reset to origin/branch
	output, err = c.runner.RunInDir(repoPath, "git", "reset", "--hard", fmt.Sprintf("origin/%s", branch))
	if err != nil {
		return fmt.Errorf("failed to reset to origin: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// AddAll adds all changes to staging
func (c *Client) AddAll(repoPath string) ([]byte, error) {
	return c.runner.RunInDir(repoPath, "git", "add", ".")
}

// Commit creates a commit with the given message
func (c *Client) Commit(repoPath, message string) ([]byte, error) {
	return c.runner.RunInDir(repoPath, "git", "commit", "-m", message)
}

// CommitAndPush commits changes and pushes to remote
func (c *Client) CommitAndPush(repoPath, message string) error {
	// Add all changes
	output, err := c.runner.RunInDir(repoPath, "git", "add", ".")
	if err != nil {
		return fmt.Errorf("failed to add changes: %w\nOutput: %s", err, string(output))
	}

	// Check if there are changes to commit
	statusOutput, err := c.runner.RunInDir(repoPath, "git", "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("failed to check status: %w", err)
	}

	if len(bytes.TrimSpace(statusOutput)) == 0 {
		return fmt.Errorf("no changes to commit")
	}

	// Commit changes
	output, err = c.runner.RunInDir(repoPath, "git", "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("failed to commit: %w\nOutput: %s", err, string(output))
	}

	// Get current branch
	branchOutput, err := c.runner.RunInDir(repoPath, "git", "branch", "--show-current")
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	branch := strings.TrimSpace(string(branchOutput))

	// If no branch, we need to create main branch
	if branch == "" {
		branch = "main"
		// Create main branch
		output, err = c.runner.RunInDir(repoPath, "git", "checkout", "-b", branch)
		if err != nil {
			return fmt.Errorf("failed to create main branch: %w\nOutput: %s", err, string(output))
		}
	}

	// Push changes (use -u for first push to set upstream)
	output, err = c.runner.RunInDir(repoPath, "git", "push", "-u", "origin", branch)
	if err != nil {
		return fmt.Errorf("failed to push: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// RepositoryExists checks if a repository exists at the given path
func (c *Client) RepositoryExists(repoPath string) bool {
	gitDir := filepath.Join(repoPath, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// CreateBranch creates a new branch and switches to it
func (c *Client) CreateBranch(repoPath, branchName string) error {
	output, err := c.runner.RunInDir(repoPath, "git", "checkout", "-b", branchName)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// PushBranch pushes a branch to remote
func (c *Client) PushBranch(repoPath, branchName string) error {
	output, err := c.runner.RunInDir(repoPath, "git", "push", "-u", "origin", branchName)
	if err != nil {
		return fmt.Errorf("failed to push branch: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// CreatePullRequest creates a pull request using GitHub CLI
func (c *Client) CreatePullRequest(repoPath, title, body string) error {
	// Create the PR with title and body
	output, err := c.runner.RunInDir(repoPath, "gh", "pr", "create", 
		"--title", title,
		"--body", body,
		"--base", "main")
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// MergePullRequest merges a pull request using GitHub CLI
func (c *Client) MergePullRequest(repoPath string) error {
	// Auto-merge the PR with squash
	output, err := c.runner.RunInDir(repoPath, "gh", "pr", "merge", "--auto", "--squash", "--delete-branch")
	if err != nil {
		return fmt.Errorf("failed to merge pull request: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// SwitchToMainBranch switches back to the main branch
func (c *Client) SwitchToMainBranch(repoPath string) error {
	output, err := c.runner.RunInDir(repoPath, "git", "checkout", "main")
	if err != nil {
		return fmt.Errorf("failed to switch to main branch: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// BranchExists checks if a branch exists locally
func (c *Client) BranchExists(repoPath, branchName string) bool {
	output, err := c.runner.RunInDir(repoPath, "git", "branch", "--list", branchName)
	return err == nil && strings.TrimSpace(string(output)) != ""
}

// RemoteBranchExists checks if a branch exists on remote
func (c *Client) RemoteBranchExists(repoPath, branchName string) bool {
	output, err := c.runner.RunInDir(repoPath, "git", "ls-remote", "--heads", "origin", branchName)
	return err == nil && strings.TrimSpace(string(output)) != ""
}

// SwitchToBranch switches to an existing branch
func (c *Client) SwitchToBranch(repoPath, branchName string) error {
	output, err := c.runner.RunInDir(repoPath, "git", "checkout", branchName)
	if err != nil {
		return fmt.Errorf("failed to switch to branch: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// PullBranch pulls changes from remote branch
func (c *Client) PullBranch(repoPath, branchName string) error {
	output, err := c.runner.RunInDir(repoPath, "git", "pull", "origin", branchName)
	if err != nil {
		return fmt.Errorf("failed to pull branch: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// PRExistsForBranch checks if a PR exists for the given branch
func (c *Client) PRExistsForBranch(repoPath, branchName string) (bool, string) {
	output, err := c.runner.RunInDir(repoPath, "gh", "pr", "list", "--head", branchName, "--json", "number", "--jq", ".[0].number")
	if err != nil || strings.TrimSpace(string(output)) == "" {
		return false, ""
	}
	return true, strings.TrimSpace(string(output))
}

// UpdatePullRequest updates the body of an existing PR
func (c *Client) UpdatePullRequest(repoPath, prNumber, body string) error {
	output, err := c.runner.RunInDir(repoPath, "gh", "pr", "edit", prNumber, "--body", body)
	if err != nil {
		return fmt.Errorf("failed to update pull request: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// MergePullRequestByNumber merges a PR by its number
func (c *Client) MergePullRequestByNumber(repoPath, prNumber string) error {
	output, err := c.runner.RunInDir(repoPath, "gh", "pr", "merge", prNumber, "--squash", "--delete-branch")
	if err != nil {
		return fmt.Errorf("failed to merge pull request: %w\nOutput: %s", err, string(output))
	}
	return nil
}