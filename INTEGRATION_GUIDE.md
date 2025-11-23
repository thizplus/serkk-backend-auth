# Auth Service Integration Guide

**สำหรับ: Social Monolith Backend Team**
**วันที่: 2025-11-24**
**Version: 1.0**

---

## 📋 สารบัญ

1. [ภาพรวม Architecture](#ภาพรวม-architecture)
2. [Communication Methods](#communication-methods)
3. [NATS Events Integration](#nats-events-integration)
4. [Event Schema](#event-schema)
5. [การติดตั้งและ Setup](#การติดตั้งและ-setup)
6. [Code Examples](#code-examples)
7. [Testing](#testing)
8. [Troubleshooting](#troubleshooting)

---

## 📐 ภาพรวม Architecture

### Current State

```
┌─────────────────────┐          ┌──────────────────────┐
│   Auth Service      │          │  Social Monolith     │
│   (Microservice)    │          │  (Backend)           │
│                     │          │                      │
│  - User Register    │          │  - User Profiles     │
│  - User Login       │          │  - Posts & Comments  │
│  - OAuth (G/FB/L)   │          │  - Friends & Chat    │
│  - JWT Generation   │          │  - Notifications     │
└─────────────────────┘          └──────────────────────┘
         │                                  ▲
         │        NATS JetStream            │
         └──────────► Events ───────────────┘
              (user.events.*)
```

### คำอธิบาย

- **Auth Service** = Microservice แยกออกมาดูแลเฉพาะ Authentication & Authorization
- **Social Monolith** = Backend เดิมที่ดูแล User Profiles, Social Features
- **NATS JetStream** = Message broker สำหรับ Event-Driven Communication

---

## 🔌 Communication Methods

### 1. Event-Driven (Primary) ✅ แนะนำ

**ใช้ NATS JetStream** เป็นหลักสำหรับ sync ข้อมูล user

**ข้อดี:**
- ✅ **Async & Non-blocking** - ไม่ทำให้ Auth Service ช้า
- ✅ **Reliable** - JetStream เก็บ message persistence
- ✅ **Decoupled** - Services ไม่ต้องรู้จักกันโดยตรง
- ✅ **Scalable** - รองรับ multiple subscribers

**ข้อเสีย:**
- ❌ ต้องติดตั้ง NATS Server
- ❌ เพิ่ม complexity

### 2. HTTP API (Fallback) ⚠️ สำรอง

Auth Service ยังมี **HTTP sync** สำรอง (ถ้า NATS ล่ม)

**การตั้งค่า:**
```env
USE_EVENT_SYNC=false          # ปิดใช้ Events, ใช้ HTTP
BACKEND_SYNC_URL=http://your-backend:3000/api/sync/users
```

> **หมายเหตุ:** แนะนำให้ใช้ **Events** เป็นหลัก HTTP เป็น fallback เท่านั้น

---

## 📡 NATS Events Integration

### Event Types

Auth Service publish 3 events หลัก:

| Event Topic | เมื่อไหร่ | Payload |
|-------------|----------|---------|
| `user.events.created` | มี user register ใหม่ | User data + metadata |
| `user.events.updated` | User update profile | Updated user data |
| `user.events.deleted` | User ถูกลบ | User ID + metadata |

### NATS Configuration

**Server:** `nats://localhost:4222` (ถ้าติดตั้งแยก server ให้แก้ URL)
**Stream Name:** `USER_EVENTS`
**Subject Pattern:** `user.events.*`
**Storage:** File Storage (Persistent)
**Retention:** WorkQueue (ลบหลัง ack)

---

## 📦 Event Schema

### Minimal Identity Event Design

Auth Service ส่ง **minimal identity event** เท่านั้น โดย:
- ✅ ส่งเฉพาะ **identity data** (id, email, username)
- ❌ **ไม่ส่ง** profile data (displayName, avatar, bio, etc.)
- 🎯 **Downstream services** (Social Monolith) เป็นผู้ enrich ข้อมูล profile

### User Event Payload Structure

```json
{
  // === Minimal Identity Data ===
  "id": "4017047b-9360-491a-bcc6-8bc9a91d9086",
  "email": "user@example.com",
  "username": "john_doe",
  "action": "created",  // "created" | "updated" | "deleted"

  // === Observability Metadata ===
  "request_id": "uuid-request-id-here",      // Distributed tracing
  "timestamp": "2025-11-24T02:27:50Z",       // ISO 8601 format
  "service_name": "gofiber-auth",            // Source service
  "sequence": 2                               // NATS sequence (optional)
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string (UUID) | ✅ | User ID (Primary Key) |
| `email` | string | ✅ | Email address (identifier) |
| `username` | string | ✅ | Username (identifier) |
| `action` | string | ✅ | "created" \| "updated" \| "deleted" |
| `request_id` | string (UUID) | ✅ | Correlation ID for distributed tracing |
| `timestamp` | string | ✅ | Event timestamp (ISO 8601) |
| `service_name` | string | ✅ | "gofiber-auth" |
| `sequence` | integer | ❌ | NATS JetStream sequence number |

### Fields Removed (Managed by Downstream Services)

| Field | Reason | Who Manages |
|-------|--------|-------------|
| `displayName` | Profile data | Social/Profile Service |
| `avatar` | Profile data | Social/Profile Service |
| `bio` | Profile data | Social/Profile Service |
| `role` | Auth internal | Auth Service only |
| `isActive` | Auth internal | Auth Service only |
| `permissions` | Auth internal | Auth Service only |

> **หมายเหตุ:** Social Monolith ควรเก็บ user profile แยกจาก identity data และใช้ `id` เป็น foreign key

---

## 🛠 การติดตั้งและ Setup

### Step 1: ติดตั้ง NATS Server

#### Option A: Docker (แนะนำ)

```bash
docker run -d --name nats-server \
  -p 4222:4222 \
  -p 8222:8222 \
  nats:latest -js
```

#### Option B: ติดตั้งแบบ Standalone

```bash
# Download NATS Server
# https://github.com/nats-io/nats-server/releases

# Run with JetStream enabled
./nats-server -js
```

**ตรวจสอบ:**
```bash
curl http://localhost:8222/healthz
# ควรได้ "OK"
```

---

### Step 2: ติดตั้ง NATS Client Library

เลือกตาม tech stack ของคุณ:

#### Go
```bash
go get github.com/nats-io/nats.go
```

#### Node.js
```bash
npm install nats
```

#### Python
```bash
pip install nats-py
```

---

### Step 3: Subscribe to Events

เลือก code example ตามภาษาที่ใช้:

---

## 💻 Code Examples

### Example 1: Go Subscriber

```go
package main

import (
    "encoding/json"
    "log"
    "github.com/nats-io/nats.go"
)

// UserEvent matches Auth Service minimal identity event payload
type UserEvent struct {
    // Minimal Identity Data
    ID       string `json:"id"`
    Email    string `json:"email"`
    Username string `json:"username"`
    Action   string `json:"action"`

    // Observability Metadata
    RequestID   string `json:"request_id"`
    Timestamp   string `json:"timestamp"`
    ServiceName string `json:"service_name"`
    Sequence    uint64 `json:"sequence,omitempty"`
}

func main() {
    // Connect to NATS
    nc, err := nats.Connect("nats://localhost:4222")
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()

    // Create JetStream context
    js, err := nc.JetStream()
    if err != nil {
        log.Fatal(err)
    }

    // Subscribe to all user events
    _, err = js.Subscribe("user.events.*", func(msg *nats.Msg) {
        var event UserEvent
        if err := json.Unmarshal(msg.Data, &event); err != nil {
            log.Printf("Error unmarshaling: %v", err)
            return
        }

        // Process event based on action
        switch event.Action {
        case "created":
            handleUserCreated(event)
        case "updated":
            handleUserUpdated(event)
        case "deleted":
            handleUserDeleted(event)
        }

        // Acknowledge message
        msg.Ack()
    },
        nats.Durable("social-backend-consumer"), // Durable name (สำคัญ!)
        nats.ManualAck(),                        // Manual ack
    )

    if err != nil {
        log.Fatal(err)
    }

    log.Println("✅ Subscribed to user.events.*")
    select {} // Keep running
}

func handleUserCreated(event UserEvent) {
    log.Printf("🆕 New user: %s (%s)", event.Username, event.Email)

    // TODO: บันทึกลง database ของคุณ
    // db.CreateUserProfile(event)
}

func handleUserUpdated(event UserEvent) {
    log.Printf("♻️  Updated user: %s", event.Username)

    // TODO: Update user profile
    // db.UpdateUserProfile(event.ID, event)
}

func handleUserDeleted(event UserEvent) {
    log.Printf("🗑️  Deleted user: %s", event.Username)

    // TODO: Soft delete or archive
    // db.SoftDeleteUser(event.ID)
}
```

---

### Example 2: Node.js Subscriber

```javascript
const { connect, StringCodec } = require('nats');

// UserEvent interface (TypeScript) - Minimal Identity Event
interface UserEvent {
    // Minimal Identity Data
    id: string;
    email: string;
    username: string;
    action: 'created' | 'updated' | 'deleted';

    // Observability Metadata
    request_id: string;
    timestamp: string;
    service_name: string;
    sequence?: number;
}

async function main() {
    // Connect to NATS
    const nc = await connect({
        servers: 'nats://localhost:4222'
    });

    console.log('✅ Connected to NATS');

    // Create JetStream client
    const js = nc.jetstream();
    const sc = StringCodec();

    // Subscribe to user events
    const consumer = await js.consumers.get('USER_EVENTS', 'social-backend-consumer');

    const messages = await consumer.consume();

    for await (const msg of messages) {
        try {
            const event: UserEvent = JSON.parse(sc.decode(msg.data));

            console.log(`📨 Received: ${event.action} - ${event.username}`);

            // Process based on action
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

            // Acknowledge
            msg.ack();
        } catch (error) {
            console.error('Error processing message:', error);
            msg.nak(); // Negative ack (retry)
        }
    }
}

async function handleUserCreated(event: UserEvent) {
    console.log(`🆕 Creating user profile: ${event.username}`);
    // TODO: Insert into your database
    // await db.users.create({ ... });
}

async function handleUserUpdated(event: UserEvent) {
    console.log(`♻️  Updating user profile: ${event.username}`);
    // TODO: Update in database
    // await db.users.update({ id: event.id }, { ... });
}

async function handleUserDeleted(event: UserEvent) {
    console.log(`🗑️  Deleting user: ${event.username}`);
    // TODO: Soft delete
    // await db.users.update({ id: event.id }, { deletedAt: new Date() });
}

main().catch(console.error);
```

---

### Example 3: Python Subscriber

```python
import asyncio
import json
from nats.aio.client import Client as NATS
from nats.js import JetStreamContext

async def message_handler(msg):
    """Handle incoming user events"""
    try:
        # Parse event
        event = json.loads(msg.data.decode())

        action = event.get('action')
        username = event.get('username')

        print(f"📨 Received: {action} - {username}")

        # Process based on action
        if action == 'created':
            await handle_user_created(event)
        elif action == 'updated':
            await handle_user_updated(event)
        elif action == 'deleted':
            await handle_user_deleted(event)

        # Acknowledge
        await msg.ack()

    except Exception as e:
        print(f"Error processing message: {e}")
        await msg.nak()  # Negative ack (retry)

async def handle_user_created(event):
    print(f"🆕 Creating user: {event['username']}")
    # TODO: Insert into database
    # db.users.insert_one(event)

async def handle_user_updated(event):
    print(f"♻️  Updating user: {event['username']}")
    # TODO: Update database
    # db.users.update_one({'id': event['id']}, {'$set': event})

async def handle_user_deleted(event):
    print(f"🗑️  Deleting user: {event['username']}")
    # TODO: Soft delete
    # db.users.update_one({'id': event['id']}, {'$set': {'deletedAt': datetime.now()}})

async def main():
    # Connect to NATS
    nc = NATS()
    await nc.connect(servers=["nats://localhost:4222"])

    print("✅ Connected to NATS")

    # Create JetStream context
    js: JetStreamContext = nc.jetstream()

    # Subscribe to user events
    await js.subscribe(
        subject="user.events.*",
        cb=message_handler,
        durable="social-backend-consumer",
        manual_ack=True
    )

    print("👂 Listening for user events...")

    # Keep running
    while True:
        await asyncio.sleep(1)

if __name__ == '__main__':
    asyncio.run(main())
```

---

## 🧪 Testing

### 1. ทดสอบการรับ Events

**ขั้นตอน:**

1. Start NATS Server
2. Start Social Monolith Subscriber
3. Register user ใหม่ใน Auth Service:

```bash
curl -X POST http://localhost:8088/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "Password123",
    "displayName": "Test User"
  }'
```

4. ตรวจสอบ log ของ Subscriber - ควรเห็น event `user.events.created`

---

### 2. ตรวจสอบ NATS Stream

**ดู stream info:**
```bash
nats stream info USER_EVENTS
```

**ดูจำนวน messages:**
```bash
nats stream view USER_EVENTS
```

**List consumers:**
```bash
nats consumer list USER_EVENTS
```

---

### 3. ทดสอบด้วย Test Subscriber

Auth Service มี test subscriber พร้อมใช้งาน:

```bash
cd gofiber-auth
go run cmd/test_subscriber/main.go
```

Output ตัวอย่าง:
```
✅ Connected to NATS
✅ JetStream context created
📊 Stream: USER_EVENTS
📊 Messages: 2
👂 Listening for events on user.events.*

🔔 Received event:
Subject: user.events.created
Payload:
{
  "id": "4017047b-9360-491a-bcc6-8bc9a91d9086",
  "email": "test@example.com",
  "username": "testuser",
  ...
}
✅ Message acknowledged
```

---

## 🚨 Troubleshooting

### Problem 1: ไม่ได้รับ Events

**สาเหตุที่เป็นไปได้:**

1. **NATS Server ไม่ทำงาน**
   ```bash
   # Check NATS is running
   curl http://localhost:8222/healthz
   ```

2. **Subject pattern ไม่ตรง**
   ```go
   // ❌ Wrong
   js.Subscribe("user.events", ...)

   // ✅ Correct
   js.Subscribe("user.events.*", ...)
   ```

3. **Stream ไม่มี**
   ```bash
   nats stream ls
   # ถ้าไม่มี USER_EVENTS ให้ restart Auth Service
   ```

---

### Problem 2: Duplicate Events

**สาเหตุ:** ไม่ได้ใช้ `Durable` consumer name

**แก้ไข:**
```go
// ต้องระบุ Durable name เพื่อ track progress
js.Subscribe("user.events.*", handler,
    nats.Durable("social-backend-consumer"),  // ← สำคัญ!
    nats.ManualAck(),
)
```

---

### Problem 3: Messages หาย

**สาเหตุ:** ไม่ได้ `Ack()` message

**แก้ไข:**
```go
_, err = js.Subscribe("user.events.*", func(msg *nats.Msg) {
    // Process message
    processEvent(msg.Data)

    // ❗ ต้อง Ack เสมอ
    msg.Ack()
}, nats.ManualAck())
```

---

### Problem 4: Too Many Retries

**สาเหตุ:** Error ใน handler ทำให้ NAK และ retry ไม่รู้จบ

**แก้ไข:**
```go
_, err = js.Subscribe("user.events.*", func(msg *nats.Msg) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Panic: %v", r)
            msg.Term() // Terminate message (ไม่ retry)
        }
    }()

    if err := processEvent(msg.Data); err != nil {
        log.Printf("Error: %v", err)

        // Check retry count
        meta, _ := msg.Metadata()
        if meta.NumDelivered > 5 {
            msg.Term() // Stop retrying after 5 attempts
        } else {
            msg.Nak() // Retry
        }
        return
    }

    msg.Ack() // Success
}, nats.ManualAck())
```

---

## 📊 Monitoring & Observability

### Request ID Tracing

ทุก event มี `request_id` สำหรับ distributed tracing:

```json
{
  "request_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "action": "created",
  "username": "john"
}
```

**การใช้งาน:**
1. เก็บ `request_id` ใน log ของคุณ
2. เมื่อเกิด error สามารถ trace กลับไปหา Auth Service ได้
3. ใช้ใน centralized logging (ELK, Loki)

---

### Sequence Numbers

NATS JetStream ให้ sequence number อัตโนมัติ:

```bash
# View messages with sequence
nats stream view USER_EVENTS

[1] Subject: user.events.created ...
[2] Subject: user.events.created ...
[3] Subject: user.events.updated ...
```

**ใช้ตรวจสอบ:**
- Message หาย (gap ใน sequence)
- Duplicate processing

---

## 🔐 Security Considerations

### 1. Authentication

ถ้า NATS Server อยู่ต่าง network ควรใช้:

```go
nc, err := nats.Connect("nats://localhost:4222",
    nats.UserInfo("username", "password"),
    nats.RootCAs("./certs/ca.pem"),
)
```

### 2. Authorization

ตั้งค่า NATS ACL:

```conf
# nats-server.conf
authorization {
  users = [
    {
      user: "social-backend"
      password: "secret"
      permissions: {
        subscribe: ["user.events.>"]
        publish: []  # ห้าม publish
      }
    }
  ]
}
```

---

## 📚 Additional Resources

### Documentation
- [NATS JetStream Docs](https://docs.nats.io/nats-concepts/jetstream)
- [NATS Go Client](https://github.com/nats-io/nats.go)
- [Auth Service README](./README.md)

### Tools
- **NATS CLI:** `brew install nats-io/nats-tools/nats`
- **NATS Top:** Monitor streams real-time
- **Prometheus:** Auth Service exposes `/metrics` endpoint

---

## 🤝 Support

หากมีปัญหาในการ integrate:

1. Check Auth Service logs: `http://localhost:8088/metrics`
2. Check NATS stream: `nats stream info USER_EVENTS`
3. Contact Auth Service team

---

## 📝 Changelog

### v1.0 (2025-11-24)
- ✅ Initial release
- ✅ NATS JetStream integration
- ✅ Event schema v1
- ✅ Go/Node.js/Python examples
- ✅ Observability (request_id, sequence numbers)

---

**Happy Integration! 🚀**
