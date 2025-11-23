# Auth Service Cleanup Plan

## เป้าหมาย
ลบส่วนที่ไม่จำเป็นออกจาก starter template เพื่อให้เหลือแค่ส่วนที่เกี่ยวข้องกับ **Authentication Service** เท่านั้น

---

## ไฟล์ที่ต้องลบ

### 1. Task Management (ไม่จำเป็น)

#### Domain Layer
- [ ] `domain/models/task.go`
- [ ] `domain/repositories/task_repository.go`
- [ ] `domain/services/task_service.go`
- [ ] `domain/dto/task.go`

#### Application Layer
- [ ] `application/serviceimpl/task_service_impl.go`

#### Infrastructure Layer
- [ ] `infrastructure/postgres/task_repository_impl.go`

#### Interface Layer
- [ ] `interfaces/api/handlers/task_handler.go`
- [ ] `interfaces/api/routes/task_routes.go`

---

### 2. Job Management (ไม่จำเป็น)

#### Domain Layer
- [ ] `domain/models/job.go`
- [ ] `domain/repositories/job_repository.go`
- [ ] `domain/services/job_service.go`
- [ ] `domain/dto/job.go`

#### Application Layer
- [ ] `application/serviceimpl/job_service_impl.go`

#### Infrastructure Layer
- [ ] `infrastructure/postgres/job_repository_impl.go`

#### Interface Layer
- [ ] `interfaces/api/handlers/job_handler.go`
- [ ] `interfaces/api/routes/job_routes.go`

---

### 3. File Management (เลือกได้)

**คำถาม**: Auth Service ต้องการ file upload สำหรับ avatar หรือไม่?

#### Option A: ลบทิ้งทั้งหมด (ถ้าไม่ต้องการ file upload)
- [ ] `domain/models/file.go`
- [ ] `domain/repositories/file_repository.go`
- [ ] `domain/services/file_service.go`
- [ ] `domain/dto/file.go`
- [ ] `application/serviceimpl/file_service_impl.go`
- [ ] `infrastructure/postgres/file_repository_impl.go`
- [ ] `infrastructure/storage/bunny_storage.go`
- [ ] `interfaces/api/handlers/file_handler.go`
- [ ] `interfaces/api/routes/file_routes.go`
- [ ] `pkg/utils/path.go`

#### Option B: เก็บไว้แต่ปรับให้เฉพาะ Avatar Upload
- [ ] เก็บ `infrastructure/storage/bunny_storage.go`
- [ ] ปรับ user service ให้มี `UploadAvatar()` method
- [ ] ลบ file management ทั้งหมดออก

**คำแนะนำ**: **เลือก Option A** และใช้ URL string สำหรับ avatar (ให้ frontend upload เอง หรือใช้ OAuth provider's avatar)

---

### 4. WebSocket (ไม่จำเป็น)

- [ ] `infrastructure/websocket/websocket.go`
- [ ] `interfaces/api/websocket/websocket_handler.go`
- [ ] `interfaces/api/routes/websocket_routes.go`

**เหตุผล**: Auth Service ไม่จำเป็นต้องมี real-time communication

---

### 5. Scheduler (อาจไม่จำเป็น)

- [ ] `pkg/scheduler/scheduler.go`

**เหตุผล**: Auth Service อาจไม่ต้องการ background jobs (ยกเว้นถ้าต้องการ cleanup expired tokens/sessions)

**คำแนะนำ**: **เก็บไว้** เพราะอาจใช้สำหรับ:
- Clean up expired sessions
- Clean up expired password reset tokens
- Clean up expired email verification tokens

---

## ไฟล์ที่ต้องแก้ไข (หลังลบ)

### 1. Routes Setup

**File**: `interfaces/api/routes/routes.go`

```go
// ก่อนลบ
func SetupRoutes(app *fiber.App, h *handlers.Handlers) {
    SetupHealthRoutes(app)
    api := app.Group("/api/v1")

    SetupAuthRoutes(api, h)
    SetupUserRoutes(api, h)
    SetupTaskRoutes(api, h)      // ❌ ลบ
    SetupFileRoutes(api, h)      // ❌ ลบ
    SetupJobRoutes(api, h)       // ❌ ลบ
    SetupWebSocketRoutes(app)    // ❌ ลบ
}

// หลังลบ
func SetupRoutes(app *fiber.App, h *handlers.Handlers) {
    SetupHealthRoutes(app)
    api := app.Group("/api/v1")

    SetupAuthRoutes(api, h)
    SetupUserRoutes(api, h)
}
```

---

### 2. Handlers Struct

**File**: `interfaces/api/handlers/handlers.go`

```go
// ก่อนลบ
type Handlers struct {
    UserHandler *UserHandler
    TaskHandler *TaskHandler  // ❌ ลบ
    FileHandler *FileHandler  // ❌ ลบ
    JobHandler  *JobHandler   // ❌ ลบ
}

// หลังลบ
type Handlers struct {
    UserHandler *UserHandler
}
```

---

### 3. DI Container

**File**: `pkg/di/container.go`

ต้องลบการ initialize services/repositories ที่เกี่ยวกับ:
- Task
- Job
- File
- WebSocket

```go
// ลบส่วนนี้ออก
taskRepo := postgres.NewTaskRepository(c.db)
taskService := serviceimpl.NewTaskService(taskRepo)
// ... etc
```

---

### 4. Database Migrations

ถ้ามีไฟล์ migration หรือ auto-migration ต้องลบการสร้าง tables:
- `tasks`
- `jobs`
- `files`

---

## ไฟล์ที่ควรเก็บไว้

### Core Auth Components
```
✅ domain/models/user.go
✅ domain/repositories/user_repository.go
✅ domain/services/user_service.go
✅ domain/dto/auth.go
✅ domain/dto/user.go
✅ domain/dto/common.go
✅ domain/dto/mappers.go
✅ application/serviceimpl/user_service_impl.go
✅ infrastructure/postgres/user_repository_impl.go
✅ infrastructure/postgres/database.go
✅ infrastructure/redis/redis.go
✅ interfaces/api/handlers/user_handler.go
✅ interfaces/api/routes/auth_routes.go
✅ interfaces/api/routes/user_routes.go
✅ interfaces/api/routes/health_routes.go
✅ interfaces/api/middleware/auth_middleware.go
✅ interfaces/api/middleware/cors_middleware.go
✅ interfaces/api/middleware/logger_middleware.go
✅ interfaces/api/middleware/error_middleware.go
✅ pkg/config/config.go
✅ pkg/utils/jwt.go
✅ pkg/utils/validator.go
✅ pkg/utils/response.go
✅ pkg/di/container.go
✅ pkg/scheduler/scheduler.go (เก็บไว้สำหรับ cleanup tasks)
✅ cmd/api/main.go
```

---

## สรุปจำนวนไฟล์ที่ต้องลบ

| หมวดหมู่ | จำนวนไฟล์ |
|----------|-----------|
| Task Management | 7 ไฟล์ |
| Job Management | 7 ไฟล์ |
| File Management | 10 ไฟล์ (ถ้าเลือก Option A) |
| WebSocket | 3 ไฟล์ |
| **รวม** | **27 ไฟล์** |

---

## ขั้นตอนการ Cleanup

### Step 1: Backup
```bash
git checkout -b cleanup/remove-starter-features
git add .
git commit -m "Backup before cleanup"
```

### Step 2: ลบไฟล์ Task Management
```bash
rm domain/models/task.go
rm domain/repositories/task_repository.go
rm domain/services/task_service.go
rm domain/dto/task.go
rm application/serviceimpl/task_service_impl.go
rm infrastructure/postgres/task_repository_impl.go
rm interfaces/api/handlers/task_handler.go
rm interfaces/api/routes/task_routes.go
```

### Step 3: ลบไฟล์ Job Management
```bash
rm domain/models/job.go
rm domain/repositories/job_repository.go
rm domain/services/job_service.go
rm domain/dto/job.go
rm application/serviceimpl/job_service_impl.go
rm infrastructure/postgres/job_repository_impl.go
rm interfaces/api/handlers/job_handler.go
rm interfaces/api/routes/job_routes.go
```

### Step 4: ลบไฟล์ File Management (Option A)
```bash
rm domain/models/file.go
rm domain/repositories/file_repository.go
rm domain/services/file_service.go
rm domain/dto/file.go
rm application/serviceimpl/file_service_impl.go
rm infrastructure/postgres/file_repository_impl.go
rm infrastructure/storage/bunny_storage.go
rm interfaces/api/handlers/file_handler.go
rm interfaces/api/routes/file_routes.go
rm pkg/utils/path.go
```

### Step 5: ลบไฟล์ WebSocket
```bash
rm infrastructure/websocket/websocket.go
rm interfaces/api/websocket/websocket_handler.go
rm interfaces/api/routes/websocket_routes.go
```

### Step 6: แก้ไขไฟล์ที่เหลือ
- [ ] แก้ `interfaces/api/routes/routes.go`
- [ ] แก้ `interfaces/api/handlers/handlers.go`
- [ ] แก้ `pkg/di/container.go`
- [ ] ลบ imports ที่ไม่ใช้แล้ว

### Step 7: Test Build
```bash
go mod tidy
go build ./cmd/api
```

### Step 8: Commit
```bash
git add .
git commit -m "chore: remove starter template features (task, job, file, websocket)"
```

---

## หลังจาก Cleanup เสร็จ

โครงสร้าง project จะเหลือประมาณ:

```
gofiber-auth/
├── cmd/api/main.go
├── domain/
│   ├── models/user.go
│   ├── repositories/user_repository.go
│   ├── services/user_service.go
│   └── dto/
│       ├── auth.go
│       ├── user.go
│       └── common.go
├── application/serviceimpl/
│   └── user_service_impl.go
├── infrastructure/
│   ├── postgres/
│   │   ├── database.go
│   │   └── user_repository_impl.go
│   └── redis/redis.go
├── interfaces/api/
│   ├── handlers/
│   │   └── user_handler.go
│   ├── routes/
│   │   ├── routes.go
│   │   ├── auth_routes.go
│   │   ├── user_routes.go
│   │   └── health_routes.go
│   └── middleware/
│       ├── auth_middleware.go
│       ├── cors_middleware.go
│       ├── logger_middleware.go
│       └── error_middleware.go
├── pkg/
│   ├── config/config.go
│   ├── di/container.go
│   ├── scheduler/scheduler.go
│   └── utils/
│       ├── jwt.go
│       ├── validator.go
│       └── response.go
├── .env
├── go.mod
└── go.sum
```

**สะอาดและพร้อมสำหรับการพัฒนา Auth Service ต่อ!** 🎉

---

## คำถามก่อนเริ่ม Cleanup

1. **File Management**: ต้องการเก็บ file upload สำหรับ avatar ไหม?
   - ✅ เก็บไว้ → Option B (ปรับให้เป็น avatar upload อย่างเดียว)
   - ❌ ไม่เก็บ → Option A (ลบทิ้งทั้งหมด, ใช้ URL string)

2. **Scheduler**: ต้องการ background jobs ไหม? (สำหรับ cleanup expired tokens)
   - ✅ เก็บไว้ → จะใช้ cleanup expired sessions/tokens
   - ❌ ไม่เก็บ → ลบทิ้ง

3. **WebSocket**: แน่ใจว่าไม่ต้องการ real-time features ใน Auth Service?
   - ✅ แน่ใจ → ลบทิ้ง
   - ❌ ต้องการ → เก็บไว้ (แต่ไม่แนะนำ)

---

**คำแนะนำ**:
- File Management: **ลบทิ้ง** (Option A) ใช้ OAuth avatar หรือ URL string
- Scheduler: **เก็บไว้** สำหรับ cleanup tasks
- WebSocket: **ลบทิ้ง** Auth Service ไม่ควรมี real-time

---

**สร้างเมื่อ**: 2025-01-22
**Version**: 1.0
