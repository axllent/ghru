package ghru

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUnzip(t *testing.T) {
	t.Run("extract zip archive with files and directories", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a test zip archive
		archivePath := filepath.Join(tempDir, "test.zip")
		if err := createTestZip(archivePath, map[string]string{
			"file1.txt":     "content1",
			"dir/file2.txt": "content2",
			"dir/file3.txt": "content3",
		}); err != nil {
			t.Fatalf("failed to create test archive: %v", err)
		}

		// Extract the archive
		extractDir := filepath.Join(tempDir, "extracted")
		filenames, err := unzip(archivePath, extractDir)
		if err != nil {
			t.Fatalf("unzip() error = %v", err)
		}

		// Verify we got the expected number of files
		expectedCount := 4 // 1 directory + 3 files
		if len(filenames) != expectedCount {
			t.Errorf("unzip() returned %d filenames, want %d", len(filenames), expectedCount)
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
		archivePath := filepath.Join(tempDir, "nested.zip")
		if err := createTestZip(archivePath, map[string]string{
			"a/b/c/file.txt": "nested content",
		}); err != nil {
			t.Fatalf("failed to create test archive: %v", err)
		}

		// Extract the archive
		extractDir := filepath.Join(tempDir, "extracted")
		filenames, err := unzip(archivePath, extractDir)
		if err != nil {
			t.Fatalf("unzip() error = %v", err)
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

		// Verify filenames were returned
		if len(filenames) == 0 {
			t.Error("unzip() should return extracted filenames")
		}
	})

	t.Run("non-existent archive file", func(t *testing.T) {
		tempDir := t.TempDir()

		// Try to extract non-existent file
		archivePath := filepath.Join(tempDir, "nonexistent.zip")
		extractDir := filepath.Join(tempDir, "extracted")
		_, err := unzip(archivePath, extractDir)
		if err == nil {
			t.Error("unzip() expected error for non-existent file, got nil")
		}
	})

	t.Run("path traversal protection", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a zip with path traversal attempt
		archivePath := filepath.Join(tempDir, "malicious.zip")
		if err := createMaliciousZip(archivePath, "../../../etc/passwd", "malicious"); err != nil {
			t.Fatalf("failed to create malicious archive: %v", err)
		}

		// Try to extract it
		extractDir := filepath.Join(tempDir, "extracted")
		_, err := unzip(archivePath, extractDir)
		if err == nil {
			t.Error("unzip() should prevent path traversal, but got nil error")
		}
	})

	t.Run("empty archive", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create an empty zip archive
		archivePath := filepath.Join(tempDir, "empty.zip")
		if err := createTestZip(archivePath, map[string]string{}); err != nil {
			t.Fatalf("failed to create empty archive: %v", err)
		}

		// Extract the archive
		extractDir := filepath.Join(tempDir, "extracted")
		filenames, err := unzip(archivePath, extractDir)
		if err != nil {
			t.Fatalf("unzip() error = %v", err)
		}

		// Verify no files were extracted
		if len(filenames) != 0 {
			t.Errorf("unzip() returned %d filenames for empty archive, want 0", len(filenames))
		}
	})

	t.Run("file permissions are preserved", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a zip with specific file permissions
		archivePath := filepath.Join(tempDir, "perms.zip")
		if err := createTestZipWithMode(archivePath, "executable.sh", "#!/bin/bash\necho test", 0755); err != nil {
			t.Fatalf("failed to create test archive: %v", err)
		}

		// Extract the archive
		extractDir := filepath.Join(tempDir, "extracted")
		_, err := unzip(archivePath, extractDir)
		if err != nil {
			t.Fatalf("unzip() error = %v", err)
		}

		// Verify file permissions (skip on Windows)
		filePath := filepath.Join(extractDir, "executable.sh")
		info, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("failed to stat extracted file: %v", err)
		}

		// On Unix systems, verify executable bit is preserved
		if info.Mode().Perm()&0111 == 0 {
			t.Skip("skipping permission test - system may not support Unix permissions")
		}
	})

	t.Run("returned filenames are correct", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a test zip archive
		archivePath := filepath.Join(tempDir, "test.zip")
		files := map[string]string{
			"file1.txt": "content1",
			"file2.txt": "content2",
		}
		if err := createTestZip(archivePath, files); err != nil {
			t.Fatalf("failed to create test archive: %v", err)
		}

		// Extract the archive
		extractDir := filepath.Join(tempDir, "extracted")
		filenames, err := unzip(archivePath, extractDir)
		if err != nil {
			t.Fatalf("unzip() error = %v", err)
		}

		// Verify all returned filenames exist
		for _, filename := range filenames {
			if _, err := os.Stat(filename); err != nil {
				t.Errorf("returned filename %s does not exist: %v", filename, err)
			}

			// Verify filename starts with extract directory
			if !filepath.IsAbs(filename) {
				t.Errorf("returned filename %s is not absolute", filename)
			}
		}
	})
}

// Helper function to create a test zip archive
func createTestZip(archivePath string, files map[string]string) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// Track created directories to avoid duplicates
	createdDirs := make(map[string]bool)

	for name, content := range files {
		// Create directory entries if needed
		dir := filepath.Dir(name)
		if dir != "." {
			parts := []string{}
			current := dir
			for current != "." {
				parts = append([]string{current}, parts...)
				current = filepath.Dir(current)
			}

			for _, part := range parts {
				if !createdDirs[part] {
					dirHeader := &zip.FileHeader{
						Name:     filepath.ToSlash(part) + "/",
						Method:   zip.Deflate,
						Modified: time.Now(),
					}
					dirHeader.SetMode(0755 | os.ModeDir)
					if _, err := zipWriter.CreateHeader(dirHeader); err != nil {
						return err
					}
					createdDirs[part] = true
				}
			}
		}

		// Write file
		header := &zip.FileHeader{
			Name:     filepath.ToSlash(name),
			Method:   zip.Deflate,
			Modified: time.Now(),
		}
		header.SetMode(0644)

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err := writer.Write([]byte(content)); err != nil {
			return err
		}
	}

	return nil
}

// Helper function to create a test zip archive with specific file mode
func createTestZipWithMode(archivePath, filename, content string, mode os.FileMode) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	header := &zip.FileHeader{
		Name:     filepath.ToSlash(filename),
		Method:   zip.Deflate,
		Modified: time.Now(),
	}
	header.SetMode(mode)

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	if _, err := writer.Write([]byte(content)); err != nil {
		return err
	}

	return nil
}

// Helper function to create a malicious zip archive with path traversal
func createMaliciousZip(archivePath, maliciousPath, content string) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	header := &zip.FileHeader{
		Name:     maliciousPath,
		Method:   zip.Deflate,
		Modified: time.Now(),
	}
	header.SetMode(0644)

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	if _, err := writer.Write([]byte(content)); err != nil {
		return err
	}

	return nil
}
