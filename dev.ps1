#Requires -Version 5.1
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$RootDir = $PSScriptRoot
$Mode = if ($args.Count -gt 0) { $args[0] } else { "stable" }

function Show-Usage {
    Write-Host @"
Usage:
  .\dev.ps1 [stable|live|help]

Modes:
  stable   Default. Build frontend static assets and start Wails without Vite dev server.
  live     Start the frontend dev server and connect Wails to it.
  help     Show this help.
"@
}

function Test-Command {
    param([string]$Cmd)
    if (-not (Get-Command $Cmd -ErrorAction SilentlyContinue)) {
        Write-Error "[ERROR] Missing required command: $Cmd"
        exit 1
    }
}

function Test-PortBusy {
    param([int]$Port)
    try {
        $conn = Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue
        return $null -ne $conn
    } catch {
        return $false
    }
}

function Resolve-WailsDevServer {
    $startPort = if ($env:WAILS_DEVSERVER_PORT) { [int]$env:WAILS_DEVSERVER_PORT } else { 34115 }
    $hostAddr = if ($env:WAILS_DEVSERVER_HOST) { $env:WAILS_DEVSERVER_HOST } else { "127.0.0.1" }
    $port = $startPort

    while (Test-PortBusy -Port $port) {
        $port++
    }

    $script:WailsDevServerAddress = "${hostAddr}:${port}"
}

function Prepare-Env {
    Test-Command "node"
    Test-Command "npm"
    Test-Command "go"
    Test-Command "wails"

    if ($env:DEV_PROXY_URL) {
        $env:HTTP_PROXY = $env:DEV_PROXY_URL
        $env:HTTPS_PROXY = $env:DEV_PROXY_URL
        $env:http_proxy = $env:DEV_PROXY_URL
        $env:https_proxy = $env:DEV_PROXY_URL
    }

    if ($env:DEV_NO_PROXY) {
        $env:NO_PROXY = $env:DEV_NO_PROXY
        $env:no_proxy = $env:DEV_NO_PROXY
    }

    if ($env:DEV_GOPROXY) {
        $env:GOPROXY = $env:DEV_GOPROXY
    } elseif (-not $env:GOPROXY) {
        $env:GOPROXY = "https://goproxy.cn,direct"
    }
}

function Install-FrontendDeps {
    Write-Host "Installing frontend dependencies..."
    Push-Location "$RootDir\frontend"
    try {
        npm install
    } finally {
        Pop-Location
    }
}

function Build-Frontend {
    Write-Host "Building frontend assets..."
    Push-Location "$RootDir\frontend"
    try {
        npm run build:clean
    } finally {
        Pop-Location
    }
}

function Run-Stable {
    Write-Host "========================================"
    Write-Host "  Facade - Dev Launcher"
    Write-Host "========================================"
    Write-Host ""
    Write-Host "Current workdir: $RootDir"
    Write-Host "Mode: stable"
    Write-Host "Frontend mode: stable static assets"
    Write-Host "Wails frontend dev server: disabled"
    Write-Host ""

    Prepare-Env
    Resolve-WailsDevServer
    Install-FrontendDeps
    Build-Frontend

    Write-Host "Starting Wails dev..."
    Write-Host "Wails dev server: http://$WailsDevServerAddress"
    wails dev -m -nogorebuild -noreload -s -skipbindings -assetdir "frontend/dist" -devserver $WailsDevServerAddress
}

function Run-Live {
    $frontendPort = if ($env:FRONTEND_PORT) { $env:FRONTEND_PORT } else { "5218" }
    $frontendProcess = $null

    Write-Host "========================================"
    Write-Host "  Facade - Dev Launcher"
    Write-Host "========================================"
    Write-Host ""
    Write-Host "Current workdir: $RootDir"
    Write-Host "Mode: live"
    Write-Host "Frontend URL: http://127.0.0.1:$frontendPort"
    Write-Host ""

    Prepare-Env
    Resolve-WailsDevServer

    Push-Location "$RootDir\frontend"
    try {
        npm install

        Write-Host "Starting Vite dev server..."
        $frontendProcess = Start-Process -FilePath "npm" -ArgumentList "run","dev:raw","--","--host","127.0.0.1","--port","$frontendPort" -NoNewWindow -PassThru
    } finally {
        Pop-Location
    }

    try {
        Write-Host "Starting Wails dev..."
        Write-Host "Wails dev server: http://$WailsDevServerAddress"
        wails dev -m -s -skipbindings -frontenddevserverurl "http://127.0.0.1:$frontendPort" -viteservertimeout 60 -devserver $WailsDevServerAddress
    } finally {
        if ($frontendProcess -and -not $frontendProcess.HasExited) {
            try {
                $frontendProcess.Kill()
            } catch {}
        }
    }
}

switch ($Mode) {
    "stable" { Run-Stable }
    "live"   { Run-Live }
    { $_ -in "help", "-h", "--help" } { Show-Usage }
    default {
        Write-Error "[ERROR] Unsupported mode: $Mode"
        Write-Host ""
        Show-Usage
        exit 1
    }
}
