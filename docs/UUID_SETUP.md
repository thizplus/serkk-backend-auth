# UUID Generation Options

## Option 1: PostgreSQL UUID-OSSP Extension

### Enable Extension

```sql
-- Connect to your database
\c gofiber_auth

-- Enable uuid-ossp extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Verify
SELECT uuid_generate_v4();
```

### Update Table Default

```sql
-- Alter users table to use PostgreSQL UUID generation
ALTER TABLE users 
ALTER COLUMN id SET DEFAULT uuid_generate_v4();

-- Alter oauth_providers table
ALTER TABLE oauth_providers 
ALTER COLUMN id SET DEFAULT uuid_generate_v4();
```

### Update Go Models

```go
// domain/models/user.go
type User struct {
    ID uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
    // ... rest of fields
}

// Remove BeforeCreate hook if using PostgreSQL default
```

## Option 2: PostgreSQL pgcrypto Extension (gen_random_uuid)

### Enable Extension

```sql
-- Connect to database
\c gofiber_auth

-- Enable pgcrypto extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Verify
SELECT gen_random_uuid();
```

### Update Table Default

```sql
ALTER TABLE users 
ALTER COLUMN id SET DEFAULT gen_random_uuid();

ALTER TABLE oauth_providers 
ALTER COLUMN id SET DEFAULT gen_random_uuid();
```

### Update Go Models

```go
// domain/models/user.go
type User struct {
    ID uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    // ... rest of fields
}
```

## Option 3: Go Application (Current Solution - Recommended)

**Pros:**
- ✅ No PostgreSQL extensions needed
- ✅ Works with any database
- ✅ Portable across environments
- ✅ Already implemented

**Current Implementation:**

```go
type User struct {
    ID uuid.UUID `gorm:"primaryKey;type:uuid"`
    // ...
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.ID == uuid.Nil {
        u.ID = uuid.New()
    }
    return nil
}
```

## Comparison

| Feature | uuid-ossp | pgcrypto | Go (Current) |
|---------|-----------|----------|--------------|
| PostgreSQL Dependency | ✓ | ✓ | ✗ |
| Performance | Fast | Fast | Fast |
| Portability | Low | Low | High |
| Setup Complexity | Medium | Medium | Low |
| Already Working | - | - | ✅ |

## Recommendation

**Keep the current Go solution** unless you have specific requirements for database-level UUID generation.

The current solution:
- ✅ Already working
- ✅ No database dependencies
- ✅ Easy to test and maintain
- ✅ Works across different PostgreSQL versions
