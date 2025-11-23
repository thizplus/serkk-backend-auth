# Event Schema V2 - Minimal Identity Event

**อัปเดตล่าสุด:** 2025-11-24
**เวอร์ชัน:** 2.0
**Breaking Change:** ✅ ใช่ (ลบ fields หลายตัว)

---

## 📋 สรุปการเปลี่ยนแปลง

### แนวคิด: Minimal Identity Event

Auth Service ปรับจาก **"Full User Data Event"** เป็น **"Minimal Identity Event"**

**หลักการ:**
- ✅ Auth Service ส่งเฉพาะ **identity data** (id, email, username)
- ❌ **ไม่ส่ง** profile data (displayName, avatar, bio)
- ❌ **ไม่ส่ง** authorization data (role, isActive, permissions)
- 🎯 Downstream services **enrich** ข้อมูล profile เอง

---

## 🔄 Schema Comparison

### Before (V1) - Full User Data

```json
{
  // User Data
  "id": "uuid-here",
  "email": "user@example.com",
  "username": "john_doe",
  "displayName": "John Doe",          ❌ ลบออก
  "avatar": "https://cdn.../pic.jpg", ❌ ลบออก
  "role": "user",                     ❌ ลบออก
  "isActive": true,                   ❌ ลบออก
  "action": "created",

  // Metadata
  "request_id": "uuid",
  "timestamp": "2025-11-24T...",
  "service_name": "gofiber-auth"
}
```

### After (V2) - Minimal Identity Event

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
  "sequence": 42  // NATS sequence number
}
```

---

## ✂️ Fields ที่ลบออก

| Field | Type | เหตุผลที่ลบ | ใครเป็นเจ้าของ |
|-------|------|-------------|----------------|
| `displayName` | string | Profile data | **Social/Profile Service** |
| `avatar` | string | Profile data | **Social/Profile Service** |
| `bio` | string | Profile data | **Social/Profile Service** |
| `role` | string | Authorization internal | **Auth Service internal** |
| `isActive` | boolean | Authorization internal | **Auth Service internal** |
| `permissions` | array | Authorization internal | **Auth Service internal** |

---

## ✅ Fields ที่เก็บไว้

| Field | Type | Required | เหตุผลที่เก็บไว้ |
|-------|------|----------|-----------------|
| `id` | string (UUID) | ✅ | Primary key / Foreign key |
| `email` | string | ✅ | Unique identifier |
| `username` | string | ✅ | Unique identifier |
| `action` | string | ✅ | Event type |
| `request_id` | string (UUID) | ✅ | Distributed tracing |
| `timestamp` | string | ✅ | Event timestamp |
| `service_name` | string | ✅ | Source service |
| `sequence` | integer | ❌ | NATS sequence (optional) |

---

## 🎯 ผลกระทบต่อ Downstream Services

### Social Monolith Backend ต้องทำอะไรบ้าง:

#### 1. Database Schema

**แยกตาราง identity และ profile:**

```sql
-- Auth Service เป็นเจ้าของ (via events)
CREATE TABLE users_identity (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
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
    created_at TIMESTAMP,
    updated_at TIMESTAMP,

    FOREIGN KEY (id) REFERENCES users_identity(id)
);
```

#### 2. Event Handler

**เมื่อได้รับ `user.events.created`:**

```go
func handleUserCreated(event UserEvent) {
    // 1. บันทึก identity data
    db.Exec(`
        INSERT INTO users_identity (id, email, username, created_at)
        VALUES ($1, $2, $3, NOW())
    `, event.ID, event.Email, event.Username)

    // 2. สร้าง profile ว่างๆ (ให้ user มา update เอง)
    db.Exec(`
        INSERT INTO users_profile (id, display_name, created_at)
        VALUES ($1, $2, NOW())
    `, event.ID, event.Username) // ใช้ username เป็น default display_name

    log.Printf("✅ Created user profile for %s", event.Username)
}
```

**เมื่อได้รับ `user.events.updated`:**

```go
func handleUserUpdated(event UserEvent) {
    // Update identity data only
    db.Exec(`
        UPDATE users_identity
        SET email = $1, username = $2, updated_at = NOW()
        WHERE id = $3
    `, event.Email, event.Username, event.ID)

    log.Printf("✅ Updated user identity for %s", event.Username)
}
```

**เมื่อได้รับ `user.events.deleted`:**

```go
func handleUserDeleted(event UserEvent) {
    // Soft delete or cascade delete
    db.Exec(`
        UPDATE users_profile SET deleted_at = NOW() WHERE id = $1
    `, event.ID)

    db.Exec(`
        UPDATE users_identity SET deleted_at = NOW() WHERE id = $1
    `, event.ID)

    log.Printf("✅ Deleted user %s", event.Username)
}
```

---

## 📊 ตัวอย่าง Flow

### Scenario: User สมัครใหม่

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

## 🔐 ข้อดีของการแยก Identity และ Profile

### 1. **Separation of Concerns**
- Auth Service = Authentication & Authorization only
- Social Service = User profiles, social features
- ชัดเจน ไม่ overlap

### 2. **Reduced Coupling**
- Event payload เล็กลง (~50% reduction)
- Auth Service ไม่ต้องรู้ว่า Social Service จัดการ profile ยังไง
- เปลี่ยน profile schema ได้โดยไม่กระทบ Auth Service

### 3. **Scalability**
- Auth Service scale ได้อิสระจาก Social Service
- Profile data (avatar, bio) ไม่ซ้ำซ้อนระหว่าง services
- ลด event size → ลด bandwidth

### 4. **Security**
- Auth Service ไม่ leak ข้อมูล role, isActive ออกไป
- Profile data (avatar URL) ไม่ถูก broadcast ผ่าน events
- Authorization logic เก็บไว้ใน Auth Service เท่านั้น

---

## 🚨 Breaking Changes

### สำหรับ Downstream Services ที่ integrate อยู่แล้ว:

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

### Migration Path:

**ขั้นตอนการ migrate:**

1. **Deploy new Schema V2** (Auth Service)
2. **Update Subscribers** ให้รองรับ schema ใหม่
3. **Test กับ new registrations**
4. **Backfill existing users** (optional)

---

## 📝 Changelog

### Version 2.0 (2025-11-24)

**Added:**
- ✅ `sequence` field (NATS JetStream sequence number)

**Removed:**
- ❌ `displayName` - ย้ายไป Social/Profile Service
- ❌ `avatar` - ย้ายไป Social/Profile Service
- ❌ `role` - เก็บเป็น Auth Service internal only
- ❌ `isActive` - เก็บเป็น Auth Service internal only

**Changed:**
- 🔄 Event philosophy: Full User Data → Minimal Identity Event
- 🔄 Payload size reduction: ~50%

---

## 🔗 Related Documentation

- [INTEGRATION_GUIDE.md](./INTEGRATION_GUIDE.md) - อ่านก่อนเพื่อ integrate
- [README.md](./README.md) - Auth Service overview

---

## 🤝 สำหรับ Social Monolith Team

**คำถามที่พบบ่อย:**

### Q1: ผม register user ใหม่แล้ว displayName หายไปไหน?

**A:** Auth Service ไม่ส่ง displayName ใน event อีกต่อไป คุณต้อง:
1. รับ event → สร้าง users_identity (id, email, username)
2. รับ API call จาก client → สร้าง users_profile (displayName, avatar)

### Q2: ถ้า user update email/username ใน Auth Service จะเกิดอะไรขึ้น?

**A:** Auth Service จะส่ง `user.events.updated` พร้อม email/username ใหม่
- คุณต้อง update ตาราง `users_identity`
- **ไม่ต้อง** update `users_profile` (displayName, avatar ไม่เปลี่ยน)

### Q3: ผมจะเก็บ role, permissions ของ user ยังไง?

**A:** Auth Service เก็บ role/permissions เป็น internal data
- ถ้าต้องการ authorization → เรียก Auth Service API: `GET /api/users/{id}/permissions`
- **ไม่มี** ใน events เพราะเป็น security sensitive

### Q4: Avatar URL ควรเก็บที่ไหน?

**A:** Social/Profile Service เท่านั้น
- Upload avatar → Social Service upload ไป CDN
- เก็บ URL ใน `users_profile.avatar`
- Auth Service **ไม่รู้จัก** avatar เลย

---

**Happy Migrating! 🚀**
