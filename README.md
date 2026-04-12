# EduAccess API

Multi-tenant School Management SaaS backend built with Go, Echo, GORM, and PostgreSQL (Supabase-ready).

---

## Table of Contents

- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Environment Variables](#environment-variables)
  - [Option A — Local PostgreSQL (Docker Compose)](#option-a--local-postgresql-docker-compose)
  - [Option B — Connect to Supabase](#option-b--connect-to-supabase)
  - [Run Without Docker](#run-without-docker)
- [Database Setup](#database-setup)
- [Roles & Permissions](#roles--permissions)
- [API Reference](#api-reference)
  - [Authentication](#authentication)
  - [Users & Profile](#users--profile)
  - [Schools](#schools)
  - [Students](#students)
  - [Parents](#parents)
  - [Academic Structure](#academic-structure)
- [Authentication Flow](#authentication-flow)
- [Response Format](#response-format)
- [Swagger / Interactive Docs](#swagger--interactive-docs)
- [Docker](#docker)
- [Contributing](#contributing)

---

## Overview

EduAccess is a multi-tenant API that powers school management for multiple schools from a single deployment. Each school is a tenant; data is scoped by `school_id`. A **superadmin** manages the platform across all tenants; each school has its own **admin_sekolah**.

---

## Tech Stack

| Layer       | Technology                              |
|-------------|------------------------------------------|
| Language    | Go 1.25+                                 |
| HTTP        | Echo v4                                  |
| ORM         | GORM (PostgreSQL driver via pgx)         |
| Auth        | JWT (HS256) — access + refresh tokens    |
| Database    | PostgreSQL 15+ / Supabase                |
| API Docs    | Swagger (swaggo/swag)                    |
| Container   | Docker + Docker Compose                  |

---

## Project Structure

```
eduaccess-api/
├── cmd/
│   └── main.go                  # Entrypoint — wires all modules
├── database/
│   └── migrations/
│       └── 001_initial_schema.sql
├── docs/                        # Auto-generated Swagger docs
├── internal/
│   ├── auth/                    # Register, login, refresh, logout
│   ├── school/                  # School CRUD, rules, subscriptions
│   ├── student/                 # Students, parents, academic structure
│   ├── user/                    # User management & profile
│   └── shared/
│       ├── apperror/            # Domain error types
│       ├── middleware/          # JWT auth middleware
│       ├── response/            # Consistent JSON response helpers
│       └── validator/           # Request binding & validation
├── pkg/
│   ├── database/                # GORM connection setup
│   └── jwt/                     # Token generation & parsing
├── .env.example                 # Copy this to .env
├── docker-compose.yml
└── Dockerfile
```

Each domain follows a clean architecture layout:

```
internal/<domain>/
├── application/   # Use-case handlers (business logic)
├── delivery/http/ # HTTP handlers & DTOs
├── domain/        # Entities, interfaces, constants
└── infrastructure/# GORM repositories
```

---

## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose (for local DB)
- `swag` CLI (only needed to regenerate Swagger docs — the Dockerfile handles this automatically)

### Environment Variables

Copy the example file and fill in the values:

```bash
cp .env.example .env
```

| Variable            | Required | Description                                              |
|---------------------|----------|----------------------------------------------------------|
| `APP_ENV`           | No       | `development` (enables SQL logging) or `production`      |
| `APP_PORT`          | No       | HTTP port, default `8080`                                |
| `DATABASE_URL`      | Either   | Full Postgres DSN — use this for Supabase / Railway      |
| `DB_HOST`           | Either   | Individual DB connection vars (alternative to above)     |
| `DB_PORT`           | Either   | Default `5432`                                           |
| `DB_USER`           | Either   | Database user                                            |
| `DB_PASSWORD`       | Either   | Database password                                        |
| `DB_NAME`           | Either   | Database name                                            |
| `DB_SSLMODE`        | No       | `disable` (local) or `require` (Supabase)                |
| `DB_MAX_OPEN_CONNS` | No       | Max open DB connections, default `25`                    |
| `DB_MAX_IDLE_CONNS` | No       | Max idle DB connections, default `5`                     |
| `JWT_SECRET`        | **Yes**  | Secret key for signing JWTs — use a long random string   |

> **Never commit your `.env` file.** It is already in `.gitignore`. Share secrets through a password manager or your team's secrets vault.

---

### Option A — Local PostgreSQL (Docker Compose)

This spins up both the API and a local Postgres instance:

```bash
# 1. Copy and configure environment
cp .env.example .env
# Set JWT_SECRET to a random string, leave DATABASE_URL empty

# 2. Start everything
docker compose up --build

# API is available at http://localhost:8080
# Swagger UI at        http://localhost:8080/swagger/index.html
```

The compose file mounts `database/migrations/` into Postgres so the schema is applied automatically on first start.

---

### Option B — Connect to Supabase

1. Create a project at [supabase.com](https://supabase.com).
2. In the Supabase dashboard go to **Settings → Database → Connection string → URI** and copy the connection string.
3. Set it in your `.env`:

```dotenv
DATABASE_URL=postgresql://postgres:[YOUR-PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres?sslmode=require
JWT_SECRET=your-long-random-secret
```

4. Apply the initial migration. You can paste the contents of `database/migrations/001_initial_schema.sql` into the Supabase SQL editor, or run it via `psql`:

```bash
psql "$DATABASE_URL" -f database/migrations/001_initial_schema.sql
```

5. Start the API:

```bash
go run ./cmd/main.go
```

> Supabase passwords and connection strings are secrets. Store them only in `.env` (which is gitignored) or your CI/CD secrets store. Never paste them in chat or commit history.

---

### Run Without Docker

```bash
# Install dependencies
go mod download

# (Optional) Regenerate Swagger docs after changing annotations
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/main.go --output docs

# Run
go run ./cmd/main.go
```

---

## Database Setup

The full schema lives in [database/migrations/001_initial_schema.sql](database/migrations/001_initial_schema.sql).

Key tables:

| Table                   | Purpose                                            |
|-------------------------|----------------------------------------------------|
| `users`                 | All user accounts (all roles)                      |
| `roles`                 | Role definitions                                   |
| `model_has_roles`       | User ↔ role assignment                             |
| `refresh_tokens`        | JWT refresh token store                            |
| `schools`               | School tenants                                     |
| `school_users`          | User ↔ school membership (provides `school_id`)    |
| `school_rules`          | Key-value config per school                        |
| `subscriptions`         | School subscription & plan                         |
| `student_profiles`      | Student-specific data                              |
| `student_parent_links`  | Many-to-many student ↔ parent                     |
| `academic_levels`       | Grade levels (e.g. SD, SMP)                        |
| `classrooms`            | Classes within a level                             |
| `sub_classrooms`        | Sub-classes / sections                             |

When using **Docker Compose**, the schema is applied automatically on first start. When using **Supabase**, apply it once via the SQL editor or `psql`.

---

## Roles & Permissions

| Role              | Constant (`domain` package) | Access                              |
|-------------------|-----------------------------|-------------------------------------|
| `superadmin`      | `RoleSuperadmin`            | Full platform access; no school_id  |
| `admin_sekolah`   | `RoleAdminSekolah`          | Full access within their school     |
| `kepala_sekolah`  | `RoleKepalaSekolah`         | Read/manage within their school     |
| `guru`            | `RoleGuru`                  | Teacher access                      |
| `staff`           | `RoleStaff`                 | Staff access                        |
| `orangtua`        | `RoleOrangTua`              | Parent (linked to students)         |
| `siswa`           | `RoleSiswa`                 | Student                             |

Role-based rules are enforced at the application layer (use-case handlers), not just at the route level. The JWT payload carries the role, so each handler can check it without a DB round-trip.

---

## API Reference

Base path: `/api/v1`

All protected routes require the header:

```
Authorization: Bearer <access_token>
```

---

### Authentication

| Method | Path              | Auth | Description                            |
|--------|-------------------|------|----------------------------------------|
| POST   | `/auth/register`  | No   | Register a new user                    |
| POST   | `/auth/login`     | No   | Login, returns access + refresh tokens |
| POST   | `/auth/refresh`   | No   | Rotate refresh token, get new pair     |
| POST   | `/auth/logout`    | No   | Revoke refresh token                   |

**Register**
```json
POST /api/v1/auth/register
{
  "name":     "Budi Santoso",
  "username": "budi",
  "email":    "budi@sekolah.id",
  "password": "secret123",
  "role":     "admin_sekolah"
}
```

> `superadmin` accounts cannot be created via this endpoint.

**Login**
```json
POST /api/v1/auth/login
{
  "email":    "budi@sekolah.id",
  "password": "secret123"
}
```
Returns:
```json
{
  "success": true,
  "message": "login successful",
  "data": {
    "access_token":  "<jwt>",
    "refresh_token": "<jwt>"
  }
}
```

Access tokens expire in **15 minutes**. Refresh tokens expire in **7 days**.

**Refresh**
```json
POST /api/v1/auth/refresh
{
  "refresh_token": "<your-refresh-token>"
}
```

**Logout**
```json
POST /api/v1/auth/logout
{
  "refresh_token": "<your-refresh-token>"
}
```

---

### Users & Profile

| Method | Path                  | Auth | Description               |
|--------|-----------------------|------|---------------------------|
| GET    | `/users`              | Yes  | List users (paginated)    |
| GET    | `/users/:id`          | Yes  | Get user by ID            |
| PUT    | `/users/:id`          | Yes  | Update user name/avatar   |
| DELETE | `/users/:id`          | Yes  | Soft-deactivate user      |
| PUT    | `/users/:id/password` | Yes  | Change password           |
| GET    | `/profile`            | Yes  | Get own profile           |
| PUT    | `/profile`            | Yes  | Update own profile        |

Query params for `GET /users`: `role`, `search`, `page`, `per_page`

---

### Schools

| Method | Path                         | Auth | Description                     |
|--------|------------------------------|------|---------------------------------|
| POST   | `/schools`                   | Yes  | Create school (superadmin only) |
| GET    | `/schools`                   | Yes  | List schools (paginated)        |
| GET    | `/schools/:id`               | Yes  | Get school by ID                |
| PUT    | `/schools/:id`               | Yes  | Update school                   |
| DELETE | `/schools/:id`               | Yes  | Soft-deactivate school          |
| GET    | `/schools/:id/rules`         | Yes  | Get school key-value rules      |
| PUT    | `/schools/:id/rules`         | Yes  | Create/update school rules      |
| GET    | `/schools/:id/subscription`  | Yes  | Get school subscription & plan  |

Query params for `GET /schools`: `search`, `status` (`active`|`nonactive`), `page`, `per_page`

---

### Students

| Method | Path                               | Auth | Description               |
|--------|------------------------------------|------|---------------------------|
| POST   | `/students`                        | Yes  | Create student             |
| GET    | `/students`                        | Yes  | List students (paginated)  |
| GET    | `/students/:id`                    | Yes  | Get student by ID          |
| PUT    | `/students/:id`                    | Yes  | Update student             |
| DELETE | `/students/:id`                    | Yes  | Soft-deactivate student    |
| POST   | `/students/:id/parents`            | Yes  | Link parent to student     |
| DELETE | `/students/:id/parents/:parent_id` | Yes  | Unlink parent from student |

---

### Parents

| Method | Path           | Auth | Description              |
|--------|----------------|------|--------------------------|
| POST   | `/parents`     | Yes  | Create parent account    |
| GET    | `/parents`     | Yes  | List parents (paginated) |
| GET    | `/parents/:id` | Yes  | Get parent by ID         |
| PUT    | `/parents/:id` | Yes  | Update parent            |
| DELETE | `/parents/:id` | Yes  | Soft-deactivate parent   |

---

### Academic Structure

| Method | Path                       | Auth | Description      |
|--------|----------------------------|------|------------------|
| POST   | `/academic/levels`         | Yes  | Create level     |
| GET    | `/academic/levels`         | Yes  | List levels      |
| PUT    | `/academic/levels/:id`     | Yes  | Update level     |
| DELETE | `/academic/levels/:id`     | Yes  | Delete level     |
| POST   | `/academic/classes`        | Yes  | Create class     |
| GET    | `/academic/classes`        | Yes  | List classes     |
| PUT    | `/academic/classes/:id`    | Yes  | Update class     |
| DELETE | `/academic/classes/:id`    | Yes  | Delete class     |
| POST   | `/academic/subclasses`     | Yes  | Create sub-class |
| GET    | `/academic/subclasses`     | Yes  | List sub-classes |
| PUT    | `/academic/subclasses/:id` | Yes  | Update sub-class |
| DELETE | `/academic/subclasses/:id` | Yes  | Delete sub-class |

---

## Authentication Flow

```
Client                            API
  |                                |
  |-- POST /auth/login ---------->|
  |<-- access_token (15 min) -----|
  |<-- refresh_token (7 days) ----|
  |                                |
  |-- GET /api/v1/... ----------->|  Authorization: Bearer <access_token>
  |<-- 200 OK --------------------|
  |                                |
  |   (access_token expires)       |
  |-- POST /auth/refresh -------->|  body: { "refresh_token": "..." }
  |<-- new access_token -----------|
  |<-- new refresh_token ----------|  old refresh_token is revoked
  |                                |
  |-- POST /auth/logout ---------->|  body: { "refresh_token": "..." }
  |<-- 200 OK --------------------|
```

Tokens are signed with **HS256** using `JWT_SECRET`. The payload includes `user_id`, `school_id` (nil for superadmin), `role`, and `token_type`.

---

## Response Format

All endpoints return a consistent JSON envelope.

**Success (single object)**
```json
{
  "success": true,
  "message": "user retrieved",
  "data": { ... }
}
```

**Success (paginated list)**
```json
{
  "success":  true,
  "message":  "students retrieved",
  "data":     [ ... ],
  "page":     1,
  "per_page": 20,
  "total":    150
}
```

**Error**
```json
{
  "success": false,
  "message": "user not found"
}
```

Common HTTP status codes:

| Code | Meaning               |
|------|-----------------------|
| 200  | OK                    |
| 201  | Created               |
| 400  | Bad Request           |
| 401  | Unauthorized          |
| 403  | Forbidden             |
| 404  | Not Found             |
| 409  | Conflict              |
| 422  | Unprocessable Entity  |
| 500  | Internal Server Error |

---

## Swagger / Interactive Docs

After starting the server, open:

```
http://localhost:8080/swagger/index.html
```

To use protected endpoints:
1. Call `POST /api/v1/auth/login` to get an access token.
2. Click **Authorize** (top right) and enter `Bearer <your_access_token>`.

To regenerate docs after changing Swagger annotations:

```bash
swag init -g cmd/main.go --output docs
```

---

## Docker

**Build and run with Docker Compose (recommended for local dev):**

```bash
docker compose up --build
```

**Build the image only:**

```bash
docker build -t eduaccess-api .
```

**Run the container standalone (requires an external DB):**

```bash
docker run -p 8080:8080 \
  -e DATABASE_URL="postgresql://..." \
  -e JWT_SECRET="your-secret" \
  eduaccess-api
```

The Dockerfile is multi-stage (Go builder → Alpine runtime) and generates Swagger docs during the build. Timezone is set to `Asia/Jakarta`.

---

## Health Check

```
GET /health
→ 200 { "status": "ok" }
```

---

## Contributing

1. Create a feature branch from `main`.
2. Follow the existing clean architecture layout — add new features inside `internal/<domain>/`.
3. Keep business logic in `application/`, not in HTTP handlers.
4. Run `go vet ./...` and confirm the server starts before opening a PR.
5. Never commit `.env` or any file containing real credentials.
