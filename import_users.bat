@echo off
echo ========================================
echo    User Migration Script
echo    Import users from CSV to Auth DB
echo ========================================
echo.

REM Check if users.csv exists
if not exist "users.csv" (
    echo [ERROR] users.csv not found!
    echo Please put users.csv in the project root directory
    pause
    exit /b 1
)

echo [INFO] Found users.csv
echo.

REM Check .env file
if not exist ".env" (
    echo [ERROR] .env file not found!
    echo Please create .env file with database configuration
    pause
    exit /b 1
)

echo [INFO] Found .env file
echo.

echo [INFO] Starting user import...
echo.

REM Run the migration script
go run cmd/migrate/import_users.go

echo.
echo ========================================
echo    Migration Complete
echo ========================================
echo.

pause
