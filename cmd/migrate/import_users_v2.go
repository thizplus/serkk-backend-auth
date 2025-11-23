package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// User model - matches existing table structure
type User struct {
	ID            uuid.UUID  `gorm:"primaryKey;type:uuid;column:id"`
	Email         string     `gorm:"uniqueIndex;not null;column:email"`
	Username      string     `gorm:"uniqueIndex;not null;column:username"`
	Password      *string    `gorm:"default:null;column:password"`
	FirstName     string     `gorm:"column:first_name"`
	LastName      string     `gorm:"column:last_name"`
	DisplayName   string     `gorm:"column:display_name"`
	Avatar        string     `gorm:"column:avatar"`
	Role          string     `gorm:"default:'user';column:role"`
	IsActive      bool       `gorm:"default:true;column:is_active"`
	IsOAuthUser   bool       `gorm:"default:false;column:is_o_auth_user"`
	OAuthProvider string     `gorm:"index;column:o_auth_provider"`
	OAuthID       string     `gorm:"index;column:o_auth_id"`
	EmailVerified bool       `gorm:"default:false;index;column:email_verified"`
	LastLoginAt   *time.Time `gorm:"column:last_login_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "users"
}

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Connect to database
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_NAME", "gofiber_auth"),
		getEnv("DB_SSL_MODE", "disable"),
	)

	log.Printf("Connecting to database: %s@%s:%s/%s\n",
		getEnv("DB_USER", "postgres"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "gofiber_auth"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Connected to database")

	// Check if users table exists
	var tableExists bool
	db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')").Scan(&tableExists)
	if !tableExists {
		log.Fatal("Users table does not exist. Please run the main application first to create the table.")
	}
	log.Println("✅ Users table found")

	// Check current user count
	var existingCount int64
	db.Model(&User{}).Count(&existingCount)
	log.Printf("Current users in database: %d\n", existingCount)

	// Open CSV file
	csvFile, err := os.Open("users.csv")
	if err != nil {
		log.Fatal("Failed to open CSV file:", err)
	}
	defer csvFile.Close()

	log.Println("✅ Opened users.csv")

	// Read CSV
	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Failed to read CSV:", err)
	}

	// Skip header row
	records = records[1:]

	log.Printf("Found %d users to import\n", len(records))
	log.Println("=====================================")

	// Import users
	successCount := 0
	skipCount := 0
	errorCount := 0

	for i, record := range records {
		if len(record) < 19 {
			log.Printf("Row %d: Invalid record length, skipping\n", i+1)
			errorCount++
			continue
		}

		user, err := parseUserFromCSV(record)
		if err != nil {
			log.Printf("Row %d: Failed to parse user: %v\n", i+1, err)
			errorCount++
			continue
		}

		// Check if user already exists
		var existingUser User
		result := db.Where("id = ?", user.ID).First(&existingUser)
		if result.Error == nil {
			log.Printf("Row %d: ⏭️  User %s already exists, skipping\n", i+1, user.Email)
			skipCount++
			continue
		}

		// Insert user
		if err := db.Create(&user).Error; err != nil {
			log.Printf("Row %d: ❌ Failed to insert user %s: %v\n", i+1, user.Email, err)
			errorCount++
			continue
		}

		oauthInfo := ""
		if user.IsOAuthUser {
			oauthInfo = fmt.Sprintf(" [OAuth: %s]", user.OAuthProvider)
		}
		log.Printf("Row %d: ✅ Imported user: %s (%s)%s\n", i+1, user.Email, user.Username, oauthInfo)
		successCount++
	}

	log.Println("=====================================")
	log.Println("=== Import Summary ===")
	log.Printf("✅ Success: %d\n", successCount)
	log.Printf("⏭️  Skipped: %d\n", skipCount)
	log.Printf("❌ Errors:  %d\n", errorCount)
	log.Printf("📊 Total:   %d\n", len(records))
	log.Println("=====================================")

	if successCount > 0 {
		log.Println("\n🎉 Migration completed successfully!")
		log.Println("\nNext steps:")
		log.Println("1. Test login with existing credentials")
		log.Println("2. Test OAuth login")
		log.Println("3. Verify JWT tokens work with backend service")
	}
}

func parseUserFromCSV(record []string) (*User, error) {
	// Parse UUID
	id, err := uuid.Parse(record[0])
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	// Parse created_at
	createdAt, err := time.Parse("2006-01-02 15:04:05.999999-07", record[17])
	if err != nil {
		createdAt, err = time.Parse("2006-01-02 15:04:05", record[17])
		if err != nil {
			// Try without timezone
			parts := strings.Split(record[17], "+")
			if len(parts) > 0 {
				createdAt, err = time.Parse("2006-01-02 15:04:05.999999", parts[0])
			}
			if err != nil {
				createdAt = time.Now()
			}
		}
	}

	// Parse updated_at
	updatedAt, err := time.Parse("2006-01-02 15:04:05.999999-07", record[18])
	if err != nil {
		updatedAt, err = time.Parse("2006-01-02 15:04:05", record[18])
		if err != nil {
			parts := strings.Split(record[18], "+")
			if len(parts) > 0 {
				updatedAt, err = time.Parse("2006-01-02 15:04:05.999999", parts[0])
			}
			if err != nil {
				updatedAt = time.Now()
			}
		}
	}

	// Parse boolean for is_oauth_user
	isOAuthUser := strings.ToLower(record[6]) == "true"

	// Parse boolean for is_active
	isActive := strings.ToLower(record[16]) == "true"

	// Parse display_name to extract first_name and last_name
	displayName := record[7]
	firstName, lastName := splitDisplayName(displayName)

	// Handle password (can be empty for OAuth users)
	var password *string
	if record[3] != "" {
		password = &record[3]
	}

	user := &User{
		ID:            id,
		Email:         record[1],
		Username:      record[2],
		Password:      password,
		FirstName:     firstName,
		LastName:      lastName,
		DisplayName:   displayName,
		Avatar:        record[8],
		Role:          record[15],
		IsActive:      isActive,
		IsOAuthUser:   isOAuthUser,
		OAuthProvider: record[4],
		OAuthID:       record[5],
		EmailVerified: true,
		LastLoginAt:   &updatedAt,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}

	return user, nil
}

func splitDisplayName(displayName string) (firstName, lastName string) {
	// Split by space
	parts := strings.Fields(displayName)

	if len(parts) == 0 {
		return "User", ""
	}

	if len(parts) == 1 {
		return parts[0], ""
	}

	// First word = firstName, rest = lastName
	firstName = parts[0]
	lastName = strings.Join(parts[1:], " ")

	return firstName, lastName
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
