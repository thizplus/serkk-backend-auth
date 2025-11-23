package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Extension struct {
	Name    string
	Version string
}

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

	fmt.Println("========================================")
	fmt.Println("  PostgreSQL Extensions Check")
	fmt.Println("========================================\n")

	// Check installed extensions
	var extensions []Extension
	db.Raw("SELECT extname as name, extversion as version FROM pg_extension ORDER BY extname").Scan(&extensions)

	fmt.Printf("Installed Extensions: %d\n\n", len(extensions))

	hasUUIDOSSP := false
	hasPgcrypto := false

	for _, ext := range extensions {
		fmt.Printf("  ✅ %s (v%s)\n", ext.Name, ext.Version)
		if ext.Name == "uuid-ossp" {
			hasUUIDOSSP = true
		}
		if ext.Name == "pgcrypto" {
			hasPgcrypto = true
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("  UUID Generation Options")
	fmt.Println("========================================\n")

	// Test uuid-ossp
	if hasUUIDOSSP {
		var uuid string
		db.Raw("SELECT uuid_generate_v4()::text").Scan(&uuid)
		fmt.Printf("✅ uuid-ossp available\n")
		fmt.Printf("   Function: uuid_generate_v4()\n")
		fmt.Printf("   Sample: %s\n\n", uuid)
	} else {
		fmt.Println("❌ uuid-ossp NOT installed")
		fmt.Println("   Install: CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";\n")
	}

	// Test pgcrypto
	if hasPgcrypto {
		var uuid string
		db.Raw("SELECT gen_random_uuid()::text").Scan(&uuid)
		fmt.Printf("✅ pgcrypto available\n")
		fmt.Printf("   Function: gen_random_uuid()\n")
		fmt.Printf("   Sample: %s\n\n", uuid)
	} else {
		fmt.Println("❌ pgcrypto NOT installed")
		fmt.Println("   Install: CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";\n")
	}

	fmt.Println("========================================")
	fmt.Println("  Current Setup")
	fmt.Println("========================================\n")

	// Check current table defaults
	type ColumnDefault struct {
		TableName     string
		ColumnName    string
		ColumnDefault *string
	}

	var defaults []ColumnDefault
	db.Raw(`
		SELECT
			table_name,
			column_name,
			column_default
		FROM information_schema.columns
		WHERE table_name IN ('users', 'oauth_providers')
		AND column_name = 'id'
	`).Scan(&defaults)

	for _, d := range defaults {
		fmt.Printf("Table: %s\n", d.TableName)
		fmt.Printf("  Column: %s\n", d.ColumnName)
		if d.ColumnDefault != nil {
			fmt.Printf("  Default: %s\n", *d.ColumnDefault)
		} else {
			fmt.Printf("  Default: NULL (handled by Go BeforeCreate hook)\n")
		}
		fmt.Println()
	}

	fmt.Println("========================================")
	fmt.Println("  Recommendations")
	fmt.Println("========================================\n")

	if hasUUIDOSSP || hasPgcrypto {
		fmt.Println("You have UUID extensions available!")
		fmt.Println("\nOption 1: Keep current Go implementation (Recommended)")
		fmt.Println("  ✅ Portable")
		fmt.Println("  ✅ Already working")
		fmt.Println("  ✅ No changes needed")
		fmt.Println("\nOption 2: Use PostgreSQL UUID generation")
		fmt.Println("  ⚙️  Requires ALTER TABLE")
		fmt.Println("  ⚙️  Database-dependent")
		fmt.Println("  ⚙️  Slightly faster for bulk inserts")
	} else {
		fmt.Println("No UUID extensions found.")
		fmt.Println("Current Go implementation is the best option.")
	}

	fmt.Println()
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
