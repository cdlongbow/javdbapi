# javdbapi

javdbapi 是一个用于查询 JavDB 网站的 Go SDK 和命令行工具。提供类型安全的 API、自动限流重试、智能缓存、演员名解析，以及便捷的 CLI 命令。

## 核心特性

- **类型安全**：每种资源 ID（视频、演员、片商等）都是独立的强类型，防止混用
- **演员名解析**：支持中/日文演员名（简繁、异体字）自动匹配到对应演员 ID
- **自动限流**：内置令牌桶限流器，默认每秒 1 个请求，防止被封
- **智能重试**：指数退避 + 抖动，自动处理 429/502/503/504
- **独立缓存**：Detail 和 Reviews 分别缓存，互不影响
- **演员名归档**：Actor 命令自动按演员名创建子目录
- **番号解析**：Video 命令支持直接输入番号，自动查找对应 ID

## 快速开始

```bash
go install github.com/RPbro/javdbapi/cmd/javdbapi@latest

# 查看视频详情
javdbapi video --id ZNdEbV --output console

# 搜索视频
javdbapi search --keyword "VR" --output console

# 查看演员作品（自动按演员名创建子目录）
javdbapi actor --id WE4e --output both
```

## 使用 SDK

### 初始化 Client

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

所有配置项都有合理的默认值：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `BaseURL` | `https://javdb573.com` | 站点地址 |
| `HTTP.Timeout` | 30s | 请求超时 |
| `RateLimit.RequestsPerSecond` | 1 | 每秒请求数 |
| `Retry.MaxAttempts` | 3 | 最大重试次数 |

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

### 演员查找

通过演员名（支持简体/繁体/日文异体字）自动解析演员 ID：

```go
// 内部流程：简繁转换 → 异体字替换 → f=actor 搜索 → 名称匹配
id, err := client.ActorByName(ctx, "筱田优")
// id = "WE4e"

// 支持简体、繁体、日文异体字
client.ActorByName(ctx, "樱空桃")    // 简体 → 櫻空桃 ✓
client.ActorByName(ctx, "藤森里穗")  // 穗→穂 ✓
client.ActorByName(ctx, "深田咏美")  // 咏→詠 ✓
```

名称归一化规则（`ids.go`）：

1. 简繁转换（`s2tMap`）：樱→櫻、咏→詠、桥→橋、泽→澤 等
2. 异体字修正（`aliasMap`）：筱→篠、穗→穂、理→裏、户→戸
3. 搜索页 `f=actor` 标签直接搜演员，1 次请求即可拿到演员 ID

### 演员搜索

底层使用 `f=actor` 搜索标签页，返回演员列表：

```go
results, _ := client.SearchActors(ctx, "樱空桃", 1)
for _, item := range results.Items {
    fmt.Printf("%s: %s (%v)\n", item.ID, item.Name, item.Aliases)
}
// bvWB: 櫻空桃 [Sakura Momo, 桜空もも]
```

每个结果包含演员 ID、显示名、以及别名列表（来自链接的 title 属性）。

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

// 番号自动解析（使用 Client 方法，尊重 proxy/rate limit 等配置）
client, _ := javdbapi.NewClient(javdbapi.ClientConfig{})
resolvedID, _ := client.ResolveVideoID(ctx, "DLDSS-271")
```

## CLI 命令行工具

### 命令一览

| 命令 | 说明 | 必填参数 |
|------|------|----------|
| `video` | 查看视频详情（支持番号自动解析） | `--id` |
| `search` | 搜索视频 | `--keyword` |
| `actor` | 查看演员作品 | `--id` |
| `actor-detail` | 查看演员姓名和别名 | `--id` |
| `maker` | 查看片商作品 | `--id` |
| `home` | 浏览首页列表 | 无 |
| `ranking` | 查看排行榜 | `--period`, `--type` |

### 快速示例

```bash
# 视频详情（番号自动解析）
javdbapi video --id DLDSS-271 --output console

# 搜索 + 翻页
javdbapi search --keyword "VR" --page 1 --max-pages 3 --output both

# 演员作品 + 中字过滤
javdbapi actor --id neRNX --filter cnsub,download

# 排行榜
javdbapi ranking --period weekly --type censored --output console

# 通过代理
javdbapi video --id ZNdEbV --proxy-url http://127.0.0.1:7890 --output console
```

### 共享参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--output` | `file` | 输出模式：`file` / `console` / `both` |
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
cmd/javdbapi/              -- CLI 命令入口
├── app.go                 -- 命令注册、fetcher 构建
├── command_home.go        -- 首页浏览
├── command_search.go      -- 搜索
├── command_video.go       -- 视频详情 + 番号解析
├── command_actor.go       -- 演员作品列表
├── command_actordetail.go -- 演员信息（姓名、别名）
├── command_maker.go       -- 片商作品列表
├── command_ranking.go     -- 排行榜
├── flags.go               -- 共享参数定义
├── flagspec.go            -- 枚举参数解析
└── main.go                -- 入口

internal/
├── cliapp/                -- CLI 业务逻辑
│   ├── run.go             -- 分页、去重、并发 worker pool
│   ├── run_actordetail.go -- 演员详情抓取
│   ├── fetcher.go         -- Fetcher 接口（SDK 适配器）
│   └── types.go           -- 共享类型定义
├── clioutput/             -- 输出持久化
│   ├── model.go           -- Document schema v2
│   ├── store.go           -- 原子写入缓存 + 新鲜度判断
│   └── path.go            -- 文件路径生成 + 防注入
├── fetch/                 -- HTTP 客户端
│   ├── client.go          -- 限流、重试、响应大小上限
│   └── retry.go           -- 指数退避 + Retry-After 头
├── scrape/                -- HTML 解析
│   ├── actor_search.go    -- 演员搜索页解析 (f=actor)
│   ├── list.go            -- 列表/搜索页解析
│   ├── detail.go          -- 视频详情页解析
│   ├── reviews.go         -- 评论页解析
│   ├── score.go           -- 评分文本解析
│   ├── size.go            -- 磁力大小/文件数解析
│   ├── selectors.go       -- CSS 选择器 + 中文标签常量
│   ├── href.go            -- URL 路径 ID 提取
│   ├── text.go            -- 文本工具函数
│   ├── types.go           -- 内部数据类型
│   └── result.go          -- Result[T] 包装器
└── siteurl/               -- URL 构建
    └── build.go           -- 安全校验 + 防路径注入

javdbapi/                  -- SDK 公共 API
├── client.go              -- Client 结构体、fetchList 共享流程
├── config.go              -- 配置和默认值
├── models.go              -- 数据模型 (VideoSummary, VideoDetail, Actor, Magnet...)
├── queries.go             -- 查询参数和枚举
├── list.go                -- 列表类 API (Home, Search, SearchActors, ActorByName...)
├── detail.go              -- 视频详情 API
├── reviews.go             -- 评论 API
├── ids.go                 -- 强类型 ID + 演员名归一化 (NormalizeName)
├── page.go                -- 泛型分页容器
└── errors.go              -- 错误类型
```

### 数据流

```
CLI 命令
  → Fetcher 接口 (cliapp.Fetcher)
    → SDK Client (javdbapi.Client)
      → HTTP Client (internal/fetch) — 限流、重试
        → 站点 HTML
      → Parser (internal/scrape) — goquery 解析
    → 数据模型 (强类型 ID)
  → Store (internal/clioutput) — 原子写入缓存
  → stdout (JSON)
```

### 演员名解析流程

```
输入: "筱田优"
  → NormalizeName: 简繁转换 (无变化) → 异体字修正 (筱→篠) → "篠田優"
    → SearchActors("篠田優"): 请求 /search?f=actor&q=篠田優
      → 解析 HTML 中所有 <a href="/actors/{id}">
      → 返回 [{ID: "WE4e", Name: "篠田優", Aliases: ["篠田優(シノダユウ)", "篠田ゆう"]}]
    → 三级匹配:
      1. 精确匹配: Name=="篠田優" → 命中, 返回 WE4e
      2. 包含匹配 (备选)
      3. 字符重叠匹配 (备选)
```

## 依赖

| 依赖 | 用途 |
|------|------|
| `github.com/PuerkitoBio/goquery` | HTML 解析（jQuery 风格） |
| `github.com/urfave/cli/v3` | CLI 框架 |
| `golang.org/x/time` | 令牌桶限流器 |
| `github.com/stretchr/testify` | 测试断言 |