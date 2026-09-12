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

- `migrations/005_create_referrals.{up,down}.sql`: SQL schema for `referrals` and
  `user_referral_codes`. Both tables had GORM models and a registered handler but
  no SQL migration, so they existed only under `APP_ENV=development` and every
  referral endpoint failed elsewhere. (KPD-56)

### Changed

- Go module path is now `github.com/niaga-labs/niaga-labs-pet-service-identity` (was
  `github.com/Kilat-Pet-Delivery/service-identity`), and every import of lib-common and lib-proto follows. The repos moved to the niaga-labs org on
  2026-09-12 (HQ-40). The `replace => ../lib-*` targets are unchanged: local folders keep their
  short names. (HQ-41)
- README: the "Running the Service" section now gives the exact environment for
  the shared dev-infra stack and explains which schema each migration mode owns. (KPD-2)
- `cmd/server`: the development-only GORM `AutoMigrate` branch is gone entirely.
  Every model in this service now has a SQL migration, so the migrations own the
  schema in all environments and development still gets it automatically at
  startup -- one path, nothing left to drift. (KPD-56)
