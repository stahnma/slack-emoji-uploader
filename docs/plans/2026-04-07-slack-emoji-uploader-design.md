# Slack Emoji Uploader — Design

## Overview

A Go CLI tool that bulk-uploads emoji images from a local directory to a Slack workspace. Designed for free-tier Slack users (no paid API access required). Uses Slack's undocumented `emoji.add` endpoint with browser session credentials.

## Goals

- Upload a directory tree of images as custom emoji
- Work on Slack's free tier (no `admin.emoji.add` API)
- Idempotent — tracks upload state, safe to re-run
- Handle name conflicts with logging and auto-suffix support
- Single static binary, minimal dependencies

## Authentication

The tool uses a `xoxc-*` session token and `d` session cookie extracted from the browser. These are required because the official emoji upload API is restricted to paid plans.

**Precedence:** CLI flags > environment variables > `.env` file

| Method | Token | Cookie |
|--------|-------|--------|
| Flags | `--token` | `--cookie` |
| Env vars | `SLACK_TOKEN` | `SLACK_COOKIE` |
| `.env` file | `SLACK_TOKEN=...` | `SLACK_COOKIE=...` |

The workspace name is provided via `--team` flag or `SLACK_TEAM` in env/.env.

## CLI Interface

Three subcommands:

### `upload <dir>`

Walks the directory tree, uploads emoji, tracks state.

**Flags:**
- `--team` — Slack workspace subdomain (e.g., `mycompany`)
- `--token` — `xoxc-*` session token
- `--cookie` — `d=...` session cookie
- `--auto-suffix` — On name conflict, retry with `name2`, `name3`, etc.
- `--delay` — Delay between uploads (default: `1s`)
- `--dry-run` — Show what would be uploaded without uploading

### `resolve`

Interactive conflict resolution for emoji that failed due to name collisions.

Walks through `emoji-conflicts.json`, prompts for a new name or skip for each entry.

### `status`

Shows upload progress: total, uploaded, conflicts, remaining.

## Upload Flow

1. **Load auth** — flags > env > `.env`. Exit with clear error if missing.
2. **Load state** — read `emoji-state.json` (create if absent).
3. **Walk directory** — find all `.png`, `.gif`, `.jpg`, `.jpeg` files recursively. Emoji name = filename without extension, lowercased, spaces replaced with `-`.
4. **Filter** — skip files already recorded in `emoji-state.json`.
5. **Upload** — POST multipart form to `https://{team}.slack.com/api/emoji.add` with token and cookie. Wait `--delay` between uploads.
6. **Handle responses:**
   - **Success** — record in `emoji-state.json`
   - **Name taken** — if `--auto-suffix`, retry as `name2`, `name3`, etc. Otherwise log to `emoji-conflicts.json`
   - **Rate limited** — exponential backoff, retry
   - **Auth error** — stop immediately with clear message
7. **Print summary** — X uploaded, Y skipped, Z conflicts.

## Name Derivation

Emoji names come from the filename only (subdirectories are ignored):

```
emoji/
  partyparrot.gif       → :partyparrot:
  cats/
    cat-thumbsup.png    → :cat-thumbsup:
    cat-wave.png        → :cat-wave:
```

## Auto-Suffix Behavior

When `--auto-suffix` is set and a name is taken:

- `:partyparrot:` taken → try `:partyparrot2:`
- `:partyparrot2:` taken → try `:partyparrot3:`
- Continue until success or max attempts

## State Files

Stored in the current working directory (not hidden).

**`emoji-state.json`** — successfully uploaded emoji:
```json
{
  "emoji/partyparrot.gif": {
    "name": "partyparrot",
    "uploaded_at": "2026-04-07T12:00:00Z"
  }
}
```

**`emoji-conflicts.json`** — failed uploads due to name collisions:
```json
{
  "emoji/partyparrot.gif": {
    "name": "partyparrot",
    "error": "error_name_taken",
    "attempted": ["partyparrot"],
    "last_attempt": "2026-04-07T12:00:00Z"
  }
}
```

## Rate Limiting

- Default delay between uploads: 1 second (configurable via `--delay`)
- On `429` / rate limit response: exponential backoff with retry
- Backoff caps at a reasonable maximum (e.g., 60 seconds)

## Project Structure

```
slack_emoji_uploader/
  cmd/
    root.go          # CLI setup, auth loading, .env parsing
    upload.go        # upload subcommand
    resolve.go       # resolve subcommand
    status.go        # status subcommand
  internal/
    slack/
      client.go      # HTTP client — emoji.add, rate limit/backoff
    state/
      state.go       # Read/write state and conflict files
    walker/
      walker.go      # Directory walking, filename-to-name logic
  main.go
  go.mod
```

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- `github.com/joho/godotenv` — `.env` file loading
- Standard library for HTTP, JSON, filepath

## Documentation Requirements

The README must include clear step-by-step instructions for extracting Slack credentials from the browser:

1. Open Slack workspace in browser
2. Open DevTools → Network tab
3. Navigate to `/customize/emoji`
4. Find `xoxc-*` token in API request headers/body
5. Find `d` cookie in Application → Cookies
6. Save to `.env` file
