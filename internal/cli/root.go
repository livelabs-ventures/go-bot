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
	rootCmd.Flags().StringVar(&nameFlag, "name", "", "Override configured name (useful for testing)")
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
		return commands.RunStandupDirect(cfg)
	}
	return commands.RunStandupPR(cfg)
}