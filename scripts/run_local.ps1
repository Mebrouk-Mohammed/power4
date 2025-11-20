<#
Lance l'application Go en définissant les variables d'environnement DB nécessaires.
Usage (PowerShell):
  .\run_local.ps1 -DbUser root -DbPass "" -DbHost 127.0.0.1 -DbPort 3306 -DbName power4
#>
param(
    [string]$DbUser = "root",
    [string]$DbPass = "",
    [string]$DbHost = "127.0.0.1",
    [string]$DbPort = "3306",
    [string]$DbName = "power4"
)

Write-Host "Export des variables d'environnement pour la session PowerShell..."
$env:DB_USER = $DbUser
$env:DB_PASS = $DbPass
$env:DB_HOST = $DbHost
$env:DB_PORT = $DbPort
$env:DB_NAME = $DbName

Write-Host "Lancement de l'application Go (go run main.go)..."
# Lance la commande go run dans le répertoire courant
go run .\main.go
