package serviceimpl

import (
	"context"
	"errors"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
	"gofiber-template/domain/services"
	"gofiber-template/pkg/contextutil"
	"gofiber-template/pkg/logger"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserServiceImpl struct {
	userRepo    repositories.UserRepository
	jwtSecret   string
	syncService *SyncService
}

func NewUserService(userRepo repositories.UserRepository, jwtSecret string, syncService *SyncService) services.UserService {
	return &UserServiceImpl{
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		syncService: syncService,
	}
}

func (s *UserServiceImpl) Register(ctx context.Context, req *dto.CreateUserRequest) (*models.User, error) {
	startTime := time.Now()
	requestID := contextutil.GetRequestID(ctx)
	log := logger.GetLogger()

	log.Info("User registration started", map[string]interface{}{
		"request_id": requestID,
		"action":     "register",
		"email":      req.Email,
		"username":   req.Username,
	})

	existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		log.Warn("Registration failed: email already exists", map[string]interface{}{
			"request_id": requestID,
			"action":     "register",
			"email":      req.Email,
		})
		return nil, errors.New("email already exists")
	}

	existingUser, _ = s.userRepo.GetByUsername(ctx, req.Username)
	if existingUser != nil {
		log.Warn("Registration failed: username already exists", map[string]interface{}{
			"request_id": requestID,
			"action":     "register",
			"username":   req.Username,
		})
		return nil, errors.New("username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Password hashing failed", map[string]interface{}{
			"request_id": requestID,
			"action":     "register",
			"error":      err.Error(),
		})
		return nil, err
	}

	passwordStr := string(hashedPassword)
	user := &models.User{
		ID:          uuid.New(),
		Email:       req.Email,
		Username:    req.Username,
		Password:    &passwordStr,
		DisplayName: req.DisplayName,
		Role:        "user",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		log.Error("User creation failed", map[string]interface{}{
			"request_id": requestID,
			"action":     "register",
			"error":      err.Error(),
		})
		return nil, err
	}

	duration := time.Since(startTime).Milliseconds()
	log.Info("User registered successfully", map[string]interface{}{
		"request_id":  requestID,
		"action":      "register",
		"user_id":     user.ID.String(),
		"username":    user.Username,
		"duration_ms": duration,
	})

	// Sync to backend (async with context)
	go s.syncService.SyncUserWithRetry(ctx, user, "created")

	return user, nil
}

func (s *UserServiceImpl) Login(ctx context.Context, req *dto.LoginRequest) (string, *models.User, error) {
	startTime := time.Now()
	requestID := contextutil.GetRequestID(ctx)
	log := logger.GetLogger()

	log.Info("User login attempt", map[string]interface{}{
		"request_id": requestID,
		"action":     "login",
		"email":      req.Email,
	})

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		log.Warn("Login failed: user not found", map[string]interface{}{
			"request_id": requestID,
			"action":     "login",
			"email":      req.Email,
		})
		return "", nil, errors.New("invalid email or password")
	}

	if !user.IsActive {
		log.Warn("Login failed: account disabled", map[string]interface{}{
			"request_id": requestID,
			"action":     "login",
			"user_id":    user.ID.String(),
			"email":      req.Email,
		})
		return "", nil, errors.New("account is disabled")
	}

	if user.Password == nil {
		log.Warn("Login failed: OAuth-only account", map[string]interface{}{
			"request_id": requestID,
			"action":     "login",
			"user_id":    user.ID.String(),
			"email":      req.Email,
		})
		return "", nil, errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password))
	if err != nil {
		log.Warn("Login failed: invalid password", map[string]interface{}{
			"request_id": requestID,
			"action":     "login",
			"user_id":    user.ID.String(),
			"email":      req.Email,
		})
		return "", nil, errors.New("invalid email or password")
	}

	token, err := s.GenerateJWT(user)
	if err != nil {
		log.Error("JWT generation failed", map[string]interface{}{
			"request_id": requestID,
			"action":     "login",
			"user_id":    user.ID.String(),
			"error":      err.Error(),
		})
		return "", nil, err
	}

	duration := time.Since(startTime).Milliseconds()
	log.Info("User logged in successfully", map[string]interface{}{
		"request_id":  requestID,
		"action":      "login",
		"user_id":     user.ID.String(),
		"username":    user.Username,
		"duration_ms": duration,
	})

	return token, user, nil
}

func (s *UserServiceImpl) GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *UserServiceImpl) UpdateProfile(ctx context.Context, userID uuid.UUID, req *dto.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	user.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, userID, user)
	if err != nil {
		return nil, err
	}

	// Sync to backend (async with context)
	go s.syncService.SyncUserWithRetry(ctx, user, "updated")

	return user, nil
}

func (s *UserServiceImpl) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// Get user first for sync
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	err = s.userRepo.Delete(ctx, userID)
	if err != nil {
		return err
	}

	// Sync deletion to backend (async with context)
	go s.syncService.SyncUserWithRetry(ctx, user, "deleted")

	return nil
}

func (s *UserServiceImpl) ListUsers(ctx context.Context, offset, limit int) ([]*models.User, int64, error) {
	users, err := s.userRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, count, nil
}

func (s *UserServiceImpl) GenerateJWT(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *UserServiceImpl) ValidateJWT(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return nil, errors.New("invalid token claims")
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, errors.New("invalid user ID in token")
		}

		user, err := s.userRepo.GetByID(context.Background(), userID)
		if err != nil {
			return nil, errors.New("user not found")
		}

		return user, nil
	}

	return nil, errors.New("invalid token")
}
