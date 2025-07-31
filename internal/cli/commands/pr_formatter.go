package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FormatDailyPRBody formats the PR body with all standups for the day
func FormatDailyPRBody(repoPath string, date time.Time) string {
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

// extractTodayStandup extracts today's standup entry from the file content
func extractTodayStandup(content string, date time.Time) string {
	dateStr := date.Format("2006-01-02")
	lines := strings.Split(content, "\n")
	
	inTodaySection := false
	var todayContent []string
	
	for _, line := range lines {
		// Check if we found a date header matching today
		if strings.HasPrefix(line, "## ") && strings.Contains(line, dateStr) {
			inTodaySection = true
			continue
		}
		
		// Check if we hit the next date section or separator (stop collecting)
		if inTodaySection {
			if strings.HasPrefix(line, "## ") || line == "---" {
				break
			}
			// Collect content for today
			todayContent = append(todayContent, line)
		}
	}
	
	return strings.TrimSpace(strings.Join(todayContent, "\n"))
}