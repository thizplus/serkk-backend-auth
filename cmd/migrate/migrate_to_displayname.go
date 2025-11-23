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
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Database connection
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("🔄 Migrating firstName + lastName to displayName...")

	// Update users where displayName is empty
	result := db.Exec(`
		UPDATE users
		SET display_name = TRIM(first_name || ' ' || last_name)
		WHERE (display_name = '' OR display_name IS NULL)
		  AND (first_name != '' OR last_name != '')
	`)

	if result.Error != nil {
		log.Fatal("Migration failed:", result.Error)
	}

	log.Printf("✅ Migrated %d users", result.RowsAffected)

	// Show sample data
	var users []struct {
		Username    string
		FirstName   string
		LastName    string
		DisplayName string
	}

	db.Raw("SELECT username, first_name, last_name, display_name FROM users LIMIT 5").Scan(&users)

	log.Println("\n📊 Sample data:")
	for _, u := range users {
		log.Printf("  %s: firstName='%s', lastName='%s' → displayName='%s'",
			u.Username, u.FirstName, u.LastName, u.DisplayName)
	}

	log.Println("\n✅ Migration complete!")
	log.Println("⚠️  You can now safely remove first_name and last_name columns")
}
