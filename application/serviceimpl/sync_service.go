package serviceimpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gofiber-template/domain/models"
)

type SyncService struct {
	backendURL string
	httpClient *http.Client
}

func NewSyncService() *SyncService {
	return &SyncService{
		backendURL: os.Getenv("BACKEND_SYNC_URL"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type UserSyncPayload struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Role        string `json:"role"`
	IsActive    bool   `json:"isActive"`
	Action      string `json:"action"` // "created", "updated", "deleted"
}

func (s *SyncService) SyncUser(user *models.User, action string) error {
	if s.backendURL == "" {
		log.Println("⚠️  BACKEND_SYNC_URL not configured, skipping sync")
		return nil
	}

	payload := UserSyncPayload{
		ID:          user.ID.String(),
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		Role:        user.Role,
		IsActive:    user.IsActive,
		Action:      action,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", s.backendURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("❌ Failed to sync user %s to backend: %v", user.Username, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("❌ Backend sync failed for user %s: HTTP %d", user.Username, resp.StatusCode)
		return fmt.Errorf("backend returned status %d", resp.StatusCode)
	}

	log.Printf("✅ User %s synced to backend (action: %s)", user.Username, action)
	return nil
}

// SyncUserWithRetry tries to sync with exponential backoff
func (s *SyncService) SyncUserWithRetry(user *models.User, action string) {
	maxRetries := 3
	baseDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		err := s.SyncUser(user, action)
		if err == nil {
			return // Success
		}

		if i < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<uint(i))
			log.Printf("Retrying sync in %v (attempt %d/%d)", delay, i+2, maxRetries)
			time.Sleep(delay)
		}
	}

	log.Printf("❌ Failed to sync user %s after %d attempts", user.Username, maxRetries)
	// TODO: Send to dead letter queue or alert monitoring
}
