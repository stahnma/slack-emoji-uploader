package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/stahnma/slack_emoji_uploader/internal/slack"
	"github.com/stahnma/slack_emoji_uploader/internal/state"
	"github.com/stahnma/slack_emoji_uploader/internal/walker"
)

var (
	flagAutoSuffix bool
	flagDelay      time.Duration
	flagDryRun     bool
)

var uploadCmd = &cobra.Command{
	Use:   "upload <directory>",
	Short: "Upload emoji from a directory to Slack",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpload,
}

func init() {
	uploadCmd.Flags().BoolVar(&flagAutoSuffix, "auto-suffix", false, "Automatically append numeric suffix on name conflicts")
	uploadCmd.Flags().DurationVar(&flagDelay, "delay", 1*time.Second, "Delay between uploads")
	uploadCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Show what would be uploaded without uploading")
	rootCmd.AddCommand(uploadCmd)
}

func runUpload(cmd *cobra.Command, args []string) error {
	dir := args[0]

	token, cookie, team, err := resolveAuth()
	if err != nil {
		return err
	}

	entries, err := walker.WalkDir(dir)
	if err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}
	if len(entries) == 0 {
		fmt.Println("No emoji files found.")
		return nil
	}

	st, err := state.Load("emoji-state.json")
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}
	conflicts, err := state.LoadConflicts("emoji-conflicts.json")
	if err != nil {
		return fmt.Errorf("loading conflicts: %w", err)
	}

	client := slack.NewClient(token, cookie, team, flagDelay)

	var uploaded, skipped, conflicted int

	for _, entry := range entries {
		if st.IsUploaded(entry.Path) {
			skipped++
			continue
		}

		if flagDryRun {
			fmt.Printf("[dry-run] Would upload: %s → :%s:\n", entry.Path, entry.Name)
			continue
		}

		name := entry.Name
		attempted := []string{}
		success := false

		imageData, err := os.ReadFile(filepath.Join(dir, entry.Path))
		if err != nil {
			return fmt.Errorf("reading %s: %w", entry.Path, err)
		}

		for suffix := 0; suffix < 100; suffix++ {
			candidateName := name
			if suffix > 0 {
				candidateName = fmt.Sprintf("%s%d", name, suffix+1)
			}
			attempted = append(attempted, candidateName)

			fmt.Printf("Uploading: %s → :%s: ... ", entry.Path, candidateName)
			result, err := client.UploadEmoji(candidateName, imageData, filepath.Base(entry.Path))
			if err != nil {
				return fmt.Errorf("uploading %s: %w", entry.Path, err)
			}

			if result.OK {
				fmt.Println("OK")
				st.RecordSuccess(entry.Path, candidateName)
				if err := st.Save(); err != nil {
					return fmt.Errorf("saving state: %w", err)
				}
				uploaded++
				success = true
				break
			}

			if result.Error == "error_name_taken" {
				fmt.Printf("conflict (%s)\n", candidateName)
				if !flagAutoSuffix {
					break
				}
				continue
			}

			if result.Error == "not_authed" || result.Error == "invalid_auth" || result.Error == "token_revoked" {
				return fmt.Errorf("authentication failed: %s — check your token and cookie", result.Error)
			}

			return fmt.Errorf("unexpected error uploading %s: %s", entry.Path, result.Error)
		}

		if !success {
			conflicts.RecordConflict(entry.Path, name, attempted)
			if err := conflicts.Save(); err != nil {
				return fmt.Errorf("saving conflicts: %w", err)
			}
			conflicted++
		}

		time.Sleep(client.Delay())
	}

	fmt.Printf("\nDone: %d uploaded, %d skipped, %d conflicts\n", uploaded, skipped, conflicted)
	return nil
}
