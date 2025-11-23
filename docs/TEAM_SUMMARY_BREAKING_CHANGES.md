# 🚨 Breaking Changes: User Field Simplification

**Date:** 2025-11-23
**Version:** 2.0
**Status:** ✅ Completed

---

## 📋 Overview

เราได้ทำการลดความซับซ้อนของ User Model โดยการลบฟิลด์ `firstName` และ `lastName` ออก และใช้เพียง **`displayName`** เดียวแทน

### เหตุผล
- ลดความซับซ้อนในการจัดการข้อมูล
- Frontend และ Backend ไม่จำเป็นต้องแยก firstName/lastName
- ผู้ใช้สามารถกำหนดชื่อแสดงผลได้เองตามต้องการ
- เพิ่มความยืดหยุ่นในการตั้งชื่อ (เช่น ชื่อเล่น, ชื่อธุรกิจ)

---

## 🔴 Breaking Changes

### ฟิลด์ที่ถูกลบออก
```diff
- firstName (string)
- lastName (string)
```

### ฟิลด์ที่เพิ่มเข้ามา/คงอยู่
```diff
+ displayName (string) - REQUIRED for registration
+ avatar (string) - Optional
```

---

## 🎯 Impact Analysis

### 1. Frontend Impact (HIGH)

#### Registration API (`POST /api/v1/auth/register`)

**เดิม (Old):**
```json
{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "SecurePass123",
  "firstName": "John",
  "lastName": "Doe"
}
```

**ใหม่ (New):**
```json
{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "SecurePass123",
  "displayName": "John Doe"
}
```

**Validation Rules:**
- `displayName`: **required**, min=1, max=100 characters

---

#### Update Profile API (`PATCH /api/v1/users/profile`)

**เดิม (Old):**
```json
{
  "firstName": "John",
  "lastName": "Smith",
  "avatar": "https://example.com/avatar.jpg"
}
```

**ใหม่ (New):**
```json
{
  "displayName": "John Smith",
  "avatar": "https://example.com/avatar.jpg"
}
```

**Validation Rules:**
- `displayName`: optional (omitempty), min=1, max=100 characters
- `avatar`: optional, must be valid URL, max=500 characters

---

#### User Response (GET endpoints)

**เดิม (Old):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "username": "johndoe",
  "firstName": "John",
  "lastName": "Doe",
  "displayName": "John Doe",
  "avatar": "",
  "role": "user",
  "isActive": true,
  "createdAt": "2025-11-23T10:00:00Z",
  "updatedAt": "2025-11-23T10:00:00Z"
}
```

**ใหม่ (New):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "username": "johndoe",
  "displayName": "John Doe",
  "avatar": "",
  "role": "user",
  "isActive": true,
  "createdAt": "2025-11-23T10:00:00Z",
  "updatedAt": "2025-11-23T10:00:00Z"
}
```

**ลบออก:** `firstName`, `lastName`
**คงเหลือ:** `displayName` (always present, fallback to username if empty)

---

### 2. Backend Impact (MEDIUM)

#### Webhook Payload Changes

**เดิม (Old Payload):**
```json
{
  "action": "created",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "username": "johndoe",
    "firstName": "John",
    "lastName": "Doe",
    "displayName": "John Doe",
    "avatar": "",
    "role": "user",
    "isActive": true
  }
}
```

**ใหม่ (New Payload):**
```json
{
  "action": "created",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "username": "johndoe",
    "displayName": "John Doe",
    "avatar": "",
    "role": "user",
    "isActive": true
  }
}
```

**Backend Team Action Required:**
- ✅ ลบ `firstName` และ `lastName` ออกจาก User Model ใน Backend Service
- ✅ ใช้ `displayName` แทน
- ✅ อัพเดท validation และ DTO ให้สอดคล้องกัน

---

## 🗄️ Database Migration

### Migration Script
เราได้ทำการ migrate ข้อมูลเดิมแล้วด้วยคำสั่ง:

```sql
UPDATE users
SET display_name = TRIM(first_name || ' ' || last_name)
WHERE (display_name = '' OR display_name IS NULL)
  AND (first_name != '' OR last_name != '');
```

**ผลลัพธ์:** 1 user ถูก migrate สำเร็จ ✅

### ลบ Columns (หลัง Deploy ไปแล้ว)
```sql
ALTER TABLE users DROP COLUMN first_name;
ALTER TABLE users DROP COLUMN last_name;
```

**⚠️ คำเตือน:** รอให้ Frontend และ Backend อัพเดทโค้ดให้เรียบร้อยก่อนทำการลบ columns

---

## 📝 Implementation Checklist

### Frontend Team
- [ ] อัพเดท Registration Form ให้ใช้ `displayName` แทน `firstName + lastName`
- [ ] อัพเดท Profile Edit Form
- [ ] อัพเดท UI ที่แสดง User Information
- [ ] ลบ TypeScript interfaces ที่มี `firstName`, `lastName`
- [ ] เพิ่ม validation สำหรับ `displayName` (required, 1-100 chars)
- [ ] ทดสอบ Registration Flow
- [ ] ทดสอบ Profile Update Flow

### Backend Team
- [ ] อัพเดท User Model ให้ลบ `firstName`, `lastName`
- [ ] อัพเดท DTO/Request Objects
- [ ] อัพเดท Webhook Handler ที่รับข้อมูลจาก Auth Service
- [ ] ปรับ validation rules
- [ ] ทดสอบ Webhook Sync (action: created, updated, deleted)
- [ ] อัพเดท API Documentation

### Database Team
- [x] Migrate existing data (firstName + lastName → displayName) ✅
- [ ] Verify migration results
- [ ] Drop `first_name` and `last_name` columns (รอหลัง deploy)

---

## 🔄 Rollback Plan

หากพบปัญหา สามารถ rollback ได้ดังนี้:

1. **Code Rollback:** Deploy version เก่ากลับไป
2. **Database Rollback:**
   ```sql
   -- เพิ่ม columns กลับมา
   ALTER TABLE users ADD COLUMN first_name VARCHAR(50);
   ALTER TABLE users ADD COLUMN last_name VARCHAR(50);

   -- Split displayName กลับเป็น firstName + lastName (ถ้าจำเป็น)
   -- (แนะนำให้ใช้ backup database แทน)
   ```

---

## 📌 Testing Guide

### Test Cases for Frontend

#### 1. Registration
```bash
POST /api/v1/auth/register
{
  "email": "test@example.com",
  "username": "testuser",
  "password": "Test1234",
  "displayName": "Test User"
}

Expected: 201 Created
Response: UserResponse with displayName = "Test User"
```

#### 2. Profile Update
```bash
PATCH /api/v1/users/profile
Authorization: Bearer <token>
{
  "displayName": "Updated Name"
}

Expected: 200 OK
Response: UserResponse with updated displayName
```

#### 3. Get Current User
```bash
GET /api/v1/users/me
Authorization: Bearer <token>

Expected: 200 OK
Response: UserResponse (no firstName/lastName fields)
```

### Test Cases for Backend

#### 1. Webhook - User Created
```bash
# Auth Service จะส่ง POST request ไปที่
POST http://backend-service:8080/internal/users/sync
{
  "action": "created",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "username",
    "displayName": "Display Name",
    "avatar": "",
    "role": "user",
    "isActive": true
  }
}

Expected: 200 OK
Backend should create/update user with displayName
```

---

## 🔗 Related Documentation

- [Phase 5 Webhook Implementation](./PHASE5_WEBHOOK_IMPLEMENTATION.md)
- [Missing DisplayName Field Fix](./MISSING_DISPLAYNAME_FIELD.md)

---

## 💬 Support

หากมีคำถามหรือพบปัญหา ติดต่อ:
- **Auth Service Owner:** [Your Name/Team]
- **Repository:** https://github.com/thizplus/serkk-backend-auth

---

## 📅 Timeline

| Date | Action | Status |
|------|--------|--------|
| 2025-11-23 | Data Migration (firstName+lastName → displayName) | ✅ Done |
| 2025-11-23 | Code Changes (Model, DTO, Mappers) | ✅ Done |
| TBD | Frontend Update | ⏳ Pending |
| TBD | Backend Update | ⏳ Pending |
| TBD | Deploy to Production | ⏳ Pending |
| TBD | Drop old columns | ⏳ Pending |

---

**สรุป:**
การเปลี่ยนแปลงนี้จะทำให้ระบบเรียบง่ายขึ้น และง่ายต่อการดูแลรักษา ทั้ง Frontend และ Backend ควรอัพเดทโค้ดให้เร็วที่สุด เพื่อให้สอดคล้องกับ Auth Service เวอร์ชันใหม่

**หากมีคำถาม กรุณา comment ใน issue หรือติดต่อทีม Auth Service โดยตรง**
