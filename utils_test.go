package ghru

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetectFileType(t *testing.T) {
	tests := []struct {
		name      string
		remainder string
		want      string
		wantErr   bool
	}{
		{
			name:      "tar.gz file",
			remainder: "archive.tar.gz",
			want:      "tar.gz",
			wantErr:   false,
		},
		{
			name:      "tgz file",
			remainder: "archive.tgz",
			want:      "tar.gz",
			wantErr:   false,
		},
		{
			name:      "tar.bz2 file",
			remainder: "archive.tar.bz2",
			want:      "tar.bz2",
			wantErr:   false,
		},
		{
			name:      "zip file",
			remainder: "archive.zip",
			want:      "zip",
			wantErr:   false,
		},
		{
			name:      "uppercase tar.gz",
			remainder: "ARCHIVE.TAR.GZ",
			want:      "tar.gz",
			wantErr:   false,
		},
		{
			name:      "uppercase TGZ",
			remainder: "ARCHIVE.TGZ",
			want:      "tar.gz",
			wantErr:   false,
		},
		{
			name:      "uppercase tar.bz2",
			remainder: "ARCHIVE.TAR.BZ2",
			want:      "tar.bz2",
			wantErr:   false,
		},
		{
			name:      "uppercase ZIP",
			remainder: "ARCHIVE.ZIP",
			want:      "zip",
			wantErr:   false,
		},
		{
			name:      "mixed case tar.gz",
			remainder: "Archive.Tar.Gz",
			want:      "tar.gz",
			wantErr:   false,
		},
		{
			name:      "unsupported .tar extension",
			remainder: "archive.tar",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "unsupported .rar extension",
			remainder: "archive.rar",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "unsupported .7z extension",
			remainder: "archive.7z",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "no extension",
			remainder: "archive",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "empty string",
			remainder: "",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "file with path tar.gz",
			remainder: "path/to/archive.tar.gz",
			want:      "tar.gz",
			wantErr:   false,
		},
		{
			name:      "file with path tgz",
			remainder: "path/to/archive.tgz",
			want:      "tar.gz",
			wantErr:   false,
		},
		{
			name:      "file with path zip",
			remainder: "path/to/archive.zip",
			want:      "zip",
			wantErr:   false,
		},
		{
			name:      "partial match .gz only",
			remainder: "file.gz",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "partial match .bz2 only",
			remainder: "file.bz2",
			want:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := detectFileType(tt.remainder)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectFileType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("detectFileType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceFile(t *testing.T) {
	t.Run("successful file replacement", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir := t.TempDir()

		// Create destination file with original content
		dstPath := filepath.Join(tempDir, "testBinary")
		originalContent := []byte("original content")
		if err := os.WriteFile(dstPath, originalContent, 0755); err != nil {
			t.Fatalf("failed to create destination file: %v", err)
		}

		// Create source file with new content
		srcPath := filepath.Join(tempDir, "newSource")
		newContent := []byte("new content")
		if err := os.WriteFile(srcPath, newContent, 0644); err != nil {
			t.Fatalf("failed to create source file: %v", err)
		}

		// Replace the file
		if err := replaceFile(dstPath, srcPath); err != nil {
			t.Fatalf("replaceFile() error = %v", err)
		}

		// Verify destination has new content
		gotContent, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("failed to read destination file: %v", err)
		}
		if string(gotContent) != string(newContent) {
			t.Errorf("destination content = %q, want %q", string(gotContent), string(newContent))
		}

		// Verify source file was removed
		if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
			t.Errorf("source file should be removed, but still exists")
		}

		// Verify .old file was cleaned up
		oldPath := filepath.Join(tempDir, "testBinary.old")
		if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
			if runtime.GOOS != "windows" {
				t.Errorf(".old file should be removed on non-Windows, but still exists")
			}
		}

		// Verify .new file was cleaned up
		newPath := filepath.Join(tempDir, "testBinary.new")
		if _, err := os.Stat(newPath); !os.IsNotExist(err) {
			t.Errorf(".new file should be removed, but still exists")
		}
	})

	t.Run("preserves file permissions", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping permission test on Windows")
		}

		// Create a temporary directory for testing
		tempDir := t.TempDir()

		// Create destination file with specific permissions
		dstPath := filepath.Join(tempDir, "testBinary")
		if err := os.WriteFile(dstPath, []byte("original"), 0700); err != nil {
			t.Fatalf("failed to create destination file: %v", err)
		}

		// Create source file
		srcPath := filepath.Join(tempDir, "newSource")
		if err := os.WriteFile(srcPath, []byte("new"), 0644); err != nil {
			t.Fatalf("failed to create source file: %v", err)
		}

		// Replace the file
		if err := replaceFile(dstPath, srcPath); err != nil {
			t.Fatalf("replaceFile() error = %v", err)
		}

		// Verify permissions are preserved
		info, err := os.Stat(dstPath)
		if err != nil {
			t.Fatalf("failed to stat destination file: %v", err)
		}
		if info.Mode().Perm() != 0700 {
			t.Errorf("destination permissions = %o, want %o", info.Mode().Perm(), 0700)
		}
	})

	t.Run("source file does not exist", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create destination file
		dstPath := filepath.Join(tempDir, "testBinary")
		if err := os.WriteFile(dstPath, []byte("content"), 0755); err != nil {
			t.Fatalf("failed to create destination file: %v", err)
		}

		// Try to replace with non-existent source
		srcPath := filepath.Join(tempDir, "nonexistent")
		err := replaceFile(dstPath, srcPath)
		if err == nil {
			t.Error("replaceFile() expected error for non-existent source, got nil")
		}
	})

	t.Run("handles different file sizes", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create destination file with small content
		dstPath := filepath.Join(tempDir, "testBinary")
		if err := os.WriteFile(dstPath, []byte("small"), 0755); err != nil {
			t.Fatalf("failed to create destination file: %v", err)
		}

		// Create source file with larger content
		srcPath := filepath.Join(tempDir, "newSource")
		largeContent := make([]byte, 10240) // 10KB
		for i := range largeContent {
			largeContent[i] = byte(i % 256)
		}
		if err := os.WriteFile(srcPath, largeContent, 0644); err != nil {
			t.Fatalf("failed to create source file: %v", err)
		}

		// Replace the file
		if err := replaceFile(dstPath, srcPath); err != nil {
			t.Fatalf("replaceFile() error = %v", err)
		}

		// Verify destination has new content
		gotContent, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("failed to read destination file: %v", err)
		}
		if len(gotContent) != len(largeContent) {
			t.Errorf("destination size = %d, want %d", len(gotContent), len(largeContent))
		}
	})
}
