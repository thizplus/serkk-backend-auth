# Auth Service - Quick Start Guide

> คู่มือเริ่มต้นใช้งานอย่างรวดเร็ว สำหรับ Frontend และ Backend

---

## 🎯 สิ่งสำคัญที่ต้องรู้

### 1. Service URLs

| Environment | Auth Service | Backend Service |
|-------------|--------------|-----------------|
| Development | `http://localhost:8088/api/v1` | `http://localhost:8080/api/v1` |
| Production | `https://auth.suekk.com/api/v1` | `https://api.suekk.com/api/v1` |

### 2. JWT Secret (⚠️ สำคัญมาก!)

**Auth Service และ Backend Service ต้องใช้ JWT_SECRET เดียวกัน!**

```env
# ทั้ง 2 services
JWT_SECRET=Log2Window$P@ssWord
```

---

## 🚀 Frontend - วิธีใช้งาน

### การ Login

```javascript
// 1. Login
const response = await fetch('http://localhost:8088/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'password123'
  })
})

const data = await response.json()

// 2. เก็บ JWT Token
localStorage.setItem('token', data.data.token)
localStorage.setItem('user', JSON.stringify(data.data.user))
```

### การเรียก Backend API

```javascript
// 3. ใช้ JWT Token เรียก Backend
const token = localStorage.getItem('token')

const posts = await fetch('http://localhost:8080/api/v1/posts', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
})
```

### Google OAuth

```javascript
// 1. Get Google Auth URL
const response = await fetch('http://localhost:8088/api/v1/auth/google')
const data = await response.json()

// 2. Redirect to Google
window.location.href = data.authURL

// 3. Handle callback (ที่หน้า /oauth/callback)
const urlParams = new URLSearchParams(window.location.search)
const code = urlParams.get('code')

const authResponse = await fetch(
  `http://localhost:8088/api/v1/auth/google/callback?code=${code}`
)
const authData = await authResponse.json()

// 4. เก็บ token
localStorage.setItem('token', authData.accessToken)
```

---

## 🔧 Backend - วิธีใช้งาน

### Go Fiber

```go
// 1. Middleware สำหรับตรวจสอบ JWT
func Protected(jwtSecret string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(jwtSecret), nil
        })

        if err != nil || !token.Valid {
            return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
        }

        claims := token.Claims.(jwt.MapClaims)
        userID, _ := uuid.Parse(claims["user_id"].(string))

        c.Locals("user_id", userID)
        c.Locals("user_email", claims["email"].(string))
        c.Locals("user_role", claims["role"].(string))

        return c.Next()
    }
}

// 2. ใช้ใน routes
app.Use(Protected(os.Getenv("JWT_SECRET")))

app.Post("/posts", func(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(uuid.UUID)

    // ใช้ userID ในการ query database
    post := &Post{
        UserID: userID,
        Title:  c.FormValue("title"),
    }

    db.Create(post)
    return c.JSON(post)
})
```

### Express.js

```javascript
// 1. Middleware
const jwt = require('jsonwebtoken')

const protect = (req, res, next) => {
  const token = req.headers.authorization?.replace('Bearer ', '')

  try {
    const decoded = jwt.verify(token, process.env.JWT_SECRET)
    req.user = {
      id: decoded.user_id,
      email: decoded.email,
      role: decoded.role
    }
    next()
  } catch (error) {
    res.status(401).json({ error: 'Unauthorized' })
  }
}

// 2. ใช้ใน routes
app.post('/api/posts', protect, (req, res) => {
  const userId = req.user.id

  // ใช้ userId ในการ query database
  const post = {
    userId: userId,
    title: req.body.title
  }

  // Save to database...
  res.json(post)
})
```

---

## 📍 API Endpoints

### Authentication

```bash
# Register
POST /auth/register
Body: { email, username, password, firstName, lastName }

# Login
POST /auth/login
Body: { email, password }
Response: { token, user }

# Google OAuth
GET /auth/google
Response: { authURL }

GET /auth/google/callback?code=xxx
Response: { accessToken, user, isNewUser }
```

### User Management (ต้องมี JWT Token)

```bash
# Get Profile
GET /users/profile
Headers: Authorization: Bearer <token>

# Update Profile
PUT /users/profile
Headers: Authorization: Bearer <token>
Body: { firstName, lastName, avatar }

# Delete Account
DELETE /users/profile
Headers: Authorization: Bearer <token>
```

---

## ⚠️ สิ่งที่ต้องระวัง

### 1. JWT Secret ต้องเหมือนกัน!

```env
# Auth Service .env
JWT_SECRET=Log2Window$P@ssWord

# Backend Service .env
JWT_SECRET=Log2Window$P@ssWord  # ⚠️ ต้องเหมือนกัน 100%!
```

### 2. CORS Configuration

```env
# Auth Service .env
CORS_ALLOWED_ORIGINS=http://localhost:3000

# Backend Service .env
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

### 3. User ID Format

JWT token มี `user_id` เป็น **UUID string**:

```javascript
// JWT Claims
{
  "user_id": "4aa10e1b-06c4-4b09-8bd9-5cea94cd3723",  // UUID string
  "username": "johndoe",
  "email": "user@example.com",
  "role": "user",
  "exp": 1234567890
}
```

**ใน Backend ต้อง parse เป็น UUID:**

```go
// Go
userID, _ := uuid.Parse(claims["user_id"].(string))

// JavaScript
const userId = decoded.user_id  // เป็น string อยู่แล้ว
```

---

## 🧪 ทดสอบระบบ

### Test Flow

```bash
# 1. Start Auth Service
cd gofiber-auth
go run cmd/api/main.go
# Listening on :8088

# 2. Start Backend Service
cd gofiber-backend
go run cmd/api/main.go
# Listening on :8080

# 3. Test Register
curl -X POST http://localhost:8088/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "password123",
    "firstName": "Test",
    "lastName": "User"
  }'

# 4. Test Login
curl -X POST http://localhost:8088/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}'

# Copy token from response

# 5. Test Backend API
TOKEN="eyJhbGciOiJIUzI1NiIs..."

curl http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer $TOKEN"
```

---

## 🐛 Troubleshooting

### ปัญหา: 401 Unauthorized

**สาเหตุ:**
- JWT_SECRET ไม่ตรงกัน
- Token หมดอายุ
- Token format ไม่ถูก (ต้องเป็น `Bearer <token>`)

**วิธีแก้:**
```bash
# ตรวจสอบ JWT_SECRET
echo $JWT_SECRET  # Auth Service
echo $JWT_SECRET  # Backend Service
# ต้องเหมือนกัน!
```

### ปัญหา: CORS Error

**สาเหตุ:**
- Frontend origin ไม่ได้รับอนุญาต

**วิธีแก้:**
```env
# เพิ่ม Frontend URL ใน CORS_ALLOWED_ORIGINS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
```

### ปัญหา: User ID ไม่ตรงกัน

**สาเหตุ:**
- UUID format ผิด

**วิธีแก้:**
```go
// ต้อง parse UUID
userIDStr := claims["user_id"].(string)
userID, err := uuid.Parse(userIDStr)
if err != nil {
    return errors.New("invalid user ID")
}
```

---

## 📚 Documentation

- **Full Guide:** [INTEGRATION_GUIDE.md](./docs/INTEGRATION_GUIDE.md)
- **Architecture:** [MICROSERVICES_ARCHITECTURE.md](./docs/MICROSERVICES_ARCHITECTURE.md)
- **Migration:** [USER_MIGRATION_GUIDE.md](./docs/USER_MIGRATION_GUIDE.md)

---

## ✅ Checklist สำหรับ Deploy

### Auth Service
- [ ] ตั้งค่า `.env` (JWT_SECRET, Database, OAuth)
- [ ] Enable PostgreSQL extensions (pgcrypto)
- [ ] Migrate database
- [ ] Test endpoints
- [ ] Configure CORS
- [ ] Setup SSL/HTTPS

### Backend Service
- [ ] ตั้งค่า `.env` (JWT_SECRET เหมือนกับ Auth Service!)
- [ ] Implement JWT middleware
- [ ] Test JWT validation
- [ ] Configure CORS
- [ ] Setup SSL/HTTPS

### Frontend
- [ ] Configure API URLs
- [ ] Implement auth service
- [ ] Handle OAuth callback
- [ ] Store/retrieve JWT tokens
- [ ] Handle 401 errors
- [ ] Test full auth flow

---

**Need Help?** อ่าน [INTEGRATION_GUIDE.md](./docs/INTEGRATION_GUIDE.md) ฉบับเต็ม

**Last Updated:** 2025-11-23
