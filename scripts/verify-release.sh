#!/usr/bin/env bash
# ============================================================================
# JDM 发布资产校验脚本
# ============================================================================
# 用法: ./scripts/verify-release.sh v<版本>
# 示例: ./scripts/verify-release.sh v0.1.3
#
# 校验 Gitee 与 GitHub 两个平台的 release（见 docs/RELEASE_OPERATIONS.md 1.6/7.1）:
#   1. 三个资产齐全: jdm-setup-<版本>.exe、jdm-windows-amd64.exe、checksums.txt
#   2. checksums.txt 覆盖前两个文件
#   3. 每个文件的实际下载字节 SHA-256 与 checksums.txt 条目一致
#
# 任何一项 FAIL 都意味着发布未完成，不得宣布发布成功。
# 依赖: curl, sha256sum, awk（Git Bash 自带）。无需认证（公开仓库）。
# ============================================================================
set -u

TAG="${1:-}"
if [ -z "$TAG" ]; then
    echo "usage: $0 v<版本>   (例: $0 v0.1.3)" >&2
    exit 2
fi

GITEE_API="${JDM_VERIFY_GITEE_API:-https://gitee.com/api/v5/repos/xjtuer216/jdm/releases/tags/$TAG}"
GITEE_DL="${JDM_VERIFY_GITEE_DL:-https://gitee.com/xjtuer216/jdm/releases/download/$TAG}"
GITHUB_API="${JDM_VERIFY_GITHUB_API:-https://api.github.com/repos/xjtuer216/jdm/releases/tags/$TAG}"
GITHUB_DL="${JDM_VERIFY_GITHUB_DL:-https://github.com/xjtuer216/jdm/releases/download/$TAG}"

SETUP_EXE="jdm-setup-${TAG#v}.exe"
CLI_EXE="jdm-windows-amd64.exe"
SUMS="checksums.txt"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
FAILED=0

check() {
    if [ "$2" -eq 0 ]; then
        printf 'PASS  %s\n' "$1"
    else
        printf 'FAIL  %s\n' "$1"
        FAILED=1
    fi
}

has_asset() { # has_asset <meta.json> <asset name>（-F 固定串匹配，避免文件名中的点被当正则）
    grep -qF "\"name\":\"$2\"" "$1" || grep -qF "\"name\": \"$2\"" "$1"
}

for platform in gitee github; do
    if [ "$platform" = gitee ]; then
        api="$GITEE_API"; dl="$GITEE_DL"
    else
        api="$GITHUB_API"; dl="$GITHUB_DL"
    fi

    meta="$TMP/$platform.json"
    if ! curl -sfL "$api" -o "$meta"; then
        echo "FAIL  $platform: release $TAG 元数据不可达 ($api)"
        FAILED=1
        continue
    fi
    grep -qF "\"tag_name\":\"$TAG\"" "$meta" || grep -qF "\"tag_name\": \"$TAG\"" "$meta"
    check "$platform: release $TAG 存在（tag_name 匹配）" $?

    for asset in "$SETUP_EXE" "$CLI_EXE" "$SUMS"; do
        has_asset "$meta" "$asset"
        check "$platform: 资产 $asset 存在" $?
    done

    if ! curl -sfL "$dl/$SUMS" -o "$TMP/$platform-$SUMS"; then
        echo "FAIL  $platform: $SUMS 下载失败"
        FAILED=1
        continue
    fi

    for asset in "$SETUP_EXE" "$CLI_EXE"; do
        want="$(awk -v f="$asset" '$2==f {print $1}' "$TMP/$platform-$SUMS")"
        [ -n "$want" ]
        check "$platform: $SUMS 含 $asset 条目" $?
        [ -n "$want" ] || continue

        if ! curl -sfL "$dl/$asset" -o "$TMP/$asset"; then
            # 下载 CDN 不可达时（如 objects.githubusercontent.com 被墙），
            # 回退用 release 元数据中的 per-asset digest 核对：GitHub API 资产
            # 自带 "digest": "sha256:<hex>"。元数据在前面已拉取，无需再联网下载文件。
            if [ "$platform" = github ] && grep -qF "sha256:$want" "$meta"; then
                check "$platform: $asset SHA-256 与 $SUMS 一致（API digest 兜底，未逐字节下载）" 0
                continue
            fi
            echo "FAIL  $platform: $asset 下载失败"
            FAILED=1
            continue
        fi
        got="$(sha256sum "$TMP/$asset" | awk '{print $1}')"
        [ "$got" = "$want" ]
        check "$platform: $asset SHA-256 与 $SUMS 一致" $?
    done
done

if [ "$FAILED" -eq 0 ]; then
    echo "OK    release $TAG 发布资产校验全部通过"
fi
exit "$FAILED"
