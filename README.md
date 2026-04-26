# EduAccess API

---

BIG DISCLAIMER
THE FIRST INITIAL COMMIT WE USE A TEMPLATE USING GO SAAS TEMPLATE AND MERGE IT WITH OUR INTERN PROJECT INSPIRATION THE FIRST INITIAL COMMIT IS NOT FULLY COMPLETED ITS BASE REFRENCE FOR OUR TEAM TO WORK ON
Multi-tenant School Management SaaS backend built with Go, Echo, GORM, and PostgreSQL (Supabase-ready).

---

## Table of Contents

-[Midtermneeds](#midtermneeds)

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
  - [Admins](#admins)
  - [Students](#students)
  - [Parents](#parents)
  - [Academic Structure](#academic-structure)
- [Authentication Flow](#authentication-flow)
- [Response Format](#response-format)
- [Swagger / Interactive Docs](#swagger--interactive-docs)
- [Docker](#docker)
- [Contributing](#contributing)

---

## Midtermneeds

**Auth is enabled on all routes. You MUST send a Bearer token on every request (except login and registration). Here's the only thing you need to do:**

## Step 1 — Login

````json
POST /api/v1/auth/login
{
  "email": "adminsekolah2@gmail.com",
  "password": "password"
}
```(this is not a superadmin account for a better case this is admin_sekolah account)

Copy the `access_token` from the response.

## Step 2 — Use the Token on Every Request

In every subsequent request, set the Authorization header:

````

Authorization: Bearer <paste_your_access_token_here>

```

In Swagger UI: click the **Authorize** button (lock icon) on the most upper right, type `Bearer <token>`, click Authorize.

## That's It. You Do NOT Need to Pass school_id Anywhere(except if your a superadmin account).

The `school_id` is **automatically embedded in your token** when you log in. The server reads it from the token you never include it manually in request bodies or headers. (except for superadmin)

> **Two account types you:**
>
> - **superadmin** — for platform-level routes: creating schools, listing all users, etc due to this roles it not attach to any schools, this roles need to inclue school_id on spesific request.
> - **admin_sekolah** (linked to a school) — for school-scoped routes: headmasters, students, etc.

---
## Overview

EduAccess is a multi-tenant API that powers school management for multiple schools from a single deployment. Each school is a tenant; data is scoped by `school_id`. A **superadmin** manages the platform across all tenants; each school has its own **admin_sekolah**.

---

## Tech Stack

| Layer     | Technology                            |
| --------- | ------------------------------------- |
| Language  | Go 1.25+                              |
| HTTP      | Echo v4                               |
| ORM       | GORM (PostgreSQL driver via pgx)      |
| Auth      | JWT (HS256) — access + refresh tokens |
| Database  | PostgreSQL 15+ / Supabase             |
| API Docs  | Swagger (swaggo/swag)                 |
| Container | Docker + Docker Compose               |

---

## Project Structure

```

eduaccess-api/
├── cmd/
│ └── main.go # Entrypoint — wires all modules
├── database/
│ └── migrations/
│ └── 001_initial_schema.sql # Full schema + seed data (roles, plans)
├── docs/ # Auto-generated Swagger docs
│ ├── docs.go
│ ├── swagger.json
│ └── swagger.yaml
├── internal/
│ ├── admin/ # Admin sekolah CRUD
│ │ ├── application/
│ │ │ ├── create_admin.go
│ │ │ ├── deactivate_admin.go
│ │ │ ├── get_admin.go
│ │ │ ├── list_admins.go
│ │ │ ├── update_admin.go
│ │ │ ├── user_creator.go
│ │ │ └── user_updater.go
│ │ ├── delivery/http/
│ │ │ ├── dto.go
│ │ │ └── handler.go
│ │ ├── domain/
│ │ │ ├── admin.go
│ │ │ └── repository.go
│ │ └── infrastructure/
│ │ └── admin_repository.go
│ ├── auth/ # Register, login, refresh, logout
│ │ ├── application/
│ │ │ ├── login.go
│ │ │ ├── logout.go
│ │ │ ├── refresh.go
│ │ │ └── register.go
│ │ ├── delivery/http/
│ │ │ ├── dto.go
│ │ │ └── handler.go
│ │ ├── domain/
│ │ │ ├── repository.go
│ │ │ └── user.go
│ │ └── infrastructure/
│ │ ├── refresh_token_repository.go
│ │ ├── user_model.go
│ │ └── user_repository.go
│ ├── headmaster/ # Kepala sekolah CRUD
│ │ ├── application/
│ │ │ ├── create_headmaster.go
│ │ │ ├── deactivate_headmaster.go
│ │ │ ├── get_headmaster.go
│ │ │ ├── list_headmasters.go
│ │ │ ├── school_updater.go
│ │ │ ├── update_headmaster.go
│ │ │ └── user_creator.go
│ │ ├── delivery/http/
│ │ │ ├── dto.go
│ │ │ └── handler.go
│ │ ├── domain/
│ │ │ ├── headmaster.go
│ │ │ └── repository.go
│ │ └── infrastructure/
│ │ └── headmaster_repository.go
│ ├── parent/ # Parent CRUD
│ │ ├── application/
│ │ │ ├── create_parent.go
│ │ │ ├── deactivate_parent.go
│ │ │ ├── get_parent.go
│ │ │ ├── list_parents.go
│ │ │ ├── update_parent.go
│ │ │ └── user_creator.go
│ │ ├── delivery/http/
│ │ │ ├── dto.go
│ │ │ └── handler.go
│ │ ├── domain/
│ │ │ ├── parent.go
│ │ │ └── repository.go
│ │ └── infrastructure/
│ │ └── parent_repository.go
│ ├── school/ # School CRUD, rules, subscriptions
│ │ ├── application/
│ │ │ ├── create_school.go
│ │ │ ├── deactivate_school.go
│ │ │ ├── get_school.go
│ │ │ ├── get_subscription.go
│ │ │ ├── list_schools.go
│ │ │ ├── manage_rules.go
│ │ │ └── update_school.go
│ │ ├── delivery/http/
│ │ │ ├── dto.go
│ │ │ └── handler.go
│ │ ├── domain/
│ │ │ ├── repository.go
│ │ │ └── school.go
│ │ └── infrastructure/
│ │ └── school_repository.go
│ ├── shared/ # Cross-cutting utilities
│ │ ├── apperror/
│ │ │ └── apperror.go # Domain error types
│ │ ├── middleware/
│ │ │ └── auth.go # JWT auth middleware
│ │ ├── response/
│ │ │ └── response.go # Consistent JSON response helpers
│ │ └── validator/
│ │ └── validator.go # Request binding & validation
│ ├── student/ # Students, parents (linked), academic structure
│ │ ├── application/
│ │ │ ├── academic_handlers.go # Level / class / sub-class CRUD
│ │ │ ├── create_parent.go
│ │ │ ├── create_student.go
│ │ │ ├── deactivate_student.go
│ │ │ ├── get_student.go
│ │ │ ├── list_students.go
│ │ │ ├── manage_parent_link.go # Link / unlink parent ↔ student
│ │ │ ├── parent_handlers.go
│ │ │ ├── update_student.go
│ │ │ └── user_creator.go
│ │ ├── delivery/http/
│ │ │ ├── dto.go
│ │ │ ├── handler.go
│ │ │ └── student_handler.go
│ │ ├── domain/
│ │ │ ├── academic.go
│ │ │ ├── parent.go
│ │ │ ├── repository.go
│ │ │ ├── student_profile.go
│ │ │ └── student_repository.go
│ │ └── infrastructure/
│ │ ├── academic_repository.go
│ │ ├── parent_repository.go
│ │ └── student_profile_repository.go
│ └── user/ # Platform user management & profile
│ ├── application/
│ │ ├── change_password.go
│ │ ├── deactivate_user.go
│ │ ├── get_user.go
│ │ ├── list_users.go
│ │ ├── repository.go
│ │ └── update_user.go
│ ├── delivery/http/
│ │ ├── dto.go
│ │ └── handler.go
│ └── infrastructure/
│ └── user_repository.go
├── pkg/
│ ├── database/
│ │ └── database.go # GORM connection setup
│ └── jwt/
│ └── jwt.go # Token generation & parsing
├── .env.example
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── go.sum

```

Each domain module follows the same clean architecture layout:

```

internal/<domain>/
├── application/ # Use-case handlers — business logic, no HTTP concerns
├── delivery/http/ # Echo handlers + request/response DTOs
├── domain/ # Entities, repository interfaces, domain constants
└── infrastructure/ # GORM repository implementations

````

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
````

| Variable             | Required | Description                                            |
| -------------------- | -------- | ------------------------------------------------------ |
| `APP_ENV`            | No       | `development` (enables SQL logging) or `production`    |
| `APP_PORT`           | No       | HTTP port, default `8080`                              |
| `CORS_ALLOW_ORIGINS` | No       | Comma-separated allowlist of origins (default `*`)     |
| `DATABASE_URL`       | Either   | Full Postgres DSN — use this for Supabase / Railway    |
| `DB_HOST`            | Either   | Individual DB connection vars (alternative to above)   |
| `DB_PORT`            | Either   | Default `5432`                                         |
| `DB_USER`            | Either   | Database user                                          |
| `DB_PASSWORD`        | Either   | Database password                                      |
| `DB_NAME`            | Either   | Database name                                          |
| `DB_SSLMODE`         | No       | `disable` (local) or `require` (Supabase)              |
| `DB_MAX_OPEN_CONNS`  | No       | Max open DB connections, default `25`                  |
| `DB_MAX_IDLE_CONNS`  | No       | Max idle DB connections, default `5`                   |
| `JWT_SECRET`         | **Yes**  | Secret key for signing JWTs — use a long random string |

---

<!-- ### Option A — Local PostgreSQL (Docker Compose)

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

The compose file mounts `database/migrations/` into Postgres so the schema is applied automatically on first start. --> Work in progress for midterm assedment use option b, we will provide the lecture the env on the lms.

---

### Run Without Docker

Steps for anyone cloning this repo for the first time:

```bash
# 1. Clone the repo
git clone https://github.com/your-org/eduaccess-api.git
cd eduaccess-api

# Install / tidy Go dependencies
go mod tidy

# Copy the environment file and fill in your values
cp .env.example .env
# Open .env and set DATABASE_URL and JWT_SECRET at minimum

swag init -g cmd/main.go --output docs

# 6. Run the server
go run ./cmd/main.go
```

The server starts at `http://localhost:8080` and Swagger UI is at `http://localhost:8080/swagger/index.html`.

### Run Together With Flutter Frontend

Backend base API path is `/api/v1`, and Flutter should point to this base URL.

1. Run backend:

```bash
cp .env.example .env
# set JWT_SECRET before running
go run ./cmd/main.go
```

2. Run Flutter with the correct API base URL:

```bash
# Web/Desktop
flutter run --dart-define=EDUACCESS_BASE_URL=http://localhost:8080/api/v1

# Android emulator
flutter run --dart-define=EDUACCESS_BASE_URL=http://10.0.2.2:8080/api/v1
```

For Flutter Web in development, you can restrict CORS safely instead of `*`:

```dotenv
CORS_ALLOW_ORIGINS=http://localhost:3000,http://localhost:5000
```

---

## Database Setup

The full schema lives in [database/migrations/001_initial_schema.sql](database/migrations/001_initial_schema.sql).
this is the ERD we plan to impelment (their might be changes in the future)
https://drive.google.com/file/d/1Rt9KfwXE1S2RZ9Zu3CM6iX4B6pqTg5y0/view?usp=sharing

Key tables:

| Table                  | Purpose                                         |
| ---------------------- | ----------------------------------------------- |
| `users`                | All user accounts (all roles)                   |
| `roles`                | Role definitions                                |
| `model_has_roles`      | User ↔ role assignment                          |
| `refresh_tokens`       | JWT refresh token store                         |
| `schools`              | School tenants                                  |
| `school_users`         | User ↔ school membership (provides `school_id`) |
| `school_rules`         | Key-value config per school                     |
| `subscriptions`        | School subscription & plan                      |
| `student_profiles`     | Student-specific data                           |
| `student_parent_links` | Many-to-many student ↔ parent                   |
| `academic_levels`      | Grade levels (e.g. SD, SMP)                     |
| `classrooms`           | Classes within a level                          |
| `sub_classrooms`       | Sub-classes / sections                          |
| `headmaster_profiles`  | head master data                                |

When using **Docker Compose**, the schema is applied automatically on first start. When using **Supabase**, apply it once via the SQL editor or `psql`.

---

## Roles & Permissions

| Role             | Constant (`domain` package) | Access                             |
| ---------------- | --------------------------- | ---------------------------------- |
| `superadmin`     | `RoleSuperadmin`            | Full platform access; no school_id |
| `admin_sekolah`  | `RoleAdminSekolah`          | Full access within their school    |
| `kepala_sekolah` | `RoleKepalaSekolah`         | Read/manage within their school    |
| `guru`           | `RoleGuru`                  | Teacher access                     |
| `staff`          | `RoleStaff`                 | Staff access                       |
| `orangtua`       | `RoleOrangTua`              | Parent (linked to students)        |
| `siswa`          | `RoleSiswa`                 | Student                            |

---

## API Reference

Base path: `/api/v1`

All protected routes require the header:

```
Authorization: Bearer <access_token>
```

---

### Authentication

| Method | Path             | Auth | Description                            |
| ------ | ---------------- | ---- | -------------------------------------- |
| POST   | `/auth/register` | No   | Register a new user                    |
| POST   | `/auth/login`    | No   | Login, returns access + refresh tokens |
| POST   | `/auth/refresh`  | No   | Rotate refresh token, get new pair     |
| POST   | `/auth/logout`   | No   | Revoke refresh token                   |

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
    "access_token": "<jwt>",
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

| Method | Path                  | Auth | Description             |
| ------ | --------------------- | ---- | ----------------------- |
| GET    | `/users`              | Yes  | List users (paginated)  |
| GET    | `/users/:id`          | Yes  | Get user by ID          |
| PUT    | `/users/:id`          | Yes  | Update user name/avatar |
| DELETE | `/users/:id`          | Yes  | Soft-deactivate user    |
| PUT    | `/users/:id/password` | Yes  | Change password         |
| GET    | `/profile`            | Yes  | Get own profile         |
| PUT    | `/profile`            | Yes  | Update own profile      |

Query params for `GET /users`: `role`, `search`, `page`, `per_page`

---

### Schools

| Method | Path                        | Auth | Description                     |
| ------ | --------------------------- | ---- | ------------------------------- |
| POST   | `/schools`                  | Yes  | Create school (superadmin only) |
| GET    | `/schools`                  | Yes  | List schools (paginated)        |
| GET    | `/schools/:id`              | Yes  | Get school by ID                |
| PUT    | `/schools/:id`              | Yes  | Update school                   |
| DELETE | `/schools/:id`              | Yes  | Soft-deactivate school          |
| GET    | `/schools/:id/rules`        | Yes  | Get school key-value rules      |
| PUT    | `/schools/:id/rules`        | Yes  | Create/update school rules      |
| GET    | `/schools/:id/subscription` | Yes  | Get school subscription & plan  |

Query params for `GET /schools`: `search`, `status` (`active`|`nonactive`), `page`, `per_page`

---

### Admins

| Method | Path          | Auth | Description             |
| ------ | ------------- | ---- | ----------------------- |
| POST   | `/admins`     | Yes  | Create admin profile    |
| GET    | `/admins`     | Yes  | List admins (paginated) |
| GET    | `/admins/:id` | Yes  | Get admin by ID         |
| PUT    | `/admins/:id` | Yes  | Update admin            |
| DELETE | `/admins/:id` | Yes  | Soft-deactivate admin   |

Query params for `GET /admins`: `search`, `page`, `per_page`

---

### Students

| Method | Path                               | Auth | Description                |
| ------ | ---------------------------------- | ---- | -------------------------- |
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
| ------ | -------------- | ---- | ------------------------ |
| POST   | `/parents`     | Yes  | Create parent account    |
| GET    | `/parents`     | Yes  | List parents (paginated) |
| GET    | `/parents/:id` | Yes  | Get parent by ID         |
| PUT    | `/parents/:id` | Yes  | Update parent            |
| DELETE | `/parents/:id` | Yes  | Soft-deactivate parent   |

---

### Academic Structure

| Method | Path                       | Auth | Description      |
| ------ | -------------------------- | ---- | ---------------- |
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
| ---- | --------------------- |
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

<!-- ## Docker

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

The Dockerfile is multi-stage (Go builder → Alpine runtime) and generates Swagger docs during the build. Timezone is set to `Asia/Jakarta`. -->still work in progress

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
