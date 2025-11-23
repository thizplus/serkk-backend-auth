# OAuth Implementation - Remaining Fixes

## สถานะ: 90% เสร็จสมบูรณ์ ✅

OAuth implementation สำหรับ Google, Facebook, และ LINE เกือบเสร็จแล้ว
เหลือแก้ไข bugs เล็กน้อยเท่านั้น

---

## ✅ สิ่งที่เสร็จแล้ว

1. ✅ Models (User, OAuthProvider)
2. ✅ Repositories (OAuth repository)
3. ✅ DTOs (OAuth DTOs)
4. ✅ Configuration (OAuth config)
5. ✅ Service layer (OAuth service สำหรับ 3 providers)
6. ✅ Handlers (OAuth handlers)
7. ✅ Routes (OAuth routes)
8. ✅ DI Container
9. ✅ Dependencies installed

---

## 🔧 สิ่งที่ต้องแก้ (Remaining Bugs)

### 1. Fix `oauth_service_impl.go`

**File**: `application/serviceimpl/oauth_service_impl.go`

#### Issue 1: Unused imports
```go
// ลบ imports ที่ไม่ใช้
"io"         // ❌ ลบ
"net/http"   // ❌ ลบ
```

#### Issue 2: Duplicate oauth2 import
```go
// เปลี่ยนจาก
import (
    ...
    "google.golang.org/api/oauth2/v2"
    googleOAuth2 "google.golang.org/api/oauth2/v2"  // ❌ ซ้ำ
)

// เป็น
import (
    ...
    googleOAuth2 "google.golang.org/api/oauth2/v2"  // ✅ เก็บแค่อันนี้
)
```

#### Issue 3: VerifiedEmail type mismatch (Line 116)
```go
// เปลี่ยนจาก
googleUserInfo := &dto.GoogleUserInfo{
    ...
    VerifiedEmail: userInfo.VerifiedEmail,  // ❌ *bool
}

// เป็น
googleUserInfo := &dto.GoogleUserInfo{
    ...
    VerifiedEmail: userInfo.VerifiedEmail != nil && *userInfo.VerifiedEmail,  // ✅ bool
}
```

#### Issue 4: userRepo.Update signature (Lines 264, 364)
```go
// เปลี่ยนจาก
s.userRepo.Update(ctx, user)  // ❌ ไม่ถูกต้อง

// เป็น
s.userRepo.Update(ctx, user.ID, user)  // ✅ ถูกต้อง
```

#### Issue 5: FindByEmail method (Line 307)
```go
// เปลี่ยนจาก
existingUser, err := s.userRepo.FindByEmail(ctx, email)  // ❌ method ไม่มี
if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {  // ❌ constant ไม่มี
    ...
}

// เป็น
existingUser, err := s.userRepo.GetByEmail(ctx, email)  // ✅ ใช้ GetByEmail
if err != nil && existingUser == nil {  // ✅ check nil แทน
    // ไม่มี user
}
```

**Full fix for Line 307-313:**
```go
// Check if email already exists
existingUser, _ := s.userRepo.GetByEmail(ctx, email)

var user *models.User
isNewUser := false

if existingUser != nil {
    // Link OAuth to existing user
    user = existingUser
} else {
    // Create new user
    ...
}
```

---

### 2. Fix `oauth_handler.go`

**File**: `interfaces/api/handlers/oauth_handler.go`

#### Issue: utils.ErrorResponse signature

ตรวจสอบ signature ของ `utils.ErrorResponse()` ใน `pkg/utils/response.go`

**ถ้า signature คือ:**
```go
func ErrorResponse(c *fiber.Ctx, status int, message string, err error) error
```

**แก้ทุก calls (Lines 55, 60, 105, 110, 155, 160):**
```go
// เปลี่ยนจาก
return utils.ErrorResponse(c, fiber.StatusBadRequest, "Missing authorization code")

// เป็น
return utils.ErrorResponse(c, fiber.StatusBadRequest, "Missing authorization code", nil)
```

**หรือถ้า signature คือ:**
```go
func ErrorResponse(c *fiber.Ctx, status int, message string) error
```

ก็ไม่ต้องแก้อะไร (ใช้ได้เลย)

---

### 3. Add `FindByEmail` method to User Repository (Optional)

**ถ้าต้องการให้ consistent:**

**File**: `domain/repositories/user_repository.go`

เพิ่ม method:
```go
FindByEmail(ctx context.Context, email string) (*models.User, error)
```

**File**: `infrastructure/postgres/user_repository_impl.go`

เพิ่ม implementation:
```go
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
    var user models.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil  // Return nil instead of error
        }
        return nil, err
    }
    return &user, nil
}
```

**หรือใช้ `GetByEmail` ที่มีอยู่แล้วก็ได้** (แนะนำ)

---

## 📝 Step-by-Step Fix Guide

### Step 1: Fix oauth_service_impl.go

```bash
# Edit application/serviceimpl/oauth_service_impl.go
```

1. ลบ imports: `"io"`, `"net/http"`
2. ลบ duplicate import: `"google.golang.org/api/oauth2/v2"` (เก็บแค่ `googleOAuth2`)
3. Line 116: แก้ `VerifiedEmail: userInfo.VerifiedEmail != nil && *userInfo.VerifiedEmail`
4. Line 264: แก้ `s.userRepo.Update(ctx, user.ID, user)`
5. Line 307: แก้ `s.userRepo.GetByEmail(ctx, email)` และลบ `repositories.ErrUserNotFound`
6. Line 364: แก้ `s.userRepo.Update(ctx, user.ID, user)`

### Step 2: Fix oauth_handler.go (ถ้าจำเป็น)

```bash
# Edit interfaces/api/handlers/oauth_handler.go
```

ตรวจสอบ `utils.ErrorResponse` signature และแก้ตาม

### Step 3: Test Build

```bash
cd "D:\Admin\Desktop\MY PROJECT\__serkk\gofiber-auth"
go mod tidy
go build ./cmd/api
```

### Step 4: Test Run

```bash
go run cmd/api/main.go
```

ควรเห็น output:
```
✓ Configuration loaded
✓ Database connected
✓ Database migrated
✓ Redis connected
✓ Repositories initialized
✓ Services initialized
✓ Event scheduler started
🚀 Server starting on port 4000
```

---

## 🧪 Testing OAuth Flow

### Test Google OAuth

```bash
# 1. Get auth URL
curl http://localhost:4000/api/v1/auth/google

# Response:
# {
#   "auth_url": "https://accounts.google.com/o/oauth2/auth?..."
# }

# 2. Open URL in browser
# 3. Authorize
# 4. Get redirected to callback with code
# 5. Exchange code for token (automatic in callback handler)
```

### Test Facebook OAuth

```bash
curl http://localhost:4000/api/v1/auth/facebook
```

### Test LINE OAuth

```bash
curl http://localhost:4000/api/v1/auth/line
```

---

## 📋 Environment Variables Required

Add to `.env`:

```env
# Google OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:4000/api/v1/auth/google/callback

# Facebook OAuth
FACEBOOK_CLIENT_ID=your-facebook-app-id
FACEBOOK_CLIENT_SECRET=your-facebook-app-secret
FACEBOOK_REDIRECT_URL=http://localhost:4000/api/v1/auth/facebook/callback

# LINE OAuth
LINE_CLIENT_ID=your-line-channel-id
LINE_CLIENT_SECRET=your-line-channel-secret
LINE_REDIRECT_URL=http://localhost:4000/api/v1/auth/line/callback
```

---

## 🎯 API Endpoints Available

### Standard Auth
```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
```

### OAuth - Google
```
GET    /api/v1/auth/google                # Get auth URL
GET    /api/v1/auth/google/callback       # Callback (automatic)
```

### OAuth - Facebook
```
GET    /api/v1/auth/facebook              # Get auth URL
GET    /api/v1/auth/facebook/callback     # Callback (automatic)
```

### OAuth - LINE
```
GET    /api/v1/auth/line                  # Get auth URL
GET    /api/v1/auth/line/callback         # Callback (automatic)
```

---

## 📊 Progress

- [x] Database models (100%)
- [x] Repositories (100%)
- [x] DTOs (100%)
- [x] Configuration (100%)
- [x] Service layer (95% - minor bugs)
- [x] Handlers (95% - minor bugs)
- [x] Routes (100%)
- [x] DI Container (100%)
- [ ] Bug fixes (pending - 5%)
- [ ] Testing (pending)

**Overall: 90% Complete** 🎉

---

**สร้างเมื่อ**: 2025-01-22
**Status**: Ready for final bug fixes
**Estimated time to complete**: 10-15 minutes
