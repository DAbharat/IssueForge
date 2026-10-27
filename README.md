# IssueForge Backend API

IssueForge is a high-performance, production-ready issue tracking and project management RESTful API built with **Go 1.26.4**. It provides comprehensive capabilities for multi-tenant team workspaces, project governance, issue lifecycle management, collaborative discussions, file attachments, and audit trails.

The backend leverages a clean layered architecture, combining PostgreSQL with **sqlc** for compile-time type-safe persistence, **Redis** for multi-level caching, distributed rate limiting, and session handoffs, **GoBullMQ** for background job queues, **Google OIDC** for federated authentication, and **Prometheus** for RED-principle system observability.

---

## Table of Contents

- [Features](#features)
  - [Core Domain Features](#core-domain-features)
  - [Technical & System Features](#technical--system-features)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Architecture & Data Flow](#architecture--data-flow)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Configuration & Environment Variables](#configuration--environment-variables)
- [API Reference](#api-reference)
  - [Authentication & User Management](#1-authentication--user-management)
  - [Workspaces](#2-workspaces)
  - [Workspace Members](#3-workspace-members)
  - [Workspace Invitations](#4-workspace-invitations)
  - [Projects](#5-projects)
  - [Project Members](#6-project-members)
  - [Issues](#7-issues)
  - [Comments](#8-comments)
  - [Issue Activity & Audit Trail](#9-issue-activity--audit-trail)
  - [Issue & Comment Attachments](#10-issue--comment-attachments)
  - [Labels](#11-labels)
  - [Archive & Soft-Delete Recovery](#12-archive--soft-delete-recovery)
  - [Health & Metrics](#13-health--metrics)
- [Database & Migrations](#database--migrations)
- [Caching & Asynchronous Queues](#caching--asynchronous-queues)
- [Authentication & Security](#authentication--security)
  - [Google OIDC & Redis Handoff Exchange](#google-oidc--redis-handoff-exchange)
  - [Role-Based Access Control (RBAC)](#role-based-access-control-rbac)
- [Rate Limiting](#rate-limiting)
- [Monitoring & Observability](#monitoring--observability)
  - [RED Metrics Architecture](#red-metrics-architecture)
  - [Production PromQL Queries](#production-promql-queries)
  - [Metrics Exposition & Scrape Configuration](#metrics-exposition--scrape-configuration)
- [Performance & Benchmarks](#performance--benchmarks)
- [Docker Deployment](#docker-deployment)
- [Development & Code Quality](#development--code-quality)
- [Contributing](#contributing)
- [License](#license)

---

## Features

### Core Domain Features

- **Multi-Tenant Workspaces**: Create, manage, rename, soft-delete, and restore team workspaces.
- **Role-Based Membership**: Granular workspace roles (`owner`, `admin`, `member`) with role promotion endpoints.
- **Workspace Invitations**: Asynchronous email/username invitation workflow with invitation tokens, lifecycle transitions (`pending`, `accepted`, `declined`, `canceled`), and Redis caching.
- **Project Organization**: Organize projects within workspaces, assign project leads, manage project-specific members, and soft-delete/restore projects.
- **Issue Lifecycle Tracking**: Manage issues with priority levels (`low`, `medium`, `high`, `urgent`), statuses (`backlog`, `todo`, `in_progress`, `done`, `canceled`), assignees, due dates, and soft-delete recovery.
- **Activity History & Audit Trail**: Automated audit logging capturing every attribute transition (title, status, priority, assignee, due date changes).
- **Nested Comments**: Add, update, and delete issue comments with author tracking and attachment links.
- **Custom Labels**: Project-scoped labels with hex color codes, and batch label attachment/detachment on issues.
- **File & Image Attachments**: Secure multi-file uploads via Cloudinary integration with background cleanup workers.
- **Archive Center**: Global soft-delete audit queries with aggregate count rollups and individual restoration endpoints.

### Technical & System Features

- **Dual-Mode Authentication**: Secure email/password login (bcrypt) + Google OAuth 2.0 / OpenID Connect (OIDC).
- **Secure Redis Handoff Exchange**: One-time crypto-random 32-byte handoff codes stored in Redis with atomic single-use consumption (`POST /api/auth/exchange`) to prevent token leakage in URL redirects.
- **Automated Account Linking**: Seamlessly binds Google OIDC identities to existing accounts with matching verified emails or auto-provisions new users with sanitized handles.
- **Type-Safe Persistence with sqlc**: Zero-reflection, high-speed PostgreSQL query layer generated with `sqlc` on `pgx/v5` connection pools.
- **Multi-Level Redis Caching**: Read-through and invalidation caching for projects, workspaces, issues, and pending invitations.
- **Asynchronous Task Queues**: GoBullMQ background worker for out-of-band Cloudinary media deletion with graceful context cancellation.
- **Distributed Rate Limiting**: Multi-tier Redis token-bucket rate limiters applied at router boundaries (Strict, Auth, Read, Write, Issues, Delete, Create, Patch, Attachments).
- **RED Principle Observability**: Full Prometheus integration (`/api/metrics`) measuring Rate, Errors, and Duration, combined with Go runtime and process collectors.
- **Graceful Shutdown**: Intercepts `SIGINT`/`SIGTERM` to safely drain background queue workers, close PostgreSQL pools, and shut down HTTP servers with timeouts.

---

## Tech Stack

| Category | Technology | Purpose |
|---|---|---|
| **Language** | Go 1.26.4 | High-concurrency backend runtime |
| **HTTP Routing** | Gorilla Mux (`mux v1.8.1`) | URL pattern matching and sub-routing |
| **CORS & Middleware** | Gorilla Handlers (`handlers v1.5.2`) | Cross-origin resource sharing & security |
| **Database** | PostgreSQL 17 + `pgx/v5` (v5.10.0) | Relational database & high-performance connection pool |
| **SQL Code Generation** | `sqlc` | Type-safe Go models and database access from raw SQL |
| **Caching & KV Store** | Redis 7 + `go-redis/v9` (v9.21.0) | Multi-level cache, session store, rate limits & OIDC handoff |
| **Background Queues** | GoBullMQ (`github.com/DAbharat/gobullmq`) | Redis-backed distributed queue for asynchronous asset cleanup |
| **Authentication** | JWT (`golang-jwt/jwt/v5`), bcrypt (`x/crypto`) | Stateless access tokens (15m) & Redis refresh tokens (7d) |
| **Federated Identity** | `coreos/go-oidc/v3`, `golang.org/x/oauth2` | Google OpenID Connect ID token verification |
| **Media Storage** | Cloudinary Go SDK (`cloudinary-go/v2`) | Cloud asset storage for images and attachments |
| **Monitoring** | Prometheus Go Client (`client_golang v1.24.1`) | Application and RED HTTP request metrics |
| **Configuration** | `godotenv` | Twelve-factor `.env` configuration management |

---

## Project Structure

```text
IssueForge/
├── cmd/
│   └── api/
│       └── main.go                         # Application composition, dependencies, workers, & server lifecycle
├── internal/
│   ├── auth/                               # Authentication & cryptographic helpers
│   │   ├── google_oauth.go                 # Google OIDC provider & verifier setup
│   │   ├── jwt.go                          # JWT generation, signing, and verification
│   │   ├── token.go                        # Crypto-random token generation utilities
│   │   └── types.go                        # Auth token payload definitions
│   ├── db/                                 # Database connectivity & persistence
│   │   ├── config/                         # Environment configuration parser & validator
│   │   │   └── config.go                   # Config struct and missing env checks
│   │   ├── migrations/                     # 21 versioned SQL schema migrations
│   │   ├── postgres.go                     # pgxpool connection initialization
│   │   ├── queries/                        # SQL source queries consumed by sqlc
│   │   └── sqlc/                           # Auto-generated type-safe PostgreSQL database client
│   ├── dto/                                # Request & response Data Transfer Objects
│   │   ├── archive_dto.go                  # Archive count & restoration DTOs
│   │   ├── comment.go                      # Comment payload DTOs
│   │   ├── issue.go                        # Issue CRUD & filter DTOs
│   │   ├── issue_activity.go               # Activity trail DTOs
│   │   ├── issue_attachments.go            # Attachment upload DTOs
│   │   ├── labels.go                       # Label definition & attachment DTOs
│   │   ├── project.go                      # Project management DTOs
│   │   ├── project_members.go              # Project membership DTOs
│   │   ├── user.go                         # Auth, user profile, & OIDC handoff DTOs
│   │   ├── workspace.go                    # Workspace DTOs
│   │   ├── workspace_invitations.go        # Workspace invitation DTOs
│   │   └── workspace_members.go            # Workspace member DTOs
│   ├── handler/                            # HTTP controllers / request handlers
│   │   ├── archive_handler.go              # Archive listing & stats endpoints
│   │   ├── comment_handler.go              # Comment endpoints
│   │   ├── health_handler.go               # System health & status checks
│   │   ├── issue_activity_handler.go       # Audit trail query endpoints
│   │   ├── issue_attachments_handler.go    # Media upload & soft-delete endpoints
│   │   ├── issues_handler.go               # Issue management endpoints
│   │   ├── labels_handler.go               # Label definition & assignment endpoints
│   │   ├── project_handler.go              # Project lifecycle endpoints
│   │   ├── project_members_handler.go      # Project membership endpoints
│   │   ├── user_handler.go                 # Auth, OIDC, handoff exchange, & user endpoints
│   │   ├── workspace_handler.go            # Workspace lifecycle endpoints
│   │   ├── workspace_invitations_handler.go# Invitation dispatch & acceptance endpoints
│   │   └── workspace_members_handler.go    # Workspace member endpoints
│   ├── httpx/                              # Standardized JSON response formatting
│   │   └── response.go                     # JSON and error serialization helpers
│   ├── middleware/                         # HTTP middleware pipeline
│   │   ├── auth.go                         # JWT Bearer token authentication
│   │   ├── metrics.go                      # Prometheus request latency and error metrics
│   │   └── rate_limiter.go                 # Distributed Redis rate limit enforcement
│   ├── redis/                              # Redis clients, caching, & queues
│   │   ├── cache/                          # Read-through domain caching
│   │   │   ├── issue_cache.go              # Issue caching layer
│   │   │   ├── project_cache.go            # Project caching layer
│   │   │   ├── workspace_cache.go          # Workspace caching layer
│   │   │   └── workspace_invitation_cache.go # Invitation caching layer
│   │   ├── handoff/                        # Google OIDC handoff store
│   │   │   └── store.go                    # Single-use temporary token storage (TTL: 60s)
│   │   ├── queue/                          # Asynchronous task queues
│   │   │   ├── delete_queue.go             # BullMQ attachment deletion producer
│   │   │   ├── delete_worker.go            # BullMQ background worker for Cloudinary asset deletion
│   │   │   └── jobs.go                     # Queue job payloads
│   │   ├── ratelimit/                      # Redis rate limit scripts
│   │   │   └── ratelimit.go                # Token-bucket / sliding-window Redis storage
│   │   ├── refreshtoken/                   # Refresh token session storage
│   │   │   └── refresh_token.go            # Refresh token persistence and revocation
│   │   └── redis.go                        # Redis client factory
│   ├── repository/                         # Database repository layer
│   │   ├── archive_repository.go           # Soft-deleted records retrieval
│   │   ├── authorization_repository.go     # RBAC permission queries
│   │   ├── comment_repository.go           # Comment queries
│   │   ├── errors.go                       # Repository-level error sentinels
│   │   ├── issue_activity_repository.go    # Audit log recording & retrieval
│   │   ├── issue_attachments_repository.go # Attachment metadata storage
│   │   ├── issue_repository.go             # Issue queries & filters
│   │   ├── labels_repository.go            # Label queries
│   │   ├── project_members_repository.go   # Project member queries
│   │   ├── project_repository.go           # Project queries
│   │   ├── user_repository.go              # User & OIDC identity queries
│   │   ├── workspace_invitations_repository.go # Invitation queries
│   │   ├── workspace_members_repository.go # Workspace member queries
│   │   └── workspace_repository.go         # Workspace queries
│   ├── service/                            # Core business logic layer
│   │   ├── archive_service.go              # Archive statistics & restoration logic
│   │   ├── authorization.go                # Workspace & project permission enforcement
│   │   ├── authz.go                        # Authorization contracts
│   │   ├── comment_service.go              # Comment operations
│   │   ├── errors.go                       # Business domain error sentinels
│   │   ├── issue_activity_service.go       # Audit trail generation
│   │   ├── issue_attachments_service.go    # Cloudinary upload & queue orchestration
│   │   ├── issues_service.go               # Issue domain operations
│   │   ├── labels_service.go               # Label operations
│   │   ├── project_members_service.go      # Project member operations
│   │   ├── project_service.go              # Project domain operations
│   │   ├── user_service.go                 # User auth, Google OIDC, & handoff logic
│   │   ├── workspace_invitations_service.go# Invitation lifecycle management
│   │   ├── workspace_members_service.go    # Workspace member operations
│   │   └── workspace_service.go            # Workspace domain operations
│   └── storage/                            # Media storage interfaces
│       ├── cloudinary.go                   # Cloudinary SDK client implementation
│       └── storage.go                      # Cloud storage abstraction interface
├── benchmarks/                             # Performance & load testing suite
│   ├── benchmark_issues.json               # Seed benchmark data
│   ├── create_benchmark_issues.go          # Seed issues generator
│   ├── create_tokens.go                    # Benchmark JWT generator
│   ├── create_users.go                     # Benchmark user generator
│   ├── issues.js                           # k6 load testing script
│   └── tokens.json                         # Benchmark token pool
├── prometheus/
│   └── prometheus.yml                      # Prometheus scraper configuration
├── Dockerfile                              # Multi-stage Alpine container build
├── docker-compose.yml                      # Full local stack (API, PostgreSQL 17, Redis 7, Prometheus, Grafana)
├── go.mod                                  # Go dependencies & replacement directives
├── go.sum                                  # Dependency checksums
├── sqlc.yaml                               # sqlc code generation configuration
└── README.md                               # This documentation
```

---

## Architecture & Data Flow

IssueForge strictly enforces a unidirectional, decoupled layered architecture:

```text
┌────────────────────────────────────────────────────────┐
│                      Client HTTP                       │
└───────────────────────────┬────────────────────────────┘
                            │
┌───────────────────────────▼────────────────────────────┐
│                    Middleware Layer                    │
│   (Prometheus Metrics, CORS, Redis Rate Limiters, JWT) │
└───────────────────────────┬────────────────────────────┘
                            │
┌───────────────────────────▼────────────────────────────┐
│                     Handler Layer                      │
│        (Payload decoding, status codes, JSON format)   │
└───────────────────────────┬────────────────────────────┘
                            │
┌───────────────────────────▼────────────────────────────┐
│                     Service Layer                      │
│   (Business rules, RBAC checks, cache coordination,    │
│    OIDC exchange, and background queue dispatch)       │
└─────────────┬──────────────────────────────┬───────────┘
              │                              │
┌─────────────▼─────────────┐  ┌─────────────▼───────────┐
│     Repository Layer      │  │  Redis Cache / Handoff  │
│  (sqlc type-safe queries) │  │  (Multi-level TTL cache,│
└─────────────┬─────────────┘  │   token handoff store)  │
              │                └─────────────────────────┘
┌─────────────▼─────────────┐  ┌─────────────────────────┐
│       PostgreSQL 17       │  │    GoBullMQ Worker      │
│     (pgx connection pool) │  │  (Async asset deletion) │
└───────────────────────────┘  └─────────────┬───────────┘
                                             │
                               ┌─────────────▼───────────┐
                               │     Cloudinary API      │
                               └─────────────────────────┘
```

---

## Prerequisites

- **Go**: 1.26.4 or later
- **PostgreSQL**: 17+ (or compatible Neon / AWS RDS PostgreSQL)
- **Redis**: 7+ (or compatible Upstash / AWS ElastiCache)
- **Cloudinary Account**: Required for image and file attachment storage
- **Google Cloud Console Credentials**: OAuth 2.0 Web Client ID & Secret for Google OIDC
- **golang-migrate / sqlc**: (Optional, for running schema migrations and regenerating code)
- **Docker & Docker Compose**: (Optional, for running the complete local infrastructure)

---

## Getting Started

### 1. Clone Repository & Install Dependencies

```bash
git clone https://github.com/DAbharat/IssueForge.git
cd IssueForge
go mod download
```

### 2. Configure Environment

Create your `.env` file in the project root:

```env
# Server
PORT=8080
FRONTEND_URL=http://localhost:3000
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://issueforge.app

# PostgreSQL 17
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_postgres_password
DB_NAME=issueforge
DB_SSLMODE=disable

# Redis 7
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_TTL=10m
REFRESH_TOKEN_TTL=168h

# Security
JWTSecret=your_long_cryptographically_secure_jwt_secret_min_32_chars
DISABLE_RATE_LIMIT=false

# Cloudinary Storage
CLOUDINARY_URL=cloudinary://<API_KEY>:<API_SECRET>@<CLOUD_NAME>
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret

# Google OIDC Authentication
GOOGLE_CLIENT_ID=your_google_client_id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/google/callback
```

### 3. Run Database Migrations

Run all 21 versioned migrations up to Google Sub linking:

```bash
migrate -path internal/db/migrations -database "postgres://postgres:your_password@localhost:5432/issueforge?sslmode=disable" up
```

### 4. Run the Application

```bash
go run ./cmd/api/main.go
```

The application will initialize database pools, Redis connections, Prometheus collectors, and BullMQ workers, starting on `:8080` (or `PORT`).

---

## Configuration & Environment Variables

| Variable | Type | Required | Description | Default / Example |
|---|---|---|---|---|
| `PORT` | Integer / String | No | HTTP port for the API server | `8080` |
| `FRONTEND_URL` | URL | No | Frontend URL for Google OIDC handoff redirect | `http://localhost:4000` |
| `CORS_ALLOWED_ORIGINS`| String | No | Comma-separated list of allowed origins | `http://localhost:3000` |
| `DB_HOST` | String | **Yes** | PostgreSQL host address | `localhost` |
| `DB_PORT` | Integer | **Yes** | PostgreSQL port | `5432` |
| `DB_USER` | String | **Yes** | PostgreSQL database username | `postgres` |
| `DB_PASSWORD` | String | **Yes** | PostgreSQL user password | - |
| `DB_NAME` | String | **Yes** | PostgreSQL database name | `issueforge` |
| `DB_SSLMODE` | String | No | PostgreSQL SSL connection mode (`disable`, `require`) | `disable` |
| `JWTSecret` | String | **Yes** | Secret used for HMAC-SHA256 JWT access token signing | - |
| `REDIS_ADDR` | String | **Yes** | Redis host and port combination | `localhost:6379` |
| `REDIS_PASSWORD` | String | No | Redis authentication password | `""` |
| `REDIS_DB` | Integer | **Yes** | Redis logical database index | `0` |
| `REDIS_TTL` | Duration | **Yes** | Standard domain cache TTL | `10m` |
| `REFRESH_TOKEN_TTL` | Duration | **Yes** | Refresh token lifetime in Redis | `168h` (7 days) |
| `CLOUDINARY_URL` | String | **Yes** | Cloudinary connection URI | `cloudinary://...` |
| `CLOUDINARY_CLOUD_NAME`| String | **Yes** | Cloudinary cloud identifier | - |
| `CLOUDINARY_API_KEY` | String | **Yes** | Cloudinary API Key | - |
| `CLOUDINARY_API_SECRET`| String | **Yes** | Cloudinary API Secret | - |
| `GOOGLE_CLIENT_ID` | String | **Yes** | Google Cloud Console OAuth 2.0 Web Client ID | - |
| `GOOGLE_CLIENT_SECRET` | String | **Yes** | Google Cloud Console OAuth 2.0 Client Secret | - |
| `GOOGLE_REDIRECT_URL` | URL | **Yes** | Google OAuth callback URL pointing to backend | `http://localhost:8080/api/auth/google/callback` |
| `DISABLE_RATE_LIMIT` | Boolean | No | Set to `true` to disable Redis rate limiting (benchmarks) | `false` |

---

## API Reference

All API routes are grouped under the `/api/` prefix. Protected routes require `Authorization: Bearer <access_token>`.

### 1. Authentication & User Management

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `POST` | `/api/register` | No | Auth (3/m) | Register a new user account with email, fullname, username, and password |
| `POST` | `/api/login` | No | Strict (5/m) | Authenticate with email/password; returns access token, refresh token, & user info |
| `POST` | `/api/auth/refresh` | No | Auth (3/m) | Rotate or renew expired access token using valid `refresh_token` |
| `GET` | `/api/auth/google` | No | Auth (3/m) | Initiate Google OAuth 2.0 flow; sets `google_oauth_state` cookie and redirects to Google |
| `GET` | `/api/auth/google/callback` | No | None | Callback invoked by Google; validates CSRF state, issues tokens, and executes handoff redirect |
| `POST` | `/api/auth/exchange` | No | None | Exchange one-time 32-byte OIDC handoff code for JWT access & refresh tokens |
| `POST` | `/api/users/logout` | **Yes** | Auth (3/m) | Revoke refresh token in Redis and invalidate current session |
| `GET` | `/api/me` | **Yes** | Read (200/m) | Retrieve current user profile and list of workspace memberships |
| `GET` | `/api/users/search?search={query}` | **Yes** | Read (200/m) | Search users by partial username match |
| `GET` | `/api/users/{userID}` | **Yes** | Read (200/m) | Retrieve public user profile by numeric user ID |
| `GET` | `/api/users/{username}` | **Yes** | Read (200/m) | Retrieve public user profile by username |
| `DELETE`| `/api/users/{userID}` | **Yes** | Delete (20/m) | Delete user account |

### 2. Workspaces

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `POST` | `/api/workspaces` | **Yes** | Create (30/m) | Create a new team workspace (creator automatically assigned as `owner`) |
| `GET` | `/api/workspaces` | **Yes** | Read (200/m) | List all active workspaces where the caller is a member |
| `GET` | `/api/workspaces/{workspaceID}` | **Yes** | Read (200/m) | Get workspace details by ID |
| `PATCH` | `/api/workspaces/{workspaceID}` | **Yes** | Patch (60/m) | Update workspace name (owner/admin only) |
| `DELETE`| `/api/workspaces/{workspaceID}` | **Yes** | Delete (20/m) | Soft-delete a workspace (owner only) |
| `POST` | `/api/workspace/{workspaceID}` | **Yes** | None | Restore a soft-deleted workspace |

### 3. Workspace Members

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `POST` | `/api/workspaces/{workspaceID}/members` | **Yes** | None | Add an existing user directly to a workspace |
| `GET` | `/api/workspaces/{workspaceID}/members` | **Yes** | Read (200/m) | List all members and roles in a workspace |
| `GET` | `/api/workspaces/{workspaceID}/members/{userID}` | **Yes** | Read (200/m) | Get specific workspace member profile and role |
| `DELETE`| `/api/workspaces/{workspaceID}/members/{userID}` | **Yes** | None | Remove a member from a workspace (owner/admin only) |
| `PATCH` | `/api/workspaces/{workspaceID}/member/promote` | **Yes** | Strict (5/m) | Promote a workspace member to `admin` |

### 4. Workspace Invitations

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `POST` | `/api/workspaces/{workspaceID}/invitations` | **Yes** | Create (30/m) | Invite a user to a workspace by email/username with specified role |
| `GET` | `/api/workspaces/{workspaceID}/invitations/pending` | **Yes** | Read (200/m) | List pending invitations sent from a workspace |
| `GET` | `/api/invitations/pending` | **Yes** | Read (200/m) | List pending invitations received by the authenticated user |
| `GET` | `/api/invitations/{invitationID}` | **Yes** | Read (200/m) | Get invitation details by numeric invitation ID |
| `POST` | `/api/invitations/{invitationID}/accept` | **Yes** | Create (30/m) | Accept an invitation and join the workspace |
| `POST` | `/api/invitations/{invitationID}/decline` | **Yes** | Create (30/m) | Decline a workspace invitation |
| `POST` | `/api/invitations/{invitationID}/cancel` | **Yes** | Create (30/m) | Cancel a dispatched invitation (workspace admin/owner only) |

### 5. Projects

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `POST` | `/api/workspaces/{workspaceID}/projects` | **Yes** | Create (30/m) | Create a new project within a workspace |
| `GET` | `/api/workspaces/{workspaceID}/projects` | **Yes** | Read (200/m) | List all active projects in a workspace |
| `GET` | `/api/projects/{projectID}` | **Yes** | Read (200/m) | Get project details by ID |
| `PATCH` | `/api/projects/{projectID}/details` | **Yes** | Patch (60/m) | Update project name and description |
| `PATCH` | `/api/projects/{projectID}/lead` | **Yes** | Patch (60/m) | Reassign the project lead to another member |
| `DELETE`| `/api/projects/{projectID}` | **Yes** | Delete (20/m) | Soft-delete a project |
| `POST` | `/api/projects/{projectID}/restore` | **Yes** | None | Restore a soft-deleted project |

### 6. Project Members

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `POST` | `/api/projects/{projectID}/members` | **Yes** | None | Add a workspace member to a specific project |
| `GET` | `/api/projects/{projectID}/members` | **Yes** | Read (200/m) | List all members assigned to a project |

### 7. Issues

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `POST` | `/api/projects/{projectID}/issues` | **Yes** | Create (30/m) | Create a new issue in a project |
| `GET` | `/api/projects/{projectID}/issues` | **Yes** | Read (200/m) | List all issues in a project |
| `GET` | `/api/issues/{issueID}` | **Yes** | Read (200/m) | Get full details for an issue |
| `PATCH` | `/api/issues/{issueID}/details` | **Yes** | Patch (60/m) | Update issue title and description |
| `PATCH` | `/api/issues/{issueID}/status` | **Yes** | Patch (60/m) | Update status (`backlog`, `todo`, `in_progress`, `done`, `canceled`) |
| `PATCH` | `/api/issues/{issueID}/assignee` | **Yes** | Patch (60/m) | Reassign issue to another project member |
| `PATCH` | `/api/issues/{issueID}/priority` | **Yes** | Patch (60/m) | Update priority (`low`, `medium`, `high`, `urgent`) |
| `PATCH` | `/api/issues/{issueID}/due-date` | **Yes** | Patch (60/m) | Update or clear issue target completion date |
| `GET` | `/api/users/me/issues/assigned` | **Yes** | Read (200/m) | List all issues currently assigned to authenticated user |
| `GET` | `/api/users/me/issues/created` | **Yes** | Read (200/m) | List all issues created by authenticated user |
| `DELETE`| `/api/issues/{issueID}` | **Yes** | Delete (20/m) | Soft-delete an issue |
| `PATCH` | `/api/issues/{issueID}/restore` | **Yes** | None | Restore a soft-deleted issue |

### 8. Comments

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `POST` | `/api/issues/{issueID}/comments` | **Yes** | None | Post a comment or threaded reply on an issue |
| `GET` | `/api/issues/{issueID}/comments` | **Yes** | None | List all comments associated with an issue |
| `GET` | `/api/comments/{commentID}` | **Yes** | None | Get a single comment by ID |
| `PATCH` | `/api/comments/{commentID}` | **Yes** | None | Edit comment content (author only; sets `is_edited=true`) |
| `DELETE`| `/api/comments/{commentID}` | **Yes** | None | Delete comment (author or admin only) |

### 9. Issue Activity & Audit Trail

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `GET` | `/api/issues/{issueID}/activities` | **Yes** | None | Retrieve chronological changelog of all issue modifications |

### 10. Issue & Comment Attachments

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `POST` | `/api/issues/{issueID}/attachments` | **Yes** | Attachment (10/m) | Upload attachment file to Cloudinary and link to issue/comment |
| `GET` | `/api/issues/{issueID}/attachments` | **Yes** | Read (200/m) | List all file attachments uploaded to an issue |
| `GET` | `/api/comments/{commentID}/attachments` | **Yes** | Read (200/m) | List file attachments uploaded to a specific comment |
| `GET` | `/api/attachments/{attachmentID}` | **Yes** | Read (200/m) | Get attachment metadata and public download URL |
| `DELETE`| `/api/attachments/{attachmentID}` | **Yes** | None | Soft-delete attachment metadata and enqueue async Cloudinary purge |

### 11. Labels

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `POST` | `/api/projects/{projectID}/labels` | **Yes** | None | Create a new custom label with name and hex color |
| `GET` | `/api/projects/{projectID}/labels` | **Yes** | None | List all labels defined within a project |
| `GET` | `/api/labels/{labelID}` | **Yes** | None | Get label details by ID |
| `PATCH` | `/api/labels/{labelID}` | **Yes** | None | Update label name and color |
| `DELETE`| `/api/labels/{labelID}` | **Yes** | None | Delete label from project |
| `POST` | `/api/issues/{issueID}/labels` | **Yes** | None | Attach one or more label IDs to an issue |
| `DELETE`| `/api/issues/{issueID}/labels/{labelID}` | **Yes** | None | Detach a label from an issue |
| `GET` | `/api/issues/{issueID}/labels` | **Yes** | None | List all labels attached to an issue |

### 12. Archive & Soft-Delete Recovery

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `GET` | `/api/archive/counts` | **Yes** | Read (200/m) | Aggregate count of deleted workspaces, projects, issues, & files |
| `GET` | `/api/archive/workspaces` | **Yes** | Read (200/m) | List all soft-deleted workspaces accessible to the user |
| `GET` | `/api/archive/projects` | **Yes** | Read (200/m) | List all soft-deleted projects accessible to the user |
| `GET` | `/api/archive/issues` | **Yes** | Read (200/m) | List all soft-deleted issues accessible to the user |
| `GET` | `/api/archive/attachments` | **Yes** | Read (200/m) | List all soft-deleted file attachments accessible to the user |

### 13. Health & Metrics

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `GET` | `/api/health` | No | None | System status, database ping, and uptime check |
| `GET` | `/api/metrics` | No | None | Prometheus format metrics exposition |

---

## Database & Migrations

IssueForge utilizes PostgreSQL 17 with 21 versioned migrations in `internal/db/migrations/`:

| Version | Migration Name | Key Schema Additions |
|---|---|---|
| `000001` | `create_user_role_enum` | Enum `user_role` (`owner`, `admin`, `member`) |
| `000002` | `create_workspaces_table` | `workspaces` table with name, timestamps |
| `000003` | `create_users_table` | `users` table with email, username, password_hash |
| `000004` | `create_projects_table` | `projects` table with workspace foreign key and lead_id |
| `000005` | `create_project_members_table` | `project_members` junction table |
| `000006` | `create_workspace_members_table`| `workspace_members` junction with role enum |
| `000007` | `create_issues_table` | `issues` table with status, priority, due date, creator, assignee |
| `000008` | `alter_issue_table` | Status and priority constraints |
| `000009` | `create_comments_table` | `comments` table with parent comment threading |
| `000010` | `create_issue_activities_table` | `issue_activities` audit log table |
| `000011` | `add_soft_delete_to_issues` | `deleted_at` timestamp on `issues` |
| `000012` | `create_issue_attachments_table`| `issue_attachments` table with Cloudinary public_id |
| `000013` | `alter_issues_table` | Foreign key index optimizations |
| `000014` | `create_labels_table` | `labels` table with project scope and color hex |
| `000015` | `create_issue_labels_table` | `issue_labels` many-to-many junction |
| `000016` | `alter_project_table` | `deleted_at` timestamp on `projects` |
| `000017` | `add_deleted_at_to_workspaces` | `deleted_at` timestamp on `workspaces` |
| `000018` | `alter_users_table` | User profile timestamp and constraint adjustments |
| `000019` | `create_workspace_invitations` | `workspace_invitations` table with status enum |
| `000020` | `add_public_url_to_attachments` | Direct `public_url` storage for Cloudinary media |
| `000021` | `add_google_sub_to_users` | `google_sub` column with unique partial index |

---

## Caching & Asynchronous Queues

### Multi-Level Redis Cache

Read-through caches intercept repetitive lookups with configurable TTL (`REDIS_TTL`):
- **Workspace Cache** (`cache:workspace:{id}`): Invalidated on workspace update, deletion, or restoration.
- **Project Cache** (`cache:project:{id}`): Invalidated on project rename, lead change, or deletion.
- **Issue Cache** (`cache:issue:{id}`): Invalidated on any issue field update or deletion.
- **Invitation Cache** (`cache:invitation:{id}`): Caches pending invitations with automatic invalidation on acceptance, rejection, or cancellation.

### Asynchronous BullMQ Queue

Media cleanup is completely decoupled from HTTP request latency:
1. User deletes an attachment via `DELETE /api/attachments/{id}`.
2. The attachment record is soft-deleted in PostgreSQL (`deleted_at = NOW()`).
3. An asynchronous job is dispatched to GoBullMQ queue `attachment_delete_queue`.
4. The background `AttachmentDeleteWorker` consumes the job, calls Cloudinary Destroy API, and handles retries without blocking the user.

---

## Authentication & Security

### Google OIDC & Redis Handoff Exchange

IssueForge eliminates token leakage via browser history or URL query strings by implementing the **Redis Handoff Exchange**:

```text
User                      Browser                         Backend API (Go)              Google Identity
 │                           │                                   │                             │
 │── Click "Google Sign In"─>│                                   │                             │
 │                           │── GET /api/auth/google ──────────>│                             │
 │                           │                                   │ Generates 32-byte CSRF state│
 │                           │                                   │ Sets google_oauth_state     │
 │                           │                                   │ (HttpOnly, SameSite=Lax)    │
 │                           │<── 302 Redirect ──────────────────│                             │
 │                           │                                   │                             │
 │                           │── Follow 302 redirect ─────────────────────────────────────────>│
 │                           │   (to accounts.google.com)        │                             │
 │── Authenticate & Consent ──────────────────────────────────────────────────────────────────>│
 │                           │<── 302 Redirect with code & state ──────────────────────────────│
 │                           │                                   │                             │
 │                           │── GET /api/auth/google/callback ─>│                             │
 │                           │   (Browser attaches state cookie) │                             │
 │                           │                                   │── POST /token (exchange) ──>│
 │                           │                                   │<── Returns tokens & IDToken─│
 │                           │                                   │                             │
 │                           │                                   │ Verifies ID Token (OIDC)    │
 │                           │                                   │ Resolves User in PostgreSQL │
 │                           │                                   │ Generates JWT Access Token  │
 │                           │                                   │ Stores Refresh Token (Redis)│
 │                           │                                   │                             │
 │                           │                                   │ CreateHandoff():            │
 │                           │                                   │ - Generates 32-byte code    │
 │                           │                                   │ - Saves tokens to Redis     │
 │                           │                                   │   (TTL: 60s, handoff:<code>)│
 │                           │<── 302 Found ─────────────────────│                             │
 │                           │    Location:                      │                             │
 │                           │    /auth/callback?code=HANDOFF    │                             │
 │                           │                                   │                             │
 │                           │── Next.js /auth/callback loads ───│                             │
 │                           │                                   │                             │
 │                           │── POST /api/auth/exchange ───────>│                             │
 │                           │   Body: { "code": HANDOFF }       │ Atomically consumes &       │
 │                           │                                   │ deletes handoff from Redis  │
 │                           │<── 200 OK ────────────────────────│                             │
 │                           │    { access_token, refresh_token }│                             │
 │                           │                                   │                             │
 │── Session Stored ────────>│ Passes to useAuth.setAuthTokens() │                             │
 │── Redirect to App ───────>│ Navigates to /profile             │                             │
```

- **Account Linking**: If an existing account with the verified Google email is found, `google_sub` is bound automatically.
- **Auto-Provisioning**: New Google users receive a sanitized username (`base_xxxxx`) and placeholder hash `!oidc-no-password!`.
- **Atomic Single-Use**: The handoff entry is deleted (`DEL`) from Redis immediately upon reading. Replaying or double-invoking returns `400 invalid handoff code`.

### Role-Based Access Control (RBAC)

| Permission | Owner | Admin | Member | Non-Member |
|---|:---:|:---:|:---:|:---:|
| Soft-Delete Workspace | Yes | No | No | No |
| Rename Workspace | Yes | Yes | No | No |
| Invite Members | Yes | Yes | No | No |
| Promote to Admin | Yes | No | No | No |
| Create Project | Yes | Yes | Yes | No |
| Update Project Lead | Yes | Yes (Project Lead) | No | No |
| Soft-Delete Project | Yes | Yes | No | No |
| Create / Edit Issues | Yes | Yes | Yes | No |
| Delete Comments | Yes (Any) | Yes (Any) | Author Only | No |

---

## Rate Limiting

The API implements distributed rate limiting using Redis token buckets:

| Tier | Capacity | Refill Rate | Protected Routes |
|---|---|---|---|
| **Strict** | 5 requests | 5 req / 60 sec | `POST /api/login`, `PATCH /api/workspaces/{id}/member/promote` |
| **Auth** | 3 requests | 3 req / 60 sec | `POST /api/register`, `GET /api/auth/google`, `POST /api/auth/refresh`, `POST /api/users/logout` |
| **Attachment** | 10 requests | 10 req / 60 sec | `POST /api/issues/{id}/attachments` |
| **Delete** | 20 requests | 20 req / 60 sec | `DELETE /api/users/{id}`, `DELETE /api/issues/{id}`, `DELETE /api/workspaces/{id}`, etc. |
| **Create** | 30 requests | 30 req / 60 sec | `POST /api/workspaces`, `POST /api/projects`, `POST /api/issues`, invitations |
| **Write** | 30 requests | 30 req / 60 sec | Standard mutations |
| **Patch** | 60 requests | 60 req / 60 sec | Status, priority, lead, due-date, details updates |
| **Issues** | 60 requests | 60 req / 60 sec | Issue domain actions |
| **Read** | 200 requests | 200 req / 60 sec | All collection listings and detail GET endpoints |

Set `DISABLE_RATE_LIMIT=true` in `.env` to bypass limits for automated load testing.

---

## Monitoring & Observability

IssueForge instruments all HTTP traffic according to the **RED Principle** (Rate, Errors, Duration) via custom Prometheus collectors registered in [internal/middleware/metrics.go](file:///d:/CodeFiles/IssueForge/internal/middleware/metrics.go).

### RED Metrics Architecture

| Component | Metric Name | Metric Type | Labels | Description |
|---|---|---|---|---|
| **Rate** | `issueforge_http_requests_total` | Counter | `method`, `route`, `status` | Cumulative count of incoming requests processed across routes |
| **Errors** | `issueforge_http_requests_total` | Counter | `method`, `route`, `status` | Filtered where `status=~"[45].."` to track client (4xx) and server (5xx) failures |
| **Duration** | `issueforge_http_request_duration_seconds` | Histogram | `method`, `route` | Latency distribution with buckets: `[.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10]` seconds |
| **Concurrency** | `issueforge_http_requests_in_flight` | Gauge | `method` | Real-time count of requests currently being served |
| **Payload Size** | `issueforge_http_response_size_bytes` | Histogram | `method`, `route` | Response payload size distribution in bytes (exponential buckets: `100B` to `10MB`) |

In addition to HTTP metrics, the `/api/metrics` registry includes default Go runtime and OS process collectors (`go_goroutines`, `go_memstats_*`, `process_cpu_seconds_total`, `process_resident_memory_bytes`).

### Production PromQL Queries

```promql
# 1. Total Request Rate (R - Rate, throughput across all endpoints)
sum(rate(issueforge_http_requests_total[5m]))

# 2. Request Rate Broken Down by Route
sum by (route) (rate(issueforge_http_requests_total[5m]))

# 3. Overall Error Percentage (E - Errors, ratio of non-2xx responses)
(sum(rate(issueforge_http_requests_total{status=~"[45].."}[5m])) / sum(rate(issueforge_http_requests_total[5m]))) * 100

# 4. Critical 5xx Server Error Rate
sum(rate(issueforge_http_requests_total{status=~"5.."}[5m]))

# 5. Latency Percentiles (D - Duration, p95 and p99 response times)
histogram_quantile(0.95, sum by (le, route) (rate(issueforge_http_request_duration_seconds_bucket[5m])))
histogram_quantile(0.99, sum by (le, route) (rate(issueforge_http_request_duration_seconds_bucket[5m])))

# 6. Real-Time In-Flight Requests
sum(issueforge_http_requests_in_flight)
```

### Metrics Exposition & Scrape Configuration

- **Exposition Endpoint**: `GET /api/metrics`
- **Prometheus Scraper**: Configured in `prometheus/prometheus.yml` scraping `metrics_path: /api/metrics` every `5s`.
- **Grafana Visualization**: Available at `http://localhost:3000` connected to Prometheus at `http://localhost:9090`.

---

## Performance & Benchmarks

The `benchmarks/` directory contains an end-to-end load testing suite:
- **`create_users.go`**: Generates and seeds bulk test users.
- **`create_tokens.go`**: Signs high-volume valid test JWT access tokens.
- **`create_benchmark_issues.go`**: Seeds large issue datasets for query stress testing.
- **`issues.js`**: **k6** load testing script testing issue creation, listing, and updates under high concurrency.

### Running the Benchmark Suite

The Go helper scripts are standalone utilities and can be executed individually to prepare test fixtures:

```bash
# 1. Seed benchmark users
go run ./benchmarks/create_users.go

# 2. Generate authentication tokens for load testing
go run ./benchmarks/create_tokens.go

# 3. Seed benchmark issues
go run ./benchmarks/create_benchmark_issues.go

# 4. Run the k6 load testing scenario
k6 run benchmarks/issues.js
```

---

## Docker Deployment

### Multi-Container Stack (`docker-compose.yml`)

The root `docker-compose.yml` orchestrates the complete backend infrastructure:
- **`api`**: IssueForge Go binary running on Alpine 3.22 (port `8080`).
- **`postgres`**: PostgreSQL 17 Alpine with persistent volume (`5432`).
- **`redis`**: Redis 7 Alpine with AOF persistence (`6379`).
- **`prometheus`**: Prometheus time-series collector (`9090`).
- **`grafana`**: Grafana visual dashboard (`3000`).

```bash
# Start all services in the background
docker-compose up -d

# View live container logs
docker-compose logs -f api

# Stop all containers
docker-compose down
```

### Standalone Docker Build

```bash
docker build -t issueforge:latest .
docker run -p 8080:8080 --env-file .env issueforge:latest
```

---

## Development & Code Quality

### sqlc Code Generation

When modifying SQL queries in `internal/db/queries/` or schema migrations in `internal/db/migrations/`:

```bash
# Install sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Recompile Go repository code
sqlc generate
```

### Testing & Linting

```bash
# Run unit and integration tests
go test ./...

# Run tests with race detection and coverage
go test -race -cover ./...

# Format source files
go fmt ./...

# Static analysis and vetting
go vet ./...
```

---

## Contributing

1. Fork the repository.
2. Create your feature branch (`git checkout -b feature/my-new-feature`).
3. Commit your changes with conventional commit messages (`git commit -m 'feat: add activity filtering'`).
4. Push to your branch (`git push origin feature/my-new-feature`).
5. Open a Pull Request.

---

## License

This project is licensed under the MIT License - see the LICENSE file for details.
