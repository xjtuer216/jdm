# JDM 项目发布操作手册

**文档版本**: 2.0  
**项目**: JDM (JDK Version Manager)  
**适用对象**: 发布执行者

---

## 概述

本文档描述 JDM 任意版本的完整发布流程，适用于 Gitee 和 GitHub 双平台。文中 `<版本>` 为占位符（如 `0.1.3`，不带 v 前缀），标签名格式为 `v<版本>`。请按顺序执行以下步骤。

---

## 1. 预构建检查清单

### 1.1 检查当前状态

```powershell
# 查看当前分支
git branch

# 查看未提交的更改
git status

# 查看现有标签
git tag -l
```

**预期输出**: 
- 工作目录应干净（无未提交更改）
- `git tag -l` 可看到历史标签，确认目标 `v<版本>` 不存在
- 发布从 `develop` 切出 `release/<版本>` 分支开始（见第 2 节）

### 1.2 确保版本号一致

版本号由构建命令显式传入，**不要求修改** `installer/build.bat`、`installer/jdm.iss` 的默认值。发布前核对：

| 文件 | 核对点 |
|------|--------|
| README.md / README.en.md | 版本徽章 `version-<版本>-blue` 已更新为本次版本 |
| CHANGELOG.md | 已有本次版本的 `## [<版本>]` 小节及比较链接 |

**验证命令**:
```powershell
# 检查 README 徽章
grep "version-<版本>" README.md README.en.md

# 检查 CHANGELOG 小节
grep "\[<版本>\]" CHANGELOG.md
```

### 1.3 构建带版本号的 jdm.exe

```powershell
# 进入 core 目录
cd core

# 使用版本 <版本> 构建
go build -ldflags "-X github.com/xjtuer216/jdm/internal/jdk.Version=<版本>" -o jdm.exe .

# 验证版本
.\jdm.exe version
```

**预期输出**:
```
JDM version <版本>
```

### 1.4 重新构建安装程序

```powershell
# 返回项目根目录
cd ..

# 进入 installer 目录
cd installer

# 执行构建脚本，传入版本号
.\build.bat <版本>
```

**构建过程输出**:
```
========================================
JDM Installer Build Script
========================================

Version: <版本>

[1/4] Building jdm.exe...
[OK] jdm.exe built successfully

[2/4] Verifying assets...
[OK] Assets verified

[3/4] Creating output directory...
[OK] Output directory ready: D:\workspace\gitee\jdm\installer\dist

[4/4] Building installer...
[OK] Build Complete!

Installer: D:\workspace\gitee\jdm\installer\dist\jdm-setup-<版本>.exe
```

### 1.5 验证构建产物

```powershell
# 验证 jdm.exe 存在且版本正确
.\core\jdm.exe version

# 验证安装程序存在
dir installer\dist\jdm-setup-<版本>.exe

# 验证文件大小（应有约 8MB）
dir installer\dist\jdm-setup-<版本>.exe /-C
```

**预期输出**:
```
JDM version <版本>

    目录: D:\workspace\gitee\jdm\installer\dist

Mode                 LastWriteTime         Length Name
----                 -------------         ------ ----
-a----         2026/4/18     13:08        8209155 jdm-setup-<版本>.exe
```

### 1.6 构建全部发布资产（v0.1.2 起）

**发布资产契约**：每个 Release 必须附带以下 **三个** 资产，缺一不可：

| 资产 | 用途 | 构建方式 |
|------|------|----------|
| `jdm-setup-<版本>.exe` | 图形化安装向导（Inno Setup） | `installer\build.bat <版本>` |
| `jdm-windows-amd64.exe` | `jdm update` 自升级用的 CLI 裸二进制（命名契约见 `core/internal/selfupdate/selfupdate.go`） | 见下 |
| `checksums.txt` | 上述两个文件的 SHA-256 摘要清单（sha256sum 格式，每行 `<sha256>  <文件名>`） | 见下 |

```powershell
# 构建自升级 CLI 裸二进制（在 core 目录）
cd core
go build -ldflags "-s -w -X github.com/xjtuer216/jdm/internal/jdk.Version=<版本>" -o jdm-windows-amd64.exe .

# 生成摘要清单（在 installer\dist 目录，覆盖两个资产）
sha256sum jdm-setup-<版本>.exe jdm-windows-amd64.exe | awk '{print $1"  "$2}' > checksums.txt
```

> ⚠️ **v0.1.2 事故教训**：v0.1.2 发布时只上传了 `jdm-windows-amd64.exe` 和
> `checksums.txt`，漏掉了 `jdm-setup-0.1.2.exe`。用户从 Release 页面只能下载到
> CLI 裸二进制（console 子系统程序），双击后只闪过一个 cmd 窗口，没有任何安装
> 向导。上传前务必核对三个资产齐全（见 5.7 / 6.9 / 7.1 的检查清单）。
>
> 注：v0.1.2 事后补救时删除了 `jdm-windows-amd64.exe`，属特例——`jdm update`
> 只消费 latest release 的资产且本地已最新时提前返回，v0.1.2 的该资产不再有
> 消费路径。**v0.1.3 起必须恢复三资产齐全**；对应地，v0.1.2 的 checksums.txt
> 中残留的 exe 条目为悬空行，无实际影响，勿再改动。

---

## 2. 发布分支与 Git 标签

发布流程与项目实际实践一致（v0.1.2 先例：PR #14）：从 `develop` 切出 `release/<版本>` 分支 → 准备发布提交 → PR 合入 `master` → 在合并提交上打标签。**未授权禁止 force push `master`**。

### 2.1 创建发布分支

```powershell
# 从 develop 切出发布分支
git checkout develop
git pull origin develop
git checkout -b release/<版本>

# 在发布分支上完成发布准备提交（版本徽章、CHANGELOG 等），例如：
# chore(release): prepare <版本>

git push origin release/<版本>
```

### 2.2 合入 master

在 Gitee 或 GitHub 上从 `release/<版本>` 发起 PR 合入 `master`（标题如 `chore(release): prepare <版本>`），review 后合并，**不使用 rebase/squash**，保留合并提交。

### 2.3 在合并提交上创建标签

```powershell
# 同步 master 并确认在合并提交上
git checkout master
git pull origin master

# 创建带注释的标签
git tag -a v<版本> -m "Release v<版本>"

# 验证
git show v<版本>
```

**预期输出**:
```
tag v<版本>
Tagger: your-name <your-email>
Date:   ...

Release v<版本>
```

### 2.4 推送标签到两个远程

```powershell
# 标签不会自动推送，需显式推送
git push origin v<版本>
git push github v<版本>

# 验证
git ls-remote --tags origin
```

**预期输出**:
```
...        refs/tags/v<版本>
...        refs/tags/v<版本>^{}
```

> 标签打在 master 的合并提交上，删除 release 分支不影响标签与制品追溯。

### 2.5 回灌 develop（必做）

PR 合入 master 后，必须把 master 回灌到 develop，否则 develop 缺少发布准备提交（版本徽章、CHANGELOG），下个版本会出现 CHANGELOG 内容重复（v0.1.2 → v0.1.3 之间的实际事故：develop 的 `[Unreleased]` 与 master 的 `[0.1.2]` 小节重复，发版前不得不手工对账去重）。

采用回灌而非"向 master 和 develop 同时开 PR"：发版分支只需一个 PR（→ master），避免同样内容审查两遍；develop 保持单向接收 master 的集成分支角色。

```powershell
git checkout develop
git pull origin develop
git merge origin/master -m "chore(release): 回灌 master v0.1.3 发布提交到 develop"
# 预期：无冲突（发布准备提交均出自 develop 系历史）

# 验证内容一致（应无输出）
git diff origin/master develop --stat

git push origin develop
```

**预期输出**:
```
   49b474e..c6930f3  develop -> develop
```

> 双远程注意：GitHub 上可能因各自独立的 PR 合并提交（`#N` 与 `!N` SHA 不同）导致 `git push github develop` 非快进被拒。此时不要 force push，基于 `github/develop` 做一次同步合并再推：
>
> ```powershell
> git checkout -b sync/github-develop github/develop
> git merge develop -m "chore(release): 同步 develop 发布提交"
> git push github sync/github-develop:develop
> git checkout develop
> git branch -D sync/github-develop
> ```
>
> 判定标准：`git diff github/develop develop` 无输出即内容一致，仅合并提交 SHA 不同无需强推。

---

## 3. 同步代码到 GitHub

### 3.1 检查当前远程仓库

```powershell
# 查看当前配置的远程仓库
git remote -v
```

**预期输出**:
```
origin  https://gitee.com/xjtuer216/jdm.git (fetch)
origin  https://gitee.com/xjtuer216/jdm.git (push)
```

### 3.2 添加 GitHub 远程仓库

**情况 A**: 如果 GitHub 仓库已存在

```powershell
# 添加 GitHub 作为第二个远程仓库
git remote add github https://github.com/xjtuer216/jdm.git

# 验证添加成功
git remote -v
```

**情况 B**: 如果 GitHub 仓库不存在

1. 打开浏览器，访问 https://github.com/new
2. 填写仓库信息：
   - **Repository name**: `jdm`
   - **Description**: `JDM - JDK Version Manager for Windows`
   - **Visibility**: Public
   - 不要勾选 "Add a README file" 和 "Add .gitignore"
3. 点击 "Create repository"

4. 返回终端执行：

```powershell
# 添加 GitHub 远程
git remote add github https://github.com/xjtuer216/jdm.git
```

### 3.3 推送到 GitHub

```powershell
# 推送 master 与 develop 分支（标签已在 2.4 推送）
git push github master
git push github develop
```

**预期输出**:
```
Enumerating objects: ..., done.
Writing objects: 100% (...), done.
...    master -> master
```

### 3.4 验证同步

```powershell
# 查看两个远程的引用
git ls-remote origin
git ls-remote github

# 两个平台都应看到 master、develop 分支和 v<版本> 标签
```

---

## 4. 生成发布说明（Release Notes）

### 4.1 收集提交历史

```powershell
# 生成自上个标签以来的所有提交
git log --pretty=format:"- %s" v<上一版本>..HEAD
```

### 4.2 整理发布说明

创建 `RELEASE_NOTES.md` 文件（用于复制到发布页面）。内容以 CHANGELOG.md 中本次版本的 `## [<版本>]` 小节为准，模板如下：

```markdown
## JDM v<版本>

（一句话概括本次版本主题，取自 CHANGELOG）

### 主要变更

（从 CHANGELOG 本次版本小节摘录 Added/Changed/Fixed 要点）

### 安装方式

1. 下载 `jdm-setup-<版本>.exe`
2. 双击运行安装程序
3. 按向导提示完成安装
4. 打开新终端即可使用

### 已安装用户升级

运行 `jdm update` 即可自升级（校验 SHA-256 通过后替换，失败不影响现有版本）。

### 快速开始

```powershell
# 查看可用版本
jdm ls-remote

# 安装 JDK 17
jdm install 17

# 设置默认版本
jdm default 17

# 查看版本
jdm version
```

### 完整更新日志

详见 [CHANGELOG.md](CHANGELOG.md) 或提交历史。
```

---

## 5. 在 Gitee 发布 Release

### 5.1 登录 Gitee

1. 打开浏览器，访问 https://gitee.com/xjtuer216/jdm
2. 确保已登录你的 Gitee 账号

### 5.2 进入 Releases 页面

**方法一**: 直接在浏览器地址栏输入
```
https://gitee.com/xjtuer216/jdm/releases
```

**方法二**: 通过仓库页面导航
1. 进入仓库主页
2. 点击顶部导航栏的 "发行版" (Releases)

### 5.3 创建新 Release

1. 点击绿色按钮 **"创建发行版"** (Create Release)

### 5.4 填写发布信息

**表单字段填写指南**:

| 字段 | 值 | 说明 |
|------|-----|------|
| **标签版本** | `v<版本>` | 从下拉菜单选择 |
| **版本标题** | `JDM v<版本> - <简短描述>` | 格式: 项目名 v版本 - 简短描述 |
| **是否是预发布** | 不勾选 | 正式版本 |
| **版本说明** | （粘贴上面生成的 RELEASE_NOTES.md 内容） | 支持 Markdown |

### 5.5 上传全部发布资产

1. 找到 **"上传附件"** 区域
2. 点击 **"选择文件"** 或拖放文件
3. 选择文件（**三个都要传**，见 1.6 发布资产契约）：
   - `installer/dist/jdm-setup-<版本>.exe`
   - `installer/dist/jdm-windows-amd64.exe`（v0.1.2 起必需）
   - `installer/dist/checksums.txt`（v0.1.2 起必需）
4. 等待上传完成

**截图说明**: 
- 上传区域会显示文件名和大小 (约 8MB)
- 进度条显示上传进度

### 5.6 发布

1. 检查所有信息无误
2. 点击 **"创建发行版"** 按钮
3. 等待页面跳转到发布详情页

### 5.7 验证发布

```powershell
# 验证 Release 页面可访问
# 浏览器访问: https://gitee.com/xjtuer216/jdm/releases/tag/v<版本>
```

**检查清单**:
- [ ] 版本号显示为 <版本>
- [ ] 发布说明完整显示
- [ ] 附件 `jdm-setup-<版本>.exe` 存在且可下载
- [ ] 附件 `jdm-windows-amd64.exe` 存在（v0.1.2 起，selfupdate 契约）
- [ ] 附件 `checksums.txt` 存在且覆盖上述两个文件（v0.1.2 起）
- [ ] 文件大小正确 (~8MB)

---

## 6. 在 GitHub 发布 Release

### 6.1 登录 GitHub

1. 打开浏览器，访问 https://github.com/xjtuer216/jdm
2. 确保已登录你的 GitHub 账号

### 6.2 进入 Releases 页面

**方法一**: 直接在浏览器地址栏输入
```
https://github.com/xjtuer216/jdm/releases
```

**方法二**: 通过仓库页面导航
1. 进入仓库主页
2. 点击右侧边栏的 **"Releases"** 链接
3. 或点击顶部 "Tags" 标签，然后选择 "Releases"

### 6.3 创建新 Release

1. 点击绿色按钮 **"Create a new release"**

### 6.4 选择标签

1. 点击 **"Choose a tag"** 下拉框
2. 输入 `v<版本>` 并从列表选择
3. 标签应已存在（从 Gitee 推送过来）

### 6.5 填写发布信息

**表单字段填写指南**:

| 字段 | 值 | 说明 |
|------|-----|------|
| **Release title** | `JDM v<版本> - <简短描述>` | 与 Gitee 保持一致 |
| **Describe this release** | （粘贴相同的 RELEASE_NOTES.md 内容） | 支持 Markdown |

### 6.6 设置发布选项

**选项设置**:
- [ ] **Set as a pre-release** - 不勾选（这是正式版本）
- [ ] **Set as the latest release** - 勾选（首次发布）
- [ ] **Create a discussion for this release** - 可选

### 6.7 上传全部发布资产

1. 找到 **"Attach binaries by dropping them here or selecting them."** 区域
2. 点击打开文件选择器，或拖放文件
3. 选择文件（**三个都要传**，见 1.6 发布资产契约）：
   - `installer/dist/jdm-setup-<版本>.exe`
   - `installer/dist/jdm-windows-amd64.exe`（v0.1.2 起必需）
   - `installer/dist/checksums.txt`（v0.1.2 起必需）
4. 等待上传完成（显示文件大小和进度）

**截图说明**:
```
📎 jdm-setup-<版本>.exe  7.83 MB
[====================] 100%
```

### 6.8 发布

1. 点击底部绿色按钮 **"Publish release"**
2. 等待页面跳转到发布详情页

### 6.9 验证发布

```powershell
# 验证 Release 页面可访问
# 浏览器访问: https://github.com/xjtuer216/jdm/releases/tag/v<版本>
```

**检查清单**:
- [ ] 版本号显示为 <版本>
- [ ] Release 标题正确
- [ ] 发布说明完整显示
- [ ] Assets 列表包含 `jdm-setup-<版本>.exe`
- [ ] Assets 列表包含 `jdm-windows-amd64.exe`（v0.1.2 起，selfupdate 契约）
- [ ] Assets 列表包含 `checksums.txt` 且覆盖上述两个文件（v0.1.2 起）
- [ ] 文件大小正确 (~7.8 MB)
- [ ] 下载链接有效

---

## 7. 发布后验证

### 7.1 运行发布资产校验脚本

```bash
# 在仓库根目录（Git Bash）；脚本自动校验两个平台的资产与摘要，无需认证
./scripts/verify-release.sh v<版本>
```

脚本校验项：两个平台均存在 `jdm-setup-<版本>.exe`、`jdm-windows-amd64.exe`、`checksums.txt` 三个资产；`checksums.txt` 覆盖前两者且摘要与实际下载字节一致。**任何一项失败都不得宣布发布完成**（v0.1.2 为已记录的例外，见 1.6）。

### 7.2 验证 Gitee Release

```powershell
# 使用 curl 检查 Release 存在
curl -s https://gitee.com/api/v5/repos/xjtuer216/jdm/releases/tags/v<版本> | jq '.tag_name, .name'

# 核对资产齐全：应列出 jdm-setup-<版本>.exe、jdm-windows-amd64.exe、checksums.txt
curl -s https://gitee.com/api/v5/repos/xjtuer216/jdm/releases/tags/v<版本> | jq '.assets[].name'
```

**预期输出**:
```
"v<版本>"
"JDM v<版本> - <简短描述>"
```

### 7.3 验证 GitHub Release

```powershell
# 使用 curl 检查 Release 存在
curl -s https://api.github.com/repos/xjtuer216/jdm/releases/tags/v<版本> | jq '.tag_name, .name'

# 核对资产齐全：应列出 jdm-setup-<版本>.exe、jdm-windows-amd64.exe、checksums.txt
curl -s https://api.github.com/repos/xjtuer216/jdm/releases/tags/v<版本> | jq '.assets[].name'
```

**预期输出**:
```
"v<版本>"
"JDM v<版本> - <简短描述>"
```

### 7.4 下载并测试安装程序

```powershell
# 创建测试目录
mkdir C:\temp\jdm-test
cd C:\temp\jdm-test

# 从 Gitee 下载
curl -L -o jdm-setup-gitee.exe "https://gitee.com/xjtuer216/jdm/releases/download/v<版本>/jdm-setup-<版本>.exe"

# 从 GitHub 下载
curl -L -o jdm-setup-github.exe "https://github.com/xjtuer216/jdm/releases/download/v<版本>/jdm-setup-<版本>.exe"

# 验证文件完整性（比较哈希）
Get-FileHash jdm-setup-gitee.exe -Algorithm SHA256
Get-FileHash jdm-setup-github.exe -Algorithm SHA256
```

**预期结果**: 两个文件的 SHA256 哈希值应完全相同。

### 7.4 功能验证

```powershell
# 安装后测试
jdm version
# 应显示: JDM version <版本>

jdm help
# 应显示帮助信息

jdm ls-remote --limit 5
# 应显示远程 JDK 版本列表
```

### 7.5 更新 README 徽章

如果 README 中的下载链接需要更新：

```markdown
# 更新 README.md 中的下载链接

# 找到并替换：
[下载](https://gitee.com/xjtuer216/jdm/releases)
# 改为：
[下载](https://github.com/xjtuer216/jdm/releases)
```

---

## 8. 发布后通知

### 8.1 更新项目文档

- [ ] 更新 README.md 中的版本号（如果有硬编码）
- [ ] 确保安装指南指向正确的 Release 页面
- [ ] 更新 CHANGELOG.md（如果存在）

### 8.2 社区通知

- [ ] 在相关论坛/社区发布发布通知
- [ ] 更新项目文档站点（如果有）

---

## 附录

### A. 快速命令参考

```bash
# === 构建（installer\build.bat 会自动完成 1.3 步骤） ===
cd installer && ./build.bat <版本>

# === 构建自升级 CLI 与摘要清单（见 1.6） ===
cd ../core
go build -ldflags "-s -w -X github.com/xjtuer216/jdm/internal/jdk.Version=<版本>" -o jdm-windows-amd64.exe .
mv jdm-windows-amd64.exe ../installer/dist/
cd ../installer/dist
sha256sum jdm-setup-<版本>.exe jdm-windows-amd64.exe | awk '{print $1"  "$2}' > checksums.txt

# === 发布分支与标签（见第 2 节） ===
git checkout develop && git pull origin develop
git checkout -b release/<版本>
# ...发布准备提交后...
git push origin release/<版本>
# PR 合入 master 后：
git checkout master && git pull origin master
git tag -a v<版本> -m "Release v<版本>"
git push origin v<版本> && git push github v<版本>

# 回灌 develop（见 2.5，必做）：
git checkout develop && git pull origin develop
git merge origin/master
git diff origin/master develop --stat   # 应无输出
git push origin develop

# === 发布后校验（见 7.1） ===
./scripts/verify-release.sh v<版本>
```

### B. 常见问题

**Q: 标签推送失败**
```powershell
# 先确认本地标签指向 master 的合并提交，再排查网络/认证
# 禁止对已发布的标签 force push（会使已分发制品失去源码追溯）
```

**Q: GitHub 推送需要认证**
- 使用 HTTPS: 输入用户名和个人访问令牌 (PAT)
- 或使用 SSH: `git@github.com:xjtuer216/jdm.git`

**Q: 安装程序上传失败**
- 检查文件大小（GitHub 单文件限制 2GB，Gitee 限制 100MB）
- 检查网络连接
- 尝试使用浏览器手动上传

### C. 版本号格式

- **标签格式**: `v<版本>`（带 v 前缀）
- **版本号格式**: `<版本>`（语义化版本，不带 v 前缀）
- **安装程序**: `jdm-setup-<版本>.exe`（文件名）

### D. 相关链接

| 资源 | 链接 |
|------|------|
| Gitee 仓库 | https://gitee.com/xjtuer216/jdm |
| GitHub 仓库 | https://github.com/xjtuer216/jdm |
| Gitee Release | https://gitee.com/xjtuer216/jdm/releases |
| GitHub Release | https://github.com/xjtuer216/jdm/releases |

---

**文档版本**: 2.0  
**最后更新**: 2026-11-04  
**维护者**: JDM 开发团队
