# Set the environment variable used by docker-compose.build.yml
$env:DOCKER_REGISTRY = "hariii31"

Write-Host "Starting build process for registry: $env:DOCKER_REGISTRY" -ForegroundColor Green

# Navigate to the deploy directory in case script is run from elsewhere
Set-Location $PSScriptRoot

Write-Host "`n[1/2] Building images..." -ForegroundColor Cyan
docker-compose -f docker-compose.build.yml build

if ($LASTEXITCODE -ne 0) {
    Write-Error "Build failed!"
    exit $LASTEXITCODE
}

Write-Host "`n[2/2] Pushing images..." -ForegroundColor Cyan
docker-compose -f docker-compose.build.yml push

if ($LASTEXITCODE -ne 0) {
    Write-Error "Push failed!"
    exit $LASTEXITCODE
}

Write-Host "`nSuccess! All images built and pushed to $env:DOCKER_REGISTRY" -ForegroundColor Green
