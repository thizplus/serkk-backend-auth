package logger

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// LogLevel represents log severity
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

// StructuredLogger provides JSON-formatted logging
type StructuredLogger struct {
	serviceName string
	env         string
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   string                 `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Service     string                 `json:"service"`
	Environment string                 `json:"environment"`
	Message     string                 `json:"message"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Action      string                 `json:"action,omitempty"`
	Duration    int64                  `json:"duration_ms,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

var defaultLogger *StructuredLogger

// InitLogger initializes the default structured logger
func InitLogger(serviceName, env string) {
	defaultLogger = &StructuredLogger{
		serviceName: serviceName,
		env:         env,
	}
}

// GetLogger returns the default logger
func GetLogger() *StructuredLogger {
	if defaultLogger == nil {
		InitLogger("gofiber-auth", "development")
	}
	return defaultLogger
}

// log outputs a structured log entry
func (l *StructuredLogger) log(level LogLevel, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Level:       level,
		Service:     l.serviceName,
		Environment: l.env,
		Message:     message,
	}

	// Add optional fields
	if requestID, ok := fields["request_id"].(string); ok {
		entry.RequestID = requestID
	}
	if userID, ok := fields["user_id"].(string); ok {
		entry.UserID = userID
	}
	if action, ok := fields["action"].(string); ok {
		entry.Action = action
	}
	if duration, ok := fields["duration_ms"].(int64); ok {
		entry.Duration = duration
	}
	if err, ok := fields["error"].(string); ok {
		entry.Error = err
	}

	// Add remaining fields as metadata
	metadata := make(map[string]interface{})
	for k, v := range fields {
		if k != "request_id" && k != "user_id" && k != "action" && k != "duration_ms" && k != "error" {
			metadata[k] = v
		}
	}
	if len(metadata) > 0 {
		entry.Metadata = metadata
	}

	// Output as JSON
	if os.Getenv("LOG_FORMAT") == "json" {
		jsonData, _ := json.Marshal(entry)
		log.Println(string(jsonData))
	} else {
		// Human-readable format for development
		log.Printf("[%s] %s: %s %v", entry.Level, entry.Service, entry.Message, fields)
	}
}

// Debug logs debug-level messages
func (l *StructuredLogger) Debug(message string, fields map[string]interface{}) {
	l.log(DEBUG, message, fields)
}

// Info logs info-level messages
func (l *StructuredLogger) Info(message string, fields map[string]interface{}) {
	l.log(INFO, message, fields)
}

// Warn logs warning-level messages
func (l *StructuredLogger) Warn(message string, fields map[string]interface{}) {
	l.log(WARN, message, fields)
}

// Error logs error-level messages
func (l *StructuredLogger) Error(message string, fields map[string]interface{}) {
	l.log(ERROR, message, fields)
}

// Package-level convenience functions
func Debug(message string, fields map[string]interface{}) {
	GetLogger().Debug(message, fields)
}

func Info(message string, fields map[string]interface{}) {
	GetLogger().Info(message, fields)
}

func Warn(message string, fields map[string]interface{}) {
	GetLogger().Warn(message, fields)
}

func Error(message string, fields map[string]interface{}) {
	GetLogger().Error(message, fields)
}
