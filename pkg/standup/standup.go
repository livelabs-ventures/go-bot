package standup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Entry represents a single standup entry
type Entry struct {
	Date      time.Time
	Yesterday []string
	Today     []string
	Blockers  string
}

// Manager handles standup operations
type Manager struct {
	repoPath string
}

// NewManager creates a new standup manager
func NewManager(repoPath string) *Manager {
	return &Manager{
		repoPath: repoPath,
	}
}

// CollectEntry collects standup information from the user
func (m *Manager) CollectEntry(reader io.Reader, writer io.Writer) (*Entry, error) {
	scanner := bufio.NewScanner(reader)
	entry := &Entry{
		Date: time.Now(),
	}

	// Yesterday
	fmt.Fprintln(writer, "What did you do yesterday?")
	fmt.Fprintln(writer, "(Enter multiple lines, press Enter twice to finish)")
	entry.Yesterday = m.collectMultiLineInput(scanner, writer)

	// Today
	fmt.Fprintln(writer, "\nWhat will you do today?")
	fmt.Fprintln(writer, "(Enter multiple lines, press Enter twice to finish)")
	entry.Today = m.collectMultiLineInput(scanner, writer)

	// Blockers
	fmt.Fprintln(writer, "\nAny blockers?")
	fmt.Fprint(writer, "> ")
	if scanner.Scan() {
		entry.Blockers = strings.TrimSpace(scanner.Text())
	}
	if entry.Blockers == "" {
		entry.Blockers = "None"
	}

	return entry, nil
}

// collectMultiLineInput collects multiple lines of input until an empty line
func (m *Manager) collectMultiLineInput(scanner *bufio.Scanner, writer io.Writer) []string {
	var lines []string
	for {
		fmt.Fprint(writer, "> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		if line == "" {
			break
		}
		lines = append(lines, line)
	}
	return lines
}

// SaveEntry saves the standup entry to the user's file
func (m *Manager) SaveEntry(entry *Entry, userName string) error {
	// Ensure stand-ups directory exists
	standupDir := filepath.Join(m.repoPath, "stand-ups")
	if err := os.MkdirAll(standupDir, 0755); err != nil {
		return fmt.Errorf("failed to create stand-ups directory: %w", err)
	}

	// Determine file path
	fileName := fmt.Sprintf("%s.md", strings.ToLower(userName))
	filePath := filepath.Join(standupDir, fileName)

	// Check if file exists and create header if it doesn't
	isNewFile := false
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		isNewFile = true
	}

	// Open file in append mode
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open standup file: %w", err)
	}
	defer file.Close()

	// Write header for new files
	if isNewFile {
		fmt.Fprintf(file, "# %s's Standups\n\n", userName)
	}

	// Write the entry
	fmt.Fprintf(file, "## %s\n\n", entry.Date.Format("2006-01-02"))
	
	fmt.Fprintf(file, "**Yesterday:**\n")
	if len(entry.Yesterday) == 0 {
		fmt.Fprintf(file, "- Nothing to report\n")
	} else {
		for _, item := range entry.Yesterday {
			fmt.Fprintf(file, "- %s\n", item)
		}
	}
	fmt.Fprintf(file, "\n")

	fmt.Fprintf(file, "**Today:**\n")
	if len(entry.Today) == 0 {
		fmt.Fprintf(file, "- Nothing planned\n")
	} else {
		for _, item := range entry.Today {
			fmt.Fprintf(file, "- %s\n", item)
		}
	}
	fmt.Fprintf(file, "\n")

	fmt.Fprintf(file, "**Blockers:**\n%s\n\n", entry.Blockers)
	fmt.Fprintf(file, "---\n\n")

	return nil
}

// FormatCommitMessage formats a commit message for the standup entry
func (m *Manager) FormatCommitMessage(entry *Entry, userName string) string {
	var builder strings.Builder

	// Title
	builder.WriteString(fmt.Sprintf("[Standup] %s - %s\n\n", userName, entry.Date.Format("2006-01-02")))

	// Yesterday
	builder.WriteString("Yesterday: ")
	if len(entry.Yesterday) == 0 {
		builder.WriteString("Nothing to report")
	} else {
		builder.WriteString(strings.Join(entry.Yesterday, "; "))
	}
	builder.WriteString("\n")

	// Today
	builder.WriteString("Today: ")
	if len(entry.Today) == 0 {
		builder.WriteString("Nothing planned")
	} else {
		builder.WriteString(strings.Join(entry.Today, "; "))
	}
	builder.WriteString("\n")

	// Blockers
	builder.WriteString("Blockers: ")
	builder.WriteString(entry.Blockers)

	return builder.String()
}