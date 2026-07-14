# javdbapi

javdbapi 是一个用于查询 JavDB 网站的 Go SDK 和命令行工具。它提供类型安全的 API、自动限流重试、智能缓存、以及便捷的 CLI 命令。

## 核心特性

- **类型安全**：每种资源 ID（视频、演员、片商等）都是独立的强类型，防止混用
- **自动限流**：内置令牌桶限流器，默认每秒 1 个请求，防止被封
- **智能重试**：指数退避 + 抖动，自动处理 429/502/503/504
- **独立缓存**：Detail 和 Reviews 分别缓存，互不影响
- **演员名归档**：Actor 命令自动按演员名创建子目录
- **番号解析**：Video 命令支持直接输入番号，自动查找对应 ID

## 环境要求

- Go 1.25+

## 安装

```bash
go install github.com/RPbro/javdbapi/cmd/javdbapi@latest
```

## 初始化 SDK Client

```go
package main

import (
    "log"
    "time"

    "github.com/RPbro/javdbapi"
)

func main() {
    client, err := javdbapi.NewClient(javdbapi.ClientConfig{})
    if err != nil {
        log.Fatal(err)
    }
    _ = client
}
```

所有配置项都有合理的默认值，留空即可使用默认配置：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `BaseURL` | `https://javdb573.com` | 站点地址 |
| `HTTP.Timeout` | 30s | 请求超时 |
| `RateLimit.RequestsPerSecond` | 1 | 每秒请求数 |
| `Retry.MaxAttempts` | 3 | 最大重试次数 |

## 使用 SDK

### 列表查询

以下命令都返回 `Page[VideoSummary]`，通过 `HasNext` 判断是否还有下一页：

```go
// 首页列表
home, _ := client.Home(ctx, javdbapi.HomeQuery{
    Type:   javdbapi.HomeTypeCensored,
    Filter: javdbapi.HomeFilterAll,
    Sort:   javdbapi.HomeSortPublishDate,
    Page:   1,
})

// 搜索
search, _ := client.Search(ctx, javdbapi.SearchQuery{
    Keyword: "VR",
    Page:    1,
})

// 片商作品
maker, _ := client.MakerVideos(ctx, javdbapi.MakerVideosQuery{
    MakerID: javdbapi.MakerID("7R"),
    Filter:  javdbapi.MakerFilterAll,
    Page:    1,
})

// 演员作品
actor, _ := client.ActorVideos(ctx, javdbapi.ActorVideosQuery{
    ActorID: javdbapi.ActorID("neRNX"),
    Filters: []javdbapi.ActorFilter{javdbapi.ActorFilterCNSub},
    Page:    1,
})

// 排行榜
ranking, _ := client.Ranking(ctx, javdbapi.RankingQuery{
    Period: javdbapi.RankingPeriodWeekly,
    Type:   javdbapi.RankingTypeCensored,
    Page:   1,
})
```

### 详情与评论

```go
// 视频详情（含演员、标签、截图、磁力链接）
detail, _ := client.Detail(ctx, javdbapi.VideoID("ZNdEbV"))

// 视频评论（独立请求，失败不影响详情）
reviews, _ := client.Reviews(ctx, javdbapi.VideoID("ZNdEbV"))
```

### 强类型 ID

```go
// 只能通过 Parse 函数构造，防止非法 ID
videoID, _ := javdbapi.ParseVideoID("ZNdEbV")
actorID, _ := javdbapi.ParseActorID("neRNX")
makerID, _ := javdbapi.ParseMakerID("7R")

// 番号自动解析（支持直接输入番号，自动搜索对应 ID）
resolvedID, _ := javdbapi.ResolveVideoID("DLDSS-271")
```

## CLI 命令行工具

### 命令一览

| 命令 | 说明 | 必填参数 |
|------|------|----------|
| `video` | 查看视频详情 | `--id` |
| `search` | 搜索视频 | `--keyword` |
| `actor` | 查看演员作品 | `--id` |
| `actor-detail` | 查看演员姓名和别名 | `--id` |
| `maker` | 查看片商作品 | `--id` |
| `home` | 浏览首页列表 | 无 |
| `ranking` | 查看排行榜 | `--period`, `--type` |

### 快速开始

```bash
# 查看视频详情
javdbapi video --id ZNdEbV --output console

# 搜索视频
javdbapi search --keyword "VR" --output console

# 查看演员作品（自动按演员名创建子目录）
javdbapi actor --id WE4e --output both

# 查看演员姓名和别名
javdbapi actor-detail --id WE4e --output console

# 查看排行榜
javdbapi ranking --period weekly --type censored --output console
```

### 共享参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--output` | `file` | 输出模式：`file`（仅文件）、`console`（仅控制台）、`both`（两者） |
| `--output-dir` | `.` | 输出目录 |
| `--stale-after` | `24h` | 缓存有效期，`0s` 强制刷新 |
| `--concurrency` | `2` | 并发抓取数（1-16） |
| `--timeout` | `30s` | HTTP 超时 |
| `--proxy-url` | 无 | HTTP/SOCKS5 代理地址 |
| `--base-url` | `https://javdb573.com` | 站点地址 |
| `--rate` | `1` | 每秒请求数 |
| `--burst` | `1` | 限流突发容量 |
| `--debug` | `false` | 启用调试日志 |
| `--fail-fast` | `false` | 遇到第一个错误即停止 |

### 各命令参数

**search**
```bash
javdbapi search --keyword "VR" --page 1 --max-pages 3 --output both
```

**home**
```bash
javdbapi home --type censored --filter all --sort publish
javdbapi home --sort magnet --output console
```

**actor**
```bash
javdbapi actor --id neRNX --filter cnsub,download
javdbapi actor --id neRNX --filter c,d   # 兼容旧别名
```

**maker**
```bash
javdbapi maker --id 7R --filter playable --output both
```

**ranking**
```bash
javdbapi ranking --period weekly --type censored
javdbapi ranking --period daily --type western --stale-after 0s --output console
```

**video**
```bash
javdbapi video --id ZNdEbV --output console
javdbapi video --id DLDSS-271 --output console  # 支持直接输番号
```

### 程序化使用

```bash
# 用 jq 提取番号
javdbapi video --id ZNdEbV --output console | jq '.detail.summary.code'

# 通过代理访问
javdbapi video --id ZNdEbV --proxy-url http://127.0.0.1:7890 --output console | jq '.metadata.sources'

# 排行榜输出
javdbapi ranking --period weekly --type censored --stale-after 0s --output console | jq -c '.detail.summary.code'
```

### 输出说明

- `console` / `both` 模式下，stdout 只输出 JSON 行
- `help` 和 `version` 输出文本
- stderr 是日志，不是 JSON，不要喂给 parser
- exit code `0` 且 stdout 为空 = 命中新鲜缓存
- exit code `1` 时 stdout 可能已输出部分有效 JSON

## 缓存机制

- 缓存文件以番号命名（如 `DLDSS-271.json`），番号不存在时回退到内部 ID
- 使用原子写入（先写临时文件再 rename），中途崩溃不会损坏缓存
- `Detail` 和 `Reviews` 分别缓存，互不影响
- 评论拉取失败时，`Detail` 数据依然正常保存，失败信息记录在 `partial_errors` 中
- 缓存 schema 版本为 2，旧版本缓存直接视为过期

## 错误处理

### 哨兵错误

```go
if errors.Is(err, javdbapi.ErrNotFound) {
    // 页面不存在
}
if errors.Is(err, javdbapi.ErrRateLimited) {
    // 被限流了
}
```

可用的哨兵错误：
- `ErrInvalidConfig` — 配置无效
- `ErrInvalidQuery` — 查询参数无效
- `ErrNotFound` — 页面不存在（HTTP 404）
- `ErrRateLimited` — 被限流（HTTP 429）
- `ErrEmptyResult` — 结果为空
- `ErrParse` — 解析失败

### 错误类型

- `OpError` — 包装操作名和 URL，支持 `errors.As` 提取
- `HTTPError` — 原始 HTTP 状态码，支持 `errors.Is` 匹配哨兵错误

## 测试

```bash
go test ./...              # 离线单元测试
make race                  # 竞态检测
make integration           # 真实站点集成测试（需 JAVDB_INTEGRATION=1）
make check                 # 全量检查（lint + test + race）
```

所有默认测试使用本地 fixture 和 mock HTTP 服务器，不需要外网。

## 架构

```
cmd/javdbapi/          -- CLI 命令入口
├── app.go             -- 命令组装
├── command_*.go       -- 各子命令实现
└── flags.go           -- 共享参数定义

internal/
├── cliapp/            -- CLI 业务逻辑（分页、去重、并发）
├── clioutput/         -- 输出持久化（JSON 缓存、原子写入）
├── fetch/             -- HTTP 客户端（限流、重试、退避）
├── scrape/            -- HTML 解析（goquery 选择器）
└── siteurl/           -- URL 构建（安全校验、防注入）

javdbapi/              -- SDK 公共 API
├── client.go          -- Client 结构体
├── config.go          -- 配置和默认值
├── models.go          -- 数据模型
├── queries.go         -- 查询参数和枚举
├── list.go            -- 列表类 API
├── detail.go          -- 详情 API
├── reviews.go         -- 评论 API
├── ids.go             -- ID 类型和解析
├── page.go            -- 泛型分页容器
└── errors.go          -- 错误类型
```

## 依赖

| 依赖 | 用途 |
|------|------|
| `github.com/PuerkitoBio/goquery` | HTML 解析（jQuery 风格） |
| `github.com/urfave/cli/v3` | CLI 框架 |
| `golang.org/x/time` | 令牌桶限流器 |
| `github.com/stretchr/testify` | 测试断言 |
