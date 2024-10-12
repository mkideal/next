Param(
    [string]$Version
)

Write-Host "Generating new GUIDs and replacing in installer.wxs..."

$GUID_1 = [guid]::NewGuid().ToString()
$GUID_2 = [guid]::NewGuid().ToString()
$GUID_3 = [guid]::NewGuid().ToString()
$GUID_4 = [guid]::NewGuid().ToString()

$content = Get-Content 'scripts/installer.wxs'
$content = $content -replace 'GUID-1', $GUID_1
$content = $content -replace 'GUID-2', $GUID_2
$content = $content -replace 'GUID-3', $GUID_3
$content = $content -replace 'GUID-4', $GUID_4
$content | Set-Content 'installer-temp.wxs'

Write-Host "Creating MSI packages..."

$Architectures = @('amd64', 'arm64', '386')
foreach ($Arch in $Architectures) {
    (Get-Content 'installer-temp.wxs') -replace 'ARCH', $Arch | Set-Content "installer-$Arch.wxs"
    candle.exe "installer-$Arch.wxs"
    light.exe -ext WixUIExtension "installer-$Arch.wixobj" -o "build/next$Version.windows-$Arch.msi"
}
