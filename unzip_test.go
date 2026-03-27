package ghru

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

// createZipArchive writes a zip file containing the given name->content entries.
func createZipArchive(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	zw := zip.NewWriter(f)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestUnzip(t *testing.T) {
	t.Run("extracts flat file", func(t *testing.T) {
		zipPath := createZipArchive(t, map[string]string{
			"hello.txt": "hello world",
		})
		dest := t.TempDir()

		filenames, err := unzip(zipPath, dest)
		if err != nil {
			t.Fatalf("unzip failed: %v", err)
		}
		if len(filenames) != 1 {
			t.Errorf("got %d filenames, want 1", len(filenames))
		}

		got, err := os.ReadFile(filepath.Join(dest, "hello.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "hello world" {
			t.Errorf("file content = %q, want %q", string(got), "hello world")
		}
	})

	t.Run("extracts file in subdirectory", func(t *testing.T) {
		zipPath := createZipArchive(t, map[string]string{
			"subdir/nested.txt": "nested content",
		})
		dest := t.TempDir()

		if _, err := unzip(zipPath, dest); err != nil {
			t.Fatalf("unzip failed: %v", err)
		}

		got, err := os.ReadFile(filepath.Join(dest, "subdir", "nested.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "nested content" {
			t.Errorf("nested file content = %q, want %q", string(got), "nested content")
		}
	})

	t.Run("extracts multiple files", func(t *testing.T) {
		zipPath := createZipArchive(t, map[string]string{
			"a.txt": "aaa",
			"b.txt": "bbb",
		})
		dest := t.TempDir()

		filenames, err := unzip(zipPath, dest)
		if err != nil {
			t.Fatalf("unzip failed: %v", err)
		}
		if len(filenames) != 2 {
			t.Errorf("got %d filenames, want 2", len(filenames))
		}
	})

	t.Run("rejects ZipSlip path traversal", func(t *testing.T) {
		zipPath := createZipArchive(t, map[string]string{
			"../../evil.txt": "malicious",
		})
		dest := t.TempDir()

		_, err := unzip(zipPath, dest)
		if err == nil {
			t.Fatal("expected error for ZipSlip path traversal, got nil")
		}
	})
}
