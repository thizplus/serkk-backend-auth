# Auth Service - Architecture & Project Structure

**Last Updated:** 2025-11-24
**Version:** 2.0
**Architecture:** Clean Architecture + Event-Driven Microservices

---

## 📋 Table of Contents

1. [Current Project Structure](#current-project-structure)
2. [Clean Architecture Layers](#clean-architecture-layers)
3. [Current State: Auth Microservice](#current-state-auth-microservice)
4. [Integration Points](#integration-points)
5. [Future Microservices Plan](#future-microservices-plan)
6. [Data Flow](#data-flow)
7. [Technology Stack](#technology-stack)

---

## 🗂️ Current Project Structure

```
gofiber-auth/
│
├── cmd/
│   ├── api/
│   │   └── main.go                    # 🚀 Application entry point
│   └── test_subscriber/
│       └── main.go                    # 🧪 NATS event subscriber (for testing)
│
├── domain/                            # 🏛️ BUSINESS LOGIC LAYER (CORE)
│   ├── models/                        # Entities
│   │   ├── user.go                    # User entity
│   │   └── oauth_provider.go         # OAuth provider entity
│   │
│   ├── dto/                           # Data Transfer Objects
│   │   ├── user_dto.go                # Request/Response DTOs
│   │   └── oauth_dto.go               # OAuth DTOs
│   │
│   ├── services/                      # Service interfaces (contracts)
│   │   ├── user_service.go            # User business logic interface
│   │   ├── oauth_service.go           # OAuth interface
│   │   └── event_publisher.go         # Event publishing interface
│   │
│   └── repositories/                  # Repository interfaces
│       ├── user_repository.go         # User data access interface
│       └── oauth_repository.go        # OAuth data access interface
│
├── application/                       # 📦 APPLICATION LAYER (USE CASES)
│   └── serviceimpl/                   # Service implementations
│       ├── user_service_impl.go       # User business logic
│       ├── oauth_service_impl.go      # OAuth logic (Google, Facebook, LINE)
│       └── sync_service.go            # User sync via events/HTTP
│
├── infrastructure/                    # 🔧 INFRASTRUCTURE LAYER (EXTERNAL)
│   ├── postgres/                      # Database implementations
│   │   ├── user_repository_impl.go    # User CRUD
│   │   └── oauth_repository_impl.go   # OAuth CRUD
│   │
│   └── nats/                          # Event Bus implementation
│       └── nats_publisher.go          # NATS JetStream publisher
│
├── interfaces/                        # 🌐 INTERFACE LAYER (HTTP/API)
│   └── api/
│       ├── handlers/                  # HTTP request handlers
│       │   ├── auth_handler.go        # Register, Login
│       │   ├── oauth_handler.go       # OAuth callbacks
│       │   ├── user_handler.go        # User CRUD
│       │   ├── health_handler.go      # Health check
│       │   └── metrics_handler.go     # Prometheus metrics
│       │
│       ├── middleware/                # HTTP middleware
│       │   ├── auth_middleware.go     # JWT validation
│       │   ├── cors_middleware.go     # CORS policy
│       │   ├── logger_middleware.go   # Request logging
│       │   ├── metrics_middleware.go  # Metrics collection
│       │   └── request_id_middleware.go # Request ID tracking
│       │
│       └── routes/                    # Route definitions
│           └── routes.go              # API routes setup
│
├── pkg/                               # 📚 SHARED UTILITIES
│   ├── config/                        # Configuration
│   │   ├── config.go                  # Config loader
│   │   └── database.go                # DB connection
│   │
│   ├── di/                            # Dependency Injection
│   │   └── container.go               # DI container
│   │
│   ├── logger/                        # Structured logging
│   │   └── logger.go                  # JSON logger
│   │
│   ├── metrics/                       # Prometheus metrics
│   │   └── metrics.go                 # Metrics definitions
│   │
│   ├── contextutil/                   # Context helpers
│   │   └── context.go                 # Request ID helpers
│   │
│   └── auth_code_store/               # OAuth state store
│       └── store.go                   # In-memory store
│
├── .env.example                       # Environment variables template
├── go.mod                             # Go modules
├── go.sum                             # Go dependencies
│
├── README.md                          # Project overview
└── SERVICE_INTEGRATION.md             # Complete integration guide
```

---

## 🏗️ Clean Architecture Layers

Auth Service ใช้ **Clean Architecture** (Uncle Bob) แบ่งเป็น 4 ชั้น:

```
┌─────────────────────────────────────────────────────────────┐
│  INTERFACES LAYER (HTTP/API)                                │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Handlers, Middleware, Routes                        │   │
│  │  - รับ HTTP requests                                 │   │
│  │  - แปลง request → DTO                                │   │
│  │  - เรียก Application Layer                           │   │
│  │  - ส่ง HTTP response                                 │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            ↓ Dependency
┌─────────────────────────────────────────────────────────────┐
│  APPLICATION LAYER (USE CASES)                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Service Implementations                             │   │
│  │  - User registration logic                           │   │
│  │  - OAuth flow logic                                  │   │
│  │  - JWT generation                                    │   │
│  │  - Event publishing                                  │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            ↓ Dependency
┌─────────────────────────────────────────────────────────────┐
│  DOMAIN LAYER (BUSINESS LOGIC - CORE)                       │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Entities, DTOs, Interfaces                          │   │
│  │  - User, OAuthProvider models                        │   │
│  │  - Service interfaces (contracts)                    │   │
│  │  - Repository interfaces                             │   │
│  │  - ไม่มี dependencies ภายนอก (PURE)                  │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            ↑ Implements
┌─────────────────────────────────────────────────────────────┐
│  INFRASTRUCTURE LAYER (EXTERNAL)                            │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Database, NATS, External APIs                       │   │
│  │  - PostgreSQL repository implementations             │   │
│  │  - NATS publisher implementation                     │   │
│  │  - External service integrations                     │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Dependency Rule

**สิ่งสำคัญ:** Dependencies ชี้เข้าหา **Domain Layer** เท่านั้น

```
Infrastructure → Domain ← Application ← Interfaces
                  ↑
            (Core/Center)
```

**ข้อดี:**
- ✅ Domain ไม่ depend on ใคร (testable)
- ✅ เปลี่ยน database ได้โดยไม่กระทบ business logic
- ✅ เปลี่ยน framework ได้โดยไม่กระทบ core
- ✅ Easy to test (mock interfaces)

---

## 🎯 Current State: Auth Microservice

### สถานะปัจจุบัน

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│               🎯 AUTH SERVICE (Microservice)                │
│                                                             │
│  Responsibilities:                                          │
│  ✅ User Registration (Email/Password)                      │
│  ✅ User Authentication (Login)                             │
│  ✅ OAuth Integration (Google, Facebook, LINE)              │
│  ✅ JWT Token Generation & Validation                       │
│  ✅ User Identity Management (id, email, username)          │
│  ✅ Event Publishing (user.events.*)                        │
│                                                             │
│  Technology:                                                │
│  - GoFiber (HTTP Framework)                                 │
│  - PostgreSQL (Database)                                    │
│  - NATS JetStream (Event Bus)                               │
│  - JWT (Authentication)                                     │
│  - Prometheus (Metrics)                                     │
│                                                             │
│  Ports:                                                     │
│  - HTTP: 8088                                               │
│  - Metrics: 8088/metrics                                    │
│  - NATS: 4222                                               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### ไม่รับผิดชอบ (Out of Scope)

```
❌ User Profiles (displayName, avatar, bio)
   → Social/Profile Service

❌ User Permissions/Roles Management
   → Internal only, ไม่ expose ใน events

❌ Social Features (posts, comments, follows)
   → Social Service

❌ Business Logic อื่นๆ
   → Respective services
```

---

## 🔌 Integration Points

### 1. HTTP API (Synchronous)

```
┌─────────────┐                    ┌─────────────┐
│             │   HTTP Request     │             │
│   Client    ├───────────────────→│    Auth     │
│ (Frontend)  │                    │   Service   │
│             │←───────────────────┤             │
│             │   HTTP Response    │             │
└─────────────┘   (JWT Token)      └─────────────┘
```

**Endpoints:**
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `GET /api/v1/auth/google` - OAuth URL
- `GET /api/v1/users/me` - Get current user

---

### 2. Event-Driven (Asynchronous)

```
┌─────────────┐                    ┌─────────────┐
│             │   User Event       │             │
│    Auth     ├───────────────────→│    NATS     │
│   Service   │   (Publish)        │  JetStream  │
│             │                    │             │
└─────────────┘                    └──────┬──────┘
                                          │
                                          │ Subscribe
                                          ↓
                               ┌──────────────────────┐
                               │                      │
                               │   Social Service     │
                               │   (Subscriber)       │
                               │                      │
                               │   - users_identity   │
                               │   - users_profile    │
                               │                      │
                               └──────────────────────┘
```

**Event Topics:**
- `user.events.created` - New user registered
- `user.events.updated` - User updated email/username
- `user.events.deleted` - User deleted

**Event Payload (V2 - Minimal Identity):**
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "username": "john_doe",
  "action": "created"
}
```

---

### 3. JWT Validation (Inter-Service)

```
┌─────────────┐                    ┌─────────────┐
│             │   JWT Token        │             │
│   Social    ├───────────────────→│    Auth     │
│   Service   │   Validate         │   Service   │
│             │                    │             │
│             │←───────────────────┤             │
│             │   User Info        │             │
└─────────────┘                    └─────────────┘

หรือ

┌─────────────┐
│             │   JWT Secret
│   Social    │   (Shared)
│   Service   │   Validate Locally
│             │   (Faster)
│             │
└─────────────┘
```

---

## 🚀 Future Microservices Plan

### ขณะนี้: Monolithic → Microservices (Stage 1)

```
BEFORE:
┌─────────────────────────────────────┐
│                                     │
│      Monolithic Application         │
│                                     │
│  - Auth                             │
│  - User Profiles                    │
│  - Social Features                  │
│  - Posts                            │
│  - Comments                         │
│  - Notifications                    │
│                                     │
└─────────────────────────────────────┘


NOW (Stage 1):
┌───────────────┐         ┌────────────────────────┐
│               │         │                        │
│  Auth Service │         │  Social Monolith       │
│  (Extracted)  │         │                        │
│               │         │  - User Profiles       │
│  - Register   │         │  - Social Features     │
│  - Login      │  Events │  - Posts               │
│  - OAuth      │────────→│  - Comments            │
│  - JWT        │  NATS   │  - Notifications       │
│               │         │                        │
└───────────────┘         └────────────────────────┘
     Port 8088                  Port 8080
```

---

### แผนอนาคต: Stage 2-4

#### Stage 2: Extract Profile Service

```
┌───────────────┐         ┌────────────────────┐
│  Auth Service │         │  Profile Service   │
│               │  Events │                    │
│  - Register   │────────→│  - Display Name    │
│  - Login      │  NATS   │  - Avatar          │
│  - OAuth      │         │  - Bio             │
│  - JWT        │         │  - Settings        │
└───────────────┘         └────────────────────┘
                                   │
                                   │ HTTP/Events
                                   ↓
                          ┌─────────────────────┐
                          │  Social Monolith    │
                          │                     │
                          │  - Posts            │
                          │  - Comments         │
                          │  - Follows          │
                          │  - Notifications    │
                          └─────────────────────┘
```

---

#### Stage 3: Extract Social Features

```
┌───────────────┐    ┌────────────────┐    ┌─────────────────┐
│  Auth Service │    │Profile Service │    │  Social Service │
│               │    │                │    │                 │
│  - Register   │    │  - User Profile│    │  - Posts        │
│  - Login      │    │  - Avatar      │    │  - Comments     │
│  - OAuth      │    │  - Bio         │    │  - Likes        │
│  - JWT        │    │  - Settings    │    │  - Shares       │
└───────────────┘    └────────────────┘    └─────────────────┘
        │                    │                      │
        └────────────────────┴──────────────────────┘
                            │
                      NATS JetStream
                            │
        ┌───────────────────┴───────────────────┐
        │                                       │
        ↓                                       ↓
┌───────────────────┐                  ┌──────────────────┐
│  Follow Service   │                  │Notification Svc  │
│                   │                  │                  │
│  - Follow/Unfollow│                  │  - Push Notifs   │
│  - Followers      │                  │  - Email         │
│  - Following      │                  │  - WebSocket     │
└───────────────────┘                  └──────────────────┘
```

---

#### Stage 4: Full Microservices Architecture

```
                        API Gateway / BFF
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
        ↓                      ↓                      ↓
┌───────────────┐    ┌────────────────┐    ┌─────────────────┐
│  Auth Service │    │Profile Service │    │  Social Service │
└───────┬───────┘    └────────┬───────┘    └────────┬────────┘
        │                     │                      │
        │                     │                      │
        └─────────────────────┴──────────────────────┘
                              │
                        NATS JetStream
                       (Event Bus)
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ↓                     ↓                     ↓
┌───────────────┐   ┌──────────────────┐  ┌────────────────┐
│Follow Service │   │Notification Svc  │  │  Media Service │
│               │   │                  │  │                │
│- Follow/Unfl  │   │- Push Notifs     │  │- Upload        │
│- Followers    │   │- Email           │  │- CDN           │
│- Following    │   │- WebSocket       │  │- Resize        │
└───────────────┘   └──────────────────┘  └────────────────┘
        │                     │                     │
        └─────────────────────┴─────────────────────┘
                              │
                     Shared Infrastructure
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ↓                     ↓                     ↓
    PostgreSQL            Redis Cache          Prometheus
    (Per Service)        (Shared/Per Svc)      (Metrics)
```

---

### Service Candidates (แยกได้ในอนาคต)

| Service | Priority | Complexity | Dependencies | Benefit |
|---------|----------|------------|--------------|---------|
| **Auth Service** | ✅ Done | Medium | None | High - Security isolation |
| **Profile Service** | High | Low | Auth | Medium - Independent scaling |
| **Social Service** | High | Medium | Auth, Profile | High - Core feature |
| **Follow Service** | Medium | Low | Auth | Medium - Can scale independently |
| **Notification Service** | Medium | High | All | High - Push/Email/WebSocket |
| **Media Service** | Low | High | Auth | Medium - Upload/CDN/Resize |
| **Search Service** | Low | High | All | Medium - Elasticsearch |
| **Analytics Service** | Low | Medium | All | Low - Business intelligence |

---

## 🔄 Data Flow

### User Registration Flow

```
┌─────────┐
│  Client │
└────┬────┘
     │ 1. POST /api/v1/auth/register
     │    { email, username, password, displayName }
     ↓
┌──────────────────┐
│  Auth Service    │
│  (Port 8088)     │
└────┬─────────────┘
     │ 2. Create user in DB
     │    INSERT INTO users (id, email, username, ...)
     │
     │ 3. Publish event to NATS
     │    Topic: user.events.created
     │    Payload: { id, email, username }
     ↓
┌──────────────────┐
│  NATS JetStream  │
│  (Port 4222)     │
└────┬─────────────┘
     │ 4. Deliver event to subscribers
     ↓
┌───────────────────────┐
│  Social Service       │
│  (Subscriber)         │
└───┬───────────────────┘
    │ 5a. INSERT INTO users_identity
    │     (id, email, username)
    │
    │ 5b. INSERT INTO users_profile
    │     (id, display_name) -- from client API call
    │
    ↓
  ✅ Complete!
```

---

### JWT Validation Flow

```
┌─────────┐
│  Client │
└────┬────┘
     │ 1. Request with JWT
     │    Authorization: Bearer <token>
     ↓
┌──────────────────┐
│  Social Service  │
└────┬─────────────┘
     │
     │ Option 1: Call Auth Service
     │ GET http://localhost:8088/api/v1/users/me
     ↓
┌──────────────────┐
│  Auth Service    │
│                  │
│  Validate JWT    │
│  Return user     │
└──────────────────┘

     OR

┌──────────────────┐
│  Social Service  │
│                  │
│  Option 2:       │
│  Validate JWT    │
│  Locally         │
│  (JWT_SECRET)    │
│  Faster!         │
└──────────────────┘
```

---

## 🛠️ Technology Stack

### Core Technologies

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Language** | Go 1.21+ | High performance, concurrency |
| **Framework** | GoFiber v2 | Fast HTTP framework |
| **Database** | PostgreSQL 14+ | Relational data storage |
| **ORM** | GORM | Database abstraction |
| **Event Bus** | NATS JetStream | Event-driven messaging |
| **Cache** | Redis (planned) | Session, rate limiting |
| **Metrics** | Prometheus | Monitoring |
| **Logging** | Structured JSON | Observability |

---

### Libraries

```go
// HTTP Framework
github.com/gofiber/fiber/v2

// Database
gorm.io/gorm
gorm.io/driver/postgres

// Event Bus
github.com/nats-io/nats.go

// Authentication
github.com/golang-jwt/jwt/v5
golang.org/x/crypto/bcrypt

// OAuth
golang.org/x/oauth2
google.golang.org/api/oauth2/v2

// Validation
github.com/go-playground/validator/v10

// Monitoring
github.com/prometheus/client_golang
```

---

## 📊 Database Schema

### Current (Auth Service)

```sql
-- Users table (Auth Service owns)
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255),              -- NULL for OAuth users
    display_name VARCHAR(100),          -- TODO: Remove in V3
    avatar TEXT,                        -- TODO: Remove in V3
    role VARCHAR(50) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    is_oauth_user BOOLEAN DEFAULT false,
    oauth_provider VARCHAR(50),
    oauth_id VARCHAR(255),
    email_verified BOOLEAN DEFAULT false,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- OAuth providers
CREATE TABLE oauth_providers (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    provider VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMP,
    profile_data JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(provider, provider_id)
);
```

---

### Recommended (Downstream Services)

```sql
-- Social Service owns

-- Identity data (from Auth events)
CREATE TABLE users_identity (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Profile data (Social Service manages)
CREATE TABLE users_profile (
    id UUID PRIMARY KEY REFERENCES users_identity(id),
    display_name VARCHAR(100),
    avatar TEXT,
    bio TEXT,
    location VARCHAR(100),
    website VARCHAR(255),
    followers_count INTEGER DEFAULT 0,
    following_count INTEGER DEFAULT 0,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

---

## 🎯 Design Principles

### 1. **Single Responsibility**
- Auth Service = Authentication & Authorization only
- Profile data → Profile Service (future)
- Social features → Social Service

### 2. **Event-Driven Architecture**
- Services communicate via events (NATS)
- Loose coupling
- Async processing

### 3. **API-First Design**
- REST API for synchronous operations
- Events for asynchronous updates
- Clear API contracts

### 4. **Clean Architecture**
- Domain-centric design
- Dependencies point inward
- Framework-agnostic core

### 5. **Observability**
- Request ID tracking
- Structured logging
- Prometheus metrics
- Distributed tracing ready

---

## 📈 Scalability Considerations

### Current Bottlenecks

1. **Database** - PostgreSQL single instance
   - **Solution:** Read replicas, Connection pooling

2. **NATS** - Single instance
   - **Solution:** NATS clustering (future)

3. **JWT Validation** - Call Auth Service every time
   - **Solution:** Validate locally with shared secret

---

### Horizontal Scaling

```
                    Load Balancer
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ↓                 ↓                 ↓
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ Auth Service │  │ Auth Service │  │ Auth Service │
│  Instance 1  │  │  Instance 2  │  │  Instance 3  │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │
       └─────────────────┴─────────────────┘
                         │
                         ↓
                  PostgreSQL
                  (Shared DB)
```

**Note:** Auth Service is stateless → easy to scale horizontally

---

## 🔐 Security Architecture

```
┌─────────────────────────────────────────┐
│  API Gateway / Firewall                 │
│  - Rate limiting                        │
│  - DDoS protection                      │
└─────────────────┬───────────────────────┘
                  │
                  ↓
┌─────────────────────────────────────────┐
│  Auth Service                           │
│  - JWT generation                       │
│  - Password hashing (bcrypt)            │
│  - OAuth 2.0                            │
│  - CORS policy                          │
└─────────────────┬───────────────────────┘
                  │
                  ↓
┌─────────────────────────────────────────┐
│  Database (PostgreSQL)                  │
│  - Encrypted connections (SSL)          │
│  - Row-level security (future)          │
│  - Audit logs                           │
└─────────────────────────────────────────┘
```

---

## 📝 Next Steps

### Short Term (1-2 months)
- [ ] Add Redis for session management
- [ ] Implement refresh token flow
- [ ] Add rate limiting per user
- [ ] Setup monitoring dashboard (Grafana)
- [ ] Load testing

### Medium Term (3-6 months)
- [ ] Extract Profile Service
- [ ] Implement API Gateway
- [ ] Setup CI/CD pipeline
- [ ] Kubernetes deployment
- [ ] Database read replicas

### Long Term (6-12 months)
- [ ] Extract Social Service
- [ ] Extract Notification Service
- [ ] Implement service mesh (Istio/Linkerd)
- [ ] Multi-region deployment
- [ ] Advanced analytics

---

## 📚 Related Documentation

- [SERVICE_INTEGRATION.md](./SERVICE_INTEGRATION.md) - Complete integration guide
- [README.md](./README.md) - Quick start guide
- [.env.example](./.env.example) - Environment variables

---

**Architecture Version:** 2.0
**Last Review:** 2025-11-24
**Next Review:** 2025-12-24

---

**สรุป:** Auth Service ตอนนี้เป็น **Microservice** แบบ Clean Architecture + Event-Driven ที่พร้อมสำหรับการ scale และแยก services เพิ่มเติมในอนาคต! 🚀
