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
        [string[]]$Arguments
    )

    $output = & $Command @Arguments 2>&1
    $exitCode = $LASTEXITCODE
    if ($exitCode -ne 0) {
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

    $reader = [System.IO.StreamReader]::new($Path, [System.Text.Encoding]::UTF8, $true)
    try {
        $content = $reader.ReadToEnd()
        $encoding = $reader.CurrentEncoding
    }
    finally {
        $reader.Dispose()
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
    $match = [regex]::Match(
        $file.Content,
        '(?m)^\s*const\s+Version\s*=\s*"(?<version>\d+\.\d+\.\d+)"\s*$'
    )
    if (-not $match.Success) {
        throw "无法从 $Path 中识别版本号，期望格式为 const Version = `"1.2.3`"。"
    }

    return $match.Groups['version'].Value
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
        )
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
        $replacementLimit = [int]::MaxValue
        if ($fileName -ieq 'package-lock.json') {
            $replacementLimit = 2
        }
        elseif ($fileName -ieq 'package.json') {
            $replacementLimit = 1
        }

        $updatedContent = $versionPattern.Replace(
            $file.Content,
            { param($match) $NewVersion },
            $replacementLimit
        )
        if ($updatedContent -ne $file.Content) {
            Write-TextFile -Path $relativePath -Content $updatedContent -Encoding $file.Encoding
            $updatedFiles.Add($relativePath)
        }
    }

    if ($updatedFiles.Count -eq 0) {
        throw "项目中没有找到可从 $OldVersion 更新到 $NewVersion 的版本引用。"
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

    # Read the version before performing any Git mutation, as the release input.
    $initialVersion = Get-VersionFromFile -Path $versionFile
    $initialNextVersion = Get-NextPatchVersion -Version $initialVersion
    Write-Host "识别到当前版本：$initialVersion；候选新版本：$initialNextVersion" -ForegroundColor Cyan

    $workingTreeStatus = @(
        Invoke-CheckedCommand -Command 'git' -Arguments @('status', '--porcelain', '--untracked-files=normal')
    )
    if ($workingTreeStatus.Count -gt 0) {
        throw "工作区存在未提交改动。请先提交或暂存这些改动，再重新运行脚本。`n$($workingTreeStatus -join "`n")"
    }

    Invoke-CheckedCommand -Command 'gh' -Arguments @('auth', 'status') | Out-Null
    Invoke-CheckedCommand -Command 'git' -Arguments @('checkout', 'main') | Out-Null
    Invoke-CheckedCommand -Command 'git' -Arguments @('pull', '--ff-only', 'origin', 'main') | Out-Null

    # The remote update may have changed the version. Always release from the
    # version now present on the freshly updated main branch.
    $oldVersion = Get-VersionFromFile -Path $versionFile
    $newVersion = Get-NextPatchVersion -Version $oldVersion
    if ($oldVersion -ne $initialVersion) {
        Write-Host "拉取 main 后版本已更新为 $oldVersion；新版本调整为 $newVersion。" -ForegroundColor Yellow
    }

    $releaseBranch = "release/v$newVersion"
    $releaseTitle = "release: Upgrade version to v$newVersion"

    $localBranch = @(
        Invoke-CheckedCommand -Command 'git' -Arguments @('branch', '--list', $releaseBranch)
    )
    if ($localBranch.Count -gt 0) {
        throw "本地分支已存在：$releaseBranch"
    }

    $remoteBranch = @(
        Invoke-CheckedCommand -Command 'git' -Arguments @('ls-remote', '--heads', 'origin', $releaseBranch)
    )
    if ($remoteBranch.Count -gt 0) {
        throw "远程分支已存在：$releaseBranch"
    }

    $labels = ConvertFrom-CommandJson -InputObject @(
        Invoke-CheckedCommand -Command 'gh' -Arguments @('label', 'list', '--limit', '1000', '--json', 'name')
    )
    foreach ($requiredLabel in @('📈version', '📑documentation')) {
        if ($requiredLabel -notin @($labels | ForEach-Object { $_.name })) {
            throw "GitHub 仓库中不存在标签：$requiredLabel"
        }
    }

    $previousRelease = ConvertFrom-CommandJson -InputObject @(
        Invoke-CheckedCommand -Command 'gh' -Arguments @(
            'release', 'view', '--json', 'tagName,publishedAt,url'
        )
    )
    $previousReleaseTime = [DateTimeOffset]::Parse(
        $previousRelease.publishedAt,
        [System.Globalization.CultureInfo]::InvariantCulture
    )

    $existingOpenPullRequests = @(
        ConvertFrom-CommandJson -InputObject @(
            Invoke-CheckedCommand -Command 'gh' -Arguments @(
                'pr', 'list', '--state', 'open', '--limit', '10000',
                '--json', 'number,headRefName,baseRefName,isDraft,url'
            )
        ) | Sort-Object number
    )
    $existingOpenPullRequestNumbers = @(
        $existingOpenPullRequests | ForEach-Object { $_.number }
    )
    $existingOpenPullRequestText = Format-NumberList -Numbers $existingOpenPullRequestNumbers

    Write-Host ''
    Write-Host '即将执行以下发布操作：' -ForegroundColor Yellow
    Write-Host "  1. 创建分支 $releaseBranch 并把版本更新为 $newVersion"
    Write-Host "  2. 提交并推送到 origin，提交信息为：$releaseTitle"
    Write-Host '  3. 创建带 📈version、📑documentation 标签的 Draft PR，目标为 main'
    Write-Host "  4. 将现有 $($existingOpenPullRequests.Count) 个 Open/Draft PR 的目标分支改为 $releaseBranch"
    Write-Host "     PR：$existingOpenPullRequestText"
    $confirmation = Read-Host '直接按 Enter 确认并继续；输入任意内容取消'
    if ($confirmation.Length -ne 0) {
        Write-Host '已取消，未创建发布分支，也未修改远程仓库。' -ForegroundColor Yellow
        return
    }

    Invoke-CheckedCommand -Command 'git' -Arguments @('checkout', '-b', $releaseBranch) | Out-Null

    $updatedFiles = @(Update-VersionReferences -OldVersion $oldVersion -NewVersion $newVersion)
    $unreleasedAdded = Ensure-UnreleasedHeading -Path 'CHANGELOG.md'
    Write-Host "已更新 $($updatedFiles.Count) 个版本文件。" -ForegroundColor Green
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
    if ($pendingChanges.Count -eq 0) {
        throw '版本更新后没有产生任何待提交改动。'
    }

    Invoke-CheckedCommand -Command 'git' -Arguments @('add', '--all') | Out-Null
    Invoke-CheckedCommand -Command 'git' -Arguments @('commit', '-m', $releaseTitle) | Out-Null
    Invoke-CheckedCommand -Command 'git' -Arguments @(
        'push', '--set-upstream', 'origin', $releaseBranch
    ) | Out-Null

    Invoke-CheckedCommand -Command 'gh' -Arguments @(
        'pr', 'create',
        '--draft',
        '--base', 'main',
        '--head', $releaseBranch,
        '--title', $releaseTitle,
        '--body=',
        '--label', '📈version',
        '--label', '📑documentation'
    ) | Out-Null

    $newPullRequest = ConvertFrom-CommandJson -InputObject @(
        Invoke-CheckedCommand -Command 'gh' -Arguments @(
            'pr', 'view', $releaseBranch,
            '--json', 'number,url,title,state,isDraft,headRefName,baseRefName'
        )
    )

    $pullRequestsToRetarget = @(
        ConvertFrom-CommandJson -InputObject @(
            Invoke-CheckedCommand -Command 'gh' -Arguments @(
                'pr', 'list', '--state', 'open', '--limit', '10000',
                '--json', 'number,headRefName,baseRefName,isDraft,url'
            )
        ) | Where-Object { $_.number -ne $newPullRequest.number }
    )

    foreach ($pullRequest in $pullRequestsToRetarget) {
        if ($pullRequest.baseRefName -ne $releaseBranch) {
            Invoke-CheckedCommand -Command 'gh' -Arguments @(
                'pr', 'edit', [string]$pullRequest.number, '--base', $releaseBranch
            ) | Out-Null
        }
    }

    $allIssues = @(
        ConvertFrom-CommandJson -InputObject @(
            Invoke-CheckedCommand -Command 'gh' -Arguments @(
                'issue', 'list', '--state', 'all', '--limit', '10000',
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
                'pr', 'list', '--state', 'open', '--limit', '10000', '--json', 'number'
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
   - 需要更新 CHANGELOG.md，记录非 dependabot 的改动内容。
2. 查看当前所有 Issue（$issueText），并进行实现：
   - 每实现一个功能，需要记得 commit 到 $releaseBranch 分支；
   - 有些 Issue 可能需要大改，或者对功能做出重大调整，告知我后，暂不处理；
   - 有些 Issue 可能提供的信息不够充分，自动回复该 Issue，要求提供更多信息，暂不处理；
   - 有些 Issue 可能在修改后无法进行验证，这一类 Issue 需要直接修改，并在修改完后检查代码逻辑，以取代实际验证；
   - 开发时，不要一个功能就做一次验证，这样太浪费时间了；
   - 开发全部完成后，如果有涉及到 UI 的功能，基于改动大小，决定是否使用你的 computer use 能力进行 UI 测试验证；
   - 开发全部完成后，需要更新 CHANGELOG.md，记录本次版本的改动内容；
   - PR #$($newPullRequest.number) 需要更新 description，每行一条 `Fixed #<issue number>`，记录本次版本修复的 Issue；

我本机有 `gh` 指令，你也许需要用到。
"@

    Write-Output $summary
}
catch {
    Write-Error "创建新版本失败：$($_.Exception.Message)"
    exit 1
}
finally {
    Pop-Location
}
