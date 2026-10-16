# Installs vex from its GitHub releases, or updates it:
#
#   irm https://raw.githubusercontent.com/ixexel661/vex/main/install/install.ps1 | iex
#
# -Uninstall removes vex again, together with its settings and themes:
#
#   & ([scriptblock]::Create((irm https://raw.githubusercontent.com/ixexel661/vex/main/install/install.ps1))) -Uninstall
#
# $env:VEX_VERSION picks a release, e.g. v1.2.0 (default: the latest one).
# $env:VEX_INSTALL_DIR picks the directory (default: %LOCALAPPDATA%\Programs\vex).
#
# iex runs this in the scope of the caller, so everything lives in functions
# named after vex, and errors throw instead of closing the window with exit.
param([switch]$Uninstall)

function Get-VexDir {
    if ($env:VEX_INSTALL_DIR) {
        return $env:VEX_INSTALL_DIR
    }
    return Join-Path $env:LOCALAPPDATA 'Programs\vex'
}

# Get-VexAsset returns the name of the release file for this system.
function Get-VexAsset {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    switch ($arch.ToString()) {
        'X64' { return 'vex-windows-amd64.exe' }
        'Arm64' { return 'vex-windows-arm64.exe' }
        default { throw "vex: there is no vex for $arch" }
    }
}

# Get-VexTag returns the tag of the release to install.
function Get-VexTag {
    if ($env:VEX_VERSION) {
        return 'v' + ($env:VEX_VERSION -replace '^v', '')
    }
    try {
        $release = Invoke-RestMethod -UseBasicParsing 'https://api.github.com/repos/ixexel661/vex/releases/latest'
    } catch {
        throw "vex: cannot find the latest release: $($_.Exception.Message)"
    }
    return $release.tag_name
}

# Get-VexChecksum returns the checksum of asset from a SHA256SUMS file.
function Get-VexChecksum($sums, $asset) {
    foreach ($line in Get-Content $sums) {
        $hash, $name = $line -split '\s+'
        if ($name -eq $asset) {
            return $hash
        }
    }
    return ''
}

# The PATH of the user is read and written in the registry, where entries
# like %USERPROFILE%\bin keep their variables.
function Get-VexUserPath {
    $key = Get-Item 'HKCU:\Environment'
    return ($key.GetValue('Path', '', 'DoNotExpandEnvironmentNames') -split ';' | Where-Object { $_ })
}

function Set-VexUserPath($entries) {
    Set-ItemProperty 'HKCU:\Environment' -Name Path -Value ($entries -join ';') -Type ExpandString
    # Changing a variable through .NET tells Windows that the environment
    # changed, so that new terminals get the new PATH.
    [Environment]::SetEnvironmentVariable('VEX_INSTALLER', '1', 'User')
    [Environment]::SetEnvironmentVariable('VEX_INSTALLER', $null, 'User')
}

function Test-VexPathEntry($entry, $dir) {
    return [Environment]::ExpandEnvironmentVariables($entry).TrimEnd('\') -eq $dir.TrimEnd('\')
}

function Add-VexToPath($dir) {
    $entries = @(Get-VexUserPath)
    if (-not ($entries | Where-Object { Test-VexPathEntry $_ $dir })) {
        Set-VexUserPath ($entries + $dir)
        Write-Host "Added $dir to your PATH"
    }
    # The PATH of this session, to use vex right away.
    if (-not ($env:Path -split ';' | Where-Object { Test-VexPathEntry $_ $dir })) {
        $env:Path = "$env:Path;$dir"
    }
}

function Remove-VexFromPath($dir) {
    $entries = @(Get-VexUserPath)
    $kept = @($entries | Where-Object { -not (Test-VexPathEntry $_ $dir) })
    if ($kept.Count -lt $entries.Count) {
        Set-VexUserPath $kept
        Write-Host "Removed $dir from your PATH"
    }
}

function Remove-VexItem($path) {
    if (Test-Path $path) {
        Remove-Item -Recurse -Force $path
        Write-Host "Removed $path"
    }
}

function Install-Vex {
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'
    # Windows PowerShell 5.1 does not always offer TLS 1.2, which GitHub needs.
    [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

    $dir = Get-VexDir
    $exe = Join-Path $dir 'vex.exe'
    $tag = Get-VexTag
    $asset = Get-VexAsset
    $url = "https://github.com/ixexel661/vex/releases/download/$tag"
    Write-Host "Downloading vex $tag ($asset)"

    $work = Join-Path ([IO.Path]::GetTempPath()) ([IO.Path]::GetRandomFileName())
    New-Item -ItemType Directory $work | Out-Null
    try {
        $file = Join-Path $work $asset
        $sums = Join-Path $work 'SHA256SUMS'
        try {
            Invoke-WebRequest -UseBasicParsing "$url/$asset" -OutFile $file
            Invoke-WebRequest -UseBasicParsing "$url/SHA256SUMS" -OutFile $sums
        } catch {
            throw "vex: cannot download $asset of vex ${tag}: $($_.Exception.Message)"
        }
        $want = Get-VexChecksum $sums $asset
        if (-not $want -or (Get-FileHash -Algorithm SHA256 $file).Hash -ne $want) {
            throw "vex: the checksum of $asset does not match, vex is not installed"
        }
        New-Item -ItemType Directory -Force $dir | Out-Null
        try {
            Move-Item -Force $file $exe
        } catch {
            throw "vex: cannot replace $exe, close vex and try again: $($_.Exception.Message)"
        }
    } finally {
        Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $work
    }
    Write-Host "Installed vex $tag to $exe"

    Add-VexToPath $dir
    $found = Get-Command vex.exe -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($found -and $found.Path -ne $exe) {
        Write-Host "Note: $($found.Path) comes first in PATH, so vex starts that one"
    }
}

function Uninstall-Vex {
    $ErrorActionPreference = 'Stop'
    $dir = Get-VexDir
    $exe = Join-Path $dir 'vex.exe'
    if (-not (Test-Path $exe)) {
        Write-Host "There is no vex in $dir"
    }
    Remove-VexItem $exe
    # The default directory belongs to vex alone.
    if (-not $env:VEX_INSTALL_DIR -and (Test-Path $dir) -and -not (Get-ChildItem -Force $dir)) {
        Remove-VexItem $dir
    }
    Remove-VexFromPath $dir
    Remove-VexItem (Join-Path $env:APPDATA 'vex')
}

if ($Uninstall) {
    Uninstall-Vex
} else {
    Install-Vex
}
