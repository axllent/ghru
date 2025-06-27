package ghru

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (src) to an output directory (dest).
func unzip(src string, dest string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		filePath := filepath.Join(dest, filepath.Clean(f.Name))

		// Check for ZipSlip vulnerability: Ensure the file path is within the destination directory.
		// More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(filePath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", filePath)
		}

		filenames = append(filenames, filePath)

		if f.FileInfo().IsDir() {
			// Make Folder
			if err := os.MkdirAll(filePath, os.ModePerm); /* #nosec */ err != nil {
				return filenames, fmt.Errorf("failed to create directory %s: %w", filePath, err)
			}
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); /* #nosec */ err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(filepath.Clean(filePath), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc) // #nosec - file is streamed from zip to file

		// Close the file without defer to close before next iteration of loop
		if err := outFile.Close(); err != nil {
			return filenames, err
		}

		if err := rc.Close(); err != nil {
			return filenames, err
		}

		if err != nil {
			return filenames, err
		}
	}

	return filenames, nil
}
