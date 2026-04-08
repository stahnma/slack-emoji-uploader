package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	flagToken  string
	flagCookie string
	flagTeam   string
)

var rootCmd = &cobra.Command{
	Use:   "slack-emoji-uploader",
	Short: "Bulk upload emoji to Slack workspaces",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "Slack xoxc-* session token")
	rootCmd.PersistentFlags().StringVar(&flagCookie, "cookie", "", "Slack session cookie (the value of the 'd' cookie)")
	rootCmd.PersistentFlags().StringVar(&flagTeam, "team", "", "Slack workspace subdomain")
}

func initConfig() {
	_ = godotenv.Load()
}

func resolveAuth() (token, cookie, team string, err error) {
	token = flagToken
	if token == "" {
		token = os.Getenv("SLACK_TOKEN")
	}
	cookie = flagCookie
	if cookie == "" {
		cookie = os.Getenv("SLACK_COOKIE")
	}
	team = flagTeam
	if team == "" {
		team = os.Getenv("SLACK_TEAM")
	}
	if token == "" || cookie == "" || team == "" {
		return "", "", "", fmt.Errorf("missing required auth: token, cookie, and team must be set via flags, env vars, or .env file")
	}
	// Strip "d=" prefix if user included it — the client adds it internally
	cookie = strings.TrimPrefix(cookie, "d=")
	return token, cookie, team, nil
}
