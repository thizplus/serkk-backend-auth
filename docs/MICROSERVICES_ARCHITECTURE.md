# Microservices Architecture Guide

## สถาปัตยกรรมระบบ

โปรเจคนี้แบ่งเป็น 2 Services แยกกัน:

```
┌─────────────────────────────────────────────────┐
│              Frontend Applications              │
│   - Web App (suekk.com)                        │
│   - Admin Panel (admin.suekk.com)              │
│   - Mobile App                                  │
└────────┬────────────────────────────┬───────────┘
         │                            │
         │ Auth APIs                  │ Business APIs
         │                            │
         v                            v
┌──────────────────────┐    ┌──────────────────────┐
│  Auth Service        │    │  Backend Service     │
│  Port 8088           │    │  Port 8080           │
│  auth.suekk.com      │    │  api.suekk.com       │
├──────────────────────┤    ├──────────────────────┤
│ ✅ Register          │    │ ✅ Posts             │
│ ✅ Login             │    │ ✅ Comments          │
│ ✅ OAuth             │    │ ✅ Follows           │
│ ✅ User Management   │    │ ✅ Upload            │
│ ✅ JWT Generate      │    │ ✅ AI Features       │
└──────────┬───────────┘    └──────────┬───────────┘
           │                           │
           v                           v
    ┌─────────────┐            ┌─────────────────┐
    │ Auth DB     │            │ Backend DB      │
    │ gofiber_    │            │ gofiber_social  │
    │ auth        │            │                 │
    └─────────────┘            └─────────────────┘
```

---

## 🔄 Flow การทำงาน

### 1. User Authentication

```
User → Frontend → Auth Service (8088) → Return JWT
                                      ↓
                            Store JWT in Frontend
```

### 2. API Calls

```
User → Frontend → Backend Service (8080)
                  ↑ (with JWT Token)
                  ↓
            Validate JWT → Process Request
```

---

## 📍 API Endpoints

### Auth Service (Port 8088)

**Base URL:**
- Development: `http://localhost:8088/api/v1`
- Production: `https://auth.suekk.com/api/v1`

**Endpoints:**

```
Authentication:
POST   /auth/register          - สมัครสมาชิกด้วยอีเมล
POST   /auth/login             - เข้าสู่ระบบด้วยอีเมล

OAuth:
GET    /auth/google            - Get Google OAuth URL
GET    /auth/google/callback   - Google OAuth callback
GET    /auth/facebook          - Get Facebook OAuth URL (อนาคต)
GET    /auth/facebook/callback - Facebook OAuth callback (อนาคต)
GET    /auth/line              - Get LINE OAuth URL (อนาคต)
GET    /auth/line/callback     - LINE OAuth callback (อนาคต)

User Management (ต้องมี JWT):
GET    /users/profile          - ดูข้อมูลโปรไฟล์
PUT    /users/profile          - แก้ไขโปรไฟล์
DELETE /users/profile          - ลบบัญชีผู้ใช้
```

### Backend Service (Port 8080)

**Base URL:**
- Development: `http://localhost:8080/api/v1`
- Production: `https://api.suekk.com/api/v1`

**Endpoints:**

```
Posts:
POST   /posts                  - สร้างโพสต์
GET    /posts                  - ดูโพสต์ทั้งหมด
GET    /posts/:id              - ดูโพสต์ตาม ID
PUT    /posts/:id              - แก้ไขโพสต์
DELETE /posts/:id              - ลบโพสต์

Comments:
POST   /posts/:id/comments     - แสดงความคิดเห็น
GET    /posts/:id/comments     - ดูความคิดเห็น

Follows:
POST   /follows/:user_id       - ติดตามผู้ใช้
DELETE /follows/:user_id       - เลิกติดตาม
GET    /follows/followers      - ดูผู้ติดตาม
GET    /follows/following      - ดูผู้ที่ติดตาม

... (endpoints อื่นๆ)
```

---

## 💻 วิธีเรียกใช้งาน

### Frontend Implementation

#### 1. Configuration

```javascript
// config/api.js
export const API_CONFIG = {
  // Development
  AUTH_API: process.env.NEXT_PUBLIC_AUTH_API || 'http://localhost:8088/api/v1',
  BACKEND_API: process.env.NEXT_PUBLIC_BACKEND_API || 'http://localhost:8080/api/v1',

  // Production
  // AUTH_API: 'https://auth.suekk.com/api/v1',
  // BACKEND_API: 'https://api.suekk.com/api/v1',
}
```

#### 2. Authentication Service

```javascript
// services/authService.js
import { API_CONFIG } from '@/config/api'

export const authService = {
  // Register
  async register(data) {
    const response = await fetch(`${API_CONFIG.AUTH_API}/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data)
    })
    return await response.json()
  },

  // Login
  async login(email, password) {
    const response = await fetch(`${API_CONFIG.AUTH_API}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, password })
    })

    const data = await response.json()

    if (data.success) {
      // Store JWT token
      localStorage.setItem('token', data.data.token)
      localStorage.setItem('user', JSON.stringify(data.data.user))
    }

    return data
  },

  // Google OAuth
  async googleLogin() {
    const response = await fetch(`${API_CONFIG.AUTH_API}/auth/google`)
    const data = await response.json()

    // Redirect to Google
    window.location.href = data.authURL
  },

  // Handle OAuth Callback
  async handleOAuthCallback(code) {
    const response = await fetch(
      `${API_CONFIG.AUTH_API}/auth/google/callback?code=${code}`
    )
    const data = await response.json()

    // Store JWT token
    localStorage.setItem('token', data.accessToken)
    localStorage.setItem('user', JSON.stringify(data.user))

    return data
  },

  // Logout
  logout() {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  },

  // Get current user
  getCurrentUser() {
    const user = localStorage.getItem('user')
    return user ? JSON.parse(user) : null
  },

  // Get token
  getToken() {
    return localStorage.getItem('token')
  }
}
```

#### 3. Backend API Service

```javascript
// services/apiService.js
import { API_CONFIG } from '@/config/api'
import { authService } from './authService'

export const apiService = {
  // Helper: Get headers with JWT
  getHeaders() {
    const token = authService.getToken()
    return {
      'Content-Type': 'application/json',
      'Authorization': token ? `Bearer ${token}` : ''
    }
  },

  // Posts
  async getPosts() {
    const response = await fetch(`${API_CONFIG.BACKEND_API}/posts`, {
      headers: this.getHeaders()
    })
    return await response.json()
  },

  async createPost(data) {
    const response = await fetch(`${API_CONFIG.BACKEND_API}/posts`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify(data)
    })
    return await response.json()
  },

  // Comments
  async getComments(postId) {
    const response = await fetch(
      `${API_CONFIG.BACKEND_API}/posts/${postId}/comments`,
      { headers: this.getHeaders() }
    )
    return await response.json()
  },

  async addComment(postId, content) {
    const response = await fetch(
      `${API_CONFIG.BACKEND_API}/posts/${postId}/comments`,
      {
        method: 'POST',
        headers: this.getHeaders(),
        body: JSON.stringify({ content })
      }
    )
    return await response.json()
  },

  // Profile
  async getProfile() {
    const response = await fetch(`${API_CONFIG.AUTH_API}/users/profile`, {
      headers: this.getHeaders()
    })
    return await response.json()
  },

  async updateProfile(data) {
    const response = await fetch(`${API_CONFIG.AUTH_API}/users/profile`, {
      method: 'PUT',
      headers: this.getHeaders(),
      body: JSON.stringify(data)
    })
    return await response.json()
  }
}
```

#### 4. Usage Example

```javascript
// pages/login.js
import { authService } from '@/services/authService'

export default function LoginPage() {
  const handleLogin = async (e) => {
    e.preventDefault()

    const result = await authService.login(email, password)

    if (result.success) {
      // Redirect to dashboard
      router.push('/dashboard')
    } else {
      alert('Login failed')
    }
  }

  const handleGoogleLogin = async () => {
    await authService.googleLogin()
  }

  return (
    <form onSubmit={handleLogin}>
      <input type="email" />
      <input type="password" />
      <button type="submit">Login</button>
      <button type="button" onClick={handleGoogleLogin}>
        Login with Google
      </button>
    </form>
  )
}

// pages/posts.js
import { apiService } from '@/services/apiService'

export default function PostsPage() {
  const [posts, setPosts] = useState([])

  useEffect(() => {
    const fetchPosts = async () => {
      const data = await apiService.getPosts()
      setPosts(data.posts)
    }
    fetchPosts()
  }, [])

  const handleCreatePost = async (content) => {
    await apiService.createPost({
      title: 'New Post',
      content: content
    })
  }

  return <div>...</div>
}
```

---

## 🔒 CORS Configuration

### ปัญหา: CORS Error

เมื่อ Frontend เรียก API จาก domain ต่างกัน จะเจอ CORS error:

```
Access to fetch at 'http://localhost:8088/api/v1/auth/login' from origin
'http://localhost:3000' has been blocked by CORS policy
```

### วิธีแก้: ตั้งค่า CORS ใน Backend

#### Auth Service (Port 8088)

**File:** `gofiber-auth/interfaces/api/middleware/cors_middleware.go`

```go
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CorsMiddleware() fiber.Handler {
	return cors.New(cors.Config{
		// Development: อนุญาตทุก origin
		AllowOrigins: "*",

		// Production: ระบุ origin ที่อนุญาต
		// AllowOrigins: "https://suekk.com,https://admin.suekk.com,https://app.suekk.com",

		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		MaxAge: 3600,
	})
}
```

**สำหรับ Production (แนะนำ):**

```go
package middleware

import (
	"strings"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CorsMiddleware() fiber.Handler {
	return cors.New(cors.Config{
		// อนุญาตเฉพาะ domain และ subdomain ที่กำหนด
		AllowOriginsFunc: func(origin string) bool {
			// อนุญาต localhost (Development)
			if strings.HasPrefix(origin, "http://localhost:") {
				return true
			}

			// อนุญาต suekk.com และ subdomain ทั้งหมด
			allowedDomains := []string{
				"https://suekk.com",
				"https://www.suekk.com",
				"https://admin.suekk.com",
				"https://app.suekk.com",
				"https://mobile.suekk.com",
			}

			for _, domain := range allowedDomains {
				if origin == domain {
					return true
				}
			}

			// อนุญาต subdomain pattern *.suekk.com
			if strings.HasSuffix(origin, ".suekk.com") {
				return true
			}

			return false
		},

		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		MaxAge: 3600,
	})
}
```

#### Backend Service (Port 8080)

**File:** `gofiber-backend/interfaces/api/middleware/cors_middleware.go`

```go
package middleware

import (
	"strings"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CorsMiddleware() fiber.Handler {
	return cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			// Development
			if strings.HasPrefix(origin, "http://localhost:") {
				return true
			}

			// Production - อนุญาต domain เดียวกับ Auth Service
			if strings.HasSuffix(origin, ".suekk.com") || origin == "https://suekk.com" {
				return true
			}

			return false
		},

		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		MaxAge: 3600,
	})
}
```

#### ตั้งค่าใน main.go

**Auth Service:**

```go
// cmd/api/main.go
package main

import (
	"gofiber-template/interfaces/api/middleware"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// CORS Middleware - ต้องอยู่ก่อน routes
	app.Use(middleware.CorsMiddleware())

	// Routes
	routes.SetupRoutes(app, h)

	app.Listen(":8088")
}
```

**Backend Service:**

```go
// cmd/api/main.go
package main

import (
	"gofiber-backend/interfaces/api/middleware"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// CORS Middleware
	app.Use(middleware.CorsMiddleware())

	// Routes
	routes.SetupRoutes(app, h)

	app.Listen(":8080")
}
```

---

## 🌐 Domain Configuration

### Development

```env
# Auth Service (.env)
APP_PORT=8088
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001

# Backend Service (.env)
APP_PORT=8080
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
```

### Production

```env
# Auth Service (.env)
APP_PORT=8088
APP_DOMAIN=auth.suekk.com
CORS_ALLOWED_ORIGINS=https://suekk.com,https://admin.suekk.com,https://app.suekk.com

# Backend Service (.env)
APP_PORT=8080
APP_DOMAIN=api.suekk.com
CORS_ALLOWED_ORIGINS=https://suekk.com,https://admin.suekk.com,https://app.suekk.com
```

---

## 🔑 JWT Configuration

### สิ่งสำคัญ: JWT_SECRET ต้องเหมือนกัน!

**Auth Service (.env):**
```env
JWT_SECRET=your-super-secret-key-change-in-production
```

**Backend Service (.env):**
```env
JWT_SECRET=your-super-secret-key-change-in-production
# ⚠️ ต้องเหมือนกับ Auth Service ทุกประการ!
```

### ทำไมต้องเหมือนกัน?

1. **Auth Service** สร้าง JWT ด้วย `JWT_SECRET`
2. **Backend Service** ตรวจสอบ JWT ด้วย `JWT_SECRET` เดียวกัน
3. ถ้าไม่เหมือนกัน → Validation ไม่ผ่าน → 401 Unauthorized

---

## 🧪 Testing

### 1. Test Auth Service

```bash
# Register
curl -X POST http://localhost:8088/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "password123",
    "firstName": "John",
    "lastName": "Doe"
  }'

# Login
curl -X POST http://localhost:8088/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# Response
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "uuid",
      "email": "test@example.com",
      "username": "testuser",
      "firstName": "John",
      "lastName": "Doe"
    }
  }
}
```

### 2. Test Backend Service with JWT

```bash
# Copy token from login response
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Test protected endpoint
curl http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer $TOKEN"

# Create post
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Post",
    "content": "Hello World"
  }'
```

### 3. Test CORS

```bash
# Test from browser console
fetch('http://localhost:8088/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'test@example.com',
    password: 'password123'
  })
})
.then(res => res.json())
.then(console.log)
```

---

## 📦 Deployment Checklist

### Before Deploy

- [ ] ตั้งค่า `JWT_SECRET` เป็นค่าเดียวกันทั้ง 2 service
- [ ] ตั้งค่า CORS ให้รองรับ production domains
- [ ] ตั้งค่า environment variables
- [ ] Test JWT validation ระหว่าง services
- [ ] Setup SSL certificates (HTTPS)
- [ ] Configure subdomain DNS records

### DNS Configuration

```
A     auth.suekk.com    → IP ของ Auth Service
A     api.suekk.com     → IP ของ Backend Service
A     suekk.com         → IP ของ Frontend
CNAME admin.suekk.com   → Frontend IP
CNAME app.suekk.com     → Frontend IP
```

### Nginx Configuration (ตัวอย่าง)

```nginx
# Auth Service
server {
    listen 80;
    server_name auth.suekk.com;

    location / {
        proxy_pass http://localhost:8088;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# Backend Service
server {
    listen 80;
    server_name api.suekk.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## 🚨 Common Issues

### Issue 1: CORS Error

**Problem:**
```
Access to fetch blocked by CORS policy
```

**Solution:**
- ตรวจสอบ CORS middleware ใน `main.go`
- ตรวจสอบ `AllowOrigins` ใน config
- ตรวจสอบว่า frontend origin ถูกอนุญาต

### Issue 2: JWT Validation Failed

**Problem:**
```
401 Unauthorized - Invalid token
```

**Solution:**
- ตรวจสอบ `JWT_SECRET` ต้องเหมือนกันทั้ง 2 service
- ตรวจสอบ token format: `Bearer <token>`
- ตรวจสอบ token expiry

### Issue 3: Cannot Connect to Service

**Problem:**
```
Failed to fetch / Network error
```

**Solution:**
- ตรวจสอบ service รันอยู่หรือไม่
- ตรวจสอบ port ถูกต้องหรือไม่
- ตรวจสอบ firewall

---

## 📚 References

- [Go Fiber Documentation](https://docs.gofiber.io/)
- [JWT Best Practices](https://jwt.io/introduction)
- [CORS Explained](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
- [Microservices Architecture](https://microservices.io/)

---

## 📝 License

MIT License
