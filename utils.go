package ghru

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	// Temporary directory
	tempDir string
)

// Detect file type based on the remainder of the filename passed to it.
// This only returns supported formats
func detectFileType(remainder string) (string, error) {
	remainder = strings.ToLower(remainder)

	if strings.HasSuffix(remainder, ".tar.gz") || strings.HasSuffix(remainder, ".tgz") {
		return "tar.gz", nil
	}
	if strings.HasSuffix(remainder, ".tar.bz2") {
		return "tar.bz2", nil
	}
	if strings.HasSuffix(remainder, ".zip") {
		return "zip", nil
	}

	return "", fmt.Errorf("unsupported file type: %s", remainder)
}

// DownloadToFile downloads a URL to a file
func downloadToFile(url, fileName string) error {
	// Get the data
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file from %s: %w", url, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 {
		return fmt.Errorf("failed to download file: received status code %d", resp.StatusCode)
	}

	// Create the file
	out, err := os.Create(filepath.Clean(fileName))
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", fileName, err)
	}

	defer func() { _ = out.Close() }()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)

	return err
}

// ReplaceFile replaces one file with another.
// Running files cannot be overwritten, so it has to be moved
// and the new binary saved to the original path. This requires
// read & write permissions to both the original file and directory.
// Note, on Windows it is not possible to delete a running program,
// so the old exe is renamed and moved to os.TempDir()
func replaceFile(dst, src string) error {
	// Open the source file for reading
	source, err := os.Open(filepath.Clean(src))
	if err != nil {
		return err
	}

	// Destination directory eg: /usr/local/bin
	dstDir := filepath.Dir(dst)
	// Binary filename
	binaryFilename := filepath.Base(dst)
	// Old binary tmp name
	dstOld := fmt.Sprintf("%s.old", binaryFilename)
	// New binary tmp name
	dstNew := fmt.Sprintf("%s.new", binaryFilename)
	// Absolute path of new tmp file
	newTmpAbs := filepath.Join(dstDir, dstNew)
	// Absolute path of old tmp file
	oldTmpAbs := filepath.Join(dstDir, dstOld)

	// Get src permissions, ignore errors
	fi, _ := os.Stat(dst)
	srcPerms := fi.Mode().Perm()

	// Create the new file
	tmpNew, err := os.OpenFile(filepath.Clean(newTmpAbs), os.O_CREATE|os.O_RDWR, srcPerms) // #nosec
	if err != nil {
		return err
	}

	// Copy new binary to <binary>.new
	if _, err := io.Copy(tmpNew, source); err != nil {
		return err
	}

	// Close immediately else Windows has a fit
	if err := tmpNew.Close(); err != nil {
		return err
	}

	if err := source.Close(); err != nil {
		return err
	}

	// Rename the current executable to <binary>.old
	if err := os.Rename(dst, oldTmpAbs); err != nil {
		return err
	}

	// Rename the <binary>.new to current executable
	if err := os.Rename(newTmpAbs, dst); err != nil {
		return err
	}

	// Delete the old binary
	if runtime.GOOS == "windows" {
		tmpDir := os.TempDir()
		delFile := filepath.Join(tmpDir, filepath.Base(oldTmpAbs))
		if err := os.Rename(oldTmpAbs, delFile); err != nil {
			return err
		}
	} else {
		if err := os.Remove(oldTmpAbs); err != nil {
			return err
		}
	}

	// Remove the src file
	return os.Remove(src)
}

// GetTempDir will create & return a temporary directory if one has not been specified
func getTempDir() (string, error) {
	if tempDir == "" {
		randBytes := make([]byte, 6)
		if _, err := rand.Read(randBytes); err != nil {
			return "", err
		}
		tempDir = filepath.Join(os.TempDir(), "updater-"+hex.EncodeToString(randBytes))
	}

	err := mkDirIfNotExists(tempDir)

	return tempDir, err
}

// MkDirIfNotExists will create a directory if it doesn't exist
func mkDirIfNotExists(path string) error {
	if !isDir(path) {
		return os.MkdirAll(path, os.ModePerm) // #nosec
	}

	return nil
}

// IsDir returns if a path is a directory
func isDir(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) || !info.IsDir() {
		return false
	}

	return true
}
