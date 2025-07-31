package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// MockCommandRunner implements CommandRunner for testing
type MockCommandRunner struct {
	Commands []MockCommand
	Index    int
}

type MockCommand struct {
	Name   string
	Args   []string
	Dir    string
	Output []byte
	Error  error
}

func (m *MockCommandRunner) Run(name string, args ...string) ([]byte, error) {
	if m.Index >= len(m.Commands) {
		return nil, fmt.Errorf("unexpected command: %s %v", name, args)
	}

	cmd := m.Commands[m.Index]
	m.Index++

	// Verify command matches
	if cmd.Name != name {
		return nil, fmt.Errorf("expected command %s, got %s", cmd.Name, name)
	}

	// Verify args match (loose comparison for flexibility in tests)
	if len(cmd.Args) > 0 && !argsMatch(cmd.Args, args) {
		return nil, fmt.Errorf("expected args %v, got %v", cmd.Args, args)
	}

	return cmd.Output, cmd.Error
}

func (m *MockCommandRunner) RunInDir(dir, name string, args ...string) ([]byte, error) {
	if m.Index >= len(m.Commands) {
		return nil, fmt.Errorf("unexpected command in dir %s: %s %v", dir, name, args)
	}

	cmd := m.Commands[m.Index]
	m.Index++

	// Verify command matches
	if cmd.Name != name {
		return nil, fmt.Errorf("expected command %s, got %s", cmd.Name, name)
	}

	// Verify directory if specified in mock
	if cmd.Dir != "" && cmd.Dir != dir {
		return nil, fmt.Errorf("expected dir %s, got %s", cmd.Dir, dir)
	}

	// Verify args match
	if len(cmd.Args) > 0 && !argsMatch(cmd.Args, args) {
		return nil, fmt.Errorf("expected args %v, got %v", cmd.Args, args)
	}

	return cmd.Output, cmd.Error
}

func argsMatch(expected, actual []string) bool {
	if len(expected) != len(actual) {
		return false
	}
	for i, e := range expected {
		if e != actual[i] {
			return false
		}
	}
	return true
}

func TestCheckGHInstalled(t *testing.T) {
	tests := []struct {
		name    string
		mock    MockCommand
		wantErr bool
	}{
		{
			name: "gh installed",
			mock: MockCommand{
				Name:   "gh",
				Args:   []string{"--version"},
				Output: []byte("gh version 2.40.0 (2024-01-10)\n"),
				Error:  nil,
			},
			wantErr: false,
		},
		{
			name: "gh not installed",
			mock: MockCommand{
				Name:   "gh",
				Args:   []string{"--version"},
				Output: []byte(""),
				Error:  fmt.Errorf("command not found"),
			},
			wantErr: true,
		},
		{
			name: "wrong command",
			mock: MockCommand{
				Name:   "gh",
				Args:   []string{"--version"},
				Output: []byte("some other output"),
				Error:  nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &MockCommandRunner{
				Commands: []MockCommand{tt.mock},
			}
			client := NewClientWithRunner(runner)

			err := client.CheckGHInstalled()
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckGHInstalled() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckAuthenticated(t *testing.T) {
	tests := []struct {
		name    string
		mock    MockCommand
		wantErr bool
	}{
		{
			name: "authenticated",
			mock: MockCommand{
				Name:   "gh",
				Args:   []string{"auth", "status"},
				Output: []byte("Logged in to github.com as user"),
				Error:  nil,
			},
			wantErr: false,
		},
		{
			name: "not authenticated",
			mock: MockCommand{
				Name:   "gh",
				Args:   []string{"auth", "status"},
				Output: []byte(""),
				Error:  fmt.Errorf("not logged in"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &MockCommandRunner{
				Commands: []MockCommand{tt.mock},
			}
			client := NewClientWithRunner(runner)

			err := client.CheckAuthenticated()
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckAuthenticated() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCloneRepository(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetPath := filepath.Join(tempDir, "repo")

	tests := []struct {
		name    string
		repo    string
		mock    MockCommand
		wantErr bool
	}{
		{
			name: "successful clone",
			repo: "test/repo",
			mock: MockCommand{
				Name:   "gh",
				Args:   []string{"repo", "clone", "test/repo", targetPath},
				Output: []byte("Cloning into 'repo'..."),
				Error:  nil,
			},
			wantErr: false,
		},
		{
			name: "clone fails",
			repo: "test/repo",
			mock: MockCommand{
				Name:   "gh",
				Args:   []string{"repo", "clone", "test/repo", targetPath},
				Output: []byte("repository not found"),
				Error:  fmt.Errorf("exit status 1"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &MockCommandRunner{
				Commands: []MockCommand{tt.mock},
			}
			client := NewClientWithRunner(runner)

			err := client.CloneRepository(tt.repo, targetPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("CloneRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSyncRepository(t *testing.T) {
	repoPath := "/test/repo"

	tests := []struct {
		name     string
		mocks    []MockCommand
		wantErr  bool
		errMatch string
	}{
		{
			name: "successful sync",
			mocks: []MockCommand{
				{
					Name:   "git",
					Args:   []string{"branch", "-a"},
					Dir:    repoPath,
					Output: []byte("* main\n  remotes/origin/main\n"),
					Error:  nil,
				},
				{
					Name:   "git",
					Args:   []string{"fetch", "--all"},
					Dir:    repoPath,
					Output: []byte("Fetching origin"),
					Error:  nil,
				},
				{
					Name:   "git",
					Args:   []string{"branch", "--show-current"},
					Dir:    repoPath,
					Output: []byte("main\n"),
					Error:  nil,
				},
				{
					Name:   "git",
					Args:   []string{"rev-parse", "origin/main"},
					Dir:    repoPath,
					Output: []byte("abc123"),
					Error:  nil,
				},
				{
					Name:   "git",
					Args:   []string{"reset", "--hard", "origin/main"},
					Dir:    repoPath,
					Output: []byte("HEAD is now at abc123"),
					Error:  nil,
				},
			},
			wantErr: false,
		},
		{
			name: "fetch fails",
			mocks: []MockCommand{
				{
					Name:   "git",
					Args:   []string{"branch", "-a"},
					Dir:    repoPath,
					Output: []byte("* main\n"),
					Error:  nil,
				},
				{
					Name:   "git",
					Args:   []string{"fetch", "--all"},
					Dir:    repoPath,
					Output: []byte("error: failed to fetch"),
					Error:  fmt.Errorf("exit status 1"),
				},
			},
			wantErr:  true,
			errMatch: "failed to fetch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &MockCommandRunner{
				Commands: tt.mocks,
			}
			client := NewClientWithRunner(runner)

			err := client.SyncRepository(repoPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("SyncRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMatch != "" && !strings.Contains(err.Error(), tt.errMatch) {
				t.Errorf("SyncRepository() error = %v, should contain %v", err, tt.errMatch)
			}
		})
	}
}

func TestCommitAndPush(t *testing.T) {
	repoPath := "/test/repo"
	message := "Test commit"

	tests := []struct {
		name     string
		mocks    []MockCommand
		wantErr  bool
		errMatch string
	}{
		{
			name: "successful commit and push",
			mocks: []MockCommand{
				{
					Name:   "git",
					Args:   []string{"add", "."},
					Dir:    repoPath,
					Output: []byte(""),
					Error:  nil,
				},
				{
					Name:   "git",
					Args:   []string{"status", "--porcelain"},
					Dir:    repoPath,
					Output: []byte("M file.txt\n"),
					Error:  nil,
				},
				{
					Name:   "git",
					Args:   []string{"commit", "-m", message},
					Dir:    repoPath,
					Output: []byte("[main abc123] Test commit"),
					Error:  nil,
				},
				{
					Name:   "git",
					Args:   []string{"branch", "--show-current"},
					Dir:    repoPath,
					Output: []byte("main\n"),
					Error:  nil,
				},
				{
					Name:   "git",
					Args:   []string{"push", "-u", "origin", "main"},
					Dir:    repoPath,
					Output: []byte("Everything up-to-date"),
					Error:  nil,
				},
			},
			wantErr: false,
		},
		{
			name: "no changes to commit",
			mocks: []MockCommand{
				{
					Name:   "git",
					Args:   []string{"add", "."},
					Dir:    repoPath,
					Output: []byte(""),
					Error:  nil,
				},
				{
					Name:   "git",
					Args:   []string{"status", "--porcelain"},
					Dir:    repoPath,
					Output: []byte(""),
					Error:  nil,
				},
			},
			wantErr:  true,
			errMatch: "no changes to commit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &MockCommandRunner{
				Commands: tt.mocks,
			}
			client := NewClientWithRunner(runner)

			err := client.CommitAndPush(repoPath, message)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommitAndPush() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMatch != "" && !strings.Contains(err.Error(), tt.errMatch) {
				t.Errorf("CommitAndPush() error = %v, should contain %v", err, tt.errMatch)
			}
		})
	}
}

func TestRepositoryExists(t *testing.T) {
	// Create temp directory with .git subdirectory
	tempDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	gitDir := filepath.Join(tempDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	client := NewClient()

	// Test existing repository
	if !client.RepositoryExists(tempDir) {
		t.Error("RepositoryExists() = false, want true for existing repo")
	}

	// Test non-existing repository
	nonExistentPath := filepath.Join(tempDir, "non-existent")
	if client.RepositoryExists(nonExistentPath) {
		t.Error("RepositoryExists() = true, want false for non-existent repo")
	}
}