# GHRU - GitHub Release Updater for Go applications

[![GoDoc](https://pkg.go.dev/badge/github.com/axllent/ghru/v2)](https://pkg.go.dev/github.com/axllent/ghru/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/axllent/ghru/v2)](https://goreportcard.com/report/github.com/axllent/ghru/v2)

GHRU is a Go package that integrates self-updating functionality into your applications. It provides the ability to fetch the latest newer release information (such as tag, name, release notes), and for self-updating with the latest **semantic version** (semver) from your public GitHub releases.

By default, GHRU skips pre-releases but can be configured to include them. It supports flexible naming conventions and works seamlessly with compressed release assets.

## GitHub release file requirements

To ensure compatibility with GHRU, your release files must meet the following criteria:

- **Compression Formats**: Supported formats auto-detected and include `.tar.gz`, `.tgz`, `.tar.bz2`, and `.zip`.
- **Binary Placement**: The binary must be at the top level in the release archive file (not in subdirectories).
- **Additional Files**: Additional files such as `CHANGELOG` or `README` can be included in the archive but will be ignored during self-updates.
- **File Naming**: Each file must specify the lowercased operating system and architecture in its name, and can optionally include the version number (the format is defined in your config, see below). Examples:
  - `app1-linux-amd64.tar.gz`
  - `app2-darwin-arm64.tar.bz2`
  - `app3-v1.2.3-windows-amd64.zip`
  - `windows-amd64.zip`

## Defining the archive name template

Go templating is used to parse your `ArchiveName`, which uses the following values which are inherited from the running application:

- `{{.OS}}` - Operating system (lowercased), eg: `linux`, `darwin`, `windows` etc. ([see list](https://github.com/golang/go/blob/master/src/internal/syslist/syslist.go#L17))
- `{{.Arch}}` - Architecture (lowercased), eg: `386`, `amd64`, `arm64` etc. ([see list](https://github.com/golang/go/blob/master/src/internal/syslist/syslist.go#L58))
- `{{.Version}}` - the tag version of the release

In your `ghru.Config{}` you must define the `ArchiveName`, for example `"package-{{.OS}}-{{.Arch}}"`.
GHRU will detect the supported file format based on the filename and append this to the `ArchiveName`, resulting in `package-linux-amd64.tar.gz` when executed from Linux on amd64 architecture.

## Adding the module

`go get -u github.com/axllent/ghru/v2`

## Example usage

```go
// Package main is an example application integrated with GHRU.
// Modify the ghru.Config{} variables to suite your repo.
// AppVersion variable value would typically be compiled into the application.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/axllent/ghru/v2"
)

// App version should be set at compile time, for instance as a build argument.
// It has been hardcoded here for demonstration purposes.
var AppVersion = "0.1.2"

func main() {

	c := ghru.Config{
		// GitHub <org>/<repo> - do not use the full URL
		Repo:             "myorg/myapp",
		// Archive filename template
		ArchiveName:      "myapp-{{.OS}}-{{.Arch}}",
		// The name of the binary within the archive, ".exe" is automatically appended for Windows binaries
		BinaryName:       "myapp",
		// The current version from your running application
		CurrentVersion:   AppVersion,
		// Allow pre-releases (default false)
		AllowPreReleases: false,
	}

	// Command line flags
	update := flag.Bool("u", false, "update to latest release")
	showVersion := flag.Bool("v", false, "show current version")

	flag.Parse()

	// Show version and display update information if available
	if *showVersion {
		fmt.Printf("Current version: %s\n", AppVersion)
		release, err := c.Latest()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// The latest version is the same version
		if release.Tag == AppVersion {
			fmt.Println("You are running the latest version.")
			os.Exit(0)
		}

		// A newer release is available
		fmt.Printf("Update available: %s\nRun `%s -u` to update.\n", release.Tag, os.Args[0])

		// Display release notes if available
		if release.ReleaseNotes != "" {
			fmt.Printf("\nRelease notes\n=============\n\n%s\n", release.ReleaseNotes)
		}

		os.Exit(0)
	}

	// Update the application to the latest release
	if *update {
		rel, err := c.SelfUpdate()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Updated %s to version %s\n", os.Args[0], rel.Tag)
		os.Exit(0)
	}

	fmt.Println("This is your test application")
}
```
