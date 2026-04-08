package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/stahnma/slack-emoji-uploader/internal/slack"
	"github.com/stahnma/slack-emoji-uploader/internal/state"
)

var resolveCmd = &cobra.Command{
	Use:   "resolve <directory>",
	Short: "Interactively resolve emoji name conflicts",
	Args:  cobra.ExactArgs(1),
	RunE:  runResolve,
}

func init() {
	rootCmd.AddCommand(resolveCmd)
}

func runResolve(cmd *cobra.Command, args []string) error {
	dir := args[0]

	token, cookie, team, err := resolveAuth()
	if err != nil {
		return err
	}

	st, err := state.Load("emoji-state.json")
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}
	conflicts, err := state.LoadConflicts("emoji-conflicts.json")
	if err != nil {
		return fmt.Errorf("loading conflicts: %w", err)
	}

	if len(conflicts.Entries) == 0 {
		fmt.Println("No conflicts to resolve.")
		return nil
	}

	client := slack.NewClient(token, cookie, team, 1*time.Second)
	reader := bufio.NewReader(os.Stdin)
	resolved := 0

	for filePath, entry := range conflicts.Entries {
		fmt.Printf("\nConflict: %s → :%s: already exists\n", filePath, entry.Name)
		fmt.Printf("Attempted: %s\n", strings.Join(entry.Attempted, ", "))
		fmt.Print("Enter new name (or 'skip'): ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" || input == "skip" {
			fmt.Println("Skipped.")
			continue
		}

		imageData, err := os.ReadFile(filepath.Join(dir, filePath))
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			continue
		}

		fmt.Printf("Uploading: %s → :%s: ... ", filePath, input)
		result, err := client.UploadEmoji(input, imageData, filepath.Base(filePath))
		if err != nil {
			fmt.Printf("error: %v\n", err)
			continue
		}

		if result.OK {
			fmt.Println("OK")
			st.RecordSuccess(filePath, input)
			conflicts.Remove(filePath)
			resolved++
		} else {
			fmt.Printf("failed: %s\n", result.Error)
			entry.Attempted = append(entry.Attempted, input)
			entry.LastAttempt = time.Now().UTC()
			conflicts.Entries[filePath] = entry
		}

		if err := st.Save(); err != nil {
			return fmt.Errorf("saving state: %w", err)
		}
		if err := conflicts.Save(); err != nil {
			return fmt.Errorf("saving conflicts: %w", err)
		}
		time.Sleep(client.Delay())
	}

	fmt.Printf("\nResolved: %d, Remaining: %d\n", resolved, len(conflicts.Entries))
	return nil
}
