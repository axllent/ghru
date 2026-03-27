package ghru

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

// createTarGzArchive writes a .tar.gz file containing the given name->content entries.
func createTarGzArchive(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.tar.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	for name, content := range files {
		hdr := &tar.Header{
			Name:     name,
			Typeflag: tar.TypeReg,
			Mode:     0644,
			Size:     int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestTarExtract(t *testing.T) {
	t.Run("extracts flat file from tar.gz", func(t *testing.T) {
		tarPath := createTarGzArchive(t, map[string]string{
			"hello.txt": "hello tar",
		})
		dest := t.TempDir()

		if err := tarExtract(tarPath, dest); err != nil {
			t.Fatalf("tarExtract failed: %v", err)
		}

		got, err := os.ReadFile(filepath.Join(dest, "hello.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "hello tar" {
			t.Errorf("file content = %q, want %q", string(got), "hello tar")
		}
	})

	t.Run("extracts file in subdirectory from tar.gz", func(t *testing.T) {
		tarPath := createTarGzArchive(t, map[string]string{
			"subdir/nested.txt": "nested tar content",
		})
		dest := t.TempDir()

		if err := tarExtract(tarPath, dest); err != nil {
			t.Fatalf("tarExtract failed: %v", err)
		}

		got, err := os.ReadFile(filepath.Join(dest, "subdir", "nested.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "nested tar content" {
			t.Errorf("nested file content = %q, want %q", string(got), "nested tar content")
		}
	})

	t.Run("extracts multiple files from tar.gz", func(t *testing.T) {
		tarPath := createTarGzArchive(t, map[string]string{
			"a.txt": "aaa",
			"b.txt": "bbb",
		})
		dest := t.TempDir()

		if err := tarExtract(tarPath, dest); err != nil {
			t.Fatalf("tarExtract failed: %v", err)
		}

		for _, name := range []string{"a.txt", "b.txt"} {
			if _, err := os.Stat(filepath.Join(dest, name)); err != nil {
				t.Errorf("expected %s to exist: %v", name, err)
			}
		}
	})

	t.Run("rejects TarSlip path traversal", func(t *testing.T) {
		tarPath := createTarGzArchive(t, map[string]string{
			"../../evil.txt": "malicious",
		})
		dest := t.TempDir()

		if err := tarExtract(tarPath, dest); err == nil {
			t.Fatal("expected error for TarSlip path traversal, got nil")
		}
	})
}
