package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	godotenv.Load()

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
		log.Fatal("Failed to connect:", err)
	}

	// Check if users table exists
	var exists bool
	db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')").Scan(&exists)

	if !exists {
		fmt.Println("❌ Users table does NOT exist")
		fmt.Println("✅ Ready to create new table and import data")
		return
	}

	fmt.Println("✅ Users table exists")
	fmt.Println("\nTable structure:")

	// Get column information
	type Column struct {
		ColumnName string
		DataType   string
		IsNullable string
	}

	var columns []Column
	db.Raw(`
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = 'users'
		ORDER BY ordinal_position
	`).Scan(&columns)

	for _, col := range columns {
		nullable := ""
		if col.IsNullable == "NO" {
			nullable = " NOT NULL"
		}
		fmt.Printf("  - %s (%s)%s\n", col.ColumnName, col.DataType, nullable)
	}

	// Count existing users
	var count int64
	db.Raw("SELECT COUNT(*) FROM users").Scan(&count)
	fmt.Printf("\nExisting users: %d\n", count)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
