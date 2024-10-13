Param(
    [string]$Version,
    [string]$Dir,
    [string]$GoBuild
)

Write-Host "Generating new GUIDs and replacing in installer.wxs..."
$Version = $Version -replace '^v', ''

$GUID_1 = [guid]::NewGuid().ToString()
$GUID_2 = [guid]::NewGuid().ToString()
$GUID_3 = [guid]::NewGuid().ToString()
$GUID_4 = [guid]::NewGuid().ToString()

$content = Get-Content 'scripts/installer.wxs'
$content = $content -replace 'GUID-1', $GUID_1
$content = $content -replace 'GUID-2', $GUID_2
$content = $content -replace 'GUID-3', $GUID_3
$content = $content -replace 'GUID-4', $GUID_4
$content = $content -replace 'NEXT_VERSION', $Version
$content | Set-Content 'installer-temp.wxs'

Write-Host "Creating MSI packages..."

$Architectures = @('amd64', 'arm64', '386')
foreach ($Arch in $Architectures) {
    Write-Host "Building $Dir\windows-$Arch\next..."

    Remove-Item installer-$Arch.wixobj -Force -ErrorAction SilentlyContinue
    Remove-Item "$Dir\windows-$Arch" -Recurse -Force -ErrorAction SilentlyContinue
    Remove-Item "$Dir\next$Version.windows-$Arch.wixpdb" -Force -ErrorAction SilentlyContinue
    Remove-Item "$Dir\next$Version.windows-$Arch.msi" -Force -ErrorAction SilentlyContinue
    
    # Create directory
    New-Item -ItemType Directory -Force -Path "$Dir\windows-$Arch\bin" | Out-Null
    
    # Set environment variables and compile
    $env:GOOS = "windows"
    $env:GOARCH = $Arch
    
    Write-Host "$GoBuild -o $Dir\windows-$Arch\bin\next.exe"
    Invoke-Expression "$GoBuild -o `"$Dir\windows-$Arch\bin\next.exe`""
    Invoke-Expression "$GoBuild -o `"$Dir\windows-$Arch\bin\nextls.exe`" ./cmd/nextls/"
    
    # Copy README file
    Copy-Item "README.md" -Destination "$Dir\windows-$Arch\"
    
    # Replace ARCH in installer-temp.wxs and create new file
    (Get-Content 'installer-temp.wxs') -replace 'ARCH', $Arch | Set-Content "installer-$Arch.wxs"
    
    # Run WiX toolchain commands
    & candle.exe "installer-$Arch.wxs"
    & light.exe -ext WixUIExtension "installer-$Arch.wixobj" -o "$Dir\next$Version.windows-$Arch.msi"

    Remove-Item installer-$Arch.wxs
    Remove-Item installer-$Arch.wixobj
    Remove-Item "$Dir\windows-$Arch" -Recurse -Force
    Remove-Item "$Dir\next$Version.windows-$Arch.wixpdb"
}

Remove-Item installer-temp.wxs
