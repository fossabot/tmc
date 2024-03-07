# Change Log

## [Unreleased]

### Added

- Implemented `version` command to show the version of the tm-catalog-cli
- Implemented autocompletion for most flags and arguments for the shell autocompletion script
- Added an optional flag to fetch to restore original external id to the fetched TM

### Changed

- Request results from multiple remotes concurrently instead of sequentially

### Fixed

- handle timestamp collisions on push by retrying after one second, forcing generation of new id, or reporting the error if all else fails
- Display errors when accessing remotes for list/verions instead of silently ignoring them

## [v0.0.0-alpha.4]

### Added

- REST API: added `meta.page.elements` to inventory response, reflecting number of entries in current result page

### Changed

- REST API: renamed inventory endpoint `/versions` to `/.versions`
- REST API: removed `meta.created` from inventory response
- Removed '--exact' flag to `list` and `pull`
- `list` and `pull`: match given name pattern as a prefix by complete path parts
- `list`: changed output format: put NAME column first, renamed PATH column to MPN
- `versions`: changed output format: renamed PATH column to ID

## [v0.0.0-alpha.3]

### Added

- Building docker base image for releases to enable catalog hosting
- Implemented file locking to avoid data races (guaranteed for Linux/MacOSX)
- Implemented ```pull``` command to fetch and save multiple thing models at once
- Implemented setting CORS options for API
- Implemented fetching a TM by a \[partial\] semantic version also in REST API
- Print information about used config file in `help`
- Implemented a 'tmc' remote type, which uses our own REST API as the underlying TM storage
- Added 'filter.name' parameter to REST API '/inventory' listing
- Added '--exact' flag to `list` and `pull`

### Changed

- ```create-toc```: renamed to ```update-toc``` and allow for partial updates
- ```list```: allows now listing by name pattern
- ```serve```: separate configuration of the remote(s) to be served from the target remote for push
- ```fetch```: ```--output``` now accepts only a target folder to save TM to, ```--with-path``` has been removed
- ```list, pull```: removed filter flag `filter.externalID`, search for externalID has now to be done by query search `-s`
- REST API:  removed filter parameter filter.externalID from `/inventory`, `/authors`, `/manufacturers`, `/mpns`,     
  search for externalID has now to be done by query parameter `search`
- enable/disable logging is now done only by setting a loglevel

### Fixed

- count only enabled remotes when checking if empty remote specification is unambiguous
- make fetch by partial semantic version match the most recent version beginning with given string
- (BREAKING!) ```push```: file hash calculation has been made more reliable and idempotent. Consequently, some files if pushed to TMC, may receive a new version hash, despite no change in contents
- ```fetch```: fixed "Unable to parse TMID..." error when fetching an official TM by content hash
- prevent ```serve``` from using one of remotes from config as push target when '-r' or '-d' are given
- print the actual error if updating TOC after ```push``` fails

## [v0.0.0-alpha.2] - 2023-01-15

### Fixed

- config is now created if not existing
- Adding ".exe" to Windows binaries 


## [v0.0.0-alpha.1] - 2023-01-15

This the first alpha release, which implements the most basic verbs to create and interact with a thing model catalog. 

### Added

- Verbs: create-toc, fetch, list, push, remote, serve, validate, versions 
- Target local catalogs with the '--directory' flag
- fetch can now create a file instead of printing to stdout 
- serve now exposes a REST API

See README.md for a description of all current features.
