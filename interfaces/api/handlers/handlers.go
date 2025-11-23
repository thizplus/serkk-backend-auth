package handlers

import (
	"gofiber-template/domain/services"
	"gofiber-template/pkg/config"
)

// Services contains all the services needed for handlers
type Services struct {
	UserService  services.UserService
	OAuthService services.OAuthService
	Config       *config.Config
}

// Handlers contains all HTTP handlers
type Handlers struct {
	UserHandler    *UserHandler
	OAuthHandler   *OAuthHandler
	MetricsHandler *MetricsHandler
}

// NewHandlers creates a new instance of Handlers with all dependencies
func NewHandlers(services *Services) *Handlers {
	return &Handlers{
		UserHandler:    NewUserHandler(services.UserService),
		OAuthHandler:   NewOAuthHandler(services.OAuthService, services.Config),
		MetricsHandler: NewMetricsHandler(),
	}
}