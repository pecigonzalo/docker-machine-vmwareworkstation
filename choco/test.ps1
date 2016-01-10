param(
  [string]$cpu
)

if (!$cpu) {
  $cpu = "amd64"
}
if ($cpu -eq "386") {
  $options = "-forcex86"
}

"Running tests for $cpu"
$ErrorActionPreference = "Stop"

if ($env:APPVEYOR_BUILD_VERSION) {
  # run in CI
  $version = $env:APPVEYOR_BUILD_VERSION -replace('\.[^.\\/]+$')
} else {
  # run manually
  [xml]$spec = Get-Content docker-machine-vmwareworkstation.nuspec
  $version = $spec.package.metadata.version
}

"TEST: Installation of docker-machine should work"
. choco install -y docker-machine

"TEST: Version $version in docker-machine-vmwareworkstation.nuspec file should match"
[xml]$spec = Get-Content docker-machine-vmwareworkstation.nuspec
if ($spec.package.metadata.version.CompareTo($version)) {
  Write-Error "FAIL: Wrong version in nuspec file!"
}

"TEST: Package should contain only install script"
Add-Type -assembly "system.io.compression.filesystem"
$zip = [IO.Compression.ZipFile]::OpenRead("$pwd\docker-machine-vmwareworkstation.$version.nupkg")
if ($zip.Entries.Count -ne 5) {
  Write-Error "FAIL: Wrong count in nupkg!"
}
$zip.Dispose()

"TEST: Installation of package should work"
. choco install -y docker-machine-vmwareworkstation $options -source .

"TEST: Create a machine with driver vmwareworkstation should work"
try {
  . docker-machine create -d vmwareworkstation test
} catch {
}
if (! (Test-Path $env:USERPROFILE\.docker\machine\machines\test)) {
  Write-Error "FAIL: Machine directory is missing"
}

"TEST: Uninstall show remove the binary"
. choco uninstall -y docker-machine-vmwareworkstation

"TEST: Finished"
