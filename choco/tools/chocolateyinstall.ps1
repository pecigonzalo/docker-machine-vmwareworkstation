$packageName    = 'docker-machine-vmwareworkstation'
$driverName     = 'docker-machine-driver-vmwareworkstation'
$url            = 'https://github.com/pecigonzalo/docker-machine-vmwareworkstation/releases/download/v1.0.0/docker-machine-driver-vmwareworkstation.exe'
$checksum       = '56e48f38cd34a277e5c51ec98bb3ccea'
$url64          = 'https://github.com/pecigonzalo/docker-machine-vmwareworkstation/releases/download/v1.0.0/docker-machine-driver-vmwareworkstation.exe'
$checksum64     = '56e48f38cd34a277e5c51ec98bb3ccea'
$checksumType   = 'md5'
$checksumType64 = 'md5'

$toolsDir    = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$packageDir  = "$(Split-Path -parent $toolsDir)"
$installDir  = Join-Path "$packageDir" "bin"
$installBin  = "${driverName}.exe"
$installPath = Join-Path "$installDir" "$installBin"

New-Item -ItemType Directory -Force -Path "$installDir"
Get-ChocolateyWebFile "$packageName" "$installPath" "$url" "$url64" -checksum "$checksum" -checksumType "$checksumType" -checksum64 "$checksum64" -checksumType64 "$checksumType64"
