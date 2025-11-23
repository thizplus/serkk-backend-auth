# Auth Service Integration Guide

คู่มือการเชื่อมต่อและใช้งาน Auth Service สำหรับ Frontend และ Backend

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Auth Service Setup](#auth-service-setup)
3. [API Endpoints](#api-endpoints)
4. [Frontend Integration](#frontend-integration)
5. [Backend Integration](#backend-integration)
6. [Testing](#testing)
7. [Production Deployment](#production-deployment)

---

## 🎯 Overview

### Architecture

```
┌─────────────────┐
│    Frontend     │
│  (Port 3000)    │
└────────┬────────┘
         │
         ├──────────────┬─────────────────┐
         │              │                 │
         v              v                 v
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ Auth Service│  │   Backend   │  │   Other     │
│ (Port 8088) │  │   Service   │  │  Services   │
│             │  │ (Port 8080) │  │             │
│ - Login     │  │ - Posts     │  │             │
│ - Register  │  │ - Comments  │  │             │
│ - OAuth     │  │ - Business  │  │             │
└─────────────┘  └─────────────┘  └─────────────┘
```

### Services

| Service | Port | Domain (Production) | Purpose |
|---------|------|---------------------|---------|
| Auth Service | 8088 | `auth.suekk.com` | Authentication & User Management |
| Backend Service | 8080 | `api.suekk.com` | Business Logic |
| Frontend | 3000 | `suekk.com` | User Interface |

---

## 🚀 Auth Service Setup

### 1. Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Redis (optional)

### 2. Installation

```bash
cd gofiber-auth

# Copy environment file
cp .env.example .env

# Update .env
# See configuration section below
```

### 3. Configuration

**Required Environment Variables:**

```env
# Application
APP_NAME=Auth Service
APP_PORT=8088
APP_ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=gofiber_auth
DB_SSL_MODE=disable

# JWT (IMPORTANT: Must match Backend Service)
JWT_SECRET=Log2Window$P@ssWord

# OAuth - Google
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-xxxxxxxxxxxxx
GOOGLE_REDIRECT_URL=http://localhost:8088/api/v1/auth/google/callback

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

### 4. Database Setup

```sql
-- Create database
CREATE DATABASE gofiber_auth;

-- Enable extensions (if not installed)
\c gofiber_auth
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
```

### 5. Run Service

```bash
# Development
go run cmd/api/main.go

# Production
go build -o auth-service cmd/api/main.go
./auth-service
```

**Server will start at:** `http://localhost:8088`

---

## 📍 API Endpoints

### Base URL

- Development: `http://localhost:8088/api/v1`
- Production: `https://auth.suekk.com/api/v1`

### Authentication Endpoints

#### 1. Register

**POST** `/auth/register`

**Request:**
```json
{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "password123",
  "firstName": "John",
  "lastName": "Doe"
}
```

**Response:**
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "johndoe",
    "firstName": "John",
    "lastName": "Doe",
    "role": "user",
    "isActive": true
  }
}
```

#### 2. Login

**POST** `/auth/login`

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "username": "johndoe",
      "firstName": "John",
      "lastName": "Doe",
      "role": "user"
    }
  }
}
```

#### 3. Google OAuth

**GET** `/auth/google`

**Response:**
```json
{
  "authURL": "https://accounts.google.com/o/oauth2/auth?..."
}
```

**Workflow:**
1. Frontend redirects user to `authURL`
2. User logs in with Google
3. Google redirects to `/auth/google/callback?code=xxx`
4. Auth service returns JWT token

**GET** `/auth/google/callback?code=xxx`

**Response:**
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "tokenType": "Bearer",
  "expiresIn": 604800,
  "user": {
    "id": "uuid",
    "email": "user@gmail.com",
    "username": "user_abc123",
    "firstName": "John",
    "lastName": "Doe",
    "avatar": "https://lh3.googleusercontent.com/...",
    "role": "user"
  },
  "isNewUser": true
}
```

### User Management Endpoints

All endpoints require JWT token in Authorization header.

#### 4. Get Profile

**GET** `/users/profile`

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "johndoe",
    "firstName": "John",
    "lastName": "Doe",
    "avatar": "https://...",
    "role": "user"
  }
}
```

#### 5. Update Profile

**PUT** `/users/profile`

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Request:**
```json
{
  "firstName": "Jane",
  "lastName": "Smith",
  "avatar": "https://example.com/avatar.jpg"
}
```

#### 6. Delete Account

**DELETE** `/users/profile`

**Headers:**
```
Authorization: Bearer <jwt_token>
```

---

## 💻 Frontend Integration

### Setup

```javascript
// config/api.js
export const API_CONFIG = {
  AUTH_API: process.env.NEXT_PUBLIC_AUTH_API || 'http://localhost:8088/api/v1',
  BACKEND_API: process.env.NEXT_PUBLIC_BACKEND_API || 'http://localhost:8080/api/v1'
}
```

### Authentication Service

```javascript
// services/authService.js
import { API_CONFIG } from '@/config/api'

class AuthService {
  // Register
  async register(userData) {
    const response = await fetch(`${API_CONFIG.AUTH_API}/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(userData)
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.message || 'Registration failed')
    }

    return data
  }

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

    if (response.ok && data.success) {
      // Store token
      localStorage.setItem('token', data.data.token)
      localStorage.setItem('user', JSON.stringify(data.data.user))
      return data
    }

    throw new Error(data.message || 'Login failed')
  }

  // Google OAuth Login
  async googleLogin() {
    const response = await fetch(`${API_CONFIG.AUTH_API}/auth/google`)
    const data = await response.json()

    // Redirect to Google
    window.location.href = data.authURL
  }

  // Handle OAuth Callback
  async handleOAuthCallback(code) {
    const response = await fetch(
      `${API_CONFIG.AUTH_API}/auth/google/callback?code=${code}`
    )

    const data = await response.json()

    // Store token
    localStorage.setItem('token', data.accessToken)
    localStorage.setItem('user', JSON.stringify(data.user))

    return data
  }

  // Logout
  logout() {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  // Get Token
  getToken() {
    return localStorage.getItem('token')
  }

  // Get User
  getUser() {
    const user = localStorage.getItem('user')
    return user ? JSON.parse(user) : null
  }

  // Check if authenticated
  isAuthenticated() {
    return !!this.getToken()
  }

  // Get Profile
  async getProfile() {
    const token = this.getToken()

    const response = await fetch(`${API_CONFIG.AUTH_API}/users/profile`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })

    return await response.json()
  }

  // Update Profile
  async updateProfile(data) {
    const token = this.getToken()

    const response = await fetch(`${API_CONFIG.AUTH_API}/users/profile`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(data)
    })

    return await response.json()
  }
}

export default new AuthService()
```

### API Service for Backend Calls

```javascript
// services/apiService.js
import { API_CONFIG } from '@/config/api'
import authService from './authService'

class ApiService {
  // Helper to get headers with JWT
  getHeaders() {
    const token = authService.getToken()
    return {
      'Content-Type': 'application/json',
      'Authorization': token ? `Bearer ${token}` : ''
    }
  }

  // Generic request
  async request(endpoint, options = {}) {
    const response = await fetch(`${API_CONFIG.BACKEND_API}${endpoint}`, {
      ...options,
      headers: {
        ...this.getHeaders(),
        ...options.headers
      }
    })

    if (response.status === 401) {
      // Unauthorized - redirect to login
      authService.logout()
      window.location.href = '/login'
      throw new Error('Unauthorized')
    }

    return await response.json()
  }

  // GET request
  async get(endpoint) {
    return this.request(endpoint, { method: 'GET' })
  }

  // POST request
  async post(endpoint, data) {
    return this.request(endpoint, {
      method: 'POST',
      body: JSON.stringify(data)
    })
  }

  // PUT request
  async put(endpoint, data) {
    return this.request(endpoint, {
      method: 'PUT',
      body: JSON.stringify(data)
    })
  }

  // DELETE request
  async delete(endpoint) {
    return this.request(endpoint, { method: 'DELETE' })
  }
}

export default new ApiService()
```

### Usage Examples

```javascript
// pages/login.jsx
import authService from '@/services/authService'

export default function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const handleLogin = async (e) => {
    e.preventDefault()

    try {
      await authService.login(email, password)
      router.push('/dashboard')
    } catch (error) {
      alert(error.message)
    }
  }

  const handleGoogleLogin = () => {
    authService.googleLogin()
  }

  return (
    <form onSubmit={handleLogin}>
      <input
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="Email"
      />
      <input
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder="Password"
      />
      <button type="submit">Login</button>
      <button type="button" onClick={handleGoogleLogin}>
        Login with Google
      </button>
    </form>
  )
}

// pages/oauth/callback.jsx
import { useEffect } from 'react'
import { useRouter } from 'next/router'
import authService from '@/services/authService'

export default function OAuthCallback() {
  const router = useRouter()
  const { code } = router.query

  useEffect(() => {
    if (code) {
      authService.handleOAuthCallback(code)
        .then(() => router.push('/dashboard'))
        .catch(error => {
          console.error(error)
          router.push('/login')
        })
    }
  }, [code])

  return <div>Processing login...</div>
}

// pages/dashboard.jsx
import { useEffect, useState } from 'react'
import apiService from '@/services/apiService'

export default function Dashboard() {
  const [posts, setPosts] = useState([])

  useEffect(() => {
    // Call Backend API with JWT token
    apiService.get('/posts')
      .then(data => setPosts(data.posts))
      .catch(error => console.error(error))
  }, [])

  const createPost = async (content) => {
    await apiService.post('/posts', {
      title: 'New Post',
      content: content
    })
  }

  return (
    <div>
      <h1>Dashboard</h1>
      {/* ... */}
    </div>
  )
}
```

---

## 🔧 Backend Integration

### JWT Validation

Your backend service MUST validate JWT tokens from Auth Service.

**IMPORTANT:** Use the **same JWT_SECRET** as Auth Service!

```env
# Backend .env
JWT_SECRET=Log2Window$P@ssWord  # MUST be identical to Auth Service!
```

### Go Fiber Example

```go
// middleware/auth.go
package middleware

import (
	"errors"
	"strings"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserContext struct {
	ID       uuid.UUID
	Username string
	Email    string
	Role     string
}

func Protected(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Missing authorization header",
			})
		}

		// Extract token
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Invalid or expired token",
			})
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Invalid token claims",
			})
		}

		// Parse user data
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Invalid user ID in token",
			})
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Invalid user ID format",
			})
		}

		// Set user context
		c.Locals("user", &UserContext{
			ID:       userID,
			Username: claims["username"].(string),
			Email:    claims["email"].(string),
			Role:     claims["role"].(string),
		})

		return c.Next()
	}
}

// Get user from context
func GetUser(c *fiber.Ctx) (*UserContext, error) {
	user, ok := c.Locals("user").(*UserContext)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}

// Require specific role
func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, err := GetUser(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Unauthorized",
			})
		}

		if user.Role != role {
			return c.Status(403).JSON(fiber.Map{
				"success": false,
				"message": "Forbidden: insufficient permissions",
			})
		}

		return c.Next()
	}
}
```

### Usage in Routes

```go
// main.go
func main() {
	app := fiber.New()

	jwtSecret := os.Getenv("JWT_SECRET")

	// Public routes
	app.Get("/health", healthCheck)

	// Protected routes
	api := app.Group("/api/v1")
	api.Use(middleware.Protected(jwtSecret))

	// Posts (any authenticated user)
	api.Post("/posts", createPost)
	api.Get("/posts", getPosts)

	// Admin only
	admin := api.Group("/admin")
	admin.Use(middleware.RequireRole("admin"))
	admin.Get("/users", listAllUsers)

	app.Listen(":8080")
}

// Handler example
func createPost(c *fiber.Ctx) error {
	// Get authenticated user
	user, err := middleware.GetUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Use user.ID for database operations
	post := &Post{
		UserID:  user.ID,  // UUID from JWT token
		Title:   c.FormValue("title"),
		Content: c.FormValue("content"),
	}

	// Save to database
	db.Create(post)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    post,
	})
}
```

### Express.js Example

```javascript
// middleware/auth.js
const jwt = require('jsonwebtoken')

const JWT_SECRET = process.env.JWT_SECRET

const protect = (req, res, next) => {
  const authHeader = req.headers.authorization

  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return res.status(401).json({
      success: false,
      message: 'Missing or invalid authorization header'
    })
  }

  const token = authHeader.replace('Bearer ', '')

  try {
    const decoded = jwt.verify(token, JWT_SECRET)

    req.user = {
      id: decoded.user_id,
      username: decoded.username,
      email: decoded.email,
      role: decoded.role
    }

    next()
  } catch (error) {
    return res.status(401).json({
      success: false,
      message: 'Invalid or expired token'
    })
  }
}

const requireRole = (role) => {
  return (req, res, next) => {
    if (req.user.role !== role) {
      return res.status(403).json({
        success: false,
        message: 'Forbidden: insufficient permissions'
      })
    }
    next()
  }
}

module.exports = { protect, requireRole }

// Usage in routes
const express = require('express')
const { protect, requireRole } = require('./middleware/auth')

const app = express()

// Protected route
app.post('/api/posts', protect, (req, res) => {
  const userId = req.user.id  // From JWT token

  // Create post with userId
  const post = {
    userId: userId,
    title: req.body.title,
    content: req.body.content
  }

  // Save to database...

  res.json({ success: true, data: post })
})

// Admin only route
app.get('/api/admin/users', protect, requireRole('admin'), (req, res) => {
  // Only admin can access
})
```

---

## 🧪 Testing

### Test Authentication Flow

```bash
# 1. Register
curl -X POST http://localhost:8088/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "password123",
    "firstName": "Test",
    "lastName": "User"
  }'

# 2. Login
curl -X POST http://localhost:8088/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# Copy the token from response

# 3. Test Backend API with JWT
TOKEN="eyJhbGciOiJIUzI1NiIs..."

curl http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer $TOKEN"

# 4. Create Post
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Post",
    "content": "Hello World"
  }'
```

---

## 🚀 Production Deployment

### 1. Environment Configuration

**Auth Service (.env):**
```env
APP_ENV=production
APP_PORT=8088
JWT_SECRET=your-production-secret-key-minimum-32-characters
CORS_ALLOWED_ORIGINS=https://suekk.com,https://admin.suekk.com
GOOGLE_REDIRECT_URL=https://auth.suekk.com/api/v1/auth/google/callback
```

**Backend Service (.env):**
```env
APP_ENV=production
APP_PORT=8080
JWT_SECRET=your-production-secret-key-minimum-32-characters  # SAME as Auth Service!
CORS_ALLOWED_ORIGINS=https://suekk.com,https://admin.suekk.com
```

### 2. CORS Configuration

Both services must allow Frontend origins.

**Go Fiber CORS:**
```go
// middleware/cors.go
func CorsMiddleware() fiber.Handler {
    allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")

    return cors.New(cors.Config{
        AllowOrigins:     allowedOrigins,
        AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
        AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
        AllowCredentials: true,
        MaxAge:           3600,
    })
}
```

### 3. Domain Setup

**DNS Records:**
```
auth.suekk.com  →  <Auth Service IP>
api.suekk.com   →  <Backend Service IP>
suekk.com       →  <Frontend IP>
```

**Nginx Reverse Proxy:**
```nginx
# Auth Service
server {
    listen 80;
    server_name auth.suekk.com;

    location / {
        proxy_pass http://localhost:8088;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
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

### 4. SSL/HTTPS Setup

```bash
# Install Certbot
sudo apt-get install certbot python3-certbot-nginx

# Get SSL certificates
sudo certbot --nginx -d auth.suekk.com
sudo certbot --nginx -d api.suekk.com
sudo certbot --nginx -d suekk.com
```

---

## 🔐 Security Checklist

- [ ] Use strong JWT_SECRET (minimum 32 characters)
- [ ] Same JWT_SECRET across Auth and Backend services
- [ ] Enable HTTPS in production
- [ ] Configure CORS properly (don't use `*` in production)
- [ ] Set secure cookie flags
- [ ] Implement rate limiting
- [ ] Use environment variables for secrets
- [ ] Regular security updates
- [ ] Monitor failed login attempts
- [ ] Implement token refresh mechanism

---

## 📚 Additional Resources

- [Authentication Flow Diagram](./MICROSERVICES_ARCHITECTURE.md)
- [User Migration Guide](./USER_MIGRATION_GUIDE.md)
- [OAuth Implementation Guide](./OAUTH_IMPLEMENTATION_GUIDE.md)

---

## 💬 Support

For issues or questions:
- Check documentation first
- Review error logs
- Verify JWT_SECRET matches
- Check CORS configuration
- Ensure database is accessible

---

**Last Updated:** 2025-11-23
**Version:** 1.0.0
