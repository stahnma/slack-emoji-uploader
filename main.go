package main

import (
	"os"

	"github.com/stahnma/slack_emoji_uploader/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
