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
		return fmt.Errorf("GitHub CLI not found: %w. Please install it from https://cli.github.com/", err)
	}

	// Verify it's actually gh by checking output
	if !strings.Contains(string(output), "gh version") {
		return fmt.Errorf("gh command found but appears to be incorrect: output=%s", string(output))
	}

	return nil
}

// CheckAuthenticated checks if user is authenticated with GitHub
func (c *Client) CheckAuthenticated() error {
	_, err := c.runner.Run("gh", "auth", "status")
	if err != nil {
		return fmt.Errorf("not authenticated with GitHub: %w. Please run 'gh auth login'", err)
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
	// Check if this is an empty repository
	if isEmpty, err := c.isEmptyRepository(repoPath); err != nil {
		return fmt.Errorf("failed to check repository state: %w", err)
	} else if isEmpty {
		return nil // Nothing to sync in empty repo
	}

	// Fetch all changes
	if err := c.fetchAll(repoPath); err != nil {
		return fmt.Errorf("failed to fetch remote changes: %w", err)
	}

	// Get current branch
	branch, err := c.getCurrentBranch(repoPath)
	if err != nil {
		return fmt.Errorf("failed to determine current branch: %w", err)
	}

	if branch == "" {
		return nil // Detached HEAD state, nothing to sync
	}

	// Check if remote branch exists
	if exists, err := c.remoteBranchExists(repoPath, branch); err != nil {
		return fmt.Errorf("failed to check remote branch: %w", err)
	} else if !exists {
		return nil // Remote branch doesn't exist yet
	}

	// Reset to remote branch
	if err := c.resetToRemote(repoPath, branch); err != nil {
		return fmt.Errorf("failed to sync with remote: %w", err)
	}

	return nil
}

// isEmptyRepository checks if the repository has any branches
func (c *Client) isEmptyRepository(repoPath string) (bool, error) {
	output, err := c.runner.RunInDir(repoPath, "git", "branch", "-a")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(output)) == "", nil
}

// fetchAll fetches all remote changes
func (c *Client) fetchAll(repoPath string) error {
	output, err := c.runner.RunInDir(repoPath, "git", "fetch", "--all")
	if err != nil {
		return fmt.Errorf("%w (output: %s)", err, string(output))
	}
	return nil
}

// getCurrentBranch returns the current branch name
func (c *Client) getCurrentBranch(repoPath string) (string, error) {
	output, err := c.runner.RunInDir(repoPath, "git", "branch", "--show-current")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// remoteBranchExists checks if a branch exists on the remote
func (c *Client) remoteBranchExists(repoPath, branch string) (bool, error) {
	_, err := c.runner.RunInDir(repoPath, "git", "rev-parse", fmt.Sprintf("origin/%s", branch))
	return err == nil, nil
}

// resetToRemote resets the current branch to match the remote
func (c *Client) resetToRemote(repoPath, branch string) error {
	output, err := c.runner.RunInDir(repoPath, "git", "reset", "--hard", fmt.Sprintf("origin/%s", branch))
	if err != nil {
		return fmt.Errorf("%w (output: %s)", err, string(output))
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
	// Stage all changes
	if err := c.stageAllChanges(repoPath); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Check if there are changes to commit
	hasChanges, err := c.hasUncommittedChanges(repoPath)
	if err != nil {
		return fmt.Errorf("failed to check for changes: %w", err)
	}

	if !hasChanges {
		return ErrNoChangesToCommit
	}

	// Commit changes
	if err := c.createCommit(repoPath, message); err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	// Get or create branch
	branch, err := c.ensureBranch(repoPath)
	if err != nil {
		return fmt.Errorf("failed to determine branch: %w", err)
	}

	// Push changes
	if err := c.pushWithUpstream(repoPath, branch); err != nil {
		return fmt.Errorf("failed to push to remote: %w", err)
	}

	return nil
}

// ErrNoChangesToCommit indicates there are no changes to commit
var ErrNoChangesToCommit = fmt.Errorf("no changes to commit")

// stageAllChanges adds all changes to the staging area
func (c *Client) stageAllChanges(repoPath string) error {
	output, err := c.runner.RunInDir(repoPath, "git", "add", ".")
	if err != nil {
		return fmt.Errorf("%w (output: %s)", err, string(output))
	}
	return nil
}

// hasUncommittedChanges checks if there are uncommitted changes
func (c *Client) hasUncommittedChanges(repoPath string) (bool, error) {
	output, err := c.runner.RunInDir(repoPath, "git", "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return len(bytes.TrimSpace(output)) > 0, nil
}

// createCommit creates a commit with the given message
func (c *Client) createCommit(repoPath, message string) error {
	output, err := c.runner.RunInDir(repoPath, "git", "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("%w (output: %s)", err, string(output))
	}
	return nil
}

// ensureBranch gets the current branch or creates main if in detached HEAD
func (c *Client) ensureBranch(repoPath string) (string, error) {
	branch, err := c.getCurrentBranch(repoPath)
	if err != nil {
		return "", err
	}

	if branch == "" {
		// Detached HEAD state, create main branch
		branch = "main"
		output, err := c.runner.RunInDir(repoPath, "git", "checkout", "-b", branch)
		if err != nil {
			return "", fmt.Errorf("failed to create main branch: %w (output: %s)", err, string(output))
		}
	}

	return branch, nil
}

// pushWithUpstream pushes the branch and sets upstream if needed
func (c *Client) pushWithUpstream(repoPath, branch string) error {
	output, err := c.runner.RunInDir(repoPath, "git", "push", "-u", "origin", branch)
	if err != nil {
		return fmt.Errorf("%w (output: %s)", err, string(output))
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

// PullRequestOptions contains options for creating a pull request
type PullRequestOptions struct {
	Title string
	Body  string
	Base  string
}

// CreatePullRequest creates a pull request using GitHub CLI
func (c *Client) CreatePullRequest(repoPath, title, body string) error {
	opts := PullRequestOptions{
		Title: title,
		Body:  body,
		Base:  "main",
	}
	return c.CreatePullRequestWithOptions(repoPath, opts)
}

// CreatePullRequestWithOptions creates a pull request with custom options
func (c *Client) CreatePullRequestWithOptions(repoPath string, opts PullRequestOptions) error {
	args := []string{"pr", "create", "--title", opts.Title, "--body", opts.Body}
	if opts.Base != "" {
		args = append(args, "--base", opts.Base)
	}

	output, err := c.runner.RunInDir(repoPath, "gh", args...)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w (output: %s)", err, string(output))
	}
	return nil
}

// MergeOptions contains options for merging a pull request
type MergeOptions struct {
	Auto         bool
	Squash       bool
	DeleteBranch bool
}

// MergePullRequest merges a pull request using GitHub CLI with default options
func (c *Client) MergePullRequest(repoPath string) error {
	opts := MergeOptions{
		Auto:         true,
		Squash:       true,
		DeleteBranch: true,
	}
	return c.MergePullRequestWithOptions(repoPath, opts)
}

// MergePullRequestWithOptions merges a pull request with custom options
func (c *Client) MergePullRequestWithOptions(repoPath string, opts MergeOptions) error {
	args := []string{"pr", "merge"}
	if opts.Auto {
		args = append(args, "--auto")
	}
	if opts.Squash {
		args = append(args, "--squash")
	}
	if opts.DeleteBranch {
		args = append(args, "--delete-branch")
	}

	output, err := c.runner.RunInDir(repoPath, "gh", args...)
	if err != nil {
		return fmt.Errorf("failed to merge pull request: %w (output: %s)", err, string(output))
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

// PRInfo contains information about a pull request
type PRInfo struct {
	Exists bool
	Number string
}

// PRExistsForBranch checks if a PR exists for the given branch
func (c *Client) PRExistsForBranch(repoPath, branchName string) (bool, string) {
	info := c.GetPRInfoForBranch(repoPath, branchName)
	return info.Exists, info.Number
}

// GetPRInfoForBranch retrieves PR information for a specific branch
func (c *Client) GetPRInfoForBranch(repoPath, branchName string) PRInfo {
	output, err := c.runner.RunInDir(repoPath, "gh", "pr", "list", 
		"--head", branchName, 
		"--json", "number", 
		"--jq", ".[0].number")
	
	if err != nil || strings.TrimSpace(string(output)) == "" {
		return PRInfo{Exists: false, Number: ""}
	}
	
	return PRInfo{
		Exists: true,
		Number: strings.TrimSpace(string(output)),
	}
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