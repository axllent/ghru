package ghru

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
)

func TestValidConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr string
	}{
		{
			name:    "empty repo",
			config:  Config{ArchiveName: "app", BinaryName: "app", CurrentVersion: "1.0.0"},
			wantErr: "repo must be set",
		},
		{
			name:    "repo without slash",
			config:  Config{Repo: "noslash", ArchiveName: "app", BinaryName: "app", CurrentVersion: "1.0.0"},
			wantErr: "repo must be in the format 'owner/repo'",
		},
		{
			name:    "repo with leading slash",
			config:  Config{Repo: "/owner/repo", ArchiveName: "app", BinaryName: "app", CurrentVersion: "1.0.0"},
			wantErr: "repo must be in the format 'owner/repo'",
		},
		{
			name:    "empty archive name",
			config:  Config{Repo: "owner/repo", BinaryName: "app", CurrentVersion: "1.0.0"},
			wantErr: "archive name must be set",
		},
		{
			name:    "empty binary name",
			config:  Config{Repo: "owner/repo", ArchiveName: "app", CurrentVersion: "1.0.0"},
			wantErr: "binary name must be set",
		},
		{
			name:    "empty current version",
			config:  Config{Repo: "owner/repo", ArchiveName: "app", BinaryName: "app"},
			wantErr: "current version must be set",
		},
		{
			name:   "valid config",
			config: Config{Repo: "owner/repo", ArchiveName: "app", BinaryName: "app", CurrentVersion: "1.0.0"},
		},
		{
			name:   "valid config with dashes and dots in repo",
			config: Config{Repo: "my-org/my.repo-v2", ArchiveName: "app", BinaryName: "app", CurrentVersion: "1.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validConfig()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func makeReleasesHandler(tag string, prerelease bool) http.HandlerFunc {
	os := runtime.GOOS
	arch := runtime.GOARCH
	assetName := fmt.Sprintf("app-%s-%s.tar.gz", os, arch)
	return func(w http.ResponseWriter, _ *http.Request) {
		resp := []map[string]any{
			{
				"tag_name":   tag,
				"prerelease": prerelease,
				"body":       "release notes",
				"assets": []map[string]any{
					{
						"name":                 assetName,
						"browser_download_url": "http://example.com/" + assetName,
						"size":                 1024,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func TestLatestVersionComparison(t *testing.T) {
	tests := []struct {
		name           string
		releaseTag     string
		currentVersion string
		prerelease     bool
		allowPre       bool
		wantErr        string
	}{
		{
			name:           "newer release available",
			releaseTag:     "v2.0.0",
			currentVersion: "1.0.0",
		},
		{
			name:           "same version (prefixed)",
			releaseTag:     "v1.0.0",
			currentVersion: "1.0.0",
		},
		{
			name:           "same version (exact)",
			releaseTag:     "v1.0.1",
			currentVersion: "v1.0.1",
		},
		{
			name:           "older release only",
			releaseTag:     "v0.9.0",
			currentVersion: "1.0.0",
			wantErr:        "no releases found",
		},
		{
			name:           "release tag without v prefix",
			releaseTag:     "2.0.0",
			currentVersion: "1.0.0",
		},
		{
			name:           "current version with v prefix",
			releaseTag:     "v2.0.0",
			currentVersion: "v1.0.0",
		},
		{
			name:           "patch version bump",
			releaseTag:     "v1.0.1",
			currentVersion: "1.0.0",
		},
		{
			name:           "pre-release alpha to beta",
			releaseTag:     "v1.0.1-beta",
			currentVersion: "v1.0.1-alpha",
			prerelease:     true,
			allowPre:       true,
		},
		{
			name:           "pre-release beta to alpha",
			releaseTag:     "v1.0.1-alpha",
			currentVersion: "v1.0.1-beta",
			prerelease:     true,
			allowPre:       true,
			wantErr:        "no releases found",
		},
		{
			name:           "pre-release skipped when not allowed",
			releaseTag:     "v2.0.0-beta.1",
			currentVersion: "1.0.0",
			prerelease:     true,
			allowPre:       false,
			wantErr:        "no releases found",
		},
		{
			name:           "pre-release included when allowed",
			releaseTag:     "v2.0.0-beta.1",
			currentVersion: "1.0.0",
			prerelease:     true,
			allowPre:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(makeReleasesHandler(tt.releaseTag, tt.prerelease))
			defer srv.Close()

			cfg := Config{
				Repo:             "owner/repo",
				ArchiveName:      "app-{{.OS}}-{{.Arch}}",
				BinaryName:       "app",
				CurrentVersion:   tt.currentVersion,
				AllowPreReleases: tt.allowPre,
				apiBaseURL:       srv.URL,
			}

			_, err := cfg.Latest()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
