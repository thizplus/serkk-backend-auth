-- Initialize database for GoFiber Auth Service
-- This script runs automatically when PostgreSQL container starts

-- Enable pgcrypto extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create users table will be handled by GORM AutoMigrate
-- But we ensure the extension is available

-- You can add additional initialization here if needed
