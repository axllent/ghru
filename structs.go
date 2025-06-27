package ghru

// Config is a ghru configuration
type Config struct {
	// GitHub repository in the format "owner/repo"
	Repo string

	// ArchiveName is naming convention for the archives in GitHub releases.
	// Example: "app-{{.OS}}-{{.Arch}}.{{.Ext}}"
	// It can contain the following placeholders for detected values:
	// - {{.OS}}: Operating System (e.g., "linux", "windows", "darwin")
	// - {{.Arch}}: Architecture (e.g., "amd64", "arm64")
	// - {{.Version}}: Version (e.g., "v1.0.0")
	ArchiveName string

	// The binary name within the archive.
	// If Windows is detected, it will be suffixed with ".exe".
	// Example: "app"
	BinaryName string

	// CurrentVersion is the current version of the application.
	CurrentVersion string

	// Allow pre-releases, default false
	AllowPreReleases bool
}

// Releases struct for Github releases json
type Releases []struct {
	Name       string `json:"name"`       // release name
	Tag        string `json:"tag_name"`   // release tag
	Notes      string `json:"body"`       // release notes
	Prerelease bool   `json:"prerelease"` // Github pre-release
	Assets     []struct {
		BrowserDownloadURL string `json:"browser_download_url"`
		ID                 int64  `json:"id"`
		Name               string `json:"name"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

// Release struct contains the file data for downloadable release
type Release struct {
	Name         string
	Tag          string
	ReleaseNotes string
	Prerelease   bool
	URL          string
	Size         int64
	FileType     string
}
