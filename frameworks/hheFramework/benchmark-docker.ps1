[CmdletBinding()]
param(
    [ValidateSet("plain", "seal", "helib", "tfhe")]
    [string]$Backend,

    [ValidateRange(1, 1000000)]
    [int]$Iterations = 3,

    [ValidateRange(0, 1000000)]
    [int]$Warmups = 1,

    [ValidateSet("warn", "strict")]
    [string]$Policy = "warn",

    [ValidateRange(1, 1024)]
    [int]$Jobs = [Environment]::ProcessorCount,

    [Nullable[int]]$Cpu,

    [string]$SecurityProfile = "",

    [switch]$SkipImageBuild,
    [switch]$SkipCompile,
    [switch]$SkipReport
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Invoke-Checked {
    param(
        [Parameter(Mandatory)]
        [string]$File,

        [Parameter(Mandatory)]
        [string[]]$Arguments
    )

    & $File @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "$File exited with code $LASTEXITCODE"
    }
}

function Select-HheBackend {
    Write-Host "Choose the HHE Framework backend:"
    Write-Host "  1. plain"
    Write-Host "  2. SEAL"
    Write-Host "  3. HElib"
    Write-Host "  4. TFHE"

    while ($true) {
        switch (Read-Host "Selection [1-4]") {
            "1" { return "plain" }
            "2" { return "seal" }
            "3" { return "helib" }
            "4" { return "tfhe" }
            default { Write-Warning "Enter 1, 2, 3, or 4." }
        }
    }
}

function Get-CmakeCipherList {
    param(
        [Parameter(Mandatory)]
        [string]$CmakeFile,

        [Parameter(Mandatory)]
        [string]$Variable
    )

    $content = [IO.File]::ReadAllText($CmakeFile)
    $pattern = "(?ms)set\(\s*" + [Regex]::Escape($Variable) + "\s+(?<body>.*?)\)"
    $match = [Regex]::Match($content, $pattern)
    if (-not $match.Success) {
        throw "Could not find CMake cipher list $Variable in $CmakeFile"
    }

    $withoutComments = (($match.Groups["body"].Value -split "\r?\n") |
        ForEach-Object { $_ -replace "#.*$", "" }) -join " "

    $ciphers = @($withoutComments -split "\s+" | Where-Object { $_ })
    if ($ciphers.Count -eq 0) {
        throw "CMake cipher list $Variable is empty"
    }
    return $ciphers
}

if (-not $Backend) {
    $Backend = Select-HheBackend
}
$Backend = $Backend.ToLowerInvariant()

$repoRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot "../.."))
$cmakeFile = Join-Path $PSScriptRoot "CMakeLists.txt"
$builderDockerfile = Join-Path $PSScriptRoot "docker/Dockerfile.builder"
$builderImage = "hhe-arena:hhe-framework-builder"
$toolingImage = "hhe-arena:tooling"

$backendConfig = @{
    plain = @{ List = "PLAIN_CIPHS"; Suffix = "_test" }
    seal  = @{ List = "SEAL_CIPHS";  Suffix = "_seal_test" }
    helib = @{ List = "HELIB_CIPHS"; Suffix = "_helib_test" }
    tfhe  = @{ List = "TFHE_CIPHS";  Suffix = "_tfhe_test" }
}
$config = $backendConfig[$Backend]
$ciphers = @(Get-CmakeCipherList -CmakeFile $cmakeFile -Variable $config.List)
$targets = @($ciphers | ForEach-Object { "{0}{1}" -f $_, $config.Suffix })

Invoke-Checked -File "docker" -Arguments @("version")

if (-not $SkipImageBuild) {
    Invoke-Checked -File "docker" -Arguments @(
        "build", "--pull=false",
        "-t", $builderImage,
        "-f", $builderDockerfile,
        $repoRoot
    )

    if (-not $SkipReport) {
        Invoke-Checked -File "docker" -Arguments @(
            "build", "--pull=false",
            "-t", $toolingImage,
            $repoRoot
        )
        Invoke-Checked -File "docker" -Arguments @(
            "run", "--rm", $toolingImage, "probe"
        )
    }
}

$mount = "type=bind,source=$repoRoot,target=/workspace"
$buildDir = "build/hhe-framework-$Backend"

if (-not $SkipCompile) {
    $targetText = $targets -join " "
    $buildCommand = "sed -i 's/\r$//' frameworks/hheFramework/thirdparty/build.sh && cmake -S frameworks/hheFramework -B $buildDir -DCMAKE_BUILD_TYPE=RelWithDebInfo -DBUILD_TESTS=ON && cmake --build $buildDir --parallel $Jobs --target $targetText"

    Write-Host "Compiling $($targets.Count) $Backend targets in Docker..."
    Invoke-Checked -File "docker" -Arguments @(
        "run", "--rm",
        "--mount", $mount,
        "--workdir", "/workspace",
        $builderImage,
        "bash", "-lc", $buildCommand
    )
}

$resultDirHost = Join-Path $repoRoot "results/hhe-framework/$Backend"
[IO.Directory]::CreateDirectory($resultDirHost) | Out-Null
$rawHost = Join-Path $resultDirHost "raw.jsonl"
if (Test-Path -LiteralPath $rawHost) {
    Remove-Item -LiteralPath $rawHost -Force
}

$rawContainer = "/workspace/results/hhe-framework/$Backend/raw.jsonl"
$libraryPath = @(
    "/workspace/frameworks/hheFramework/thirdparty/m4ri/installed/lib",
    "/workspace/frameworks/hheFramework/thirdparty/HElib/install/lib",
    "/workspace/frameworks/hheFramework/thirdparty/NTL/usr/local/lib",
    "/workspace/frameworks/hheFramework/thirdparty/GMP/usr/local/lib",
    "/workspace/frameworks/hheFramework/thirdparty/tfhe/installed/lib"
) -join ":"

for ($index = 0; $index -lt $ciphers.Count; $index++) {
    $cipher = $ciphers[$index]
    $target = $targets[$index]
    Write-Host ("[{0}/{1}] Benchmarking {2} ({3})" -f ($index + 1), $ciphers.Count, $cipher, $Backend)

    $runArguments = @(
        "run", "--rm",
        "--mount", $mount,
        "--workdir", "/workspace",
        "-e", "HHE_BENCH_CIPHER=$cipher",
        "-e", "HHE_BENCH_LANGUAGE=cpp",
        "-e", "HHE_BENCH_BACKEND=$Backend",
        "-e", "HHE_BENCH_IMPLEMENTATION=hhe-framework-$cipher-$Backend",
        "-e", "HHE_BENCH_ITERATIONS=$Iterations",
        "-e", "HHE_BENCH_WARMUPS=$Warmups",
        "-e", "HHE_BENCH_POLICY=$Policy",
        "-e", "HHE_BENCH_OUTPUT=$rawContainer",
        "-e", "LD_LIBRARY_PATH=$libraryPath"
    )

    if ($null -ne $Cpu) {
        $runArguments += @(
            "--cpuset-cpus", ([int]$Cpu).ToString(),
            "-e", "HHE_BENCH_CPU=$([int]$Cpu)"
        )
    }
    if ($SecurityProfile) {
        $runArguments += @("-e", "HHE_BENCH_SECURITY_PROFILE=$SecurityProfile")
    }

    $runArguments += @(
        $builderImage,
        "./$buildDir/tests/$target"
    )
    Invoke-Checked -File "docker" -Arguments $runArguments
}

if (-not (Test-Path -LiteralPath $rawHost)) {
    throw "No benchmark JSONL was produced at $rawHost"
}

if (-not $SkipReport) {
    $resultsMount = "type=bind,source=$(Join-Path $repoRoot 'results'),target=/work/results"
    $rawForReport = "results/hhe-framework/$Backend/raw.jsonl"
    $reportForContainer = "results/hhe-framework/$Backend/report"

    Invoke-Checked -File "docker" -Arguments @(
        "run", "--rm",
        "--mount", $resultsMount,
        $toolingImage,
        "report", $rawForReport,
        "--output-dir", $reportForContainer,
        "--schema", "/opt/hhe/schemas/benchmark.schema.json"
    )
}

Write-Host "Raw results: $rawHost"
if (-not $SkipReport) {
    Write-Host "Reports: $(Join-Path $resultDirHost 'report')"
}
