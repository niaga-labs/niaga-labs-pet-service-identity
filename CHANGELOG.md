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

- README: the "Running the Service" section now gives the exact environment for
  the shared dev-infra stack and explains which schema each migration mode owns. (KPD-2)
- `cmd/server`: dropped `ReferralModel` and `UserReferralCodeModel` from the
  development `AutoMigrate` list -- the SQL migrations own those tables now, so
  development and every other environment share one schema. (KPD-56)
