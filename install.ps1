$url = "https://github.com/Andrei-666/moledrop/releases/latest/download/mole-windows-amd64.exe"
$dest = "C:\Windows\System32\mole.exe"

Write-Host "Installing MoleDrop..."
Invoke-WebRequest -Uri $url -OutFile $dest
Write-Host "MoleDrop installed successfully!"
Write-Host "Run: mole send <file>"
Write-Host "Run: mole receive <code>"