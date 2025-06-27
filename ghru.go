// Package ghru is the GitHub Release Updater package
package ghru

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"
	"time"

	"golang.org/x/mod/semver"
)

// Latest returns the latest Release
func (c *Config) Latest() (Release, error) {
	latestRelease := Release{}
	if err := c.validConfig(); err != nil {
		return latestRelease, err
	}

	currentVersion := c.CurrentVersion
	if !strings.HasPrefix(currentVersion, "v") {
		// Ensure the current version starts with 'v' for semver compatibility
		currentVersion = "v" + currentVersion
	}

	releaseURL := fmt.Sprintf("https://api.github.com/repos/%s/releases", c.Repo)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(releaseURL)
	if err != nil {
		return latestRelease, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 {
		return latestRelease, fmt.Errorf("failed to download file: received status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return latestRelease, err
	}

	// AllReleases will map the semantic version to the release data.
	// The key is prefixed with a "v" (if missing) to ensure semver compatibility, and allow sorting later
	var allReleases = map[string]Release{}

	var releases releases

	if err := json.Unmarshal(body, &releases); err != nil {
		return latestRelease, fmt.Errorf("failed to parse releases: %v", err)
	}

	// Loop through releases
	for _, r := range releases {
		version := r.Tag
		// Ensure the version starts with 'v' for semver compatibility
		if !strings.HasPrefix(version, "v") {
			version = "v" + version
		}

		// Invalid semantic version, skip
		if !semver.IsValid(version) {
			continue
		}

		// Skip if pre-releases are not enabled
		if !c.AllowPreReleases && (semver.Prerelease(version) != "" || r.Prerelease) {
			continue
		}

		if semver.Compare(version, currentVersion) < 0 {
			// Only include releases that are newer than the current version
			continue
		}

		var archiveName bytes.Buffer

		templ, err := template.New("format").Parse(c.ArchiveName)
		if err != nil {
			return latestRelease, fmt.Errorf("failed to parse archive template: %w", err)
		}
		if err := templ.Execute(&archiveName, map[string]any{
			"OS":      runtime.GOOS,
			"Arch":    runtime.GOARCH,
			"Version": version,
		}); err != nil {
			return latestRelease, fmt.Errorf("failed to parse archive template: %v", err)
		}

		for _, a := range r.Assets {
			if !strings.HasPrefix(a.Name, archiveName.String()) {
				continue
			}

			fileType, err := detectFileType(a.Name[len(archiveName.String()):])
			if err != nil {
				continue
			}

			thisRelease := Release{
				Name:         a.Name,
				Tag:          r.Tag,
				Prerelease:   r.Prerelease,
				ReleaseNotes: strings.TrimSpace(r.Notes),
				URL:          a.BrowserDownloadURL,
				Size:         a.Size,
				FileType:     fileType,
			}

			allReleases[version] = thisRelease
			break
		}
	}

	if len(allReleases) == 0 {
		// No releases with suitable assets found
		return latestRelease, fmt.Errorf("no releases found")
	}

	// Create slice of versions from the map keys
	versions := make([]string, len(allReleases))
	i := 0
	for k := range allReleases {
		versions[i] = k
		i++
	}

	// Sort semantic versions ascending
	semver.Sort(versions)

	// Set the latest release to the last version in the sorted slice
	latestRelease = allReleases[versions[len(versions)-1]]

	return latestRelease, nil
}

// SelfUpdate updates the application to the latest Release.
// It returns an error if there is no newer release.
func (c *Config) SelfUpdate() (Release, error) {
	latestRelease, err := c.Latest()
	if err != nil {
		return Release{}, err
	}

	v := c.CurrentVersion
	if !strings.HasPrefix(v, "v") {
		// Ensure the current version starts with 'v' for semver compatibility
		v = "v" + v
	}

	latestCompareVer := latestRelease.Tag
	if !strings.HasPrefix(latestCompareVer, "v") {
		// Ensure the current version starts with 'v' for semver compatibility
		latestCompareVer = "v" + latestCompareVer
	}

	if latestRelease.Tag == c.CurrentVersion || (semver.IsValid(v) && semver.Compare(latestCompareVer, v) <= 0) {
		return latestRelease, fmt.Errorf("no newer releases found (current version: %s)", c.CurrentVersion)
	}

	tmpDir, err := getTempDir()
	if err != nil {
		return latestRelease, err
	}

	outFile := filepath.Join(tmpDir, latestRelease.Name)

	if err := downloadToFile(latestRelease.URL, outFile); err != nil {
		return latestRelease, err
	}

	newExec := filepath.Join(tmpDir, c.BinaryName)
	if runtime.GOOS == "windows" {
		newExec += ".exe"
	}

	switch latestRelease.FileType {
	case "tar.gz", "tar.bz2":
		if err := tarExtract(outFile, tmpDir); err != nil {
			return latestRelease, err
		}
	case "zip":
		if _, err := unzip(outFile, tmpDir); err != nil {
			return latestRelease, err
		}
	default:
		return latestRelease, fmt.Errorf("unsupported file type: %s", latestRelease.FileType)
	}

	if runtime.GOOS != "windows" {
		// Make the new binary executable if *nix or macOS
		err := os.Chmod(newExec, 0755) // #nosec
		if err != nil {
			return latestRelease, err
		}
	}

	// Get the running binary to be replaced
	oldExec, err := os.Executable()
	if err != nil {
		return latestRelease, err
	}

	if err = replaceFile(oldExec, newExec); err != nil {
		return latestRelease, err
	}

	return latestRelease, nil
}

// Validate the configuration
func (c *Config) validConfig() error {
	// Ensure the Repo is set
	if c.Repo == "" {
		return fmt.Errorf("repo must be set")
	}

	// Validate the org/repo format using a regex
	re := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]+/[a-zA-Z0-9][a-zA-Z0-9_.-]+$`)
	if !re.MatchString(c.Repo) {
		return fmt.Errorf("repo must be in the format 'owner/repo'")
	}

	// Ensure the archive name is set
	if c.ArchiveName == "" {
		return fmt.Errorf("archive name must be set")
	}

	// Ensure the binary name is set
	if c.BinaryName == "" {
		return fmt.Errorf("binary name must be set")
	}

	// Ensure the current version is set
	if c.CurrentVersion == "" {
		return fmt.Errorf("current version must be set")
	}

	return nil
}
