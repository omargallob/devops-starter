package installer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDownload_Success(t *testing.T) {
	payload := []byte("hello world binary content")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "devops-starter/1.0" {
			t.Errorf("expected User-Agent 'devops-starter/1.0', got %q", r.Header.Get("User-Agent"))
		}
		w.Header().Set("Content-Length", "26")
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
	}))
	defer srv.Close()

	destPath := filepath.Join(t.TempDir(), "downloaded")
	err := Download(context.Background(), srv.URL+"/tool.tar.gz", destPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("content mismatch: got %q, want %q", got, payload)
	}
}

func TestDownload_LargeFile(t *testing.T) {
	// Generate a 1MB payload
	payload := make([]byte, 1024*1024)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
	}))
	defer srv.Close()

	destPath := filepath.Join(t.TempDir(), "large-file")
	err := Download(context.Background(), srv.URL, destPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if len(got) != len(payload) {
		t.Errorf("size mismatch: got %d, want %d", len(got), len(payload))
	}

	// Verify integrity via checksum
	wantHash := sha256.Sum256(payload)
	gotHash := sha256.Sum256(got)
	if wantHash != gotHash {
		t.Error("content integrity mismatch")
	}
}

func TestDownload_HTTP404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	destPath := filepath.Join(t.TempDir(), "not-found")
	err := Download(context.Background(), srv.URL+"/missing", destPath)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should mention 404, got: %v", err)
	}
}

func TestDownload_HTTP500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	destPath := filepath.Join(t.TempDir(), "server-error")
	err := Download(context.Background(), srv.URL, destPath)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should mention 500, got: %v", err)
	}
}

func TestDownload_ContextCancelled(t *testing.T) {
	// Server that delays response
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("too late"))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	destPath := filepath.Join(t.TempDir(), "cancelled")
	err := Download(ctx, srv.URL, destPath)
	if err == nil {
		t.Fatal("expected error when context is cancelled")
	}
	// Should contain context-related error
	if !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "deadline") {
		t.Errorf("error should mention context/deadline, got: %v", err)
	}
}

func TestDownload_InvalidDestPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data"))
	}))
	defer srv.Close()

	// Write to a path where the parent directory doesn't exist
	destPath := filepath.Join(t.TempDir(), "nonexistent", "subdir", "file")
	err := Download(context.Background(), srv.URL, destPath)
	if err == nil {
		t.Fatal("expected error for invalid dest path")
	}
	if !strings.Contains(err.Error(), "creating file") {
		t.Errorf("error should mention 'creating file', got: %v", err)
	}
}

func TestDownload_ChecksumVerification(t *testing.T) {
	// Test that a downloaded file can be checksum-verified
	payload := []byte("verifiable content")
	h := sha256.Sum256(payload)
	expectedChecksum := hex.EncodeToString(h[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
	}))
	defer srv.Close()

	destPath := filepath.Join(t.TempDir(), "checksum-test")
	err := Download(context.Background(), srv.URL, destPath)
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}

	// Verify checksum of downloaded file
	if err := VerifyChecksum(destPath, expectedChecksum); err != nil {
		t.Errorf("checksum verification failed: %v", err)
	}
}
