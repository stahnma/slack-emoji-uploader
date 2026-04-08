# slack-emoji-uploader

A CLI tool for bulk-uploading custom emoji to free-tier Slack workspaces. It uses Slack's browser-based emoji upload endpoint (the same one the web UI uses), so it works without requiring a paid Slack plan or admin API access.

## Features

- Bulk upload emoji from a directory of image files
- Automatic API token derivation -- only a browser cookie is needed
- Idempotent state tracking -- safely resume interrupted uploads
- Automatic numeric suffix for name conflicts (`--auto-suffix`)
- Interactive conflict resolution via the `resolve` subcommand
- Dry-run mode to preview what would be uploaded
- Rate limit handling with automatic exponential backoff and retry

## Installation

```
go install github.com/stahnma/slack-emoji-uploader@latest
```

## Getting Your Slack Credentials

This tool authenticates using a browser session cookie. You only need two values: your **cookie** and **team name**. The API token is derived automatically.

### Getting the cookie (`d`)

1. Open your Slack workspace in a browser (e.g., `https://app.slack.com`).
2. Open browser DevTools (F12, or Cmd+Option+I on Mac).
3. Go to the **Application** tab (Chrome) or **Storage** tab (Firefox).
4. Expand **Cookies** in the left sidebar, click your Slack domain.
5. Find the cookie named `d` and copy its value (starts with `xoxd-`).

### Getting your team name

6. Your team name is the subdomain of your Slack URL. For example, if your workspace is at `mycompany.slack.com`, the team is `mycompany`.

### Save your credentials

7. Create a `.env` file in the directory where you run the tool (see below).

### Providing the token manually (optional)

The tool automatically fetches the `xoxc-*` API token using your cookie. If you prefer to provide it manually (e.g., from the browser console), you can set `SLACK_TOKEN` in your `.env` file or pass `--token`. To extract it manually, open the browser console on any Slack page and run:

```javascript
document.documentElement.innerHTML.match(/xoxc-[a-zA-Z0-9-]+/)[0]
```

## Configuration

Create a `.env` file in your working directory:

```
SLACK_COOKIE=xoxd-your-cookie-value-here
SLACK_TEAM=your-workspace-name
```

Optionally, you can also set the token explicitly:

```
SLACK_TOKEN=xoxc-your-token-here
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
| `--token` | global | | Slack `xoxc-*` session token (auto-derived if omitted) |
| `--cookie` | global | | Slack session cookie (value of the `d` cookie) |
| `--team` | global | | Slack workspace subdomain |
| `--auto-suffix` | upload | `false` | Append numeric suffix on name conflicts |
| `--delay` | upload | `1s` | Delay between uploads (e.g., `2s`, `500ms`) |
| `--dry-run` | upload | `false` | Show what would be uploaded without uploading |
| `--verbose` | upload | `false` | Show detailed request/response info for debugging |

## How It Works

This tool uploads emoji through Slack's undocumented `emoji.add` HTTP endpoint -- the same one the Slack web interface uses when you add a custom emoji from the browser. Because it does not use the paid Admin API (`admin.emoji.add`), it works on free-tier workspaces.

Upload progress is tracked in `emoji-state.json` so that interrupted runs can be safely resumed. Conflicts (emoji names that already exist) are recorded in `emoji-conflicts.json` for later resolution. On subsequent runs, both files are checked so that already-uploaded emoji and known conflicts are skipped without hitting the API.

If Slack rate-limits a request, the tool automatically retries with exponential backoff (2s, 4s, 8s, ..., up to 60s). Retry progress is shown inline. If all retries are exhausted, the emoji is skipped and will be retried on the next run.

## Notes

- Session cookies expire periodically. If you start getting authentication errors, re-extract your `d` cookie from the browser.
- If you have `SLACK_TOKEN` set in your shell environment (e.g., for a Slack bot), it will take precedence over auto-derivation. Unset it (`unset SLACK_TOKEN`) or pass `--token ""` to use auto-derivation.
- Be respectful of rate limits. The default 1-second delay between uploads is a reasonable starting point. Increasing the delay with `--delay` is recommended if you are uploading a large batch.
- Use `--verbose` to see token/cookie details and full API responses when troubleshooting.
