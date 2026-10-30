# ADR 0001：分离元数据源与下载源，新增 jdk_mirror

- 状态：已采纳
- 日期：2026-10-30
- 相关：IKRUWJ（IKE2K2 子任务）

## 背景

国内用户从 GitHub 下载 Temurin JDK 经常失败或速度极慢。JDM 原有的两个 URL 型配置都无法解决：

- `mirror` 是**元数据源**，指向 Adoptium API。国内镜像站不实现该 API，无法替代。
- `proxy` 是**代理**，只能给 github.com 的 URL 加前缀，无法指向按目录布局提供文件的**下载源**。

用户真正需要的是 nvm `node_mirror` 那种“一条配置切换下载来源”的体验。

## 决策

1. **元数据源与下载源分离**。新增 `jdk_mirror`（下载源基址）；`mirror` 继续表示元数据源，`proxy` 继续表示代理，三者互不替代。
2. **不复用 `download_mirror`**。它自 v0.1.0 起表示代理前缀，v0.1.1 已作为 `proxy` 的弃用别名。改语义会把升级用户的存量值（如 `https://ghproxy.net`）静默当作下载源，且无法可靠区分两种含义。
3. **代理语义不变**。仅当 URL 含 `github.com` 时拼接 `proxy`。
4. **默认下载源为清华 TUNA**（`https://mirrors.tuna.tsinghua.edu.cn/Adoptium`）。`mirror` 与 `proxy` 的默认值不变。
5. **固定镜像布局**：`{base}/{major}/jdk/x64/windows/{文件名}`，文件名逐字取自元数据源的 `package.name`，不解析 HTML 目录。首版只支持 Windows x64。
6. **候选链**：下载源 → `proxy` + 原始 GitHub 链接 → 原始 GitHub 链接，每个来源只尝试一次。
7. **完整性校验**：反序列化 `package.checksum`（SHA-256）与 `package.size`。连接/HTTP 错误、非 2xx、HTML 响应、字节数不符、摘要不符都判为失败并换源；checksum 缺失时告警并跳过哈希校验，但仍校验字节数。
8. **`none` 表示禁用**。`jdk_mirror` 与 `proxy` 都接受 `none`（大小写不敏感），原样保存以便 `config get` 明确显示，只在读取生效值时视为禁用。
9. **统一 User-Agent**：所有 HTTP 请求携带 `JDM/<version>`。部分镜像站会拒绝 Go 默认 UA。

## 影响

- 正向：国内用户开箱即用；镜像缺文件或内容异常时仍能完成安装；升级用户已有代理配置不受影响。
- 负向：安装多了一次元数据请求之外的候选尝试；镜像滞后时首次下载可能回退 GitHub，速度不稳定。
- 配置层新增一个可读写键，`config list`、shell 补全与安装目录配置模板同步更新。
- 不改变 `jdm ls-remote`、重复安装报错等现有行为。

## 备选方案

- **复用 `download_mirror` 作为下载源**：否决，存量语义是代理前缀，升级用户会被静默误解释。
- **让 `mirror` 同时承担两种职责**：否决，镜像站不实现 Adoptium API，版本解析会失败。
- **解析镜像 HTML 目录或内置版本清单**：否决，引入脆弱性与维护成本，首版只做基址 + 固定布局。
