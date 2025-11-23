# User Migration Guide

คู่มือการ Import ข้อมูล Users จาก CSV เข้าสู่ Auth Service

---

## 📋 ข้อมูลที่รองรับ

### CSV Format

```csv
id,email,username,password,o_auth_provider,o_auth_id,is_o_auth_user,
display_name,avatar,bio,location,website,karma,followers_count,
following_count,role,is_active,created_at,updated_at
```

### Field Mapping

| CSV Column | Database Column | การแปลงข้อมูล |
|-----------|----------------|---------------|
| id | id | UUID ตรง |
| email | email | ตรง |
| username | username | ตรง |
| password | password | ตรง (NULL สำหรับ OAuth users) |
| o_auth_provider | oauth_provider | ตรง |
| o_auth_id | oauth_id | ตรง |
| is_o_auth_user | is_oauth_user | Boolean |
| display_name | display_name | ตรง |
| display_name | first_name | แยกคำแรก |
| display_name | last_name | คำที่เหลือ |
| avatar | avatar | ตรง |
| role | role | ตรง |
| is_active | is_active | Boolean |
| created_at | created_at | Parse timestamp |
| updated_at | updated_at | Parse timestamp |
| updated_at | last_login_at | ใช้ค่า updated_at |
| - | email_verified | ตั้งเป็น true |

**Field ที่ไม่ใช้ (เก็บใน Backend DB):**
- bio
- location
- website
- karma
- followers_count
- following_count

---

## 🚀 วิธีใช้งาน

### ขั้นตอนที่ 1: เตรียม CSV File

วาง `users.csv` ไว้ที่ root ของโปรเจค:

```bash
D:\Admin\Desktop\MY PROJECT\__serkk\gofiber-auth\users.csv
```

### ขั้นตอนที่ 2: ตั้งค่า Database

แก้ไข `.env` ให้ชื้อไปที่ database ที่ต้องการ import:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=gofiber_auth  # Database ปลายทาง
```

### ขั้นตอนที่ 3: รัน Migration Script

```bash
# ไปที่ directory ของโปรเจค
cd "D:\Admin\Desktop\MY PROJECT\__serkk\gofiber-auth"

# รัน import script
go run cmd/migrate/import_users.go
```

### ตัวอย่าง Output

```
Found 4 users to import
Row 1: ✅ Imported user: thepthai.jm@gmail.com (thepthai)
Row 2: ✅ Imported user: devzstack@gmail.com (devzstack)
Row 3: ✅ Imported user: info@thizplus.com (info)
Row 4: ✅ Imported user: manage.karismarketing@gmail.com (managekarismarketing)

=== Import Summary ===
✅ Success: 4
⏭️  Skipped: 0
❌ Errors: 0
📊 Total: 4
```

---

## 🔍 ตัวอย่างข้อมูลที่ Import

### User 1: Register ธรรมดา

**CSV:**
```csv
4aa10e1b-06c4-4b09-8bd9-5cea94cd3723,thepthai.jm@gmail.com,thepthai,
$2a$10$MG9UcemfgSxPEkdJPLRAduwK7C4zTJeE1FnaucyQmgtBBlUypBtsG,
,,False,NOTz,,,,,0,0,0,user,True,2025-11-12 21:21:04,2025-11-12 21:21:04
```

**ผลลัพธ์ในฐานข้อมูล:**
```sql
id:              4aa10e1b-06c4-4b09-8bd9-5cea94cd3723
email:           thepthai.jm@gmail.com
username:        thepthai
password:        $2a$10$MG9UcemfgSxPEkdJPLRAduwK7C4zTJeE1FnaucyQmgtBBlUypBtsG
first_name:      NOTz
last_name:       (empty)
display_name:    NOTz
avatar:          (empty)
is_oauth_user:   false
oauth_provider:  (empty)
oauth_id:        (empty)
email_verified:  true
role:            user
is_active:       true
```

### User 2: Google OAuth

**CSV:**
```csv
8e080dfb-79a8-487a-ab52-2b0779fbe827,devzstack@gmail.com,devzstack,,
google,108233785739852432802,True,DEVZ STACK,
https://lh3.googleusercontent.com/a/ACg8ocK10xIAu_c5qq-Cp0vhyesdi0DgnczcK3YnvUgt4RAqCFnWPg=s96-c,
,,,,0,0,0,user,True,2025-11-12 00:01:00,2025-11-12 01:17:59
```

**ผลลัพธ์ในฐานข้อมูล:**
```sql
id:              8e080dfb-79a8-487a-ab52-2b0779fbe827
email:           devzstack@gmail.com
username:        devzstack
password:        NULL
first_name:      DEVZ
last_name:       STACK
display_name:    DEVZ STACK
avatar:          https://lh3.googleusercontent.com/a/ACg8ocK10xIAu_c5qq...
is_oauth_user:   true
oauth_provider:  google
oauth_id:        108233785739852432802
email_verified:  true
role:            user
is_active:       true
```

---

## ⚠️ สิ่งที่ต้องระวัง

### 1. Database Connection

ตรวจสอบให้แน่ใจว่าเชื่อมต่อ database ถูกต้อง:

```bash
# ทดสอบ connection
psql -h localhost -U postgres -d gofiber_auth -c "SELECT 1"
```

### 2. Duplicate Users

Script จะตรวจสอบ UUID ก่อน insert:
- ถ้า user ID มีอยู่แล้ว → **Skip**
- ถ้ายังไม่มี → **Insert**

### 3. Password Hash

- Password ที่ hash แล้วจาก CSV จะถูก import ตรง ๆ
- OAuth users จะมี password = NULL
- Users สามารถ login ด้วย password เดิมได้

### 4. Email Verification

- User ทุกคนที่ import จะมี `email_verified = true`
- ถือว่าเป็น existing users ที่ verified แล้ว

---

## 🔄 Import ซ้ำได้หรือไม่?

**ได้ครับ!** Script จะตรวจสอบ UUID ก่อน:

```go
// Check if user already exists
var existingUser models.User
result := db.Where("id = ?", user.ID).First(&existingUser)
if result.Error == nil {
    log.Printf("User %s already exists, skipping\n", user.Email)
    skipCount++
    continue
}
```

- User ที่มีอยู่แล้ว → Skip
- User ใหม่เท่านั้น → Insert

---

## 🧹 Clean Up (Optional)

หากต้องการเริ่มใหม่:

```sql
-- ระวัง! คำสั่งนี้จะลบข้อมูลทั้งหมด
TRUNCATE TABLE users CASCADE;
```

---

## 📊 ตรวจสอบผลลัพธ์

### SQL Query

```sql
-- ดู users ทั้งหมด
SELECT id, email, username, is_oauth_user, oauth_provider, created_at
FROM users
ORDER BY created_at DESC;

-- นับจำนวน users
SELECT COUNT(*) FROM users;

-- นับแยกตามประเภท
SELECT
  is_oauth_user,
  oauth_provider,
  COUNT(*) as total
FROM users
GROUP BY is_oauth_user, oauth_provider;
```

### Expected Result

```
is_oauth_user | oauth_provider | total
--------------|----------------|------
false         |                | 1     (thepthai)
true          | google         | 3     (devzstack, info, manage...)
```

---

## 🔧 Troubleshooting

### Error: Cannot connect to database

```
Failed to connect to database: dial tcp [::1]:5432: connect: connection refused
```

**แก้ไข:**
1. ตรวจสอบว่า PostgreSQL รันอยู่
2. ตรวจสอบ `.env` ว่า DB credentials ถูกต้อง
3. ตรวจสอบ firewall

### Error: Invalid UUID

```
Row 1: Failed to parse user: invalid UUID: invalid UUID length: 0
```

**แก้ไข:**
- ตรวจสอบ CSV format ว่าครบทุก column
- ตรวจสอบ UUID ใน column แรก

### Error: Duplicate key value

```
ERROR: duplicate key value violates unique constraint "users_email_key"
```

**แก้ไข:**
- Email ซ้ำในฐานข้อมูล
- ลบ user เก่าออกก่อน หรือ skip user นั้น

---

## 📝 Next Steps

หลังจาก import users เรียบร้อย:

1. **ทดสอบ Login**
   ```bash
   curl -X POST http://localhost:8088/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{
       "email": "thepthai.jm@gmail.com",
       "password": "password_เดิม"
     }'
   ```

2. **ทดสอบ OAuth**
   - Users ที่ login ด้วย Google จะ link กับ account เดิม
   - OAuth provider ID จะถูกเช็คจาก `oauth_id`

3. **Sync กับ Backend**
   - User IDs เหมือนกัน
   - Backend สามารถ query posts/comments ด้วย user_id ได้ตรง

---

## ✅ Summary

**สิ่งที่ import:**
- ✅ User accounts (email/password)
- ✅ OAuth connections (Google)
- ✅ User profiles (avatar, display name)
- ✅ User roles และ status

**สิ่งที่ไม่ import (เก็บใน Backend DB):**
- ❌ Social stats (karma, followers)
- ❌ Profile details (bio, location, website)

**ผลลัพธ์:**
- Users สามารถ login ด้วย credentials เดิมได้
- OAuth users สามารถ login ด้วย Google ได้
- User IDs เหมือนกัน → Backend query ได้ตรง
