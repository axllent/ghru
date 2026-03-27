package ghru

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFileType(t *testing.T) {
	tests := []struct {
		remainder string
		want      string
		wantErr   bool
	}{
		{".tar.gz", "tar.gz", false},
		{".tgz", "tar.gz", false},
		{".tar.bz2", "tar.bz2", false},
		{".zip", "zip", false},
		{"-linux-amd64.tar.gz", "tar.gz", false},
		{"-windows-amd64.zip", "zip", false},
		{".exe", "", true},
		{".tar", "", true},
		{"", "", true},
		{".gz", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.remainder, func(t *testing.T) {
			got, err := detectFileType(tt.remainder)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil (result %q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSha256Checksum(t *testing.T) {
	const content = "hello sha256"

	dir := t.TempDir()
	path := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	h := sha256.New()
	h.Write([]byte(content))
	expected := fmt.Sprintf("%x", h.Sum(nil))

	got, err := sha256Checksum(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Errorf("checksum = %q, want %q", got, expected)
	}

	t.Run("nonexistent file returns error", func(t *testing.T) {
		_, err := sha256Checksum(filepath.Join(dir, "does_not_exist.txt"))
		if err == nil {
			t.Fatal("expected error for nonexistent file, got nil")
		}
	})
}

func TestIsDir(t *testing.T) {
	dir := t.TempDir()

	t.Run("existing directory returns true", func(t *testing.T) {
		if !isDir(dir) {
			t.Errorf("isDir(%q) = false, want true", dir)
		}
	})

	t.Run("nonexistent path returns false", func(t *testing.T) {
		if isDir(filepath.Join(dir, "no_such_dir")) {
			t.Error("isDir(nonexistent) = true, want false")
		}
	})

	t.Run("file path returns false", func(t *testing.T) {
		path := filepath.Join(dir, "file.txt")
		if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		if isDir(path) {
			t.Errorf("isDir(file) = true, want false")
		}
	})
}

func TestMkDirIfNotExists(t *testing.T) {
	t.Run("creates directory when absent", func(t *testing.T) {
		newDir := filepath.Join(t.TempDir(), "newsubdir")
		if err := mkDirIfNotExists(newDir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isDir(newDir) {
			t.Error("directory was not created")
		}
	})

	t.Run("no-op when directory already exists", func(t *testing.T) {
		if err := mkDirIfNotExists(t.TempDir()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestDownloadToFile(t *testing.T) {
	t.Run("downloads content successfully", func(t *testing.T) {
		const body = "binary content data"
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, body)
		}))
		defer srv.Close()

		out := filepath.Join(t.TempDir(), "out.bin")
		if err := downloadToFile(srv.URL, out); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got, err := os.ReadFile(out)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != body {
			t.Errorf("content = %q, want %q", string(got), body)
		}
	})

	t.Run("returns error on 404", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		if err := downloadToFile(srv.URL, filepath.Join(t.TempDir(), "out.bin")); err == nil {
			t.Fatal("expected error for 404 response, got nil")
		}
	})

	t.Run("returns error on 500", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		if err := downloadToFile(srv.URL, filepath.Join(t.TempDir(), "out.bin")); err == nil {
			t.Fatal("expected error for 500 response, got nil")
		}
	})
}

func TestReplaceFile(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "binary")
	src := filepath.Join(dir, "newsrc")

	if err := os.WriteFile(dst, []byte("old content"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("new content"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := replaceFile(dst, src); err != nil {
		t.Fatalf("replaceFile failed: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read dst after replace: %v", err)
	}
	if string(got) != "new content" {
		t.Errorf("dst content = %q, want %q", string(got), "new content")
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("src file should have been removed by replaceFile")
	}

	if _, err := os.Stat(dst + ".old"); !os.IsNotExist(err) {
		t.Error(".old temp file should have been removed by replaceFile")
	}
}
