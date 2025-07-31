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

// FileSystem interface for file operations (for better testability)
type FileSystem interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)
}

// OSFileSystem implements FileSystem using os package
type OSFileSystem struct{}

func (fs *OSFileSystem) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (fs *OSFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (fs *OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (fs *OSFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Manager handles standup operations
type Manager struct {
	repoPath string
	fs       FileSystem
}

// NewManager creates a new standup manager
func NewManager(repoPath string) *Manager {
	return &Manager{
		repoPath: repoPath,
		fs:       &OSFileSystem{},
	}
}

// NewManagerWithFileSystem creates a new standup manager with custom filesystem (for testing)
func NewManagerWithFileSystem(repoPath string, fs FileSystem) *Manager {
	return &Manager{
		repoPath: repoPath,
		fs:       fs,
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

// EntryWriter handles writing standup entries
type EntryWriter interface {
	WriteEntry(entry *Entry, userName string) error
}

// FileEntryWriter writes entries to files
type FileEntryWriter struct {
	manager *Manager
}

// NewFileEntryWriter creates a new file-based entry writer
func NewFileEntryWriter(m *Manager) *FileEntryWriter {
	return &FileEntryWriter{manager: m}
}

// WriteEntry implements EntryWriter
func (w *FileEntryWriter) WriteEntry(entry *Entry, userName string) error {
	return w.manager.SaveEntry(entry, userName)
}

// SaveEntry saves the standup entry to the user's file
func (m *Manager) SaveEntry(entry *Entry, userName string) error {
	filePath, err := m.ensureStandupFile(userName)
	if err != nil {
		return err
	}

	existingContent, err := m.readExistingContent(filePath)
	if err != nil {
		return err
	}

	newContent := m.buildUpdatedContent(existingContent, entry, userName)
	return m.fs.WriteFile(filePath, []byte(newContent), 0644)
}

// GetStandupFilePath returns the path to the standup file for a user
func (m *Manager) GetStandupFilePath(userName string) (string, error) {
	standupDir := filepath.Join(m.repoPath, "stand-ups")
	fileName := fmt.Sprintf("%s.md", strings.ToLower(userName))
	return filepath.Join(standupDir, fileName), nil
}

// ensureStandupFile ensures the standup directory exists and returns the file path
func (m *Manager) ensureStandupFile(userName string) (string, error) {
	standupDir := filepath.Join(m.repoPath, "stand-ups")
	if err := m.fs.MkdirAll(standupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create stand-ups directory: %w", err)
	}

	fileName := fmt.Sprintf("%s.md", strings.ToLower(userName))
	return filepath.Join(standupDir, fileName), nil
}

// readExistingContent reads the existing file content if it exists
func (m *Manager) readExistingContent(filePath string) (string, error) {
	content, err := m.fs.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // File doesn't exist yet, that's OK
		}
		return "", fmt.Errorf("failed to read existing file: %w", err)
	}
	return string(content), nil
}

// buildUpdatedContent builds the updated content with today's entry
func (m *Manager) buildUpdatedContent(existingContent string, entry *Entry, userName string) string {
	if existingContent == "" {
		// New file
		return m.buildNewFileContent(entry, userName)
	}

	// Parse and update existing content
	parsedContent := m.parseExistingContent(existingContent, entry.Date)
	return m.assembleContent(parsedContent, entry, userName)
}

// buildNewFileContent creates content for a new standup file
func (m *Manager) buildNewFileContent(entry *Entry, userName string) string {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("# %s's Standups\n\n", userName))
	content.WriteString(m.formatEntry(entry))
	content.WriteString("\n")
	return content.String()
}

// parsedStandupContent represents parsed standup file content
type parsedStandupContent struct {
	header       string
	otherEntries []string
}

// entryParser handles parsing of standup entries
type entryParser struct {
	dateStr string
}

// parseExistingContent parses existing content and removes today's entry if it exists
func (m *Manager) parseExistingContent(content string, date time.Time) parsedStandupContent {
	parser := &entryParser{dateStr: date.Format("2006-01-02")}
	return parser.parse(content)
}

// parse processes the content and returns parsed standup content
func (p *entryParser) parse(content string) parsedStandupContent {
	lines := strings.Split(content, "\n")
	result := parsedStandupContent{}
	
	state := &parseState{
		currentEntry: &strings.Builder{},
	}

	for i, line := range lines {
		p.processLine(line, i, len(lines), state, &result)
	}

	p.finalizeEntry(state, &result)
	return result
}

// parseState holds the current state during parsing
type parseState struct {
	currentEntry *strings.Builder
	inTodayEntry bool
	headerFound  bool
}

// processLine processes a single line during parsing
func (p *entryParser) processLine(line string, index, total int, state *parseState, result *parsedStandupContent) {
	if p.isHeader(line) && !state.headerFound {
		state.headerFound = true
		result.header = line
		return
	}

	if p.shouldSkipLine(line, index, total, state) {
		return
	}

	if p.isTodayEntry(line) {
		state.inTodayEntry = true
		return
	}

	if state.inTodayEntry && p.isEntryEnd(line) {
		state.inTodayEntry = false
		if line == "---" {
			return // Skip the separator
		}
	}

	if !state.inTodayEntry && state.headerFound {
		state.currentEntry.WriteString(line)
		state.currentEntry.WriteString("\n")
	}
}

// isHeader checks if a line is a header
func (p *entryParser) isHeader(line string) bool {
	return strings.HasPrefix(line, "# ") && strings.Contains(line, "'s Standups")
}

// shouldSkipLine determines if a line should be skipped
func (p *entryParser) shouldSkipLine(line string, index, total int, state *parseState) bool {
	return state.headerFound && index < total-1 && strings.TrimSpace(line) == "" && !state.inTodayEntry
}

// isTodayEntry checks if a line starts today's entry
func (p *entryParser) isTodayEntry(line string) bool {
	return strings.HasPrefix(line, "## ") && strings.Contains(line, p.dateStr)
}

// isEntryEnd checks if a line marks the end of an entry
func (p *entryParser) isEntryEnd(line string) bool {
	return line == "---" || strings.HasPrefix(line, "## ")
}

// finalizeEntry completes the parsing and adds any remaining content
func (p *entryParser) finalizeEntry(state *parseState, result *parsedStandupContent) {
	if state.currentEntry.Len() > 0 {
		result.otherEntries = append(result.otherEntries, strings.TrimSpace(state.currentEntry.String()))
	}
}

// assembleContent assembles the final content with the new entry
func (m *Manager) assembleContent(parsed parsedStandupContent, entry *Entry, userName string) string {
	var content strings.Builder
	
	// Write header
	if parsed.header != "" {
		content.WriteString(parsed.header)
	} else {
		content.WriteString(fmt.Sprintf("# %s's Standups", userName))
	}
	content.WriteString("\n\n")
	
	// Write today's entry
	content.WriteString(m.formatEntry(entry))
	content.WriteString("\n")
	
	// Write other entries
	for _, otherEntry := range parsed.otherEntries {
		if otherEntry != "" {
			content.WriteString(otherEntry)
			content.WriteString("\n")
		}
	}
	
	return content.String()
}

// formatEntry formats a single standup entry
func (m *Manager) formatEntry(entry *Entry) string {
	var content strings.Builder
	
	fmt.Fprintf(&content, "## %s\n\n", entry.Date.Format("2006-01-02"))
	
	m.formatSection(&content, "Yesterday", entry.Yesterday, "Nothing to report")
	m.formatSection(&content, "Today", entry.Today, "Nothing planned")
	
	fmt.Fprintf(&content, "**Blockers:**\n%s\n\n---\n", entry.Blockers)

	return content.String()
}

// formatSection formats a section with items or a default message
func (m *Manager) formatSection(w *strings.Builder, title string, items []string, defaultMsg string) {
	fmt.Fprintf(w, "**%s:**\n", title)
	if len(items) == 0 {
		fmt.Fprintf(w, "- %s\n", defaultMsg)
	} else {
		for _, item := range items {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	w.WriteString("\n")
}

// FormatCommitMessage formats a commit message for the standup entry
func (m *Manager) FormatCommitMessage(entry *Entry, userName string) string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "[Standup] %s - %s\n\n", userName, entry.Date.Format("2006-01-02"))
	fmt.Fprintf(&builder, "Yesterday: %s\n", formatItems(entry.Yesterday, "Nothing to report"))
	fmt.Fprintf(&builder, "Today: %s\n", formatItems(entry.Today, "Nothing planned"))
	fmt.Fprintf(&builder, "Blockers: %s", entry.Blockers)

	return builder.String()
}

// formatItems formats a list of items or returns a default message
func formatItems(items []string, defaultMsg string) string {
	if len(items) == 0 {
		return defaultMsg
	}
	return strings.Join(items, "; ")
}