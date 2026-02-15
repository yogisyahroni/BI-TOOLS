$ErrorActionPreference = "Stop"

# Configuration
$DB_HOST = "localhost"
$DB_USER = "postgres"
$DB_NAME = "insight_engine"
$BACKUP_DIR = "./backups"
$TIMESTAMP = Get-Date -Format "yyyyMMdd_HHmmss"
$BACKUP_FILE = "$BACKUP_DIR/backup_$TIMESTAMP.sql"

# Ensure backup directory exists
if (!(Test-Path -Path $BACKUP_DIR)) {
    New-Item -ItemType Directory -Path $BACKUP_DIR | Out-Null
    Write-Host "Created backup directory: $BACKUP_DIR"
}

# Check for pg_dump availability
if (!(Get-Command pg_dump -ErrorAction SilentlyContinue)) {
    Write-Error "pg_dump command not found. Please ensure PostgreSQL bin directory is in your PATH."
    exit 1
}

try {
    Write-Host "Starting backup of database '$DB_NAME' to '$BACKUP_FILE'..."
    
    # Execute pg_dump
    # Note: PGPASSWORD environment variable should be set for non-interactive auth, 
    # or .pgpass file used.
    $env:PGPASSWORD = "your_password_here" # REPLACE WITH ENV VAR OR SECRET IF NEEDED
    
    pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME -f $BACKUP_FILE

    Write-Host "Backup completed successfully: $BACKUP_FILE"
}
catch {
    Write-Error "Backup failed: $_"
    exit 1
}
