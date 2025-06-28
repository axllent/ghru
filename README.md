# GHRU - Github Release Updater for Go

[![GoDoc](https://godoc.org/github.com/axllent/ghru?status.svg)](https://godoc.org/github.com/axllent/ghru)
[![Go Report Card](https://goreportcard.com/badge/github.com/axllent/ghru)](https://goreportcard.com/report/github.com/axllent/ghru)

### Note: this version of GHRU is deprecated. Please use github.com/axllent.ghru/v2 instead.

GHRU is a golang package that allows self-updating in your application by downloading the latest **semantic** (semver) directly from your github releases (**release assets**), replacing the running version.

By default it will skip pre-releases, either defined by "This is a pre-release" option on Github, or by the semverion git tag (eg: `1.2.3-beta1`), however this can be disabled by defining `ghru.AllowPrereleases = true` in your software.

The binaries must be attached to your Github releases (assets), compressed with bzip2 (`bz2`),
and named accordingly: `<name>_<semver>_<os>_<arch>.bz2`, eg:

```
myapp_1.2.3_linux_amd64.bz2
myapp_1.2.3_linux_386.bz2
myapp_1.2.3_darwin_386.bz2
myapp_1.2.3_darwin_amd64.bz2
myapp_1.2.3_windows_amd64.exe.bz2
myapp_1.2.3_windows_386.exe.bz2
```

## Install

`go get -u github.com/axllent/ghru`

## Example usage

The update command is `ghru.Update("myuser/myapp", "myapp", appVersion)`, where:

- `myuser/myapp` (string) is the github name (handle) and the repository
- `myapp` (string) is the name of your binary (without semversion, os, architecture or extension)
- `appVersion` (string) is the current version of the running application

How you define your current running version is entirely up to you, but you must provide it otherwise
GHRU will always indicate that there is an update.

```go
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/axllent/ghru"
)

var appVersion = "0.1.2" // current app version

func main() {

	ghru.AllowPrereleases = true // optional, default false

	update := flag.Bool("u", false, "update to latest release")
	showVersion := flag.Bool("v", false, "show current version")

	flag.Parse()

	if *showVersion {
		fmt.Println(fmt.Sprintf("Version: %s", appVersion))
		latest, _, _, err := ghru.Latest("myuser/myapp", "myapp")
		if err == nil && ghru.GreaterThan(latest, appVersion) {
			fmt.Printf("Update available: %s\nRun `%s -u` to update.\n", latest, os.Args[0])
		}
		os.Exit(0)
	}

	if *update {
		rel, err := ghru.Update("myuser/myapp", "myapp", appVersion)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Updated %s to version %s\n", os.Args[0], rel)
		os.Exit(0)
	}

	// ... rest of app
}
```
