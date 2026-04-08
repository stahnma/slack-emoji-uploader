# slack-emoji-uploader

A CLI tool for bulk-uploading custom emoji to free-tier Slack workspaces. It uses Slack's browser-based emoji upload endpoint (the same one the web UI uses), so it works without requiring a paid Slack plan or admin API access.

## Features

- Bulk upload emoji from a directory of image files
- Idempotent state tracking -- safely resume interrupted uploads
- Automatic numeric suffix for name conflicts (`--auto-suffix`)
- Interactive conflict resolution via the `resolve` subcommand
- Dry-run mode to preview what would be uploaded
- Configurable rate limiting with delay between requests

## Installation

```
go install github.com/stahnma/slack_emoji_uploader@latest
```

## Getting Your Slack Credentials

This tool authenticates using a browser session token and cookie. You will need three values: a **token**, a **cookie**, and your **team name**.

1. Open your Slack workspace in a browser (e.g., `https://your-team.slack.com`).
2. Open browser DevTools (F12, or Cmd+Option+I on Mac).
3. Go to the **Network** tab.
4. Navigate to `https://your-team.slack.com/customize/emoji`.
5. In the Network tab, look for any request to an `/api/` endpoint (like `emoji.adminList` or `client.boot`).
6. Click the request and look in the **Request Payload** or **Form Data** for a value starting with `xoxc-` -- that is your token.
7. Go to the **Application** tab, then **Cookies**, then your Slack domain. Find the cookie named `d` and copy its value.
8. Your team name is the subdomain of your Slack URL. For example, if your workspace is at `mycompany.slack.com`, the team is `mycompany`.
9. Create a `.env` file in the directory where you run the tool with these values (see below).

## Configuration

Create a `.env` file in your working directory:

```
SLACK_TOKEN=xoxc-your-token-here
SLACK_COOKIE=your-cookie-value-here
SLACK_TEAM=your-workspace-name
```

Values can also be set via CLI flags (`--token`, `--cookie`, `--team`) or environment variables (`SLACK_TOKEN`, `SLACK_COOKIE`, `SLACK_TEAM`). Flags take precedence over environment variables, which take precedence over the `.env` file.

## Usage

### Upload emoji from a directory

```
slack-emoji-uploader upload ./emoji/
```

### Upload with automatic suffix on conflicts

If an emoji name already exists, automatically try appending a numeric suffix (e.g., `wave2`, `wave3`, ...):

```
slack-emoji-uploader upload --auto-suffix ./emoji/
```

### Preview uploads without uploading

```
slack-emoji-uploader upload --dry-run ./emoji/
```

### Check upload progress

```
slack-emoji-uploader status ./emoji/
```

### Interactively resolve name conflicts

```
slack-emoji-uploader resolve ./emoji/
```

## Flags Reference

| Flag | Scope | Default | Description |
|------|-------|---------|-------------|
| `--token` | global | | Slack `xoxc-*` session token |
| `--cookie` | global | | Slack `d=` session cookie value |
| `--team` | global | | Slack workspace subdomain |
| `--auto-suffix` | upload | `false` | Append numeric suffix on name conflicts |
| `--delay` | upload | `1s` | Delay between uploads (e.g., `2s`, `500ms`) |
| `--dry-run` | upload | `false` | Show what would be uploaded without uploading |

## How It Works

This tool uploads emoji through Slack's undocumented `emoji.add` HTTP endpoint -- the same one the Slack web interface uses when you add a custom emoji from the browser. Because it does not use the paid Admin API (`admin.emoji.add`), it works on free-tier workspaces.

Upload progress is tracked in `emoji-state.json` so that interrupted runs can be safely resumed. Conflicts (emoji names that already exist) are recorded in `emoji-conflicts.json` for later resolution.

## Notes

- Session tokens (`xoxc-*`) and cookies expire periodically. If you start getting authentication errors, re-extract your token and cookie from the browser.
- Be respectful of rate limits. The default 1-second delay between uploads is a reasonable starting point. Increasing the delay with `--delay` is recommended if you are uploading a large batch.
