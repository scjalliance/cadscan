[CmdletBinding(DefaultParameterSetName = "NoSignerCertificate")]
param (
  [System.IO.DirectoryInfo]
  $BuildDir = "build",

  [switch]
  $Logging = $false,

  [Parameter(ParameterSetName = "SignerCertificate", Mandatory = $true)]
  [System.Security.Cryptography.X509Certificates.X509Certificate2]
  $SignerCertificate,

  [Parameter(ParameterSetName = "SignerCertificate", Mandatory = $false)]
  [string]
  $TimestampServer = "http://timestamp.digicert.com"
)


# IMPORTANT NOTE OF USEFULNESS:
#   if you run this script with `-InformationAction Continue`,
#   you'll see the build log in the console directly.


function Write-ThisDown {
  [CmdletBinding()]
  param (
    [Parameter(ValueFromPipeline = $true)]
    $InputObject,

    [ValidateSet("Host", "Info", "Log")]
    [string[]]
    $Destination = @("Info", "Log"),

    [ValidateSet("List", "Table", "Wide", "String", "JSON", "Raw")]
    $Format = "Raw",

    [switch]
    $ShowAllProperties = $false,

    [switch]
    $PassThru
  )

  process {
    $WriteHost = $Destination -contains "Host"
    $WriteInfo = $Destination -contains "Info" -and (-not $WriteHost) -and $InformationPreference -ne "SilentlyContinue"
    $WriteLog  = $Destination -contains "Log" -and $Logging -and $LogFilePath

    function Write-ThisDownHere {
      [CmdletBinding()]
      param (
        [Parameter(ValueFromPipeline = $true)]
        $InputObject,

        [ValidateSet("Host", "Info", "Log")]
        $Medium = "Host",

        [ValidateSet("List", "Table", "Wide", "String", "JSON", "Raw")]
        $Format = "Raw",

        [switch]
        $ShowAllProperties = $false
      )

      process {
        $EffectiveMedium = $Medium

        if (-not @("Host", "Info", "Log") -contains $EffectiveMedium) {return} # Unknown medium
        if ($EffectiveMedium -eq "Host" -and (-not $WriteHost)) {return} # Host not requested
        if ($EffectiveMedium -eq "Info" -and (-not $WriteInfo)) {return} # Info not requested or `InformationAction == SilentlyContinue`
        if ($EffectiveMedium -eq "Log"  -and (-not $WriteLog )) {return} # Log not requested or is disabled

        if ($EffectiveMedium -eq "Host" -and (-not $Host.UI.SupportsVirtualTerminal)) {$EffectiveMedium = "NonVT-Host"}

        $formatListProps = @{}
        if ($ShowAllProperties) {$formatListProps = @{"Property"="*"}}

        $convertJsonProps = @{}

        switch ("$EffectiveMedium-$Format") {
          "Host-Raw"          {$InputObject | Out-Host}
          "Host-String"       {$InputObject | Out-String | Write-Host}
          "Host-List"         {$InputObject | Format-List     @formatListProps  | Out-Host}
          "Host-Table"        {$InputObject | Format-Table    @formatListProps  | Out-Host}
          "Host-Wide"         {$InputObject | Format-Wide     @formatListProps  | Out-Host}
          "Host-JSON"         {$InputObject | ConvertTo-Json  @convertJsonProps | Out-Host}
          "NonVT-Host-Raw"    {$InputObject | Out-String | Write-Host}
          "NonVT-Host-String" {$InputObject | Out-String | Write-Host}
          "NonVT-Host-List"   {$InputObject | Format-List     @formatListProps  | Out-String | Write-Host}
          "NonVT-Host-Table"  {$InputObject | Format-Table    @formatListProps  | Out-String | Write-Host}
          "NonVT-Host-Wide"   {$InputObject | Format-Wide     @formatListProps  | Out-String | Write-Host}
          "NonVT-Host-JSON"   {$InputObject | ConvertTo-Json  @convertJsonProps | Write-Host}
          "Info-Raw"          {$InputObject | Out-String | Write-Information}
          "Info-String"       {$InputObject | Out-String | Write-Information}
          "Info-List"         {$InputObject | Format-List     @formatListProps  | Out-String | Write-Information}
          "Info-Table"        {$InputObject | Format-Table    @formatListProps  | Out-String | Write-Information}
          "Info-Wide"         {$InputObject | Format-Wide     @formatListProps  | Out-String | Write-Information}
          "Info-JSON"         {$InputObject | ConvertTo-Json  @convertJsonProps | Write-Information}
          "Log-Raw"           {$InputObject | Out-File -FilePath $LogFilePath -Append}
          "Log-String"        {$InputObject | Out-String | Out-File -FilePath $LogFilePath -Append}
          "Log-List"          {$InputObject | Format-List     @formatListProps  | Out-File -FilePath $LogFilePath -Append}
          "Log-Table"         {$InputObject | Format-Table    @formatListProps  | Out-File -FilePath $LogFilePath -Append}
          "Log-Wide"          {$InputObject | Format-Wide     @formatListProps  | Out-File -FilePath $LogFilePath -Append}
          "Log-JSON"          {$InputObject | ConvertTo-Json  @convertJsonProps | Out-File -FilePath $LogFilePath -Append}
          Default             {}
        }
      }
    }

    if ($WriteHost) {
      Write-ThisDownHere -InputObject $InputObject -Medium "Host" -Format $Format -ShowAllProperties:$ShowAllProperties
    }

    if ($WriteInfo) {
      Write-ThisDownHere -InputObject $InputObject -Medium "Info" -Format $Format -ShowAllProperties:$ShowAllProperties
    }

    if ($WriteLog) {
      Write-ThisDownHere -InputObject $InputObject -Medium "Log"  -Format $Format -ShowAllProperties:$ShowAllProperties
    }
  }

  end {
    if  ($PassThru) {
      $inputObject
    }
  }
}


# Mark the start time of the build.
$Started = Get-Date


# Define our log file path
$LogFilePath = $null
if ($Logging) {
  $LogFilePath = "$BuildDir\build.log"
}


# Define our output executable path.
$OutputExe = "$BuildDir\cadscan.exe"


# Make sure the build directory exists or create it.
if (-not (Test-Path $BuildDir)) {
  if (Test-Path $BuildDir -PathType Container) {
    throw "Build directory ($BuildDir) is not a directory; bailing out."
  }

  New-Item -ItemType Directory -Path $BuildDir | Out-Null
  if (-not (Test-Path $BuildDir -PathType Container)) {
    throw "Build directory ($BuildDir) does not exist and was not created; bailing out."
  }
}


# Mark the start of the build in the build log.
"[Started] $Started" | Write-ThisDown


# Grab the commit hash, if we're in a git repo.
$GitInfo = $null
if (Test-Path .git -PathType Container) {
  $GitInfo = [PSCustomObject]@{
    Commit = git rev-parse HEAD
    Branch = git rev-parse --abbrev-ref HEAD
    Remote = git remote -v
  }
}


# Time to build!
go build -ldflags -H=windowsgui -o $OutputExe | Write-ThisDown


# Sign the executable, if we have a signer certificate in hand.
$Authenticode = $null
if ($SignerCertificate) {
  Set-AuthenticodeSignature -FilePath $OutputExe -Certificate $SignerCertificate -TimestampServer $TimestampServer -IncludeChain all | Write-ThisDown

  $fileSignature = Get-AuthenticodeSignature -FilePath $OutputExe -Verbose  | Write-ThisDown -Format List -PassThru

  $Authenticode = [PSCustomObject]@{
    Valid = $fileSignature.Status -eq "Valid"
    Status = $fileSignature.Status
    StatusMessage = $fileSignature.StatusMessage
    SignatureType = $fileSignature.SignatureType
    Thumbprint = $fileSignature.SignerCertificate.Thumbprint
    Subject = $fileSignature.SignerCertificate.Subject
    Issuer = $fileSignature.SignerCertificate.Issuer
    TimestampCertificate = $fileSignature.TimeStamperCertificate
    SignerCertificate = $fileSignature.SignerCertificate
  }
}


# Grab the output file's properties.
$OutputFile = Get-ChildItem -Path $OutputExe | Write-ThisDown -PassThru


# Mark the end time of the build.
$Finished = Get-Date


# Use a PSCustomObject to make the output usable by other scripts
$Result = [PSCustomObject]@{
  DateTime = [PSCustomObject]@{
    Started = $Started
    Finished = $Finished
  }
  Versions = [PSCustomObject]@{
    PowerShell = $PSVersionTable.PSVersion
    Go = go version
    Git = git version
  }
  Git = $GitInfo
  Authenticode = $Authenticode
  OutputFile = $OutputFile
}


# Log that output to the build log...
$Result | Write-ThisDown -ShowAllProperties -Format List -Destination Log


# Mark the end of the build in the build log.
"[Finished] $Finished" | Write-ThisDown


# Hand this back to the caller for their consumption.
$Result


# ❤️
