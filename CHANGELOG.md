# Changelog

All notable changes to service-identity are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `cmd/migrate`: standalone command that applies the golang-migrate files in
  `migrations/` and exits. The README documented `go run cmd/migrate/main.go`
  but no such command existed, so there was no way to apply the SQL schema
  without starting the server outside `APP_ENV=development`. (KPD-2)

### Changed

- README: the "Running the Service" section now gives the exact environment for
  the shared dev-infra stack and explains which schema each migration mode owns. (KPD-2)
