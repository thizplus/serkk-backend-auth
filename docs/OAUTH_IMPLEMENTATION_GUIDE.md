# OAuth Implementation Guide

## เป้าหมาย

เพิ่มระบบ OAuth Authentication สำหรับ Auth Service รองรับ:
1. ✅ **Google OAuth**
2. ✅ **Facebook OAuth**
3. ✅ **LINE OAuth**
4. ✅ **Email/Password** (มีอยู่แล้ว)

---

## OAuth Providers ที่รองรับ

### 1. Google OAuth 2.0
- **Use Case**: Login ด้วย Google Account
- **Scope**: `email`, `profile`
- **ข้อมูลที่ได้**: email, name, picture
- **Docs**: https://developers.google.com/identity/protocols/oauth2

### 2. Facebook Login
- **Use Case**: Login ด้วย Facebook Account
- **Scope**: `email`, `public_profile`
- **ข้อมูลที่ได้**: email, name, picture
- **Docs**: https://developers.facebook.com/docs/facebook-login

### 3. LINE Login
- **Use Case**: Login ด้วย LINE Account (ยอดนิยมในไทย)
- **Scope**: `profile`, `openid`, `email`
- **ข้อมูลที่ได้**: userId, displayName, pictureUrl, email
- **Docs**: https://developers.line.biz/en/docs/line-login/

---

## Database Schema Changes

### 1. Update Users Table

```sql
ALTER TABLE users ADD COLUMN is_oauth_user BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN oauth_provider VARCHAR(50);     -- 'google', 'facebook', 'line'
ALTER TABLE users ADD COLUMN oauth_id VARCHAR(255);
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN display_name VARCHAR(255);      -- สำหรับ OAuth users
ALTER TABLE users ADD COLUMN last_login_at TIMESTAMP;

-- Indexes
CREATE INDEX idx_users_oauth_provider ON users(oauth_provider);
CREATE INDEX idx_users_oauth_id ON users(oauth_id);
CREATE INDEX idx_users_email_verified ON users(email_verified);

-- Make password nullable (เพราะ OAuth users ไม่มี password)
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;
```

### 2. Create OAuthProviders Table

```sql
CREATE TABLE oauth_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,              -- 'google', 'facebook', 'line'
    provider_id VARCHAR(255) NOT NULL,          -- OAuth provider's user ID
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMP,
    profile_data JSONB,                         -- เก็บข้อมูล profile ดิบจาก OAuth provider
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(provider, provider_id)
);

CREATE INDEX idx_oauth_providers_user_id ON oauth_providers(user_id);
CREATE INDEX idx_oauth_providers_provider ON oauth_providers(provider);
```

---

## API Endpoints

### OAuth Endpoints

```
# Google OAuth
GET    /api/v1/auth/google              # Get Google OAuth URL
GET    /api/v1/auth/google/callback     # Google OAuth callback

# Facebook OAuth
GET    /api/v1/auth/facebook            # Get Facebook OAuth URL
GET    /api/v1/auth/facebook/callback   # Facebook OAuth callback

# LINE OAuth
GET    /api/v1/auth/line                # Get LINE OAuth URL
GET    /api/v1/auth/line/callback       # LINE OAuth callback

# Standard Auth (existing)
POST   /api/v1/auth/register            # Email/Password registration
POST   /api/v1/auth/login               # Email/Password login
```

---

## OAuth Flow

### Google/Facebook/LINE Flow (เหมือนกัน)

```
┌────────┐      ┌──────────┐      ┌──────────────┐      ┌──────────┐
│ Client │      │ Frontend │      │ Auth Service │      │  OAuth   │
│        │      │          │      │              │      │ Provider │
└───┬────┘      └────┬─────┘      └──────┬───────┘      └────┬─────┘
    │                │                   │                   │
    │ Click Login    │                   │                   │
    ├───────────────>│                   │                   │
    │                │ GET /auth/google  │                   │
    │                ├──────────────────>│                   │
    │                │                   │                   │
    │                │ {auth_url}        │                   │
    │                │<──────────────────┤                   │
    │                │                   │                   │
    │ Redirect       │                   │                   │
    │<───────────────┤                   │                   │
    │                                    │                   │
    │ OAuth Login Page                   │                   │
    ├───────────────────────────────────────────────────────>│
    │                                    │                   │
    │ User Authorizes                    │                   │
    │                                    │                   │
    │ Redirect with code                 │                   │
    │<───────────────────────────────────────────────────────┤
    │                                    │                   │
    │ GET /auth/google/callback?code=xxx │                   │
    ├───────────────────────────────────>│                   │
    │                                    │ Exchange code     │
    │                                    ├──────────────────>│
    │                                    │                   │
    │                                    │ Access Token      │
    │                                    │<──────────────────┤
    │                                    │                   │
    │                                    │ Get User Info     │
    │                                    ├──────────────────>│
    │                                    │                   │
    │                                    │ {email, name,...} │
    │                                    │<──────────────────┤
    │                                    │                   │
    │  {jwt_token, user}                 │                   │
    │<───────────────────────────────────┤                   │
    │                                    │                   │
```

---

## Implementation Steps

### Step 1: Update User Model

**File**: `domain/models/user.go`

```go
type User struct {
    ID            uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Email         string     `gorm:"uniqueIndex;not null"`
    Username      string     `gorm:"uniqueIndex;not null"`
    Password      *string    `gorm:"default:null"`  // Nullable สำหรับ OAuth users
    FirstName     string
    LastName      string
    DisplayName   string     // สำหรับ OAuth (LINE, Facebook)
    Avatar        string
    Role          string     `gorm:"default:'user'"`
    IsActive      bool       `gorm:"default:true"`

    // OAuth fields
    IsOAuthUser   bool       `gorm:"default:false"`
    OAuthProvider string     // 'google', 'facebook', 'line'
    OAuthID       string     `gorm:"index"`
    EmailVerified bool       `gorm:"default:false"`
    LastLoginAt   *time.Time

    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

---

### Step 2: Create OAuthProvider Model

**File**: `domain/models/oauth_provider.go`

```go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/datatypes"
)

type OAuthProvider struct {
    ID             uuid.UUID      `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    UserID         uuid.UUID      `gorm:"not null;index"`
    Provider       string         `gorm:"not null;size:50"`  // 'google', 'facebook', 'line'
    ProviderID     string         `gorm:"not null;size:255"` // OAuth provider's user ID
    AccessToken    string         `gorm:"type:text"`
    RefreshToken   string         `gorm:"type:text"`
    TokenExpiresAt *time.Time
    ProfileData    datatypes.JSON `gorm:"type:jsonb"` // Raw profile data from provider
    CreatedAt      time.Time
    UpdatedAt      time.Time

    // Relationship
    User           User           `gorm:"foreignKey:UserID"`
}

func (OAuthProvider) TableName() string {
    return "oauth_providers"
}
```

---

### Step 3: Create OAuth DTOs

**File**: `domain/dto/oauth.go`

```go
package dto

type OAuthURLResponse struct {
    AuthURL string `json:"auth_url"`
}

type OAuthCallbackRequest struct {
    Code  string `json:"code" validate:"required"`
    State string `json:"state"`
}

type OAuthLoginResponse struct {
    AccessToken  string       `json:"access_token"`
    TokenType    string       `json:"token_type"`
    ExpiresIn    int          `json:"expires_in"`
    User         UserResponse `json:"user"`
    IsNewUser    bool         `json:"is_new_user"`
}

// Provider-specific user info
type GoogleUserInfo struct {
    ID            string `json:"id"`
    Email         string `json:"email"`
    VerifiedEmail bool   `json:"verified_email"`
    Name          string `json:"name"`
    GivenName     string `json:"given_name"`
    FamilyName    string `json:"family_name"`
    Picture       string `json:"picture"`
}

type FacebookUserInfo struct {
    ID      string `json:"id"`
    Email   string `json:"email"`
    Name    string `json:"name"`
    Picture struct {
        Data struct {
            URL string `json:"url"`
        } `json:"data"`
    } `json:"picture"`
}

type LINEUserInfo struct {
    UserID      string `json:"userId"`
    DisplayName string `json:"displayName"`
    PictureURL  string `json:"pictureUrl"`
    StatusMessage string `json:"statusMessage"`
}

type LINEIDToken struct {
    Email         string `json:"email"`
    EmailVerified bool   `json:"email_verified"`
}
```

---

### Step 4: OAuth Configuration

**File**: `pkg/config/config.go`

เพิ่มใน Config struct:

```go
type OAuthConfig struct {
    // Google
    GoogleClientID     string
    GoogleClientSecret string
    GoogleRedirectURL  string

    // Facebook
    FacebookClientID     string
    FacebookClientSecret string
    FacebookRedirectURL  string

    // LINE
    LINEClientID     string
    LINEClientSecret string
    LINERedirectURL  string
}

type Config struct {
    // ... existing fields ...
    OAuth OAuthConfig
}
```

**Environment Variables**:

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

### Step 5: OAuth Service Interface

**File**: `domain/services/oauth_service.go`

```go
package services

import (
    "context"
    "gofiber-template/domain/dto"
    "gofiber-template/domain/models"
)

type OAuthService interface {
    // Google
    GetGoogleAuthURL(state string) string
    HandleGoogleCallback(ctx context.Context, code string) (*models.User, string, bool, error)

    // Facebook
    GetFacebookAuthURL(state string) string
    HandleFacebookCallback(ctx context.Context, code string) (*models.User, string, bool, error)

    // LINE
    GetLINEAuthURL(state string) string
    HandleLINECallback(ctx context.Context, code string) (*models.User, string, bool, error)

    // Common
    FindOrCreateOAuthUser(ctx context.Context, provider string, userInfo interface{}) (*models.User, bool, error)
}
```

---

### Step 6: OAuth Repository

**File**: `domain/repositories/oauth_repository.go`

```go
package repositories

import (
    "context"
    "gofiber-template/domain/models"
    "github.com/google/uuid"
)

type OAuthRepository interface {
    Create(ctx context.Context, oauth *models.OAuthProvider) error
    FindByProviderAndProviderID(ctx context.Context, provider, providerID string) (*models.OAuthProvider, error)
    FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.OAuthProvider, error)
    Update(ctx context.Context, oauth *models.OAuthProvider) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

---

## Dependencies

```bash
go get golang.org/x/oauth2
go get google.golang.org/api/oauth2/v2
```

---

## Security Considerations

### 1. State Parameter
- ใช้ random state เพื่อป้องกัน CSRF attacks
- Validate state ใน callback

### 2. Token Storage
- เก็บ access_token และ refresh_token ใน database (encrypted ถ้าเป็นไปได้)
- ใช้สำหรับ refresh user data

### 3. Email Verification
- OAuth providers ส่วนใหญ่ verify email แล้ว
- Set `email_verified = true` สำหรับ OAuth users

### 4. Password Field
- OAuth users ไม่มี password
- ต้องทำให้ password field nullable
- Validate ว่าถ้า `is_oauth_user = true` แล้วไม่ต้องมี password

---

## Testing Checklist

### Google OAuth
- [ ] สามารถ login ด้วย Google ได้
- [ ] สร้าง user ใหม่ถ้ายังไม่เคยมี
- [ ] Login user เดิมถ้ามีอยู่แล้ว
- [ ] ได้ email, name, picture จาก Google
- [ ] JWT token ถูกสร้างและส่งกลับ

### Facebook OAuth
- [ ] สามารถ login ด้วย Facebook ได้
- [ ] สร้าง user ใหม่ถ้ายังไม่เคยมี
- [ ] Login user เดิมถ้ามีอยู่แล้ว
- [ ] ได้ email, name, picture จาก Facebook
- [ ] JWT token ถูกสร้างและส่งกลับ

### LINE OAuth
- [ ] สามารถ login ด้วย LINE ได้
- [ ] สร้าง user ใหม่ถ้ายังไม่เคยมี
- [ ] Login user เดิมถ้ามีอยู่แล้ว
- [ ] ได้ displayName, pictureUrl, email จาก LINE
- [ ] JWT token ถูกสร้างและส่งกลับ

### Edge Cases
- [ ] Handle email ซ้ำ (OAuth user vs regular user)
- [ ] Handle OAuth provider ไม่ส่ง email
- [ ] Handle network errors
- [ ] Handle invalid codes

---

## Next Steps

1. Update User model ให้รองรับ OAuth fields
2. สร้าง OAuthProvider model
3. Implement OAuth repositories
4. Implement OAuth service (Google, Facebook, LINE)
5. สร้าง OAuth handlers
6. เพิ่ม OAuth routes
7. ทดสอบแต่ละ provider
8. Update documentation

---

**สร้างเมื่อ**: 2025-01-22
**Version**: 1.0
**OAuth Providers**: Google, Facebook, LINE
