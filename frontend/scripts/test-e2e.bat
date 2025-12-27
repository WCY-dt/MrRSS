@echo off
REM MrRSS E2E Test Runner for Windows
REM This script helps you run E2E tests with the Wails backend

setlocal enabledelayedexpansion

echo.
echo ========================================
echo   MrRSS E2E Test Runner
echo ========================================
echo.

REM Change to frontend directory
cd /d "%~dp0..\"

REM Check if backend is running
curl -s http://localhost:34115/api/feeds >nul 2>&1
if %errorlevel% neq 0 (
    echo WARNING: Backend is not running on port 34115
    echo.
    echo Please start the backend in another terminal:
    echo.
    echo   cd ..^& wails3 dev
    echo.
    echo Or press Ctrl+C to exit and start it manually
    echo.
    pause

    REM Check again
    curl -s http://localhost:34115/api/feeds >nul 2>&1
    if %errorlevel% neq 0 (
        echo.
        echo ERROR: Backend still not responding. Exiting.
        pause
        exit /b 1
    )
)

echo Backend is running!
echo.

REM Parse arguments
set MODE=interactive
set SPEC_FILE=

:parse_args
if "%~1"=="" goto end_parse
if /i "%~1"=="--headless" (
    set MODE=headless
    shift
    goto parse_args
)
if /i "%~1"=="--spec" (
    set SPEC_FILE=%~2
    shift
    shift
    goto parse_args
)
if /i "%~1"=="--help" (
    echo Usage: %0 [OPTIONS]
    echo.
    echo Options:
    echo   --headless    Run tests in headless mode (no GUI)
    echo   --spec FILE   Run specific test file
    echo   --help        Show this help message
    echo.
    echo Examples:
    echo   %0                    # Interactive mode
    echo   %0 --headless         # Headless mode
    echo   %0 --spec cypress\e2e\app-smoke.cy.ts  # Run specific test
    pause
    exit /b 0
)
echo Unknown option: %~1
echo Use --help for usage information
pause
exit /b 1

:end_parse

REM Run Cypress based on mode
if "%MODE%"=="headless" (
    if not "%SPEC_FILE%"=="" (
        npx cypress run --spec %SPEC_FILE%
    ) else (
        npx cypress run
    )
) else (
    if not "%SPEC_FILE%"=="" (
        npx cypress open --spec %SPEC_FILE%
    ) else (
        npx cypress open
    )
)

pause
