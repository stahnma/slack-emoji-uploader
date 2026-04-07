package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type UploadResult struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type Client struct {
	token       string
	cookie      string
	team        string
	delay       time.Duration
	baseURL     string
	baseBackoff time.Duration
	httpClient  *http.Client
}

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

	writer.WriteField("name", name)
	writer.WriteField("token", c.token)
	writer.WriteField("mode", "data")

	part, err := writer.CreateFormFile("image", filename)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(imageData)); err != nil {
		return nil, fmt.Errorf("write image: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", c.baseURL+"/api/emoji.add", &body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Cookie", "d="+c.cookie)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result UploadResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w (body: %s)", err, string(respBody))
	}
	return &result, nil
}

func (c *Client) Delay() time.Duration {
	return c.delay
}
