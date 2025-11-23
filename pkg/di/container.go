package di

import (
	"context"
	"log"
	"gofiber-template/application/serviceimpl"
	"gofiber-template/domain/repositories"
	"gofiber-template/domain/services"
	"gofiber-template/infrastructure/postgres"
	"gofiber-template/infrastructure/redis"
	"gofiber-template/interfaces/api/handlers"
	"gofiber-template/pkg/config"
	"gofiber-template/pkg/scheduler"
	"gorm.io/gorm"
)

type Container struct {
	// Configuration
	Config *config.Config

	// Infrastructure
	DB             *gorm.DB
	RedisClient    *redis.RedisClient
	EventScheduler scheduler.EventScheduler

	// Repositories
	UserRepository  repositories.UserRepository
	OAuthRepository repositories.OAuthRepository

	// Services
	SyncService  *serviceimpl.SyncService
	UserService  services.UserService
	OAuthService services.OAuthService
}

func NewContainer() *Container {
	return &Container{}
}

func (c *Container) Initialize() error {
	if err := c.initConfig(); err != nil {
		return err
	}

	if err := c.initInfrastructure(); err != nil {
		return err
	}

	if err := c.initRepositories(); err != nil {
		return err
	}

	if err := c.initServices(); err != nil {
		return err
	}

	if err := c.initScheduler(); err != nil {
		return err
	}

	return nil
}

func (c *Container) initConfig() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	c.Config = cfg
	log.Println("✓ Configuration loaded")
	return nil
}

func (c *Container) initInfrastructure() error {
	// Initialize Database
	dbConfig := postgres.DatabaseConfig{
		Host:     c.Config.Database.Host,
		Port:     c.Config.Database.Port,
		User:     c.Config.Database.User,
		Password: c.Config.Database.Password,
		DBName:   c.Config.Database.DBName,
		SSLMode:  c.Config.Database.SSLMode,
	}

	db, err := postgres.NewDatabase(dbConfig)
	if err != nil {
		return err
	}
	c.DB = db
	log.Println("✓ Database connected")

	// Run migrations
	if err := postgres.Migrate(db); err != nil {
		return err
	}
	log.Println("✓ Database migrated")

	// Initialize Redis
	redisConfig := redis.RedisConfig{
		Host:     c.Config.Redis.Host,
		Port:     c.Config.Redis.Port,
		Password: c.Config.Redis.Password,
		DB:       c.Config.Redis.DB,
	}
	c.RedisClient = redis.NewRedisClient(redisConfig)

	// Test Redis connection
	if err := c.RedisClient.Ping(context.Background()); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	} else {
		log.Println("✓ Redis connected")
	}

	return nil
}

func (c *Container) initRepositories() error {
	c.UserRepository = postgres.NewUserRepository(c.DB)
	c.OAuthRepository = postgres.NewOAuthRepository(c.DB)
	log.Println("✓ Repositories initialized")
	return nil
}

func (c *Container) initServices() error {
	// Initialize SyncService first
	c.SyncService = serviceimpl.NewSyncService()

	// Initialize UserService and OAuthService with SyncService
	c.UserService = serviceimpl.NewUserService(c.UserRepository, c.Config.JWT.Secret, c.SyncService)
	c.OAuthService = serviceimpl.NewOAuthService(c.UserRepository, c.OAuthRepository, c.UserService, c.SyncService, c.Config)
	log.Println("✓ Services initialized")
	return nil
}

func (c *Container) initScheduler() error {
	c.EventScheduler = scheduler.NewEventScheduler()

	// Start the scheduler (will be used for cleanup tasks in the future)
	c.EventScheduler.Start()
	log.Println("✓ Event scheduler started")

	return nil
}

func (c *Container) Cleanup() error {
	log.Println("Starting cleanup...")

	// Stop scheduler
	if c.EventScheduler != nil {
		if c.EventScheduler.IsRunning() {
			c.EventScheduler.Stop()
			log.Println("✓ Event scheduler stopped")
		} else {
			log.Println("✓ Event scheduler was already stopped")
		}
	}

	// Close Redis connection
	if c.RedisClient != nil {
		if err := c.RedisClient.Close(); err != nil {
			log.Printf("Warning: Failed to close Redis connection: %v", err)
		} else {
			log.Println("✓ Redis connection closed")
		}
	}

	// Close database connection
	if c.DB != nil {
		sqlDB, err := c.DB.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Warning: Failed to close database connection: %v", err)
			} else {
				log.Println("✓ Database connection closed")
			}
		}
	}

	log.Println("✓ Cleanup completed")
	return nil
}

func (c *Container) GetServices() services.UserService {
	return c.UserService
}

func (c *Container) GetConfig() *config.Config {
	return c.Config
}

func (c *Container) GetHandlerServices() *handlers.Services {
	return &handlers.Services{
		UserService:  c.UserService,
		OAuthService: c.OAuthService,
	}
}