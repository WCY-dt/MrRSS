# wails.dev.ps1 - Windows Wails development wrapper
# This script starts Wails development server with correct configuration

Write-Host "🚀 Starting Wails development server..." -ForegroundColor Green
Write-Host "🪟 Platform: Windows" -ForegroundColor Cyan

# Windows doesn't need special tags
wails dev @args

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Wails dev failed with exit code $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}
