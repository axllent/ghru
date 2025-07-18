package ghru

import (
	"archive/tar"
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// TarExtract extracts a archive from the file inputFilePath.
// It tries to create the directory structure outputFilePath contains if it doesn't exist.
// It returns potential errors to be checked or nil if everything works.
func tarExtract(inputFilePath, outputFilePath string) (err error) {
	outputFilePath = stripTrailingSlashes(outputFilePath)
	inputFilePath, outputFilePath, err = makeAbsolute(inputFilePath, outputFilePath)
	if err != nil {
		return err
	}
	undoDir, err := mkdirAll(outputFilePath, 0750)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			undoDir()
		}
	}()

	return extractArchive(inputFilePath, outputFilePath)
}

// Creates all directories with os.MkdirAll and returns a function to remove the first created directory so cleanup is possible.
func mkdirAll(dirPath string, perm os.FileMode) (func(), error) {
	var undoDir string

	for p := dirPath; ; p = filepath.Dir(p) {
		fInfo, err := os.Stat(p)
		if err == nil {
			if fInfo.IsDir() {
				break
			}

			fInfo, err = os.Lstat(p)
			if err != nil {
				return nil, err
			}

			if fInfo.IsDir() {
				break
			}

			return nil, fmt.Errorf("mkdirAll (%s): %v", p, syscall.ENOTDIR)
		}

		if os.IsNotExist(err) {
			undoDir = p
		} else {
			return nil, err
		}
	}

	if undoDir == "" {
		return func() {}, nil
	}

	if err := os.MkdirAll(dirPath, perm); err != nil {
		return nil, err
	}

	return func() {
		if err := os.RemoveAll(undoDir); err != nil {
			panic(err)
		}
	}, nil
}

// Remove trailing slash if any.
func stripTrailingSlashes(path string) string {
	if len(path) > 0 && path[len(path)-1] == '/' {
		path = path[0 : len(path)-1]
	}

	return path
}

// Make input and output paths absolute.
func makeAbsolute(inputFilePath, outputFilePath string) (string, string, error) {
	inputFilePath, err := filepath.Abs(inputFilePath)
	if err == nil {
		outputFilePath, err = filepath.Abs(outputFilePath)
	}

	return inputFilePath, outputFilePath, err
}

// Extract the file in filePath to directory.
// it supports different archive formats like tar.gz, tgz & tar.bz2
func extractArchive(filePath string, directory string) error {
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Error closing file: %s\n", err)
		}
	}()

	var compressReader io.Reader

	fileType, err := detectFileType(filePath)
	if err != nil {
		return fmt.Errorf("error detecting file type: %w", err)
	}

	switch fileType {
	case "tar.gz", "tgz":
		// Do nothing, continue with tar.gz extraction
		compressReader, err = gzip.NewReader(bufio.NewReader(file))
		if err != nil {
			return err
		}
	case "tar.bz2":
		// Bzip2 compression
		compressReader = bzip2.NewReader(bufio.NewReader(file))
	default:
		// Unknown file format
		return fmt.Errorf("unsupported file type: %s", filePath)
	}

	tarReader := tar.NewReader(compressReader)

	// Post extraction directory permissions & timestamps
	type DirInfo struct {
		Path   string
		Header *tar.Header
	}

	// Slice to add all extracted directory info for post-processing
	postExtraction := []DirInfo{}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		fileInfo := header.FileInfo()
		// Paths could contain a '..', is used in a file system operations
		if strings.Contains(fileInfo.Name(), "..") {
			continue
		}
		dir := filepath.Join(directory, filepath.Dir(header.Name))
		filename := filepath.Join(dir, fileInfo.Name())

		if fileInfo.IsDir() {
			// Create the directory 755 in case writing permissions prohibit writing before files added
			if err := os.MkdirAll(filename, 0750); err != nil {
				return err
			}

			// Set file ownership (if allowed)
			// Chtimes() && Chmod() only set after once extraction is complete
			_ = os.Chown(filename, header.Uid, header.Gid)

			// Add directory info to slice to process afterwards
			postExtraction = append(postExtraction, DirInfo{filename, header})
			continue
		}

		// make sure parent directory exists (may not be included in tar)
		if !fileInfo.IsDir() && !isDir(dir) {
			err = os.MkdirAll(dir, 0750)
			if err != nil {
				return err
			}
		}

		file, err := os.Create(filepath.Clean(filename))
		if err != nil {
			return err
		}

		writer := bufio.NewWriter(file)

		buffer := make([]byte, 4096)
		for {
			n, err := tarReader.Read(buffer)
			if err != nil && err != io.EOF {
				panic(err)
			}
			if n == 0 {
				break
			}

			_, err = writer.Write(buffer[:n])
			if err != nil {
				return err
			}
		}

		err = writer.Flush()
		if err != nil {
			return err
		}

		err = file.Close()
		if err != nil {
			return err
		}

		// set file permissions, timestamps & uid/gid
		_ = os.Chmod(filename, os.FileMode(header.Mode)) // #nosec
		_ = os.Chtimes(filename, header.AccessTime, header.ModTime)
		_ = os.Chown(filename, header.Uid, header.Gid)
	}

	if len(postExtraction) > 0 {
		for _, dir := range postExtraction {
			_ = os.Chtimes(dir.Path, dir.Header.AccessTime, dir.Header.ModTime)
			_ = os.Chmod(dir.Path, dir.Header.FileInfo().Mode().Perm())
		}
	}

	return nil
}
