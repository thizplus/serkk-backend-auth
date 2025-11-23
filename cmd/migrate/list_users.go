package main

import (
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID            uuid.UUID `gorm:"column:id"`
	Email         string    `gorm:"column:email"`
	Username      string    `gorm:"column:username"`
	Password      *string   `gorm:"column:password"`
	FirstName     string    `gorm:"column:first_name"`
	LastName      string    `gorm:"column:last_name"`
	DisplayName   string    `gorm:"column:display_name"`
	Avatar        string    `gorm:"column:avatar"`
	Role          string    `gorm:"column:role"`
	IsActive      bool      `gorm:"column:is_active"`
	IsOAuthUser   bool      `gorm:"column:is_o_auth_user"`
	OAuthProvider string    `gorm:"column:o_auth_provider"`
	OAuthID       string    `gorm:"column:o_auth_id"`
	EmailVerified bool      `gorm:"column:email_verified"`
}

func (User) TableName() string {
	return "users"
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

	var users []User
	db.Find(&users)

	fmt.Println("========================================")
	fmt.Printf("Total Users: %d\n", len(users))
	fmt.Println("========================================\n")

	for i, user := range users {
		fmt.Printf("User #%d:\n", i+1)
		fmt.Printf("  ID:           %s\n", user.ID)
		fmt.Printf("  Email:        %s\n", user.Email)
		fmt.Printf("  Username:     %s\n", user.Username)
		fmt.Printf("  Display Name: %s\n", user.DisplayName)
		fmt.Printf("  First Name:   %s\n", user.FirstName)
		fmt.Printf("  Last Name:    %s\n", user.LastName)

		if user.Password != nil {
			fmt.Printf("  Password:     ✅ (hashed)\n")
		} else {
			fmt.Printf("  Password:     ❌ (OAuth user)\n")
		}

		fmt.Printf("  Role:         %s\n", user.Role)
		fmt.Printf("  Active:       %v\n", user.IsActive)
		fmt.Printf("  Verified:     %v\n", user.EmailVerified)

		if user.IsOAuthUser {
			fmt.Printf("  OAuth:        ✅ %s (ID: %s)\n", user.OAuthProvider, user.OAuthID)
		} else {
			fmt.Printf("  OAuth:        ❌\n")
		}

		if user.Avatar != "" {
			fmt.Printf("  Avatar:       %s\n", user.Avatar[:min(50, len(user.Avatar))]+"...")
		}

		fmt.Println()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
