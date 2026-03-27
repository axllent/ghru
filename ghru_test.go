package ghru

import (
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
