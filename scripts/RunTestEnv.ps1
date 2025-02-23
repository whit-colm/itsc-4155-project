#Requires -Version 7.0

<#
.SYNOPSIS
Starts the JAWS application with PostgreSQL in Docker containers
Translated with DeepSeek DeepThink-R1 from original Bash script
#>

$ErrorActionPreference = "Stop"

# Generate simple password (not secure, but works for testing)
$psql_passwd = -join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | ForEach-Object {[char]$_})

# Get to repo root
$repoRoot = git rev-parse --show-toplevel
Set-Location $repoRoot

# Build image with timestamp tag
$jawsTag = Get-Date -Format "yyyyMMddHHmmss"
docker build --target webnative -t "jaws:$jawsTag" .

# Run containers with automatic cleanup (--rm) and default ports
docker run --rm --name jaws-psql `
    -p 54321:5432 `
    -e POSTGRES_USER="jaws" `
    -e POSTGRES_DB="jaws" `
    -e POSTGRES_PASSWORD=$psql_passwd `
    -d postgres:17-alpine

docker run --rm --name jaws-app `
    -p 8080:80 `
    -e PG_PASSWORD=$psql_passwd `
    -e PG_USER="jaws" `
    -e PG_DATABASE="jaws" `
    -e PG_HOST="jaws-psql" `
    -e PG_PORT=54321 `
    -d "jaws:$jawsTag"

Write-Host "`nContainers running!`nApp: http://localhost:8080`nPostgres: localhost:54321`n"