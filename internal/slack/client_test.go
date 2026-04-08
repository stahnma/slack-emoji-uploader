package slack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUploadEmoji_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/emoji.add" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.FormValue("name") != "partyparrot" {
			t.Errorf("name = %q", r.FormValue("name"))
		}
		if r.FormValue("token") == "" {
			t.Error("missing token")
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	}))
	defer server.Close()

	c := NewClient("xoxc-test", "testcookie", "testteam", 0)
	c.baseURL = server.URL

	result, err := c.UploadEmoji("partyparrot", []byte("fake-image"), "partyparrot.gif")
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if !result.OK {
		t.Error("expected ok=true")
	}
}

func TestUploadEmoji_NameTaken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": "error_name_taken",
		})
	}))
	defer server.Close()

	c := NewClient("xoxc-test", "testcookie", "testteam", 0)
	c.baseURL = server.URL

	result, err := c.UploadEmoji("partyparrot", []byte("fake-image"), "partyparrot.gif")
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if result.OK {
		t.Error("expected ok=false")
	}
	if result.Error != "error_name_taken" {
		t.Errorf("error = %q", result.Error)
	}
}

func TestUploadEmoji_RateLimit(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.Header().Set("Retry-After", "1")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":    false,
				"error": "ratelimited",
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	}))
	defer server.Close()

	c := NewClient("xoxc-test", "testcookie", "testteam", 0)
	c.baseURL = server.URL
	c.baseBackoff = 10 * time.Millisecond

	result, err := c.UploadEmoji("partyparrot", []byte("fake-image"), "partyparrot.gif")
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if !result.OK {
		t.Error("expected eventual success")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}
