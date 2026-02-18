package ghru

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExtractArchive(t *testing.T) {
	t.Run("extract tar.gz archive with files and directories", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a test tar.gz archive
		archivePath := filepath.Join(tempDir, "test.tar.gz")
		if err := createTestTarGz(archivePath, map[string]string{
			"file1.txt":     "content1",
			"dir/file2.txt": "content2",
			"dir/file3.txt": "content3",
		}); err != nil {
			t.Fatalf("failed to create test archive: %v", err)
		}

		// Extract the archive
		extractDir := filepath.Join(tempDir, "extracted")
		if err := extractArchive(archivePath, extractDir); err != nil {
			t.Fatalf("extractArchive() error = %v", err)
		}

		// Verify extracted files
		tests := []struct {
			path    string
			content string
		}{
			{"file1.txt", "content1"},
			{"dir/file2.txt", "content2"},
			{"dir/file3.txt", "content3"},
		}

		for _, tt := range tests {
			fullPath := filepath.Join(extractDir, tt.path)
			content, err := os.ReadFile(fullPath)
			if err != nil {
				t.Errorf("failed to read %s: %v", tt.path, err)
				continue
			}
			if string(content) != tt.content {
				t.Errorf("content of %s = %q, want %q", tt.path, string(content), tt.content)
			}
		}

		// Verify directory exists
		dirPath := filepath.Join(extractDir, "dir")
		if info, err := os.Stat(dirPath); err != nil {
			t.Errorf("directory 'dir' not created: %v", err)
		} else if !info.IsDir() {
			t.Errorf("'dir' is not a directory")
		}
	})

	t.Run("extract archive with nested directories", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create archive with nested directories
		archivePath := filepath.Join(tempDir, "nested.tar.gz")
		if err := createTestTarGz(archivePath, map[string]string{
			"a/b/c/file.txt": "nested content",
		}); err != nil {
			t.Fatalf("failed to create test archive: %v", err)
		}

		// Extract the archive
		extractDir := filepath.Join(tempDir, "extracted")
		if err := extractArchive(archivePath, extractDir); err != nil {
			t.Fatalf("extractArchive() error = %v", err)
		}

		// Verify nested file exists
		nestedFile := filepath.Join(extractDir, "a/b/c/file.txt")
		content, err := os.ReadFile(nestedFile)
		if err != nil {
			t.Fatalf("failed to read nested file: %v", err)
		}
		if string(content) != "nested content" {
			t.Errorf("content = %q, want %q", string(content), "nested content")
		}
	})

	t.Run("unsupported file type", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a non-archive file
		filePath := filepath.Join(tempDir, "test.txt")
		if err := os.WriteFile(filePath, []byte("not an archive"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Try to extract it
		extractDir := filepath.Join(tempDir, "extracted")
		err := extractArchive(filePath, extractDir)
		if err == nil {
			t.Error("extractArchive() expected error for unsupported file type, got nil")
		}
	})

	t.Run("non-existent archive file", func(t *testing.T) {
		tempDir := t.TempDir()

		// Try to extract non-existent file
		archivePath := filepath.Join(tempDir, "nonexistent.tar.gz")
		extractDir := filepath.Join(tempDir, "extracted")
		err := extractArchive(archivePath, extractDir)
		if err == nil {
			t.Error("extractArchive() expected error for non-existent file, got nil")
		}
	})

	t.Run("path traversal protection", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create an archive with path traversal attempt
		archivePath := filepath.Join(tempDir, "malicious.tar.gz")
		if err := createMaliciousTarGz(archivePath, "../../../etc/passwd", "malicious"); err != nil {
			t.Fatalf("failed to create malicious archive: %v", err)
		}

		// Try to extract it
		extractDir := filepath.Join(tempDir, "extracted")
		err := extractArchive(archivePath, extractDir)
		if err == nil {
			t.Error("extractArchive() should prevent path traversal, but got nil error")
		}
	})

	t.Run("empty archive", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create an empty tar.gz archive
		archivePath := filepath.Join(tempDir, "empty.tar.gz")
		if err := createTestTarGz(archivePath, map[string]string{}); err != nil {
			t.Fatalf("failed to create empty archive: %v", err)
		}

		// Extract the archive
		extractDir := filepath.Join(tempDir, "extracted")
		if err := extractArchive(archivePath, extractDir); err != nil {
			t.Fatalf("extractArchive() error = %v", err)
		}

		// Empty archive extraction succeeds without creating directory
		// This is expected behavior - no files means no directory creation
	})
}

// Helper function to create a test tar.gz archive
func createTestTarGz(archivePath string, files map[string]string) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	for name, content := range files {
		// Create directory entries if needed
		dir := filepath.Dir(name)
		if dir != "." {
			current := ""
			for _, part := range filepath.SplitList(filepath.ToSlash(dir)) {
				if current == "" {
					current = part
				} else {
					current = filepath.Join(current, part)
				}

				header := &tar.Header{
					Name:     current + "/",
					Mode:     0755,
					ModTime:  time.Now(),
					Typeflag: tar.TypeDir,
				}
				if err := tarWriter.WriteHeader(header); err != nil {
					// Ignore duplicate directory errors
					continue
				}
			}
		}

		// Write file
		header := &tar.Header{
			Name:    name,
			Mode:    0644,
			Size:    int64(len(content)),
			ModTime: time.Now(),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			return err
		}
	}

	return nil
}

// Helper function to create a malicious tar.gz archive with path traversal
func createMaliciousTarGz(archivePath, maliciousPath, content string) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	header := &tar.Header{
		Name:    maliciousPath,
		Mode:    0644,
		Size:    int64(len(content)),
		ModTime: time.Now(),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}
	if _, err := tarWriter.Write([]byte(content)); err != nil {
		return err
	}

	return nil
}
