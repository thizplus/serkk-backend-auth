# Auth Service - Complete Integration Guide

**Version:** 2.0
**Last Updated:** 2025-11-24
**Service:** GoFiber Auth Microservice
**Event Schema:** V2 (Minimal Identity Event)

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Event Schema V2 - What Changed](#event-schema-v2---what-changed)
3. [Integration Methods](#integration-methods)
4. [HTTP API Integration](#http-api-integration)
5. [JWT Token Validation](#jwt-token-validation)
6. [Event-Driven Integration (NATS)](#event-driven-integration-nats)
7. [API Endpoints Reference](#api-endpoints-reference)
8. [Code Examples](#code-examples)
9. [Security Best Practices](#security-best-practices)
10. [Migration Guide](#migration-guide)
11. [FAQ](#faq)
12. [Troubleshooting](#troubleshooting)

---

## 🎯 Overview

Auth Service เป็น **Authentication & Authorization Microservice** ที่รับผิดชอบ:

- ✅ User registration & login (Email/Password, OAuth)
- ✅ JWT token generation & validation
- ✅ User identity management (id, email, username)
- ✅ OAuth integration (Google, Facebook, LINE)
- ✅ Event publishing สำหรับ user lifecycle events

**ไม่รับผิดชอบ:**
- ❌ User profiles (displayName, avatar, bio) → Social/Profile Service
- ❌ User permissions/roles → Internal only
- ❌ Business logic ของ services อื่น

---

## 🔄 Event Schema V2 - What Changed

### แนวคิด: Minimal Identity Event

Auth Service ปรับจาก **"Full User Data Event"** เป็น **"Minimal Identity Event"**

**หลักการ:**
- ✅ Auth Service ส่งเฉพาะ **identity data** (id, email, username)
- ❌ **ไม่ส่ง** profile data (displayName, avatar, bio)
- ❌ **ไม่ส่ง** authorization data (role, isActive, permissions)
- 🎯 Downstream services **enrich** ข้อมูล profile เอง

---

### Schema Comparison

#### Before (V1) - Full User Data ❌

```json
{
  "id": "uuid-here",
  "email": "user@example.com",
  "username": "john_doe",
  "displayName": "John Doe",          // ❌ ลบออก
  "avatar": "https://cdn.../pic.jpg", // ❌ ลบออก
  "role": "user",                     // ❌ ลบออก
  "isActive": true,                   // ❌ ลบออก
  "action": "created",

  "request_id": "uuid",
  "timestamp": "2025-11-24T...",
  "service_name": "gofiber-auth"
}
```

#### After (V2) - Minimal Identity Event ✅

```json
{
  // Minimal Identity Data
  "id": "uuid-here",
  "email": "user@example.com",
  "username": "john_doe",
  "action": "created",

  // Observability Metadata
  "request_id": "uuid",
  "timestamp": "2025-11-24T...",
  "service_name": "gofiber-auth",
  "sequence": 42  // NATS sequence number (optional)
}
```

**Payload Size Reduction:** ~50% 📉

---

### Fields Removed

| Field | Type | เหตุผลที่ลบ | ใครเป็นเจ้าของ |
|-------|------|-------------|----------------|
| `displayName` | string | Profile data | **Social/Profile Service** |
| `avatar` | string | Profile data | **Social/Profile Service** |
| `bio` | string | Profile data | **Social/Profile Service** |
| `role` | string | Authorization internal | **Auth Service internal** |
| `isActive` | boolean | Authorization internal | **Auth Service internal** |
| `permissions` | array | Authorization internal | **Auth Service internal** |

---

### Benefits of Minimal Identity Events

#### 1. **Separation of Concerns**
- Auth Service = Authentication & Authorization only
- Social Service = User profiles, social features
- ชัดเจน ไม่ overlap

#### 2. **Reduced Coupling**
- Event payload เล็กลง (~50% reduction)
- Auth Service ไม่ต้องรู้ว่า Social Service จัดการ profile ยังไง
- เปลี่ยน profile schema ได้โดยไม่กระทบ Auth Service

#### 3. **Scalability**
- Auth Service scale ได้อิสระจาก Social Service
- Profile data (avatar, bio) ไม่ซ้ำซ้อนระหว่าง services
- ลด event size → ลด bandwidth

#### 4. **Security**
- Auth Service ไม่ leak ข้อมูล role, isActive ออกไป
- Profile data (avatar URL) ไม่ถูก broadcast ผ่าน events
- Authorization logic เก็บไว้ใน Auth Service เท่านั้น

---

## 🔄 Integration Methods

Downstream services สามารถ integrate กับ Auth Service ได้ 2 วิธี:

### 1️⃣ **HTTP API Calls** (Synchronous)
- ใช้สำหรับ: Verify JWT tokens, get user info
- Protocol: REST API over HTTP/HTTPS
- Response: JSON

### 2️⃣ **Event Subscriptions** (Asynchronous)
- ใช้สำหรับ: User lifecycle events (created, updated, deleted)
- Protocol: NATS JetStream
- Event Schema: Minimal Identity Event (V2)

---

## 📡 HTTP API Integration

### Base URL
```
Development: http://localhost:8088
Production:  https://auth.yourdomain.com
```

### Required Headers
```http
Content-Type: application/json
Authorization: Bearer <jwt_token>  # สำหรับ protected endpoints
```

### CORS Configuration
Auth Service รองรับ CORS สำหรับ origins ต่อไปนี้:
```go
AllowOrigins:
  - http://localhost:3000      # Frontend Dev
  - http://localhost:8080      # Social Backend Dev
  - https://yourdomain.com     # Production
```

---

## 🔐 JWT Token Validation

### วิธีที่ 1: Call Auth Service API (Recommended for low traffic)

**Endpoint:** `GET /api/v1/users/me`

**Request:**
```bash
curl -X GET http://localhost:8088/api/v1/users/me \
  -H "Authorization: Bearer <jwt_token>"
```

**Response (Success):**
```json
{
  "id": "uuid-here",
  "email": "user@example.com",
  "username": "john_doe",
  "display_name": "John Doe",
  "avatar": "https://...",
  "role": "user",
  "is_active": true,
  "email_verified": true
}
```

**Response (Invalid Token):**
```json
{
  "error": "Invalid or expired token"
}
```

---

### วิธีที่ 2: Validate JWT Locally (Recommended for high traffic)

**ข้อดี:**
- ⚡ Faster - ไม่ต้องเรียก API
- 🔒 Secure - ใช้ shared secret
- 📉 Reduced load on Auth Service

**ข้อเสีย:**
- ต้อง share JWT_SECRET ระหว่าง services (ใช้ environment variable)
- ต้อง implement JWT validation logic เอง

**JWT Secret:**
```env
JWT_SECRET=your-jwt-secret-key-here-minimum-32-characters
```

**JWT Claims Structure:**
```json
{
  "user_id": "uuid-here",
  "email": "user@example.com",
  "username": "john_doe",
  "role": "user",
  "exp": 1732435200,
  "iat": 1732428000
}
```

**Example (Go):**
```go
import (
    "github.com/golang-jwt/jwt/v5"
    "os"
)

func validateToken(tokenString string) (*jwt.MapClaims, error) {
    secret := os.Getenv("JWT_SECRET")

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method")
        }
        return []byte(secret), nil
    })

    if err != nil || !token.Valid {
        return nil, err
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, fmt.Errorf("invalid claims")
    }

    return &claims, nil
}

// Usage
func AuthMiddleware(c *fiber.Ctx) error {
    authHeader := c.Get("Authorization")
    tokenString := strings.TrimPrefix(authHeader, "Bearer ")

    claims, err := validateToken(tokenString)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
    }

    c.Locals("user_id", (*claims)["user_id"])
    c.Locals("username", (*claims)["username"])

    return c.Next()
}
```

**Example (Node.js):**
```javascript
const jwt = require('jsonwebtoken');

function validateToken(token) {
  try {
    const decoded = jwt.verify(token, process.env.JWT_SECRET);
    return {
      userId: decoded.user_id,
      email: decoded.email,
      username: decoded.username,
      role: decoded.role
    };
  } catch (error) {
    throw new Error('Invalid token');
  }
}

// Express middleware
function authMiddleware(req, res, next) {
  const authHeader = req.headers.authorization;
  if (!authHeader) {
    return res.status(401).json({ error: 'No token provided' });
  }

  const token = authHeader.replace('Bearer ', '');

  try {
    const user = validateToken(token);
    req.user = user;
    next();
  } catch (error) {
    return res.status(401).json({ error: 'Unauthorized' });
  }
}
```

---

## 🎉 Event-Driven Integration (NATS)

### Why Use Events?

- ✅ **Decoupled** - Services ไม่ต้องรู้จักกัน
- ✅ **Scalable** - Handle ได้หลาก downstream services
- ✅ **Reliable** - NATS JetStream มี persistence
- ✅ **Real-time** - Update ทันทีที่มี user changes

### NATS Connection Setup

**Connection URL:**
```
nats://localhost:4222  # Development
```

**Stream Name:** `USER_EVENTS`
**Subject Pattern:** `user.events.*`

### Event Types

| Event Subject | Trigger | Description |
|--------------|---------|-------------|
| `user.events.created` | User registration (email/OAuth) | User ใหม่ถูกสร้างในระบบ |
| `user.events.updated` | User updates email/username | User แก้ไข identity data |
| `user.events.deleted` | User account deletion | User ถูกลบออกจากระบบ |

### Event Schema (V2)

```json
{
  "id": "uuid-here",
  "email": "user@example.com",
  "username": "john_doe",
  "action": "created",

  "request_id": "uuid-correlation-id",
  "timestamp": "2025-11-24T12:00:00Z",
  "service_name": "gofiber-auth"
}
```

**Fields:**
- `id` - User ID (Primary Key / Foreign Key)
- `email` - Email address (unique identifier)
- `username` - Username (unique identifier)
- `action` - Event type: `created`, `updated`, `deleted`
- `request_id` - Correlation ID สำหรับ distributed tracing
- `timestamp` - ISO 8601 timestamp
- `service_name` - Source service (always `"gofiber-auth"`)

---

## 📝 Code Examples

### Subscribe to User Events (Go)

```go
package main

import (
    "encoding/json"
    "log"
    "github.com/nats-io/nats.go"
)

type UserEvent struct {
    ID          string `json:"id"`
    Email       string `json:"email"`
    Username    string `json:"username"`
    Action      string `json:"action"`
    RequestID   string `json:"request_id"`
    Timestamp   string `json:"timestamp"`
    ServiceName string `json:"service_name"`
}

func main() {
    // Connect to NATS
    nc, err := nats.Connect("nats://localhost:4222")
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()

    // Get JetStream context
    js, err := nc.JetStream()
    if err != nil {
        log.Fatal(err)
    }

    // Subscribe to all user events
    _, err = js.Subscribe("user.events.*", func(msg *nats.Msg) {
        var event UserEvent
        if err := json.Unmarshal(msg.Data, &event); err != nil {
            log.Printf("Error unmarshaling: %v", err)
            msg.Nak()
            return
        }

        log.Printf("Received event: %s for user %s", event.Action, event.Username)

        // Process event based on action
        switch event.Action {
        case "created":
            handleUserCreated(event)
        case "updated":
            handleUserUpdated(event)
        case "deleted":
            handleUserDeleted(event)
        }

        msg.Ack()
    },
    nats.Durable("social-backend-consumer"),  // Consumer name
    nats.ManualAck())                         // Manual acknowledgment

    if err != nil {
        log.Fatal(err)
    }

    log.Println("Listening for user events...")
    select {} // Block forever
}

func handleUserCreated(event UserEvent) {
    // Insert into users_identity table
    db.Exec(`
        INSERT INTO users_identity (id, email, username, created_at)
        VALUES ($1, $2, $3, NOW())
        ON CONFLICT (id) DO NOTHING
    `, event.ID, event.Email, event.Username)

    // Create empty profile
    db.Exec(`
        INSERT INTO users_profile (id, display_name, created_at)
        VALUES ($1, $2, NOW())
        ON CONFLICT (id) DO NOTHING
    `, event.ID, event.Username)  // Use username as default display_name

    log.Printf("✅ User %s created in Social Service", event.Username)
}

func handleUserUpdated(event UserEvent) {
    // Update identity data only
    db.Exec(`
        UPDATE users_identity
        SET email = $1, username = $2, updated_at = NOW()
        WHERE id = $3
    `, event.Email, event.Username, event.ID)

    log.Printf("✅ User %s updated", event.Username)
}

func handleUserDeleted(event UserEvent) {
    // Soft delete
    db.Exec(`UPDATE users_profile SET deleted_at = NOW() WHERE id = $1`, event.ID)
    db.Exec(`UPDATE users_identity SET deleted_at = NOW() WHERE id = $1`, event.ID)

    log.Printf("✅ User %s deleted", event.Username)
}
```

---

### Subscribe to User Events (Node.js)

```javascript
const { connect, JSONCodec } = require('nats');

const jc = JSONCodec();

async function subscribeToUserEvents() {
  // Connect to NATS
  const nc = await connect({ servers: 'nats://localhost:4222' });
  console.log('✅ Connected to NATS');

  // Get JetStream client
  const js = nc.jetstream();

  // Subscribe to user events
  const sub = await js.subscribe('user.events.*', {
    config: {
      durable_name: 'social-backend-consumer',
      ack_policy: 'Explicit',
    },
  });

  console.log('📡 Listening for user events...');

  for await (const msg of sub) {
    const event = jc.decode(msg.data);
    console.log(`🔔 Received event: ${event.action} for user ${event.username}`);

    try {
      switch (event.action) {
        case 'created':
          await handleUserCreated(event);
          break;
        case 'updated':
          await handleUserUpdated(event);
          break;
        case 'deleted':
          await handleUserDeleted(event);
          break;
      }

      msg.ack();
    } catch (error) {
      console.error('❌ Error processing event:', error);
      msg.nak();
    }
  }
}

async function handleUserCreated(event) {
  // Insert into database
  await db.query(`
    INSERT INTO users_identity (id, email, username, created_at)
    VALUES ($1, $2, $3, NOW())
    ON CONFLICT (id) DO NOTHING
  `, [event.id, event.email, event.username]);

  await db.query(`
    INSERT INTO users_profile (id, display_name, created_at)
    VALUES ($1, $2, NOW())
    ON CONFLICT (id) DO NOTHING
  `, [event.id, event.username]);

  console.log(`✅ User ${event.username} created in Social Service`);
}

async function handleUserUpdated(event) {
  await db.query(`
    UPDATE users_identity
    SET email = $1, username = $2, updated_at = NOW()
    WHERE id = $3
  `, [event.email, event.username, event.id]);

  console.log(`✅ User ${event.username} updated`);
}

async function handleUserDeleted(event) {
  await db.query(`UPDATE users_profile SET deleted_at = NOW() WHERE id = $1`, [event.id]);
  await db.query(`UPDATE users_identity SET deleted_at = NOW() WHERE id = $1`, [event.id]);

  console.log(`✅ User ${event.username} deleted`);
}

subscribeToUserEvents().catch(console.error);
```

---

### Database Schema Recommendations

**แยกตาราง identity และ profile:**

```sql
-- Auth Service เป็นเจ้าของ (via events)
CREATE TABLE users_identity (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Social Service เป็นเจ้าของ
CREATE TABLE users_profile (
    id UUID PRIMARY KEY,  -- FK to users_identity.id
    display_name VARCHAR(100),
    avatar TEXT,
    bio TEXT,
    location VARCHAR(100),
    website VARCHAR(255),
    followers_count INTEGER DEFAULT 0,
    following_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (id) REFERENCES users_identity(id)
);
```

---

## 📚 API Endpoints Reference

### Public Endpoints (No Auth Required)

#### POST /api/v1/auth/register
Register new user with email/password

**Request:**
```json
{
  "email": "user@example.com",
  "username": "john_doe",
  "password": "SecurePassword123!",
  "display_name": "John Doe"
}
```

**Response:**
```json
{
  "user": {
    "id": "uuid-here",
    "email": "user@example.com",
    "username": "john_doe"
  },
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

---

#### POST /api/v1/auth/login
Login with email/password

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response:**
```json
{
  "user": { ... },
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

---

#### GET /api/v1/auth/google
Get Google OAuth URL

**Response:**
```json
{
  "url": "https://accounts.google.com/o/oauth2/v2/auth?..."
}
```

---

### Protected Endpoints (Auth Required)

#### GET /api/v1/users/me
Get current user info

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "id": "uuid-here",
  "email": "user@example.com",
  "username": "john_doe",
  "display_name": "John Doe",
  "avatar": "https://...",
  "role": "user",
  "is_active": true,
  "email_verified": true
}
```

---

#### PUT /api/v1/users/:id
Update user (email/username only)

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Request:**
```json
{
  "email": "newemail@example.com",
  "username": "new_username"
}
```

**Response:**
```json
{
  "user": { ... }
}
```

**Note:** จะ publish event `user.events.updated`

---

#### DELETE /api/v1/users/:id
Delete user account

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "message": "User deleted successfully"
}
```

**Note:** จะ publish event `user.events.deleted`

---

## 🔒 Security Best Practices

### 1. JWT Token Storage (Frontend)
```javascript
// ❌ Don't store in localStorage (XSS vulnerable)
localStorage.setItem('token', token);

// ✅ Store in httpOnly cookie (recommended)
// Backend sets: Set-Cookie: token=...; HttpOnly; Secure; SameSite=Strict
```

### 2. Token Expiration
- Default expiration: **24 hours**
- Refresh token flow: **Not implemented yet** (coming soon)
- For now: Re-login after token expires

### 3. HTTPS Only (Production)
```
Always use HTTPS in production to prevent token interception
```

### 4. Rate Limiting
Auth Service has built-in rate limiting:
- **Login:** 5 attempts per IP per 15 minutes
- **Register:** 3 attempts per IP per 15 minutes

### 5. CORS Configuration
Only allow trusted origins in production

---

## 🔄 Migration Guide

### Breaking Changes (V2)

สำหรับ Downstream Services ที่ integrate อยู่แล้ว:

**ต้องแก้ไข:**

1. **Event Handler Code**
   - ลบการ parse fields: `displayName`, `avatar`, `role`, `isActive`
   - ใช้แค่ `id`, `email`, `username`

2. **Database Schema**
   - แยกตาราง identity และ profile
   - ย้าย displayName, avatar ไปที่ users_profile

3. **Client-side Integration**
   - หลัง register สำเร็จ → เรียก API เพิ่ม profile
   - ไม่คาดหวังว่า event จะมี displayName

---

### Migration Path

**ขั้นตอนการ migrate:**

1. **Deploy new Schema V2** (Auth Service) ✅
2. **Update Subscribers** ให้รองรับ schema ใหม่
3. **Test กับ new registrations**
4. **Backfill existing users** (optional)

---

### Registration Flow Example

```
1. User กรอก form:
   - email: "john@example.com"
   - username: "john_doe"
   - password: "secret123"
   - displayName: "John Doe"  ← Auth Service ไม่เก็บใน event

2. Auth Service:
   - สร้าง user ใน database
   - Publish event:
     {
       "id": "uuid-123",
       "email": "john@example.com",
       "username": "john_doe",
       "action": "created"
     }
   - Return to client: user ID + JWT

3. Client (Frontend):
   - ได้รับ user ID จาก Auth Service
   - เรียก Social Service API:
     POST /api/users/{id}/profile
     {
       "displayName": "John Doe",
       "avatar": null,
       "bio": null
     }

4. Social Service:
   - รับ event จาก NATS → สร้าง users_identity
   - รับ API call จาก client → สร้าง users_profile
   - ✅ Complete!
```

---

## ❓ FAQ

### Q1: ผม register user ใหม่แล้ว displayName หายไปไหน?

**A:** Auth Service ไม่ส่ง displayName ใน event อีกต่อไป คุณต้อง:
1. รับ event → สร้าง users_identity (id, email, username)
2. รับ API call จาก client → สร้าง users_profile (displayName, avatar)

---

### Q2: ถ้า user update email/username ใน Auth Service จะเกิดอะไรขึ้น?

**A:** Auth Service จะส่ง `user.events.updated` พร้อม email/username ใหม่
- คุณต้อง update ตาราง `users_identity`
- **ไม่ต้อง** update `users_profile` (displayName, avatar ไม่เปลี่ยน)

---

### Q3: ผมจะเก็บ role, permissions ของ user ยังไง?

**A:** Auth Service เก็บ role/permissions เป็น internal data
- ถ้าต้องการ authorization → เรียก Auth Service API: `GET /api/users/{id}/permissions`
- **ไม่มี** ใน events เพราะเป็น security sensitive

---

### Q4: Avatar URL ควรเก็บที่ไหน?

**A:** Social/Profile Service เท่านั้น
- Upload avatar → Social Service upload ไป CDN
- เก็บ URL ใน `users_profile.avatar`
- Auth Service **ไม่รู้จัก** avatar เลย

---

### Q5: Username ถูกสร้างยังไง?

**A:** Auth Service สร้างอัตโนมัติสำหรับ OAuth users:
- Email prefix + 8 characters from UUID
- Example: `manage.karismarketing_32aeac4b`

---

## 🐛 Troubleshooting

### Issue: "Invalid or expired token"

**Causes:**
- Token expired (24h lifespan)
- Wrong JWT_SECRET
- Token tampered with

**Solutions:**
1. Check token expiration: `jwt.io` (paste token)
2. Verify JWT_SECRET matches between services
3. Re-login to get new token

---

### Issue: "NATS connection failed"

**Causes:**
- NATS server not running
- Wrong NATS URL
- Firewall blocking port 4222

**Solutions:**
```bash
# Check NATS is running
nats server list

# Start NATS (if not running)
nats-server -js

# Test connection
nats stream list
```

---

### Issue: "Events not received"

**Causes:**
- Consumer not subscribed correctly
- Wrong subject pattern
- Event not acknowledged (causing redelivery)

**Debug:**
```bash
# Check stream info
nats stream info USER_EVENTS

# Check consumer status
nats consumer ls USER_EVENTS

# Monitor events
nats sub "user.events.*"
```

---

### Issue: "User not found in users_cache"

**Causes:**
- Event handler not processing events
- Database insert failed
- Consumer not running

**Solutions:**
1. Check event handler logs
2. Verify NATS subscriber is running
3. Check database constraints
4. Monitor NATS stream: `nats stream info USER_EVENTS`

---

## 📞 Support

**Documentation:**
- This file contains everything you need
- For Auth Service code: See `README.md`

**Contact:**
- GitHub Issues: https://github.com/your-repo/gofiber-auth/issues
- Team Chat: Slack #auth-service

---

## 🔄 Version History

| Version | Date | Changes |
|---------|------|---------|
| 2.0 | 2025-11-24 | Merged all docs into single file, Event Schema V2 |
| 1.0 | 2025-11-24 | Initial microservice integration guide |

---

**Happy Integrating! 🚀**
