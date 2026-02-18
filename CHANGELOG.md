# Changelog

## [v2.1.0]

- Enhance security: prevent path traversal vulnerabilities in archive extraction
- Improve error handling for HTTP response status codes
- Add unit tests for Latest and ValidConfig functions
- Add tar & zip extraction tests
- Update golang.org/x/mod dependency from 0.25.0 to 0.31.0
- Update CI/CD dependencies

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
