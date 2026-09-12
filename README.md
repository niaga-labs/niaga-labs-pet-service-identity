# service-identity

Authentication and user management service for the Kilat Pet Runner platform.

## Description

This service handles user authentication, registration, and profile management. It provides JWT-based authentication with access and refresh token support for secure API access across the platform.

## Features

- User registration with validation
- Secure login with bcrypt password hashing
- JWT access and refresh token generation
- Token refresh mechanism
- User logout functionality
- Profile retrieval and updates
- Role-based access control (Owner/Runner/Admin)

## API Endpoints

| Method | Endpoint                  | Access | Description                    |
|--------|---------------------------|--------|--------------------------------|
| POST   | /api/v1/auth/register     | Public | Register new user              |
| POST   | /api/v1/auth/login        | Public | Authenticate user              |
| POST   | /api/v1/auth/refresh      | Public | Refresh access token           |
| POST   | /api/v1/auth/logout       | Auth   | Logout user                    |
| GET    | /api/v1/auth/profile      | Auth   | Get user profile               |
| PUT    | /api/v1/auth/profile      | Auth   | Update user profile            |

## Configuration

The service requires the following environment variables:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=identity_db
JWT_SECRET=your-secret-key
SERVICE_PORT=8004
```

## Tech Stack

- **Language**: Go 1.24
- **Web Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL
- **Authentication**: JWT (golang-jwt/jwt)
- **Password Hashing**: bcrypt

## Running the Service

```bash
# Install dependencies
go mod download

# Point at the shared dev-infra stack (cd ~/Documents/dev-infra; ./dev.ps1 up kilat)
export DB_HOST=localhost DB_PORT=5432 DB_USER=kilat DB_PASSWORD=kilat_secret
export DB_NAME=kilat_identity DB_SSL_MODE=disable

# Apply the SQL migrations -- run from the repository root, the migration
# source is resolved relative to the working directory
go run ./cmd/migrate

# Start the service
go run ./cmd/server
```

### Migrations

`migrations/` is the single source of truth for the schema, in every environment including
development. `cmd/migrate` applies the files and exits; `cmd/server` applies them at startup
too, so a fresh boot is self-sufficient.

There is deliberately no GORM `AutoMigrate` path. Two things went wrong with it: GORM renamed
the unique constraint on `runner_applications` away from the name the SQL migration gives it,
and `referrals` / `user_referral_codes` had no SQL migration at all, so they existed only on a
developer laptop (KPD-56). One path for all environments removes that class of drift.

The service will start on port 8004.

## Database Schema

- **users**: Core user table with credentials and profile information
- **refresh_tokens**: Stores refresh tokens for session management

## Security

- Passwords are hashed using bcrypt with configurable cost
- JWT tokens have configurable expiration times
- Refresh tokens are stored securely and can be revoked
- All authenticated endpoints require valid JWT in Authorization header
