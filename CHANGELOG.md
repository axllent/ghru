# Changelog

## [v2.2.1]

- Fix comparison logic to include releases equal to the current version
- Add tests for latest version comparison scenarios
- Ensure proper closure of temporary files in replaceFile function
- Ensure proper closure of files in extractArchive function
- Remove dead "tar.gz" switch in extractArchive function
- Refactor getTempDir function to eliminate global tempDir variable

## [v2.2.0]

- Add sha256 checksum validation for downloaded files
- Update Go version to 1.25.0 and update golang.org/x/mod dependency
- Improve error handling and validation in various functions
- Add comprehensive tests for configuration validation, tar extraction, zip extraction, and utility functions

## [v2.1.0]

- Update Github actions workflow to use latest versions of actions and add a test for gofmt
- Update dependabot configuration to check for updates semiannually instead of quarterly

## [v2.0.2]

- Update ArchiveName example in Config struct documentation

## [v2.0.1]

- Delete temporary directory after update

## [v2.0.0]

- Completely rewrite module (v2 is incompatible with v1)
- Multiple archive formats: `tar.gz`, `tgz`, `tar.bz2` & `zip`
- Archive name templating
- Use `golang.org/x/mod/semver` for semantic versioning
- Improve error handling

## [1.1.3]

- Bump version
- Bugfix with semver.Compare()

## [v1.1.2]

- Add godoc link
- Add GreaterThan function
- Add go mods
- Switch to axllent/semver

## [v1.0.0]

- Initial release
- Use masterminds/semver
