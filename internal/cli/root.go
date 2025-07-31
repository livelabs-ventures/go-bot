package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/standup-bot/standup-bot/internal/cli/commands"
	"github.com/standup-bot/standup-bot/pkg/config"
)

var (
	configFlag bool
	directFlag bool
	mergeFlag  bool
	nameFlag   string
	jsonFlag   string
	outputFlag string
	
	// Version information
	version string
	commit  string
	date    string
	
	rootCmd    = &cobra.Command{
		Use:   "standup-bot",
		Short: "A simple CLI tool for daily standup updates via GitHub",
		Long: `Standup Bot facilitates daily standup updates via GitHub.
It collects standup information and commits it to a shared repository,
where GitHub-Slack integration broadcasts updates to the team channel.

The tool supports both interactive and scriptable modes for easy automation.

Examples:
  # Interactive mode (default)
  standup-bot

  # Direct JSON input
  standup-bot --json '{"yesterday": ["Fixed bug"], "today": ["Write tests"], "blockers": "None"}'

  # Read JSON from stdin
  echo '{"yesterday": ["Fixed bug"], "today": ["Write tests"], "blockers": "None"}' | standup-bot --json -

  # Read JSON from file
  standup-bot --json standup.json

  # Machine-readable output
  standup-bot --json standup.json --output json

  # Direct commit mode with JSON
  standup-bot --direct --json '{"yesterday": ["Task A"], "today": ["Task B"]}' --output json`,
		RunE: runStandup,
	}
	
	mcpServerCmd = &cobra.Command{
		Use:   "mcp-server",
		Short: "Run the MCP (Model Context Protocol) server",
		Long: `Starts the standup-bot MCP server using stdio transport.
This allows AI assistants to interact with standup-bot functionality.

The server exposes these tools:
- submit_standup: Submit daily standup with yesterday/today/blockers
- create_standup_pr: Create or manage standup pull requests
- get_standup_status: Check if today's standup is complete`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.RunMCPServer()
		},
	}
)

func init() {
	rootCmd.Flags().BoolVar(&configFlag, "config", false, "Run configuration setup")
	rootCmd.Flags().BoolVar(&directFlag, "direct", false, "Use direct commit workflow (multi-line commit message)")
	rootCmd.Flags().BoolVar(&mergeFlag, "merge", false, "Merge today's standup pull request")
	rootCmd.Flags().StringVar(&nameFlag, "name", "", "Override configured name (useful for testing)")
	rootCmd.Flags().StringVar(&jsonFlag, "json", "", "Accept standup data as JSON (direct string, file path, or '-' for stdin)")
	rootCmd.Flags().StringVar(&outputFlag, "output", "", "Output format: 'json' for machine-readable output")
	
	// Set version template
	rootCmd.Version = buildVersion()
	rootCmd.SetVersionTemplate(`{{with .Name}}{{printf "%s " .}}{{end}}{{printf "version %s" .Version}}
`)
	
	// Add subcommands
	rootCmd.AddCommand(mcpServerCmd)
}

// SetVersionInfo sets the version information for the CLI
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}

// buildVersion creates the version string
func buildVersion() string {
	if version == "" {
		version = "dev"
	}
	
	versionStr := version
	if commit != "" && commit != "none" {
		versionStr += fmt.Sprintf(" (commit: %s)", commit)
	}
	if date != "" && date != "unknown" {
		versionStr += fmt.Sprintf(" built at %s", date)
	}
	
	return versionStr
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// runStandup is the main entry point for the standup command
func runStandup(cmd *cobra.Command, args []string) error {
	// Create configuration manager
	cfgManager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize config manager: %w", err)
	}

	// Check if we need to run configuration
	if configFlag || !cfgManager.Exists() {
		return commands.RunConfiguration(cfgManager)
	}

	// Load configuration
	cfg, err := cfgManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override name if flag is provided
	if nameFlag != "" {
		cfg.Name = nameFlag
		fmt.Printf("Using name override: %s\n", nameFlag)
	}

	// Handle merge command
	if mergeFlag {
		return commands.RunMergeDailyStandup(cfg)
	}


	// Run the standup workflow
	if directFlag {
		return commands.RunStandupDirect(cfg, jsonFlag, outputFlag)
	}
	return commands.RunStandupPR(cfg, jsonFlag, outputFlag)
}