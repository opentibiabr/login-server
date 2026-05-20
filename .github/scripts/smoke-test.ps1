param(
    [Parameter(Mandatory = $true)]
    [string]$BinaryPath,

    [int]$HttpPort = 18080,

    [int]$GrpcPort = 19090,

    [int]$StartupTimeoutSeconds = 20
)

$ErrorActionPreference = "Stop"

function Wait-TcpPort {
    param(
        [string]$HostName,
        [int]$Port,
        [System.Diagnostics.Process]$Process,
        [DateTime]$Deadline
    )

    while ([DateTime]::UtcNow -lt $Deadline) {
        if ($Process.HasExited) {
            throw "Process exited before ${HostName}:${Port} became available. Exit code: $($Process.ExitCode)"
        }

        $client = [System.Net.Sockets.TcpClient]::new()
        try {
            $client.Connect($HostName, $Port)
            return
        }
        catch {
            Start-Sleep -Milliseconds 250
        }
        finally {
            $client.Dispose()
        }
    }

    throw "Timed out waiting for ${HostName}:${Port}"
}

function Invoke-SmokeRequest {
    param(
        [string]$Uri,
        [string]$Body,
        [string]$ContentType,
        [int]$ExpectedStatus
    )

    $response = Invoke-WebRequest -Uri $Uri -Method Post -Body $Body -ContentType $ContentType -TimeoutSec 5
    if ($response.StatusCode -ne $ExpectedStatus) {
        throw "Unexpected HTTP status for ${Uri}: got $($response.StatusCode), expected ${ExpectedStatus}"
    }
}

$resolvedBinary = Resolve-Path -LiteralPath $BinaryPath
$listenAddress = "127.0.0.1"

$startInfo = [System.Diagnostics.ProcessStartInfo]::new()
$startInfo.FileName = $resolvedBinary.Path
$startInfo.WorkingDirectory = (Get-Location).Path
$startInfo.UseShellExecute = $false
$startInfo.RedirectStandardOutput = $true
$startInfo.RedirectStandardError = $true

$startInfo.Environment["LOGIN_IP"] = $listenAddress
$startInfo.Environment["LOGIN_HTTP_PORT"] = [string]$HttpPort
$startInfo.Environment["LOGIN_GRPC_PORT"] = [string]$GrpcPort
$startInfo.Environment["ENV_LOG_LEVEL"] = "debug"
$startInfo.Environment["SERVER_PATH"] = ""
$startInfo.Environment["MYSQL_HOST"] = "127.0.0.1"
$startInfo.Environment["MYSQL_PORT"] = "3306"
$startInfo.Environment["MYSQL_DBNAME"] = "login_server_smoke"
$startInfo.Environment["MYSQL_USER"] = "login_server_smoke"
$startInfo.Environment["MYSQL_PASS"] = "login_server_smoke"
$startInfo.Environment["RATE_LIMITER_BURST"] = "100"
$startInfo.Environment["RATE_LIMITER_RATE"] = "100"

$process = [System.Diagnostics.Process]::Start($startInfo)
$smokeSucceeded = $false

try {
    $deadline = [DateTime]::UtcNow.AddSeconds($StartupTimeoutSeconds)

    Wait-TcpPort -HostName $listenAddress -Port $GrpcPort -Process $process -Deadline $deadline
    Wait-TcpPort -HostName $listenAddress -Port $HttpPort -Process $process -Deadline $deadline

    Invoke-SmokeRequest `
        -Uri "http://${listenAddress}:${HttpPort}/crash-report" `
        -Body "" `
        -ContentType "text/plain" `
        -ExpectedStatus 204

    Invoke-SmokeRequest `
        -Uri "http://${listenAddress}:${HttpPort}/login" `
        -Body '{"type":"cacheinfo"}' `
        -ContentType "application/json" `
        -ExpectedStatus 200

    $smokeSucceeded = $true
    Write-Host "Login server smoke test passed on HTTP port ${HttpPort} and gRPC port ${GrpcPort}."
}
finally {
    if (-not $process.HasExited) {
        $process.Kill()
        $process.WaitForExit(5000) | Out-Null
    }

    $stdout = $process.StandardOutput.ReadToEnd()
    $stderr = $process.StandardError.ReadToEnd()

    if ((-not $smokeSucceeded) -and $stdout) {
        Write-Host "login-server stdout:"
        Write-Host $stdout
    }

    if ((-not $smokeSucceeded) -and $stderr) {
        Write-Host "login-server stderr:"
        Write-Host $stderr
    }
}
