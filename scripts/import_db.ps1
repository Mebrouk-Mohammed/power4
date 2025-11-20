<#
Importe la base `power4.sql` dans MySQL/MariaDB local.
Usage (PowerShell):
  .\import_db.ps1 -SqlFile .\power4.sql -DbName power4 -User root -Pass ""
#>
param(
    [string]$SqlFile = ".\power4.sql",
    [string]$DbName = "power4",
    [string]$User = "root",
    [string]$Pass = ""
)

$fullPath = Resolve-Path -Path $SqlFile -ErrorAction SilentlyContinue
if (-not $fullPath) {
    Write-Error "Fichier SQL introuvable: $SqlFile"
    exit 1
}
$fullPath = $fullPath.Path

# Cherche mysql.exe dans le PATH
$mysqlPath = (where.exe mysql 2>$null) -split "`r?`n" | Select-Object -First 1
if (-not $mysqlPath) {
    # Chemins usuels
    $candidates = @(
        "C:\Program Files\MySQL\MySQL Server 8.0\bin\mysql.exe",
        "C:\Program Files\MySQL\MySQL Server 5.7\bin\mysql.exe",
        "C:\Program Files\MariaDB 10.4\bin\mysql.exe",
        "C:\Program Files\MariaDB 10.5\bin\mysql.exe"
    )
    foreach ($c in $candidates) {
        if (Test-Path $c) { $mysqlPath = $c; break }
    }
}

if (-not $mysqlPath) {
    Write-Error "Impossible de trouver 'mysql.exe' dans le PATH ni dans les chemins usuels. Installez MySQL/MariaDB ou mettez le client dans le PATH."
    exit 2
}

# Crée la base si nécessaire
if ($Pass -ne "") { $auth = "-p$Pass" } else { $auth = "" }
$createCmd = "`"$mysqlPath`" -u $User $auth -e \"CREATE DATABASE IF NOT EXISTS `$DbName` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;\""
Write-Host "Création de la base (si nécessaire)..."
$rv = cmd /c $createCmd

# Importer le fichier via cmd.exe redirection (plus simple pour <)
$importCmd = "`"$mysqlPath`" -u $User $auth $DbName < `"$fullPath`""
Write-Host "Import du fichier SQL dans la base '$DbName'..."
$rv2 = cmd /c $importCmd

Write-Host "Import terminé. Si des erreurs sont affichées, copiez-les ici et je vous aiderai à les résoudre."