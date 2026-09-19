<#
.SYNOPSIS
Creates the next patch release branch and its draft pull request.

.DESCRIPTION
Reads the current application version, increments the patch component, updates
tracked version references, creates and pushes a release branch, opens a draft
pull request, retargets the other open pull requests, and prints a Chinese
summary for the changelog preparation step.
#>

[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Invoke-CheckedCommand {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Command,

        [Parameter(Mandatory = $true)]
        [string[]]$Arguments,

        [int[]]$AllowedExitCodes = @(0)
    )

    $output = & $Command @Arguments 2>&1
    $exitCode = $LASTEXITCODE
    if ($exitCode -notin $AllowedExitCodes) {
        $details = ($output | Out-String).Trim()
        if ($details) {
            throw "命令执行失败（退出码 $exitCode）：$Command $($Arguments -join ' ')`n$details"
        }

        throw "命令执行失败（退出码 $exitCode）：$Command $($Arguments -join ' ')"
    }

    return $output
}

function ConvertFrom-CommandJson {
    param(
        [Parameter(Mandatory = $true)]
        [object[]]$InputObject
    )

    $json = ($InputObject -join "`n").Trim()
    if (-not $json) {
        throw 'GitHub CLI 没有返回预期的 JSON 数据。'
    }

    return $json | ConvertFrom-Json
}

function Read-TextFile {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Path
    )

    $bytes = [System.IO.File]::ReadAllBytes($Path)
    $preambleLength = 0

    # Detect the original byte-order mark explicitly. StreamReader's detected
    # UTF-8 encoding may emit a BOM when written back even if the source file
    # did not have one, which makes package.json invalid for some Node tools.
    if (
        $bytes.Length -ge 4 -and
        $bytes[0] -eq 0xFF -and $bytes[1] -eq 0xFE -and
        $bytes[2] -eq 0x00 -and $bytes[3] -eq 0x00
    ) {
        $encoding = [System.Text.UTF32Encoding]::new($false, $true, $true)
        $preambleLength = 4
    }
    elseif (
        $bytes.Length -ge 4 -and
        $bytes[0] -eq 0x00 -and $bytes[1] -eq 0x00 -and
        $bytes[2] -eq 0xFE -and $bytes[3] -eq 0xFF
    ) {
        $encoding = [System.Text.UTF32Encoding]::new($true, $true, $true)
        $preambleLength = 4
    }
    elseif (
        $bytes.Length -ge 3 -and
        $bytes[0] -eq 0xEF -and $bytes[1] -eq 0xBB -and $bytes[2] -eq 0xBF
    ) {
        $encoding = [System.Text.UTF8Encoding]::new($true, $true)
        $preambleLength = 3
    }
    elseif ($bytes.Length -ge 2 -and $bytes[0] -eq 0xFF -and $bytes[1] -eq 0xFE) {
        $encoding = [System.Text.UnicodeEncoding]::new($false, $true, $true)
        $preambleLength = 2
    }
    elseif ($bytes.Length -ge 2 -and $bytes[0] -eq 0xFE -and $bytes[1] -eq 0xFF) {
        $encoding = [System.Text.UnicodeEncoding]::new($true, $true, $true)
        $preambleLength = 2
    }
    else {
        $encoding = [System.Text.UTF8Encoding]::new($false, $true)
    }

    try {
        $content = $encoding.GetString(
            $bytes,
            $preambleLength,
            $bytes.Length - $preambleLength
        )
    }
    catch {
        throw "无法按文本编码读取 $Path：$($_.Exception.Message)"
    }

    return [pscustomobject]@{
        Content  = $content
        Encoding = $encoding
    }
}

function Write-TextFile {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Path,

        [Parameter(Mandatory = $true)]
        [string]$Content,

        [Parameter(Mandatory = $true)]
        [System.Text.Encoding]$Encoding
    )

    [System.IO.File]::WriteAllText($Path, $Content, $Encoding)
}

function Get-VersionFromFile {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Path
    )

    $file = Read-TextFile -Path $Path
    return Get-VersionFromText -Content $file.Content -Source $Path
}

function Get-VersionFromText {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Content,

        [Parameter(Mandatory = $true)]
        [string]$Source
    )

    $match = [regex]::Match(
        $Content,
        '(?m)^\s*const\s+Version\s*=\s*"(?<version>\d+\.\d+\.\d+)"\s*$'
    )
    if (-not $match.Success) {
        throw "无法从 $Source 中识别版本号，期望格式为 const Version = `"1.2.3`"。"
    }

    return $match.Groups['version'].Value
}

function Get-VersionFromGitReference {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Reference,

        [Parameter(Mandatory = $true)]
        [string]$Path
    )

    $content = @(
        Invoke-CheckedCommand -Command 'git' -Arguments @('show', "${Reference}:$Path")
    ) -join "`n"
    return Get-VersionFromText -Content $content -Source "${Reference}:$Path"
}

function Get-NextPatchVersion {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Version
    )

    $parts = $Version.Split('.')
    $major = [long]$parts[0]
    $minor = [long]$parts[1]
    $patch = [long]$parts[2]
    if ($patch -eq [long]::MaxValue) {
        throw "版本号 $Version 的修订号已无法继续增加。"
    }

    return "$major.$minor.$($patch + 1)"
}

function Update-VersionReferences {
    param(
        [Parameter(Mandatory = $true)]
        [string]$OldVersion,

        [Parameter(Mandatory = $true)]
        [string]$NewVersion
    )

    $matchingFiles = @(
        Invoke-CheckedCommand -Command 'git' -Arguments @(
            'grep', '-Il', '-F', $OldVersion, '--', '.'
        ) -AllowedExitCodes @(0, 1)
    )
    $versionPattern = [regex]::new(
        "(?<![0-9])$([regex]::Escape($OldVersion))(?![0-9])"
    )
    $updatedFiles = [System.Collections.Generic.List[string]]::new()

    foreach ($relativePath in $matchingFiles) {
        $relativePath = [string]$relativePath
        $fileName = [System.IO.Path]::GetFileName($relativePath)

        # Changelogs are release history and must never have old entries rewritten.
        if ($fileName -match '(?i)^CHANGELOG(?:\.|$)') {
            continue
        }

        $file = Read-TextFile -Path $relativePath
        $updatedContent = $file.Content
        if ($fileName -ieq 'package-lock.json') {
            $versionFieldLimit = 2
        }
        elseif ($fileName -ieq 'package.json') {
            $versionFieldLimit = 1
        }
        else {
            $versionFieldLimit = 0
        }

        if ($versionFieldLimit -gt 0) {
            # Only inspect the first one/two JSON version fields. This remains
            # safe when resuming after one of those fields was already updated.
            $jsonVersionPattern = [regex]::new(
                '(?m)"version"\s*:\s*"(?<value>[^"]+)"'
            )
            $versionFieldsToUpdate = @(
                $jsonVersionPattern.Matches($file.Content) |
                    Select-Object -First $versionFieldLimit |
                    Where-Object { $_.Groups['value'].Value -eq $OldVersion }
            )
            foreach ($versionField in ($versionFieldsToUpdate | Sort-Object Index -Descending)) {
                $valueGroup = $versionField.Groups['value']
                $updatedContent = $updatedContent.Remove($valueGroup.Index, $valueGroup.Length)
                $updatedContent = $updatedContent.Insert($valueGroup.Index, $NewVersion)
            }
        }
        else {
            $updatedContent = $versionPattern.Replace(
                $file.Content,
                { param($match) $NewVersion }
            )
        }

        if ($updatedContent -ne $file.Content) {
            Write-TextFile -Path $relativePath -Content $updatedContent -Encoding $file.Encoding
            $updatedFiles.Add($relativePath)
        }
    }

    return $updatedFiles
}

function Ensure-UnreleasedHeading {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Path
    )

    if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) {
        throw "找不到 changelog：$Path"
    }

    $file = Read-TextFile -Path $Path
    if ($file.Content -match '(?m)^##\s*\[Unreleased\]\s*$') {
        return $false
    }

    $newLine = if ($file.Content.Contains("`r`n")) { "`r`n" } else { "`n" }
    $releaseHeading = [regex]::Match(
        $file.Content,
        '(?m)^##\s+\[\d+\.\d+\.\d+(?:[-+][^\]]+)?\](?:\s+-\s+\d{4}-\d{2}-\d{2})?\s*$'
    )

    if ($releaseHeading.Success) {
        $updatedContent = $file.Content.Insert(
            $releaseHeading.Index,
            "## [Unreleased]$newLine$newLine"
        )
    }
    else {
        $separator = if ($file.Content.EndsWith($newLine)) { $newLine } else { "$newLine$newLine" }
        $updatedContent = "$($file.Content)$separator## [Unreleased]$newLine"
    }

    Write-TextFile -Path $Path -Content $updatedContent -Encoding $file.Encoding
    return $true
}

function Format-NumberList {
    param(
        [AllowEmptyCollection()]
        [object[]]$Numbers
    )

    if (-not $Numbers -or $Numbers.Count -eq 0) {
        return '无'
    }

    return (($Numbers | ForEach-Object { "#$_" }) -join '、')
}

foreach ($requiredCommand in @('git', 'gh')) {
    if (-not (Get-Command $requiredCommand -ErrorAction SilentlyContinue)) {
        throw "找不到必需命令：$requiredCommand"
    }
}

$originalConsoleOutputEncoding = [Console]::OutputEncoding
$utf8OutputEncoding = [System.Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = $utf8OutputEncoding
$OutputEncoding = $utf8OutputEncoding

# Build emoji labels from Unicode code points so Windows PowerShell 5.1 does
# not depend on the source file encoding when passing them to GitHub CLI.
$versionLabel = [char]::ConvertFromUtf32(0x1F4C8) + 'version'
$documentationLabel = [char]::ConvertFromUtf32(0x1F4D1) + 'documentation'

$repositoryRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repositoryRoot
try {
    $actualRoot = (
        Invoke-CheckedCommand -Command 'git' -Arguments @('rev-parse', '--show-toplevel')
    ) -join "`n"
    $actualRoot = [System.IO.Path]::GetFullPath($actualRoot.Trim())
    if ($actualRoot.TrimEnd('\', '/') -ne [System.IO.Path]::GetFullPath($repositoryRoot).TrimEnd('\', '/')) {
        throw "脚本所在项目与 Git 仓库根目录不一致：$actualRoot"
    }

    $versionCandidates = @(
        'internal/version/version.go',
        'internal/version.go'
    )
    $versionFile = $versionCandidates |
        Where-Object { Test-Path -LiteralPath $_ -PathType Leaf } |
        Select-Object -First 1
    if (-not $versionFile) {
        throw "找不到版本文件：$($versionCandidates -join ' 或 ')"
    }

    $initialVersion = Get-VersionFromFile -Path $versionFile
    Write-Host "启动时识别到版本：$initialVersion" -ForegroundColor DarkCyan
    $startingBranch = (
        Invoke-CheckedCommand -Command 'git' -Arguments @('branch', '--show-current')
    ) -join "`n"
    $startingBranch = $startingBranch.Trim()
    if (-not $startingBranch) {
        throw '当前处于 detached HEAD，无法确定发布分支。'
    }

    $releaseBranchMatch = [regex]::Match(
        $startingBranch,
        '^release/v(?<version>\d+\.\d+\.\d+)$'
    )
    $workingTreeStatus = @(
        Invoke-CheckedCommand -Command 'git' -Arguments @('status', '--porcelain', '--untracked-files=normal')
    )
    if ($workingTreeStatus.Count -gt 0 -and -not $releaseBranchMatch.Success) {
        throw "工作区存在未提交改动。请先提交或暂存这些改动，再重新运行脚本。`n$($workingTreeStatus -join "`n")"
    }

    Invoke-CheckedCommand -Command 'gh' -Arguments @('auth', 'status') | Out-Null
    $repositoryInfo = ConvertFrom-CommandJson -InputObject @(
        Invoke-CheckedCommand -Command 'gh' -Arguments @(
            'repo', 'view', '--json', 'nameWithOwner'
        )
    )
    $githubRepository = [string]$repositoryInfo.nameWithOwner
    $isResumingRelease = $false
    $releaseBranchExists = $false
    $remoteBranchExists = $false

    if ($releaseBranchMatch.Success) {
        # A failed run can leave the user on a partially updated release branch.
        # Use that branch's version instead of incrementing it again.
        $isResumingRelease = $true
        $releaseBranchExists = $true
        $releaseBranch = $startingBranch
        $newVersion = $releaseBranchMatch.Groups['version'].Value

        Invoke-CheckedCommand -Command 'git' -Arguments @('fetch', 'origin', 'main') | Out-Null
        $oldVersion = Get-VersionFromGitReference -Reference 'origin/main' -Path $versionFile
        $expectedVersion = Get-NextPatchVersion -Version $oldVersion
        if ($newVersion -ne $expectedVersion) {
            if ($newVersion -eq $oldVersion) {
                throw "release 分支版本 $newVersion 已经出现在 main 中；该发布可能已经完成。"
            }

            throw "当前分支 $releaseBranch 与 main 的下一修订版本 $expectedVersion 不匹配。"
        }

        $remoteBranch = @(
            Invoke-CheckedCommand -Command 'git' -Arguments @('ls-remote', '--heads', 'origin', $releaseBranch)
        )
        $remoteBranchExists = $remoteBranch.Count -gt 0
        if ($remoteBranchExists -and $workingTreeStatus.Count -eq 0) {
            Invoke-CheckedCommand -Command 'git' -Arguments @(
                'pull', '--ff-only', 'origin', $releaseBranch
            ) | Out-Null
        }
    }
    else {
        Invoke-CheckedCommand -Command 'git' -Arguments @('checkout', 'main') | Out-Null
        Invoke-CheckedCommand -Command 'git' -Arguments @('pull', '--ff-only', 'origin', 'main') | Out-Null

        $oldVersion = Get-VersionFromFile -Path $versionFile
        $newVersion = Get-NextPatchVersion -Version $oldVersion
        $releaseBranch = "release/v$newVersion"

        $localBranch = @(
            Invoke-CheckedCommand -Command 'git' -Arguments @('branch', '--list', $releaseBranch)
        )
        $releaseBranchExists = $localBranch.Count -gt 0
        $remoteBranch = @(
            Invoke-CheckedCommand -Command 'git' -Arguments @('ls-remote', '--heads', 'origin', $releaseBranch)
        )
        $remoteBranchExists = $remoteBranch.Count -gt 0

        if ($releaseBranchExists) {
            Invoke-CheckedCommand -Command 'git' -Arguments @('checkout', $releaseBranch) | Out-Null
            if ($remoteBranchExists) {
                Invoke-CheckedCommand -Command 'git' -Arguments @(
                    'pull', '--ff-only', 'origin', $releaseBranch
                ) | Out-Null
            }
            $isResumingRelease = $true
        }
        elseif ($remoteBranchExists) {
            Invoke-CheckedCommand -Command 'git' -Arguments @(
                'checkout', '-b', $releaseBranch, '--track', "origin/$releaseBranch"
            ) | Out-Null
            $releaseBranchExists = $true
            $isResumingRelease = $true
        }
    }

    $releaseTitle = "release: Upgrade version to v$newVersion"
    $releaseWorktreeVersion = Get-VersionFromFile -Path $versionFile
    if ($releaseWorktreeVersion -ne $oldVersion -and $releaseWorktreeVersion -ne $newVersion) {
        throw "发布分支中的版本为 $releaseWorktreeVersion，预期为 $oldVersion 或 $newVersion。"
    }

    if ($isResumingRelease) {
        Write-Host "检测到未完成的发布流程，将继续使用 $releaseBranch（版本 $newVersion）。" -ForegroundColor Cyan
    }
    else {
        Write-Host "识别到当前版本：$oldVersion；将创建版本：$newVersion。" -ForegroundColor Cyan
    }

    $labels = ConvertFrom-CommandJson -InputObject @(
        Invoke-CheckedCommand -Command 'gh' -Arguments @(
            'label', 'list', '--repo', $githubRepository,
            '--limit', '1000', '--json', 'name'
        )
    )
    $labelNames = @($labels | ForEach-Object { [string]$_.name })
    foreach ($requiredLabel in @($versionLabel, $documentationLabel)) {
        if ($requiredLabel -notin $labelNames) {
            throw "GitHub 仓库中不存在标签：$requiredLabel"
        }
    }

    $previousRelease = ConvertFrom-CommandJson -InputObject @(
        Invoke-CheckedCommand -Command 'gh' -Arguments @(
            'release', 'view', '--repo', $githubRepository,
            '--json', 'tagName,publishedAt,url'
        )
    )
    $previousReleaseTime = [DateTimeOffset]::Parse(
        $previousRelease.publishedAt,
        [System.Globalization.CultureInfo]::InvariantCulture
    )

    $existingOpenPullRequests = @(
        ConvertFrom-CommandJson -InputObject @(
            Invoke-CheckedCommand -Command 'gh' -Arguments @(
                'pr', 'list', '--repo', $githubRepository,
                '--state', 'open', '--limit', '10000',
                '--json', 'number,headRefName,baseRefName,isDraft,url'
            )
        ) | Sort-Object number
    )
    $releasePullRequests = @(
        $existingOpenPullRequests | Where-Object { $_.headRefName -eq $releaseBranch }
    )
    if ($releasePullRequests.Count -gt 1) {
        throw "分支 $releaseBranch 对应多个 Open PR，无法安全判断应该继续哪一个。"
    }

    $hasExistingReleasePullRequest = $releasePullRequests.Count -eq 1
    if ($hasExistingReleasePullRequest) {
        $newPullRequest = $releasePullRequests[0]
        $isResumingRelease = $true
    }

    $pullRequestsForRetargetPreview = @(
        $existingOpenPullRequests |
            Where-Object { -not $hasExistingReleasePullRequest -or $_.number -ne $newPullRequest.number }
    )
    $existingOpenPullRequestNumbers = @(
        $pullRequestsForRetargetPreview | ForEach-Object { $_.number }
    )
    $existingOpenPullRequestText = Format-NumberList -Numbers $existingOpenPullRequestNumbers
    $pendingChangesText = Format-NumberList -Numbers @()
    $pendingChanges = @(
        Invoke-CheckedCommand -Command 'git' -Arguments @('status', '--porcelain')
    )
    if ($pendingChanges.Count -gt 0) {
        $pendingChangesText = $pendingChanges -join "`n     "
    }

    Write-Host ''
    Write-Host '即将执行以下发布操作：' -ForegroundColor Yellow
    if ($releaseBranchExists) {
        Write-Host "  1. 继续使用分支 $releaseBranch，并完成版本 $newVersion 的更新"
    }
    else {
        Write-Host "  1. 创建分支 $releaseBranch 并把版本更新为 $newVersion"
    }
    Write-Host "  2. 提交并推送到 origin，提交信息为：$releaseTitle"
    if ($hasExistingReleasePullRequest) {
        Write-Host "  3. 继续使用已有 Draft PR #$($newPullRequest.number)"
    }
    else {
        Write-Host "  3. 创建带 $versionLabel、$documentationLabel 标签的 Draft PR，目标为 main"
    }
    Write-Host "  4. 将现有 $($pullRequestsForRetargetPreview.Count) 个 Open/Draft PR 的目标分支改为 $releaseBranch"
    Write-Host "     PR：$existingOpenPullRequestText"
    if ($pendingChanges.Count -gt 0) {
        Write-Host "  当前 release 分支上将一并提交的改动：`n     $pendingChangesText" -ForegroundColor Yellow
    }
    $confirmation = Read-Host '直接按 Enter 确认并继续；输入任意内容取消'
    if ($confirmation.Length -ne 0) {
        Write-Host '已取消，未创建发布分支，也未修改远程仓库。' -ForegroundColor Yellow
        return
    }

    if (-not $releaseBranchExists) {
        Invoke-CheckedCommand -Command 'git' -Arguments @('checkout', '-b', $releaseBranch) | Out-Null
        $releaseBranchExists = $true
    }

    $updatedFiles = @(Update-VersionReferences -OldVersion $oldVersion -NewVersion $newVersion)
    if ($updatedFiles.Count -gt 0) {
        Write-Host "已更新 $($updatedFiles.Count) 个版本文件。" -ForegroundColor Green
    }
    else {
        Write-Host "版本引用已经是 $newVersion，跳过重复替换。" -ForegroundColor Green
    }
    $unreleasedAdded = Ensure-UnreleasedHeading -Path 'CHANGELOG.md'
    if ($unreleasedAdded) {
        Write-Host '已在 CHANGELOG.md 中添加 ## [Unreleased]。' -ForegroundColor Green
    }

    Invoke-CheckedCommand -Command 'git' -Arguments @('diff', '--check') | Out-Null
    $updatedBackendVersion = Get-VersionFromFile -Path $versionFile
    if ($updatedBackendVersion -ne $newVersion) {
        throw "版本文件更新验证失败：预期 $newVersion，实际 $updatedBackendVersion。"
    }

    $pendingChanges = @(
        Invoke-CheckedCommand -Command 'git' -Arguments @('status', '--porcelain')
    )
    if ($pendingChanges.Count -gt 0) {
        Invoke-CheckedCommand -Command 'git' -Arguments @('add', '--all') | Out-Null
        Invoke-CheckedCommand -Command 'git' -Arguments @('commit', '-m', $releaseTitle) | Out-Null
    }
    else {
        Write-Host '发布改动已经提交，跳过重复 commit。' -ForegroundColor Green
    }
    Invoke-CheckedCommand -Command 'git' -Arguments @(
        'push', '--set-upstream', 'origin', $releaseBranch
    ) | Out-Null

    if (-not $hasExistingReleasePullRequest) {
        Invoke-CheckedCommand -Command 'gh' -Arguments @(
            'pr', 'create',
            '--draft',
            '--base', 'main',
            '--head', $releaseBranch,
            '--title', $releaseTitle,
            '--body=',
            '--label', $versionLabel,
            '--label', $documentationLabel,
            '--repo', $githubRepository
        ) | Out-Null

        $newPullRequest = ConvertFrom-CommandJson -InputObject @(
            Invoke-CheckedCommand -Command 'gh' -Arguments @(
                'pr', 'view', $releaseBranch,
                '--repo', $githubRepository,
                '--json', 'number,url,title,state,isDraft,headRefName,baseRefName'
            )
        )
    }

    $pullRequestsToRetarget = @(
        ConvertFrom-CommandJson -InputObject @(
            Invoke-CheckedCommand -Command 'gh' -Arguments @(
                'pr', 'list', '--repo', $githubRepository,
                '--state', 'open', '--limit', '10000',
                '--json', 'number,headRefName,baseRefName,isDraft,url'
            )
        ) | Where-Object { $_.number -ne $newPullRequest.number }
    )

    foreach ($pullRequest in $pullRequestsToRetarget) {
        if ($pullRequest.baseRefName -ne $releaseBranch) {
            Invoke-CheckedCommand -Command 'gh' -Arguments @(
                'pr', 'edit', [string]$pullRequest.number,
                '--repo', $githubRepository, '--base', $releaseBranch
            ) | Out-Null
        }
    }

    $allIssues = @(
        ConvertFrom-CommandJson -InputObject @(
            Invoke-CheckedCommand -Command 'gh' -Arguments @(
                'issue', 'list', '--repo', $githubRepository,
                '--state', 'all', '--limit', '10000',
                '--json', 'number,createdAt'
            )
        )
    )
    $issueNumbers = @(
        $allIssues |
            Where-Object {
                [DateTimeOffset]::Parse(
                    $_.createdAt,
                    [System.Globalization.CultureInfo]::InvariantCulture
                ) -gt $previousReleaseTime
            } |
            Sort-Object number |
            ForEach-Object { $_.number }
    )

    $currentOpenPullRequests = @(
        ConvertFrom-CommandJson -InputObject @(
            Invoke-CheckedCommand -Command 'gh' -Arguments @(
                'pr', 'list', '--repo', $githubRepository,
                '--state', 'open', '--limit', '10000', '--json', 'number'
            )
        ) |
            Where-Object { $_.number -ne $newPullRequest.number } |
            Sort-Object number
    )
    $openPullRequestNumbers = @($currentOpenPullRequests | ForEach-Object { $_.number })

    $issueText = Format-NumberList -Numbers $issueNumbers
    $pullRequestText = Format-NumberList -Numbers $openPullRequestNumbers
    $publishedAtText = $previousReleaseTime.ToString('yyyy-MM-dd HH:mm:ss zzz')
    $summary = @"
帮我维护项目，准备 v$newVersion 版本：

1. 先帮我检查 GitHub 上现有的 PR（$pullRequestText），并将它们合并进入 $releaseBranch 分支（注意不要默认合并进 main）：
   - 如果是 dependabot 的 PR，需要确认自动化 check 通过。对于未通过的，需要进行修改，直至其通过。
   - 对于其他 PR，需要检查自动化 check 是否通过、功能实现是否正确、代码是否符合规范、是否有潜在的安全问题等。对于不符合要求的 PR，需要进行修改，直至其通过；如果该 PR 的改动过大，或者对功能做出重大调整，告知我后，暂不处理。
   - 需要更新 CHANGELOG.md，记录非 dependabot 的改动内容。如果非 WCY-dt 或 dependabot 贡献了该 PR，则需要标注上贡献者名字。如果 PR 解决了某个 issue，则需要标注 issue 编号。
   - PR #$($newPullRequest.number) 需要更新 description，每行一条 `Fixed #<issue number>`，记录修复的 Issue。

2. 查看当前所有 Issue（$issueText），并进行实现：
   - 每实现一个功能，需要记得 commit 到 $releaseBranch 分支。
   - 有些 Issue 可能需要大改，或者对功能做出重大调整，告知我后，由我决定是否处理。
   - 有些 Issue 可能想法和功能很好，但细节上不够好，你可以主动分析并提出新的方案，然后按照新方案实现。
   - 有些 Issue 可能提供的信息不够充分，自动回复该 Issue，要求提供更多信息，暂不处理。
   - 有些 Issue 可能在修改后无法进行验证，这一类 Issue 需要直接修改，并在修改完后检查代码逻辑，以取代实际验证。
   - 开发时，不要一个功能就做一次验证，这样太浪费时间了。但需要每个 issue 或者功能进行一次 commit。
   - 开发全部完成后，需要更新 CHANGELOG.md，记录本次版本的改动内容。全部结束前要去掉 `## [Unreleased]`。
   - 开发全部完成后，如果有涉及到 UI 的功能，无需走测，但需要总结并告知我改动位置与触发条件，方便我手动测试。
   - PR #$($newPullRequest.number) 需要更新 description，每行一条 `Fixed #<issue number>`，记录修复的 Issue。

我本机有 `gh` 指令，你也许需要用到。

最终需要确保 PR 的自动化 check 通过。
"@

    Write-Output $summary
}
catch {
    Write-Error "创建新版本失败：$($_.Exception.Message)"
    exit 1
}
finally {
    Pop-Location
    [Console]::OutputEncoding = $originalConsoleOutputEncoding
}
