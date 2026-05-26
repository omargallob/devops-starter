package updater

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheck_DevVersion(t *testing.T) {
	result := Check("dev")
	if result.UpdateAvailable {
		t.Error("should not report update for dev version")
	}
	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestCheck_EmptyVersion(t *testing.T) {
	result := Check("")
	if result.UpdateAvailable {
		t.Error("should not report update for empty version")
	}
}

func TestCheck_UpdateAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := githubRelease{TagName: "v2.0.0"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	result := checkWithURL(server.URL, "1.0.0")
	if !result.UpdateAvailable {
		t.Error("should report update available")
	}
	if result.LatestVersion != "2.0.0" {
		t.Errorf("expected latest 2.0.0, got %s", result.LatestVersion)
	}
}

func TestCheck_NoUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := githubRelease{TagName: "v1.0.0"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	result := checkWithURL(server.URL, "1.0.0")
	if result.UpdateAvailable {
		t.Error("should not report update when versions match")
	}
}

func TestCheck_OlderRemote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := githubRelease{TagName: "v0.9.0"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	result := checkWithURL(server.URL, "1.0.0")
	if result.UpdateAvailable {
		t.Error("should not report update when remote is older")
	}
}

func TestCheck_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	result := checkWithURL(server.URL, "1.0.0")
	if result.Error == nil {
		t.Error("should report error on server failure")
	}
	if result.UpdateAvailable {
		t.Error("should not report update on error")
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "2.0.0", -1},
		{"1.2.0", "1.1.0", 1},
		{"1.0.1", "1.0.0", 1},
		{"0.1.5", "0.1.4", 1},
		{"10.0.0", "9.0.0", 1},
	}

	for _, tt := range tests {
		got := compareVersions(tt.a, tt.b)
		if (got > 0 && tt.want <= 0) || (got < 0 && tt.want >= 0) || (got == 0 && tt.want != 0) {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
