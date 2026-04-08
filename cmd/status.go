package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stahnma/slack_emoji_uploader/internal/state"
	"github.com/stahnma/slack_emoji_uploader/internal/walker"
)

var statusCmd = &cobra.Command{
	Use:   "status <directory>",
	Short: "Show upload progress",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	dir := args[0]

	entries, err := walker.WalkDir(dir)
	if err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	st, err := state.Load("emoji-state.json")
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}
	conflicts, err := state.LoadConflicts("emoji-conflicts.json")
	if err != nil {
		return fmt.Errorf("loading conflicts: %w", err)
	}

	total := len(entries)
	uploaded := len(st.Entries)
	conflicted := len(conflicts.Entries)
	remaining := total - uploaded - conflicted

	fmt.Printf("Total:      %d\n", total)
	fmt.Printf("Uploaded:   %d\n", uploaded)
	fmt.Printf("Conflicts:  %d\n", conflicted)
	fmt.Printf("Remaining:  %d\n", remaining)

	return nil
}
