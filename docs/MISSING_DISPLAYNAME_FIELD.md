# Issue: Missing `displayName` Field in UserResponse

**Date:** 2025-11-23
**Priority:** Medium
**Status:** Need to Fix

---

## 🐛 Problem

`UserResponse` DTO is missing the `displayName` field which is used throughout the application for user display purposes.

### Current Response

```json
{
  "id": "4aa10e1b-06c4-4b09-8bd9-5cea94cd3723",
  "email": "thepthai.jm@gmail.com",
  "username": "thepthai",
  "firstName": "NOTz",
  "lastName": "",
  "avatar": "",
  "role": "user",
  "isActive": true,
  "createdAt": "2025-11-12T21:21:04.112539+07:00",
  "updatedAt": "2025-11-12T21:21:04.112539+07:00"
}
```

**Missing:** `displayName` field

---

## 📊 Impact

### Backend Service
Backend Service uses `displayName` in:
- `users_cache` table (has `display_name` column)
- Webhook sync endpoint expects `displayName`
- All social features (posts, comments, messages) display `displayName`

### Frontend
Frontend expects `displayName` for:
- User profile display
- Post author names
- Comment author names
- Chat messages
- Notifications

### Current Workaround
Frontend must manually concatenate:
```javascript
const displayName = `${user.firstName} ${user.lastName}`.trim();
```

This creates inconsistency and extra work.

---

## ✅ Proposed Solution

### Option 1: Add `displayName` to UserResponse (Recommended)

**File:** `domain/dto/user.go`

```go
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	DisplayName string    `json:"displayName"`  // ← ADD THIS
	Avatar      string    `json:"avatar"`
	Role        string    `json:"role"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
```

**Mapper:** `domain/dto/mappers.go`

```go
func UserToUserResponse(user *models.User) *UserResponse {
	if user == nil {
		return nil
	}

	// Generate displayName from firstName + lastName
	displayName := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if displayName == "" {
		displayName = user.Username // Fallback to username
	}

	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DisplayName: displayName,  // ← ADD THIS
		Avatar:      user.Avatar,
		Role:        user.Role,
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
```

---

### Option 2: Add DisplayName Column to Database (Alternative)

**Migration:** Add `display_name` column to `users` table

```sql
ALTER TABLE users ADD COLUMN display_name VARCHAR(100);

-- Update existing records
UPDATE users SET display_name = TRIM(first_name || ' ' || last_name);

-- Set default for empty display_name
UPDATE users SET display_name = username WHERE display_name = '' OR display_name IS NULL;
```

**Model:** Update `models.User`

```go
type User struct {
	// ... existing fields
	FirstName   string `gorm:"size:50;not null"`
	LastName    string `gorm:"size:50;not null"`
	DisplayName string `gorm:"size:100;not null"` // ← ADD THIS
	// ... rest of fields
}
```

**Service:** Auto-generate on create/update

```go
func (s *UserService) Register(ctx context.Context, req *dto.RegisterRequest) (*models.User, error) {
	user := &models.User{
		Email:       req.Email,
		Username:    req.Username,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DisplayName: strings.TrimSpace(req.FirstName + " " + req.LastName), // ← ADD THIS
		// ...
	}
	// ...
}
```

---

## 🎯 Recommended Approach

**Use Option 1** (Add to DTO only) because:
- ✅ Quick fix, no database migration needed
- ✅ Works immediately
- ✅ Consistent with current architecture
- ✅ Can be done in 5 minutes

**Option 2** can be done later if needed for:
- Search optimization
- Indexing
- User-customizable display names

---

## 📝 Implementation Steps

1. **Update UserResponse DTO**
   - File: `domain/dto/user.go`
   - Add `DisplayName string` field

2. **Update Mapper**
   - File: `domain/dto/mappers.go` (or wherever UserToUserResponse is)
   - Add logic to generate displayName

3. **Update Webhook Sync**
   - Ensure webhook payload includes `displayName`
   - Backend Service expects this field

4. **Test**
   - Login response should include displayName
   - Profile endpoint should include displayName
   - OAuth callback should include displayName

---

## 🧪 Testing

### Before Fix
```bash
curl http://localhost:8088/api/v1/auth/login \
  -d '{"email":"user@example.com","password":"password"}'

# Response (missing displayName):
{
  "firstName": "John",
  "lastName": "Doe"
  // ❌ no displayName
}
```

### After Fix
```bash
curl http://localhost:8088/api/v1/auth/login \
  -d '{"email":"user@example.com","password":"password"}'

# Response (with displayName):
{
  "firstName": "John",
  "lastName": "Doe",
  "displayName": "John Doe"  // ✅ added
}
```

---

## 🔗 Related Issues

- Backend Service `users_cache` table has `display_name` column
- Webhook sync expects `displayName` field
- Frontend MIGRATION GUIDE assumes `displayName` exists
- All user-facing features need display names

---

## ⏰ Timeline

**Priority:** Medium
**Estimated Time:** 10 minutes
**Impact:** Frontend & Backend compatibility

---

**Contact:** Backend Team
**Document Created:** 2025-11-23
