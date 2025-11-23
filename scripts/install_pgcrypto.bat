@echo off
echo ========================================
echo   Installing pgcrypto Extension
echo ========================================
echo.

set PGPASSWORD=n147369
psql -h localhost -U postgres -d gofiber_auth -c "CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";"

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo   ✅ pgcrypto installed successfully!
    echo ========================================
    echo.
    echo Testing gen_random_uuid...
    psql -h localhost -U postgres -d gofiber_auth -c "SELECT gen_random_uuid();"
    echo.
    echo ✅ Your server should work now!
) else (
    echo.
    echo ❌ Failed to install pgcrypto
    echo Please check PostgreSQL connection
)

pause
