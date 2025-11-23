# 🏗️ GoFiber Auth Service - Architecture Documentation

> **สรุปโครงสร้างและการทำงานของ Auth Service ทั้งหมด**

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Architecture Pattern](#architecture-pattern)
3. [Directory Structure](#directory-structure)
4. [Data Flow](#data-flow)
5. [Layer Details](#layer-details)
6. [Dependencies](#dependencies)
7. [Configuration](#configuration)
8. [Key Features](#key-features)

---

## Overview

**Auth Service** คือ Authentication & Authorization microservice ที่สร้างด้วย:
- **Framework:** GoFiber v2
- **Database:** PostgreSQL (GORM)
- **Cache:** Redis
- **Authentication:** JWT + OAuth 2.0
- **Architecture:** Clean Architecture

**Port:** `8088` (default)

**Main Features:**
- ✅ User Registration & Login (Email/Password)
- ✅ OAuth 2.0 (Google, Facebook, LINE)
- ✅ JWT Token Generation & Validation
- ✅ User Sync to Backend Service
- ✅ Graceful Shutdown
- ✅ Health Check Endpoints

---

## Architecture Pattern

ใช้ **Clean Architecture** แบ่งเป็น 4 layers:

```
┌─────────────────────────────────────────────────────────┐
│                   cmd/api (main.go)                     │
│                    Entry Point                          │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│              interfaces/ (Adapters)                     │
│  ┌─────────────────────────────────────────────────┐   │
│  │  API Layer (Handlers, Routes, Middleware)      │   │
│  │  - HTTP Handlers                                │   │
│  │  - Routes Setup                                 │   │
│  │  - CORS, Auth, Logger Middleware                │   │
│  └─────────────────────────────────────────────────┘   │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│           application/ (Use Cases)                      │
│  ┌─────────────────────────────────────────────────┐   │
│  │  Service Implementations                        │   │
│  │  - UserServiceImpl                              │   │
│  │  - OAuthServiceImpl                             │   │
│  │  - SyncService                                  │   │
│  └─────────────────────────────────────────────────┘   │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│              domain/ (Business Logic)                   │
│  ┌──────────────────┬──────────────────┬───────────┐   │
│  │    Models        │   Repositories   │  Services │   │
│  │  (Entities)      │   (Interfaces)   │  (Ports)  │   │
│  ├──────────────────┼──────────────────┼───────────┤   │
│  │  - User          │  - UserRepo      │  - User   │   │
│  │  - OAuthProvider │  - OAuthRepo     │  - OAuth  │   │
│  └──────────────────┴──────────────────┴───────────┘   │
│  ┌─────────────────────────────────────────────────┐   │
│  │  DTOs (Data Transfer Objects)                   │   │
│  │  - Request/Response structures                  │   │
│  └─────────────────────────────────────────────────┘   │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│          infrastructure/ (External)                     │
│  ┌──────────────────┬──────────────────┬───────────┐   │
│  │   PostgreSQL     │      Redis       │  Storage  │   │
│  │ (Repository      │   (Caching)      │  (Bunny   │   │
│  │  Implementations)│                  │   CDN)    │   │
│  └──────────────────┴──────────────────┴───────────┘   │
└─────────────────────────────────────────────────────────┘
```

---

## Directory Structure

```
gofiber-auth/
│
├── cmd/                          # Entry points
│   ├── api/
│   │   └── main.go              # Main application entry
│   └── migrate/                 # Migration utilities
│       ├── check_db.go
│       ├── import_users.go
│       └── migrate_to_displayname.go
│
├── domain/                       # Business Logic Layer
│   ├── dto/                     # Data Transfer Objects
│   │   ├── auth.go             # Auth DTOs (Login, Register)
│   │   ├── oauth.go            # OAuth DTOs
│   │   ├── user.go             # User DTOs
│   │   ├── common.go           # Common response DTOs
│   │   └── mappers.go          # Model ↔ DTO mappers
│   │
│   ├── models/                  # Domain Entities
│   │   ├── user.go             # User entity
│   │   └── oauth_provider.go  # OAuth provider entity
│   │
│   ├── repositories/            # Repository Interfaces
│   │   ├── user_repository.go
│   │   └── oauth_repository.go
│   │
│   └── services/                # Service Interfaces (Ports)
│       ├── user_service.go
│       └── oauth_service.go
│
├── application/                  # Use Cases Layer
│   └── serviceimpl/             # Service Implementations
│       ├── user_service_impl.go
│       ├── oauth_service_impl.go
│       └── sync_service.go      # Backend sync service
│
├── infrastructure/               # External Dependencies
│   ├── postgres/                # PostgreSQL implementations
│   │   ├── database.go         # DB connection
│   │   ├── user_repository_impl.go
│   │   └── oauth_repository_impl.go
│   │
│   ├── redis/                   # Redis client
│   │   └── redis.go
│   │
│   ├── storage/                 # File storage (Bunny CDN)
│   │   └── bunny.go
│   │
│   └── websocket/               # WebSocket (future)
│       └── websocket.go
│
├── interfaces/                   # Adapters Layer
│   └── api/
│       ├── handlers/            # HTTP Handlers
│       │   ├── handlers.go     # Handler constructor
│       │   ├── user_handler.go
│       │   └── oauth_handler.go
│       │
│       ├── middleware/          # HTTP Middleware
│       │   ├── auth_middleware.go
│       │   ├── cors_middleware.go
│       │   ├── error_middleware.go
│       │   └── logger_middleware.go
│       │
│       └── routes/              # Route definitions
│           ├── routes.go       # Main router
│           ├── auth_routes.go
│           ├── user_routes.go
│           └── health_routes.go
│
├── pkg/                          # Shared Packages
│   ├── config/                  # Configuration
│   │   └── config.go
│   │
│   ├── di/                      # Dependency Injection
│   │   └── container.go        # DI Container
│   │
│   ├── auth_code_store/         # Authorization Code Storage
│   │   └── store.go
│   │
│   ├── scheduler/               # Event Scheduler
│   │   └── scheduler.go
│   │
│   └── utils/                   # Utilities
│       ├── jwt.go              # JWT utilities
│       ├── response.go         # Response helpers
│       └── validator.go        # Validation helpers
│
├── docs/                         # Documentation
├── microservice_plan/           # Microservice planning docs
├── postman/                     # Postman collections
├── scripts/                     # Utility scripts
│
├── .env                         # Environment variables
├── .env.example                # Environment template
├── go.mod                       # Go modules
├── go.sum
├── Dockerfile                   # Docker configuration
├── docker-compose.yml          # Docker Compose
└── Makefile                     # Build commands
```

---

## Data Flow

### 1. HTTP Request Flow

```
┌─────────┐
│ Client  │
└────┬────┘
     │ HTTP Request
     │
┌────▼─────────────────────────────────────────────────┐
│  Fiber App (cmd/api/main.go)                         │
│  ┌────────────────────────────────────────────────┐  │
│  │  Middleware Chain                              │  │
│  │  1. LoggerMiddleware   (log request)           │  │
│  │  2. CorsMiddleware     (CORS headers)          │  │
│  │  3. AuthMiddleware     (JWT validation)        │  │
│  │  4. ErrorHandler       (catch errors)          │  │
│  └───────────────────┬────────────────────────────┘  │
└──────────────────────┼───────────────────────────────┘
                       │
┌──────────────────────▼────────────────────────────────┐
│  Routes (interfaces/api/routes/)                      │
│  - Match URL to Handler                               │
│  - /api/v1/auth/login → UserHandler.Login()          │
└──────────────────────┬────────────────────────────────┘
                       │
┌──────────────────────▼────────────────────────────────┐
│  Handlers (interfaces/api/handlers/)                  │
│  - Parse request (validate DTO)                       │
│  - Call Service                                       │
│  - Return response                                    │
└──────────────────────┬────────────────────────────────┘
                       │
┌──────────────────────▼────────────────────────────────┐
│  Services (application/serviceimpl/)                  │
│  - Business logic                                     │
│  - Call Repository                                    │
│  - Generate JWT                                       │
│  - Sync to Backend (async)                            │
└──────────────────────┬────────────────────────────────┘
                       │
┌──────────────────────▼────────────────────────────────┐
│  Repositories (infrastructure/postgres/)              │
│  - Execute SQL queries (via GORM)                     │
│  - Return domain models                               │
└──────────────────────┬────────────────────────────────┘
                       │
┌──────────────────────▼────────────────────────────────┐
│  Database (PostgreSQL)                                │
│  - Store/Retrieve data                                │
└───────────────────────────────────────────────────────┘
```

### 2. User Registration Flow

```
Client
  │
  │ POST /api/v1/auth/register
  │ { email, username, password }
  │
  ▼
UserHandler.Register()
  │
  │ 1. Validate request
  │ 2. Check email/username unique
  │
  ▼
UserService.Register()
  │
  │ 1. Hash password (bcrypt)
  │ 2. Create user record
  │ 3. Generate JWT token
  │
  ▼
UserRepository.Create()
  │
  │ INSERT INTO users
  │
  ▼
Database (PostgreSQL)
  │
  ▼ (async goroutine)
SyncService.SyncUserWithRetry()
  │
  │ POST http://localhost:8080/internal/users/sync
  │ Retry: 3 times with exponential backoff
  │
  ▼
Backend Service (User Cache)
```

### 3. OAuth Login Flow (Google)

```
Client
  │
  │ 1. GET /api/v1/auth/google
  │
  ▼
OAuthHandler.GetGoogleAuthURL()
  │
  │ - Generate state (CSRF token)
  │ - Set cookie: oauth_state
  │ - Return Google OAuth URL
  │
  ▼
Client redirects to Google
  │
  │ User authenticates with Google
  │
  ▼
Google redirects back
  │
  │ 2. GET /api/v1/auth/google/callback?code=xxx&state=xxx
  │
  ▼
OAuthHandler.HandleGoogleCallback()
  │
  │ 1. Validate state (optional if cookie exists)
  │ 2. Exchange code with Google
  │ 3. Get user info from Google
  │
  ▼
OAuthService.GoogleCallback()
  │
  │ 1. Find or create user
  │ 2. Link OAuth provider
  │ 3. Generate authorization code
  │ 4. Store code in memory (5 min expiry)
  │
  ▼
Redirect to Frontend
  │
  │ http://localhost:3000/auth/callback?code=OUR_CODE
  │
  ▼
Client
  │
  │ 3. POST /api/v1/auth/exchange
  │    { code: OUR_CODE }
  │
  ▼
OAuthHandler.ExchangeCodeForToken()
  │
  │ 1. Validate code
  │ 2. Return stored JWT token
  │ 3. Delete code (one-time use)
  │
  ▼
Client receives JWT token
```

---

## Layer Details

### 🔵 1. Domain Layer (`domain/`)

**ไม่มี dependencies กับ layer อื่น** - เป็นศูนย์กลางของ business logic

#### Models (Entities)
```go
// domain/models/user.go
type User struct {
    ID            uuid.UUID
    Email         string
    Username      string
    Password      *string     // Nullable for OAuth users
    DisplayName   string
    Avatar        string
    Role          string      // "user" or "admin"
    IsActive      bool
    IsOAuthUser   bool
    OAuthProvider string      // "google", "facebook", "line"
    OAuthID       string
    EmailVerified bool
    LastLoginAt   *time.Time
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

```go
// domain/models/oauth_provider.go
type OAuthProvider struct {
    ID             uuid.UUID
    UserID         uuid.UUID
    Provider       string      // "google", "facebook", "line"
    ProviderID     string
    AccessToken    string
    RefreshToken   string
    TokenExpiresAt *time.Time
    ProfileData    datatypes.JSON
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

#### DTOs (Data Transfer Objects)
- `dto/auth.go` - Login, Register requests/responses
- `dto/oauth.go` - OAuth URL, Callback, Exchange DTOs
- `dto/user.go` - User response DTO
- `dto/common.go` - Standard API response format
- `dto/mappers.go` - Convert Models ↔ DTOs

#### Repository Interfaces
```go
// domain/repositories/user_repository.go
type UserRepository interface {
    Create(user *models.User) error
    FindByID(id uuid.UUID) (*models.User, error)
    FindByEmail(email string) (*models.User, error)
    FindByUsername(username string) (*models.User, error)
    Update(user *models.User) error
    Delete(id uuid.UUID) error
}
```

#### Service Interfaces
```go
// domain/services/user_service.go
type UserService interface {
    Register(req *dto.RegisterRequest) (*dto.AuthResponse, error)
    Login(req *dto.LoginRequest) (*dto.AuthResponse, error)
    GetUserByID(id uuid.UUID) (*dto.UserResponse, error)
    // ...
}
```

---

### 🟢 2. Application Layer (`application/`)

**Business logic implementations**

#### UserServiceImpl
- Register (with password hashing)
- Login (with JWT generation)
- User CRUD operations
- **Calls SyncService** to push user data to Backend

#### OAuthServiceImpl
- Generate OAuth URLs (Google, Facebook, LINE)
- Handle OAuth callbacks
- Exchange authorization codes
- **Calls UserService** for user creation
- **Calls SyncService** to sync OAuth users

#### SyncService
- Push user data to Backend Service
- HTTP POST with retry mechanism
- Exponential backoff (1s, 2s, 4s)
- Max retries: 3
- Runs asynchronously (goroutine)

---

### 🟡 3. Infrastructure Layer (`infrastructure/`)

**External dependencies implementations**

#### PostgreSQL (`infrastructure/postgres/`)
- `database.go` - Database connection & migration
- `user_repository_impl.go` - Implements UserRepository
- `oauth_repository_impl.go` - Implements OAuthRepository

**Connection Pool:**
```go
MaxIdleConns:    10
MaxOpenConns:    100
ConnMaxLifetime: 1 hour
```

#### Redis (`infrastructure/redis/`)
- Redis client wrapper
- Used for caching (planned)
- Currently: Warning if connection fails, but app continues

#### Storage (`infrastructure/storage/`)
- Bunny CDN integration (planned for avatar uploads)

---

### 🔴 4. Interfaces Layer (`interfaces/`)

**HTTP API adapters**

#### Handlers (`interfaces/api/handlers/`)
- Parse HTTP requests
- Validate input (DTOs)
- Call services
- Return HTTP responses
- Error handling

#### Middleware (`interfaces/api/middleware/`)

**LoggerMiddleware:**
- Log every request (method, path, status, latency)

**CorsMiddleware:**
```go
AllowOrigins:     "http://localhost:3000,http://localhost:3030"
AllowCredentials: true
AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS"
```

**AuthMiddleware:**
- Extract JWT from `Authorization: Bearer <token>`
- Validate JWT signature
- Extract user ID and role
- Store in context: `c.Locals("userID")`, `c.Locals("role")`

**ErrorMiddleware:**
- Catch all errors
- Return standard error response

#### Routes (`interfaces/api/routes/`)

**Auth Routes:**
```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
GET    /api/v1/auth/google
GET    /api/v1/auth/google/callback
POST   /api/v1/auth/exchange
GET    /api/v1/auth/facebook
GET    /api/v1/auth/facebook/callback
GET    /api/v1/auth/line
GET    /api/v1/auth/line/callback
```

**User Routes (Protected):**
```
GET    /api/v1/users/me          (Auth required)
PUT    /api/v1/users/me          (Auth required)
DELETE /api/v1/users/me          (Auth required)
```

**Health Routes:**
```
GET    /health
GET    /
```

---

### 🟣 5. Shared Packages (`pkg/`)

#### Config (`pkg/config/`)
Loads configuration from `.env`:
- App (Name, Port, Env, FrontendURL)
- Database (Host, Port, User, Password, DBName)
- Redis (Host, Port, Password, DB)
- JWT (Secret)
- OAuth (Google, Facebook, LINE credentials)
- Bunny (CDN configuration)

#### DI Container (`pkg/di/`)
**Dependency Injection Container** - จัดการ lifecycle ของ dependencies:

```go
func (c *Container) Initialize() error {
    1. initConfig()        // Load .env
    2. initInfrastructure() // Connect DB, Redis
    3. initRepositories()  // Create repo instances
    4. initServices()      // Create service instances
    5. initScheduler()     // Start background scheduler
}
```

**Cleanup on shutdown:**
- Stop scheduler
- Close Redis connection
- Close database connection

#### Auth Code Store (`pkg/auth_code_store/`)
**In-memory temporary storage** for OAuth authorization codes:
- Stores: Token, User data, State
- Expiry: 5 minutes
- One-time use (deleted after exchange)
- Auto-cleanup of expired codes

#### Scheduler (`pkg/scheduler/`)
Background event scheduler (currently used for cleanup tasks)

#### Utils (`pkg/utils/`)
- `jwt.go` - JWT generation & validation
- `response.go` - Standard response helpers
- `validator.go` - Input validation

---

## Dependencies

### External Dependencies (go.mod)

**Framework:**
- `github.com/gofiber/fiber/v2` - HTTP framework

**Database:**
- `gorm.io/gorm` - ORM
- `gorm.io/driver/postgres` - PostgreSQL driver

**Cache:**
- `github.com/redis/go-redis/v9` - Redis client

**Authentication:**
- `github.com/golang-jwt/jwt/v5` - JWT
- `golang.org/x/crypto/bcrypt` - Password hashing

**OAuth:**
- `golang.org/x/oauth2` - OAuth 2.0 client
- `google.golang.org/api/oauth2/v2` - Google OAuth

**Utilities:**
- `github.com/google/uuid` - UUID generation
- `github.com/joho/godotenv` - .env loader

---

## Configuration

### Environment Variables (`.env`)

```env
# Application
APP_NAME=GoFiber Auth Service
APP_PORT=8088
APP_ENV=development
FRONTEND_URL=http://localhost:3000

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=n147369
DB_NAME=gofiber_auth
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=Log2Window$P@ssWord

# OAuth - Google
GOOGLE_CLIENT_ID=xxx
GOOGLE_CLIENT_SECRET=xxx
GOOGLE_REDIRECT_URL=http://localhost:8088/api/v1/auth/google/callback

# OAuth - Facebook
FACEBOOK_CLIENT_ID=xxx
FACEBOOK_CLIENT_SECRET=xxx
FACEBOOK_REDIRECT_URL=http://localhost:8088/api/v1/auth/facebook/callback

# OAuth - LINE
LINE_CLIENT_ID=xxx
LINE_CLIENT_SECRET=xxx
LINE_REDIRECT_URL=http://localhost:8088/api/v1/auth/line/callback

# Bunny CDN
BUNNY_STORAGE_ZONE=
BUNNY_ACCESS_KEY=
BUNNY_BASE_URL=https://storage.bunnycdn.com
BUNNY_CDN_URL=
```

---

## Key Features

### ✅ Authentication

**Email/Password:**
- Registration with email/username
- Password hashing (bcrypt, cost: 10)
- JWT token generation (HS256, 7-day expiry)
- Login with email + password

**OAuth 2.0:**
- Google, Facebook, LINE support
- Authorization Code Exchange pattern
- Account linking (OAuth → existing email)
- CSRF protection (state parameter)

### ✅ Authorization

**JWT-based:**
- Middleware extracts & validates JWT
- User ID and role stored in context
- Protected routes require valid JWT

**Roles:**
- `user` (default)
- `admin`

### ✅ User Sync

**Push Pattern:**
- Async sync to Backend Service
- HTTP POST to `/internal/users/sync`
- Retry with exponential backoff
- Actions: `created`, `updated`, `deleted`

### ✅ Graceful Shutdown

**SIGTERM/SIGINT handling:**
1. Stop scheduler
2. Close Redis connection
3. Close database connection
4. Exit gracefully

### ✅ Health Check

```
GET /health

Response:
{
  "status": "healthy",
  "timestamp": "2024-11-23T10:00:00Z"
}
```

---

## How It Works

### Startup Sequence

```
1. main.go
   ↓
2. NewContainer()
   ↓
3. container.Initialize()
   ├─ Load Config (.env)
   ├─ Connect Database (PostgreSQL)
   ├─ Run Migrations (AutoMigrate)
   ├─ Connect Redis
   ├─ Initialize Repositories
   ├─ Initialize Services
   └─ Start Scheduler
   ↓
4. Create Fiber App
   ↓
5. Setup Middleware
   ├─ Logger
   ├─ CORS
   └─ Error Handler
   ↓
6. Create Handlers (from Services)
   ↓
7. Setup Routes
   ↓
8. Start Server (Listen on port 8088)
   ↓
9. Setup Graceful Shutdown
   └─ Listen for SIGTERM/SIGINT
```

### Request Lifecycle

```
HTTP Request
  ↓
Middleware Chain
  ├─ Logger (log request)
  ├─ CORS (add headers)
  └─ Auth (validate JWT if protected route)
  ↓
Route Matching
  ↓
Handler
  ├─ Parse request body
  ├─ Validate DTO
  └─ Call Service
  ↓
Service (Business Logic)
  ├─ Process data
  ├─ Call Repository
  └─ Sync to Backend (async)
  ↓
Repository
  ├─ Query Database (GORM)
  └─ Return Models
  ↓
Handler
  ├─ Convert Model → DTO
  └─ Return Response
  ↓
Middleware (Error Handler)
  └─ Catch any errors
  ↓
HTTP Response
```

---

## 🎯 Summary

| Layer | Directory | Responsibility |
|-------|-----------|----------------|
| **Entry Point** | `cmd/api/` | Start application |
| **Interfaces** | `interfaces/api/` | HTTP handlers, routes, middleware |
| **Application** | `application/` | Business logic implementations |
| **Domain** | `domain/` | Models, DTOs, interfaces (core) |
| **Infrastructure** | `infrastructure/` | Database, Redis, external services |
| **Shared** | `pkg/` | Config, DI, utilities |

**Key Principles:**
- ✅ Clean Architecture (dependency inversion)
- ✅ Dependency Injection (via DI Container)
- ✅ Separation of Concerns (layers)
- ✅ Interface-based design (easy to test)
- ✅ SOLID principles
- ✅ Graceful shutdown
- ✅ Error handling at every layer

**Technology Stack:**
- Language: Go 1.21
- Framework: Fiber v2
- Database: PostgreSQL 14
- Cache: Redis 7
- ORM: GORM
- Auth: JWT (HS256)
- Password: bcrypt

**Current Status:** ✅ Production-ready for Auth features
**Next Steps:** See `microservice_plan/` for future microservices architecture
