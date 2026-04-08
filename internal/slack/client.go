package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"time"
)

// UploadResult represents the Slack API response.
type UploadResult struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// Client handles Slack emoji API calls.
type Client struct {
	token       string
	cookie      string
	team        string
	delay       time.Duration
	baseURL     string
	baseBackoff time.Duration
	httpClient  *http.Client
	Verbose     bool
}

// NewClient creates a Slack client. The cookie parameter should be the raw
// cookie value (without the "d=" prefix).
func NewClient(token, cookie, team string, delay time.Duration) *Client {
	return &Client{
		token:       token,
		cookie:      cookie,
		team:        team,
		delay:       delay,
		baseURL:     fmt.Sprintf("https://%s.slack.com", team),
		baseBackoff: 2 * time.Second,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// UploadEmoji uploads an emoji image to Slack.
// It retries automatically on rate limits with exponential backoff.
func (c *Client) UploadEmoji(name string, imageData []byte, filename string) (*UploadResult, error) {
	maxRetries := 5
	backoff := c.baseBackoff

	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, err := c.doUpload(name, imageData, filename)
		if err != nil {
			return nil, err
		}
		if result.Error != "ratelimited" {
			return result, nil
		}
		if attempt == maxRetries {
			return result, nil
		}
		fmt.Printf("rate limited, retrying in %s... ", backoff)
		time.Sleep(backoff)
		backoff *= 2
		if backoff > 60*time.Second {
			backoff = 60 * time.Second
		}
	}
	return nil, fmt.Errorf("unreachable")
}

func (c *Client) doUpload(name string, imageData []byte, filename string) (*UploadResult, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("name", name); err != nil {
		return nil, fmt.Errorf("write field name: %w", err)
	}
	if err := writer.WriteField("token", c.token); err != nil {
		return nil, fmt.Errorf("write field token: %w", err)
	}
	if err := writer.WriteField("mode", "data"); err != nil {
		return nil, fmt.Errorf("write field mode: %w", err)
	}

	part, err := writer.CreateFormFile("image", filename)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(imageData)); err != nil {
		return nil, fmt.Errorf("write image: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/emoji.add", &body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Cookie", "d="+c.cookie)
	req.Header.Set("Origin", c.baseURL)
	req.Header.Set("Referer", c.baseURL+"/customize/emoji")

	if c.Verbose {
		fmt.Printf("[verbose] POST %s/api/emoji.add\n", c.baseURL)
		fmt.Printf("[verbose] Token: %s...%s\n", c.token[:10], c.token[len(c.token)-6:])
		fmt.Printf("[verbose] Cookie: d=%s...%s\n", c.cookie[:10], c.cookie[len(c.cookie)-6:])
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if c.Verbose {
		fmt.Printf("[verbose] Response: %s\n", string(respBody))
	}

	var result UploadResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w (body: %s)", err, string(respBody))
	}
	return &result, nil
}

// FetchToken derives the xoxc-* API token from the workspace page using the
// session cookie. This allows users to provide only the cookie and team name.
func FetchToken(cookie, team string) (string, error) {
	url := fmt.Sprintf("https://%s.slack.com", team)
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Cookie", "d="+cookie)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	re := regexp.MustCompile(`xoxc-[a-zA-Z0-9-]+`)
	match := re.Find(body)
	if match == nil {
		return "", fmt.Errorf("could not find xoxc-* token in page — cookie may be expired")
	}
	return string(match), nil
}

// Delay returns the configured delay between uploads.
func (c *Client) Delay() time.Duration {
	return c.delay
}
