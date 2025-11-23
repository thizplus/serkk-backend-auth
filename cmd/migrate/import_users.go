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

	"gofiber-template/domain/models"
)

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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate - skip for now, assume table exists
	log.Println("Skipping auto-migration, using existing table schema")

	// Check if users table exists
	var tableExists bool
	db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')").Scan(&tableExists)
	if !tableExists {
		log.Fatal("Users table does not exist. Please run the main application first to create tables.")
	}
	log.Println("✅ Users table found")

	// Open CSV file
	csvFile, err := os.Open("users.csv")
	if err != nil {
		log.Fatal("Failed to open CSV file:", err)
	}
	defer csvFile.Close()

	// Read CSV
	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Failed to read CSV:", err)
	}

	// Skip header row
	records = records[1:]

	log.Printf("Found %d users to import\n", len(records))

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
		var existingUser models.User
		result := db.Where("id = ?", user.ID).First(&existingUser)
		if result.Error == nil {
			log.Printf("Row %d: User %s already exists, skipping\n", i+1, user.Email)
			skipCount++
			continue
		}

		// Insert user
		if err := db.Create(&user).Error; err != nil {
			log.Printf("Row %d: Failed to insert user %s: %v\n", i+1, user.Email, err)
			errorCount++
			continue
		}

		log.Printf("Row %d: ✅ Imported user: %s (%s)\n", i+1, user.Email, user.Username)
		successCount++
	}

	log.Println("\n=== Import Summary ===")
	log.Printf("✅ Success: %d\n", successCount)
	log.Printf("⏭️  Skipped: %d\n", skipCount)
	log.Printf("❌ Errors: %d\n", errorCount)
	log.Printf("📊 Total: %d\n", len(records))
}

func parseUserFromCSV(record []string) (*models.User, error) {
	// Parse UUID
	id, err := uuid.Parse(record[0])
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	// Parse created_at
	createdAt, err := time.Parse("2006-01-02 15:04:05.999999-07", record[17])
	if err != nil {
		// Try alternative format
		createdAt, err = time.Parse("2006-01-02 15:04:05", record[17])
		if err != nil {
			createdAt = time.Now()
		}
	}

	// Parse updated_at
	updatedAt, err := time.Parse("2006-01-02 15:04:05.999999-07", record[18])
	if err != nil {
		updatedAt, err = time.Parse("2006-01-02 15:04:05", record[18])
		if err != nil {
			updatedAt = time.Now()
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

	user := &models.User{
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
		EmailVerified: true,        // Assume existing users are verified
		LastLoginAt:   &updatedAt,  // Use updated_at as last login
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
