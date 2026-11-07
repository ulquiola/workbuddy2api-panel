<p align="center">
  <img src="https://raw.githubusercontent.com/DGZSbot/ai-icon/refs/heads/main/WorkBuddy.png" alt="WorkBuddy2API" width="120">
</p>

<h1 align="center">WorkBuddy2API Panel</h1>

<p align="center">
  <b>把腾讯 CodeBuddy 账号变成 OpenAI 兼容 API 的多账号网关 · 附 Web 管理面板</b><br>
  Web 面板 · OAuth 浏览器登录 · 账号池轮转 · 熔断与冷却 · 会话粘性 · **联网检索** · 定时签到 / 活跃 / 旅行 / 保活 · <b>成长任务一键完成（17/18）</b> · 流式 / 非流式
</p>

<p align="center">
  <img alt="Go" src="https://img.shields.io/badge/Go-1.26.10-00ADD8?logo=go&logoColor=white&style=flat-square">
  <img alt="API" src="https://img.shields.io/badge/API-OpenAI%20%7C%20Anthropic-412991?style=flat-square">
  <img alt="Deploy" src="https://img.shields.io/badge/Deploy-Single_Binary%20%7C%20Docker-2496ED?style=flat-square">
  <img alt="License" src="https://img.shields.io/badge/License-MIT-green?style=flat-square">
</p>

---

> **本项目源自 [Sliverkiss/workbuddy2api](https://github.com/Sliverkiss/workbuddy2api)，当前以 [misakano7545/workbuddy2api-panel](https://github.com/misakano7545/workbuddy2api-panel) 为基线**（fork）。
> 旧分支 [linguo2625469/workbuddy2api-panel](https://github.com/linguo2625469/workbuddy2api-panel) 已于 2026-10-21 停止维护，其历史保存在本仓 `archive/local-fork` 分支；当前开发主线为 `dev`。
> 差异概览见 [与上游的差异](#-与上游的差异)；上游设计的精巧之处（账号池调度、错误分类、提示词体系）原样保留，详见下文与上游 README。
>
> 感谢 [linguo2625469/workbuddy2api-panel](https://github.com/linguo2625469/workbuddy2api-panel/)、[misakano7545/workbuddy2api-panel](https://github.com/misakano7545/workbuddy2api-panel)，以及原仓库删库后的延续 [HanawaBanana/workbuddy2api](https://github.com/HanawaBanana/workbuddy2api)。

> [!IMPORTANT]
> **本仓的源头 [Sliverkiss/workbuddy2api](https://github.com/Sliverkiss/workbuddy2api) 早已删库停更**：原仓库连同代码、issue 与提交历史整体消失，不再可访问。本仓链路为 `Sliverkiss/workbuddy2api` → `linguo2625469/workbuddy2api-panel` → `misakano7545/workbuddy2api-panel` → 本仓，**中间每一环都没有官方支持与后续保障**。
>
> **旧分支 [linguo2625469/workbuddy2api-panel](https://github.com/linguo2625469/workbuddy2api-panel) 已于 2026-10-21 宣布停止维护**：不再更新代码、不跟进上游变化、不再处理新的 issue / PR。本仓已切换到 [misakano7545/workbuddy2api-panel](https://github.com/misakano7545/workbuddy2api-panel) 的新基线上继续演进（旧历史保留在 `archive/local-fork`）。
>
> **更新与发布**：版本号采用上海时间标签（`YYYY.M.D.HHMM`）；更新检查与 Release 均以**本仓库**（`ulquiola/workbuddy2api-panel`）为准，不再跟随上游。容器部署请更新镜像。
>
> 本仓是自用增强分支，**仍在按自己的节奏维护**——已合入的修复不回退，问题与 PR 在本仓处理。
>
> **AutoBot CI（开发中）**：本仓正在开发 `.github/workflows/autobot.yml` + 根目录 `AutoBot.yml`——每天定时 clone 一批同类实现（含本仓上游与其他参考仓库），由 agent 读代码写差异报告，再按报告写 Go 实现并开 PR。**该流水线目前处于开发阶段**：分片策略、报告质量与自动 PR 的筛选规则都还在调整，可能误报、漏报、开出需要人工打回的 PR，请勿把它的产出当作既成能力依赖；稳定性与信任度以后续说明为准。
>
> 风险自评：**上游随时可能彻底收紧导致通道不可用**；因使用非官方客户端导致的账号风控 / 封禁，由使用者自行承担。

> [!WARNING]
> 请勿将本项目用于对外中转、批量小号、售卖额度等滥用场景；由此产生的账号风控与封禁由使用者自行承担。

## 项目简介

WorkBuddy2API 是一个自托管的 **OpenAI 兼容反向代理网关**，将腾讯 CodeBuddy（`copilot.tencent.com`）账号包装为统一服务。入站支持 `chat_completions`（`POST /v1/chat/completions`）、`codex_responses`（`POST /v1/responses`）与 `anthropic_messages`（`POST /v1/messages`，同样接受 `POST /messages`）。

- 官方不提供 OpenAI 形态的开放 API，本项目通过 **OAuth 设备授权**（面板「添加账号」或 `login.sh`）获取账号凭证，在网关侧做 token 自动刷新、账号池调度与流量治理；
- 面向 **个人多账号** 场景：多账号共享、单号故障自动换号、冷却 / 熔断防止雪崩、会话粘性保证多轮上下文不跳号；
- 对客户端暴露 `chat_completions`、`codex_responses` 与 `anthropic_messages`。模型名原样交给上游，不把 `claude-*` 改写成别的模型。

> ⚠️ 合规须知：本项目是**非官方**网关，使用 CodeBuddy 账号作为上游，**仅限本人授权账号、本机 / 私有环境测试**。详细边界见[安全与合规](#安全与合规)。

## 核心能力

| 能力 | 说明 |
|---|---|
| 🔑 **OAuth 一键登录** | `login.sh` / 面板设备授权流程，三站点可选（CN / 官方国际站 / 旧桌面端），自动落盘凭证并加载新账号 |
| 🔄 **多账号池** | 快过期积分加权 + 成本分层 + 加权随机选号，Top-5 候选 + 防惊群 |
| 🛡️ **熔断与冷却** | 429 软冷却 600s 起指数退避（封顶 `soft_rate_max`）、404 固定 60s 短冷却、402 硬冷却至次日 04:00、连续失败熔断、在途租约限流 |
| 🧲 **会话粘性** | 同一会话（`conversation_id`）尽量绑定同一账号，TTL 滚动续期，失败自动解绑，可镜像 Redis 防重启丢失 |
| ⏰ **定时任务** | 签到（09/21 点，末尾自动跑**连登管家**：兑换已解锁档位 + 抽完抽奖次数）+ 活跃上报（10 点，点亮连登 / 解锁领养 + streak 自检；国际号另跑**网页会话腿**，见下）+ 猫猫旅行（09/21 点，独立排程）+ token 保活（22 点）+ **成长任务队列**（11 点，自动跑一遍任务中心待办），各类独立开关 |
| ⚡ **流式 + 非流式** | 出站强制 `stream:true`；SSE 帧按规范白名单重建；非流式由本地聚合为单响应 |
| 🧠 **推理模型兼容** | DeepSeek 思维链注入（`thinking.type=enabled` + 默认档）、`reasoning_content` 多轮回填、effort 档位自动对齐到模型支持的**最近档**（同距优先模型默认档，行里显示生效档位）；网关侧 `thinking.*` 档位策略（客户端优先 / 网关优先 + 作用域「只档位 / 档位与开关」，含按模型覆盖、`off` 尊重与保存后热生效） |
| 💬 **系统提示词体系** | 网关自有提示词替换客户端 system（默认 `custom`），从源头消灭 system 来源的内容误报；`passthrough` 遇拦截自动降级重试 |
| 🗑️ **指纹脱敏** | 出站请求体黑名单指纹字段清洗（可关闭），与提示词体系两层叠加 |
| 📊 **可观测** | 每请求一行表格日志（TTFB / token 速率 / uid）；`/healthz` 带 `service` 身份标识可接负载均衡 / 宿主探活 |
| 💾 **状态持久化** | 池状态本地原子落盘 + Upstash Redis 异步镜像（可选），重启择新恢复 |
| 🎬 **媒体生成 + MCP** | 生图 / 改图 / 出视频（异步任务，可阻塞等结果）与 `POST /mcp`（MCP 工具：生成图片、改图、生成视频、查任务、列媒体模型）；记账、换号、错误策略与对话走同一条路径。另有**可选**的「裁掉右下角水印」后处理（默认关闭） |
| 🖥️ **Web 管理面板** | 内嵌单页面板（明暗主题），账号运维 / 模型档位查询 / 在线改配置（热生效）/ 运行日志 / 积分任务，见 [Web 管理面板](#-web-管理面板) |
| 🧩 **工具调用合并** | 出站 payload 把背靠背的 `assistant.tool_calls` 合并为一条（部分 agent 客户端回放并行调用的报文形态），修复上游 deepseek 系 `11148` |
| 🏷️ **本地备注名** | 账号池点击昵称就地编辑；备注存 `state.json`（`alias`），与上游昵称独立、昵称同步不覆盖；账号池 / 用量 / 到期提醒 / 任务中心 / 套餐统一显示 |
| 🔄 **版本与更新** | 版本号用上海时间标签；更新检查以本仓库 Release 为准；仓库暂无 Release 时显示提示而不是「更新失败」；容器部署更新镜像 |

## 🎬 媒体生成：生图 / 改图 / 出视频 / MCP

网关除了对话，还把 WorkBuddy 的**媒体模型**包成 OpenAI 形态的 HTTP 接口，并额外给出一个 **MCP（Model Context Protocol）** 端点——Claude Code / Codex / Cursor / DeepSeekRuntime（DSH）这类 agent 配一行就能拿到「出图 / 出视频」工具，不必自己实现一遍鉴权、轮询和记账。

### 接口一览

| 方法 | 路径 | 用途 |
|---|---|---|
| POST | `/v1/images/generations` | 文生图 |
| POST | `/v1/images/edits` | 改图 / 图生图；`image` 可传公网 URL、`data:` URL 或 multipart 文件 |
| POST | `/v1/videos` | 提交视频任务（OpenAI Videos 资源的路径，等价于下一行）。缺省 `wait=true`：阻塞到出片后回 200；`wait=false`：秒回 `task_id` |
| POST | `/v1/videos/generations` | 同上 |
| GET | `/v1/videos/{task_id}` | 查任务（`queued` / `in_progress` / `completed` / `failed`） |
| GET | `/v1/videos/{task_id}/content` | 取成品字节（流式转发签名直链，不落盘；任务没出片回 409） |
| POST | `/v1/videos/tasks` | 查任务，body `{"task_id":"..."}`（给不方便用路径参数的客户端） |
| POST | `/v1/images/strip-watermark` | 裁掉底部纯白留白（顺带去水印），默认关闭；见「去水印」 |
| POST | `/mcp` | MCP（JSON-RPC 2.0，Streamable HTTP 形态），见下 |

视频回执按 OpenAI Videos 的 Video 资源出（`object:"video"`、`status` 四态、`seconds` 是字符串、`progress` 只有 0/100、`error` 是 `{code,message}` 对象），另带 `data[].url` / `videos[].url` / `usage.credit` 这几个自解释字段。OpenAI 的 Videos API 本身已于 2026-09-24 关停，这里对齐的是它还留在各家 SDK 里的形状。图片侧 `response_format:"b64_json"` 由网关取直链后转 base64（上游无视该参数）；`n>1` 直接回 400 —— 上游一次只出一张，宁可拒绝也不静默少给。

模型名照旧带域前缀：`cn:hunyuan-image-alpha`、`cn:hunyuan-image-alpha-edit`、`cn:minimax-video-h3`、`cn:kling-video-v3`。`/v1/models` 会把媒体模型单列并带 `image_generation` / `video_generation` 标记（`modalities.output` 分别是 `["image"]` / `["video"]`）；没配 `model` 时生图用 `cn:hunyuan-image-alpha`、生视频用 `cn:minimax-video-h3`。

### 生图 / 改图

```bash
curl -s http://127.0.0.1:7863/v1/images/generations \
  -H "Authorization: Bearer $WB2API_KEY" -H 'Content-Type: application/json' \
  -d '{"prompt":"一只戴礼帽的柴犬，像素风","size":"1024x1024"}'

# 改图：image 支持公网 URL / data: URL / multipart 文件
curl -s http://127.0.0.1:7863/v1/images/edits \
  -H "Authorization: Bearer $WB2API_KEY" -H 'Content-Type: application/json' \
  -d '{"model":"cn:hunyuan-image-alpha-edit","prompt":"把背景换成星空","image":"https://…/in.png"}'
```

回执**原样透传**上游（`{created, data:[{url}], usage:{credit, …}}`），网关不转存字节——图是签名直链，约 24h 有效，要长期保存请自行落盘。实测 5.71 积分/张（`cn:hunyuan-image-alpha`）。

> 改图的输入图必须是上游**取得到**的地址。拿「刚生成的那张图的签名直链」直接回投给改图接口，实测上游会回 `invoke model error`（它取不到/不认那个 URL）；用 `data:` URL 或自己可控的公网地址最稳。

### 出视频（异步任务）

```bash
# 提交（不等待）
curl -s http://127.0.0.1:7863/v1/videos/generations \
  -H "Authorization: Bearer $WB2API_KEY" -H 'Content-Type: application/json' \
  -d '{"prompt":"一杯热咖啡，热气缓缓上升","seconds":4,"wait":false}'
# → {"id":"1312831544-WandVideo-…","object":"video","task_id":"…","status":"queued", …}

# 查（阻塞到出片；不用 wait 参数时 POST 就是阻塞语义）
curl -s http://127.0.0.1:7863/v1/videos/1312831544-WandVideo-… \
  -H "Authorization: Bearer $WB2API_KEY"
```

请求参数：`prompt`（必填）、`model`、`seconds`（实测 4-15）、`resolution`、`aspect_ratio`（`16:9` / `9:16` / `1:1`）、`negative_prompt`、`enable_audio`（默认 `true`）、`watermark`（默认 `true`）、`image`（图生视频，公网/签名 URL）、`last_image`（尾帧）、`wait`（缺省 `true`）、`timeout_seconds`（缺省 900，封顶 1800）。除 `wait` / `timeout_seconds` 外都原样交给上游。

回执（出片后）：

```json
{"id":"…","object":"video","task_id":"…","status":"completed","model":"cn:minimax-video-h3",
 "created_at":1793801279,"completed_at":1793801403,"size":"768P","seconds":4,
 "data":[{"url":"https://aigc-output-video-….cos.ap-guangzhou.myqcloud.com/….mp4?q-signature=…"}],
 "videos":[{"url":"…"}],
 "usage":{"credit":57.14,"output_video_tokens":200000,"output_video_duration":4,"output_video_coefficient":1},
 "hint":"视频地址是签名直链（约 12 小时有效）；需要长期保存请自行落盘。"}
```

实测口径（2026-11-04，CN 池）：

| 模型 | 请求 | 实际扣费 | 出片耗时 | 回执 size |
|---|---|---|---|---|
| `cn:minimax-video-h3` | 4s / 16:9 / 有声 | 57.14 | ≈124s | `768P` |
| `cn:minimax-video-h3` | 4s 图生视频 | 57.14 | ≈124s | `768P` |
| `cn:kling-video-v3` | 4s（实际出 5s） | 102.86 | ≈473s | 不回该字段 |

几个必须知道的坑：

- **上游不校验 `aspect_ratio` / `resolution`**：`4:3`、`999x999` 都能提交成功并正常出片。网关同样不假装校验（只把合法值写进文档），免得把上游未来支持的比例挡在外面。
- **上游 `usage` 只回一次**：第一次查到 `completed` 时回，之后重复查询没有 `usage`（kling 连 `size` 都不回）。网关在任务台账里缓存这次用量，重复查询照样能拿到 `credit` / `size` / `seconds`。
- **计费只记一次**：提交不扣钱，出片完成那一刻按 `usage.credit` 记一次（任务台账里的「已记账」标记 + 同步/异步两条路径共用同一把锁），重复查询、并发查询都不会重复记账。
- **任务是账号级可见**：上游任务只对提交它的账号可见，账号结构里没有数字租户 id 可以反推，所以网关自己维护「task → 账号」台账；**台账在进程内存里**，重启后查询会退化成逐个账号试、命中即收养（此时若上游不再回 `usage`，这次查询就不带积分明细，但也不会重复扣费）。
- **kling 慢一个量级**：MCP 默认 `wait=true`，客户端超时要给到 15 分钟以上；只想拿任务号就传 `wait=false`。

### MCP 端点（`POST /mcp`）

MCP 端点走 **Streamable HTTP** 的纯 POST 形态（不提供 SSE 流、不维护 session）：客户端发一条 JSON-RPC 就回一条 JSON-RPC。鉴权与 REST 相同（`Authorization: Bearer <api_key>`）；`GET /mcp` 回 405 + `Allow: POST`（客户端据此退回纯 POST）。

```bash
curl -s http://127.0.0.1:7863/mcp \
  -H "Authorization: Bearer $WB2API_KEY" -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"c","version":"0"}}}'
```

工具（5 个）：

| 工具 | 说明 |
|---|---|
| `list_generation_models` | 列可用的生图 / 出视频模型（读目录旁路缓存，不打上游），带计费与直链时效说明 |
| `generate_image` | 文生图，参数 `prompt`（必填）、`model`、`size`、`quality`、`n`；同步回 URL |
| `edit_image` | 改图，参数 `prompt`（必填）、`image`（URL 或 `data:` URL，必填）、`model` |
| `generate_video` | 出视频，参数 `prompt`（必填）、`model`、`seconds`、`resolution`、`aspect_ratio`、`negative_prompt`、`enable_audio`、`image`、`last_image`、`wait`（缺省 `true`） |
| `get_video_task` | 查任务 / 收片，参数 `task_id` |

客户端配置（Claude Code / DSH / 任何支持 HTTP MCP 的宿主）：

```json
{"mcpServers":{"wb2api":{"type":"http","url":"http://127.0.0.1:7863/mcp",
  "headers":{"Authorization":"Bearer <api_key>"}}}}
```

约定：**工具执行失败**（上游 4xx/5xx、任务 failed）回 `result.isError=true` + 上游原话进 `content[0].text`，让模型自己读到「时长只支持 4-15 秒」这类可行动信息；**参数不合法**（缺 `prompt` / 缺 `task_id` / 未知工具）才走 JSON-RPC `-32602`。工具调用内部直接复用网关自己的 HTTP handler（`httptest` 形态），所以换号、错误策略、记账、请求日志与 REST 完全同一条路径——不会出现「REST 记了账、MCP 漏了账」这种分叉。

### 去水印（可选，默认关闭）

成品右下角有「AI生成 / WORKBUDD>.」水印。它是**白色半透明**、锚在右下角（距右边 / 下边各约 10px）：1024 / 1536 边长档量出来固定 **111×51px**，1344×1792 档约 126×80px，**不随画布缩放**。

两条路，按需选：

1. **白底上它本来就不显眼**——纯白区像素均值 255.0 / 标准差 0.1，成品最终排在白底上时肉眼看不出来，可以不处理。
2. **要彻底去干净就裁**：让画面底部真有一条纯白留白，再把这条留白整条裁掉，水印必然一起消失（水印最高约 80px，留白 ≥ 90px 才安全）。

模型不总肯留白（实测：杯子 38.5%、猫 31.2%、深色木桌那张只有 1.0%），所以稳的打法是两步：

```bash
# 1) 生成时直接要留白（不给白边就补一刀）
curl -s http://127.0.0.1:7863/v1/images/generations   -H "Authorization: Bearer $KEY" -H "Content-Type: application/json"   -d '{"model":"cn:hunyuan-image-alpha","size":"1024x1536",
       "prompt":"……；画面最下方三分之一是完全纯白留白，不要有任何物体、阴影、反光、文字"}'

# 留白不够？用改图补一条白边（约 5.71 credit）
curl -s http://127.0.0.1:7863/v1/images/edits   -H "Authorization: Bearer $KEY"   -F image=@photo.png   -F 'prompt=把图片最下方约三分之一区域涂成纯白色（#FFFFFF）空白留白，不要改变上方画面'

# 2) 裁掉留白（本地解码裁剪，不扣积分）
curl -s http://127.0.0.1:7863/v1/images/strip-watermark   -H "Authorization: Bearer $KEY" -H "Content-Type: application/json"   -d '{"image":"https://….myqcloud.com/….png","min_blank_px":90}'
```

回执（`data_url` 是裁好的图，`return:"meta"` 时只回数字不返回图）：

```json
{
  "object": "image.strip",
  "width": 1024, "height": 1536, "out_width": 1024, "out_height": 944,
  "boundary": 944, "blank_px": 592, "blank_ratio": 0.385, "blank_lum": 254.4,
  "min_blank": 90, "white_lum": 248, "safe": true, "format": "png",
  "data_url": "data:image/png;base64,…",
  "usage": {"credit": 0, "hint": "本地后处理，不消耗上游积分"}
}
```

参数：`image`（必填：上游产物域 URL / `data:` URL / 裸 base64）、`min_blank_px`（默认 90）、`keep_blank`（裁完在内容下方留几像素白边，默认 0）、`format`（`png` / `jpeg`，省略跟随输入）、`quality`（JPEG 默认 92）、`return`（`data_url` 默认；`meta` 只回数字，适合先校验留白够不够）。

留白不够时回 **422 `no_blank_strip`**，`gateway_hint` 里带实测数字和下一步（先 `edit_image` 涂白）。MCP 侧同名工具 `strip_watermark`，回 MCP `image` content（base64）+ 文本元数据。

开关 `features.watermark_strip_enabled`（默认 `false`，面板保存后热生效）：关着时端点回 **403 `watermark_strip_disabled`**，MCP 的 `strip_watermark` 也不出现在 `tools/list` 里，其它生成接口完全不受影响。默认关是因为它要把成品图回读、解码、再 base64 回传（1024×1536 PNG 约 2.5MB → base64 3.4MB），只在真要彻底去水印时划算。

边界：只裁「整条纯白留白」，不做 inpainting；`min_blank_px` 调小会切到画面内容；裁切会改画幅，介意就 `keep_blank` 少裁一点。图片地址只允许上游产物域（`.myqcloud.com`），防止被当成 SSRF 跳板。

## 代理池与账号绑定

面板「代理池」支持一行一个批量添加，接受 `http`、`https`、`socks5`、`socks5h` URL，包括用户名密码。空行忽略，重复项跳过，无效行单独报告，其他有效行仍保存。例如：

```text
http://user:password@127.0.0.1:8080
https://127.0.0.1:8443
socks5://user:password@127.0.0.1:1080
socks5h://127.0.0.1:1081
```

账号池每行的代理菜单保存后立即生效：

| 模式 | 出站行为 |
|---|---|
| 默认随机 | 每个 HTTP 请求独立从可用节点中随机选择；该请求的重定向和响应流保持同一节点 |
| 直连 | 使用原有直连传输 |
| 固定代理 | 只用绑定节点；节点故障时保留绑定并返回代理错误 |

分区默认出口（面板「代理池 → 分区默认出口」）按 realm 指定：`cn` 与 `global` 各一条，账号行没单独指定代理时取它。查表顺序是**账号绑定 > realm 默认 > 随机池 > 直连**——账号行里点过「直连」的号不会被 realm 默认覆盖。

池为空时默认随机使用直连。池中已有节点但没有可用节点时返回错误。代理故障不处罚账号，不触发额外换号重试。聊天、图片、搜索、嵌入、token 刷新、余额和账号后台任务使用同一账号路由入口。未关联账号的 OAuth 授权与更新下载沿用原出站行为。

HTTP/HTTPS 使用标准代理传输；`socks5` 在本机解析目标域名，`socks5h` 由代理解析。各节点复用独立连接池，沿用现有 DNS、TLS、响应头超时和流式空闲控制。禁用或删除只关闭空闲连接，在途请求继续完成。

自动测活默认每 60 秒运行，单次超时 10 秒，最多 8 个节点并发。默认检测 `https://www.workbuddy.cn/`，最终响应为 2xx/3xx 即成功。新增节点与重启后的节点先检测，首次成功后进入候选；连续失败 3 次自动禁用，之后连续成功 2 次恢复。面板可修改 URL、间隔、超时与阈值，也可单独或全部检测。手动禁用优先，检测成功仍保持禁用。业务限流、余额不足等响应不影响测活状态。

节点、手动启停、账号绑定与测活设置原子保存到 `state_file` 同目录的 `proxy-pool.json`，健康状态仅保存在内存。代理列表、接口响应和错误日志不返回密码；编辑时密码留空保持原值，从 URL 删除用户名可移除认证。被账号或 realm 默认引用的节点禁止删除（HTTP 409），需先解除引用；从面板删除账号时同步清理绑定。

管理接口沿用面板鉴权：

| 方法 | 路径 | 用途 |
|---|---|---|
| GET / POST | `/panel/api/proxies` | 列表与批量添加（`{"lines":"..."}`） |
| PATCH / DELETE | `/panel/api/proxies/{id}` | 编辑（`name`、`url`、`password`、`enabled`）与删除 |
| POST | `/panel/api/proxies/{id}/check` | 排队检测单个节点 |
| POST | `/panel/api/proxies/check` | 排队检测所有节点 |
| PUT | `/panel/api/proxies/health` | 保存 `url`、`interval_seconds`、`timeout_seconds`、`failure_threshold`、`success_threshold` |
| PUT | `/panel/api/accounts/{uid}/proxy` | 保存 `{"mode":"random"}`、`{"mode":"direct"}` 或 `{"mode":"fixed","proxy_id":"..."}` |
| PUT | `/panel/api/proxies/realm/{realm}` | 同上格式，保存 `cn` / `global` 的分区默认出口 |

## 🎯 成长任务一键完成（17/18）

官方「成长计划」的 18 个成长任务中，**17 个可在面板上一键纯 API 完成**——无需安装官方客户端、无需人工交互，点一下「一键完成」即自动推进进度、等待异步计分落定并**自动领奖**。剩余任务展示操作指引。

### 任务覆盖与奖励

| 任务 | 奖励 | 一键完成方式 |
|---|---|---|
| `first_buddy` | +300c +8e | 解锁上报 → 同意协议 → 领养第一只 Buddy |
| `create_canvas` | +300c +5e | 设计画布创建事件组（Ardot 遥测） |
| `chat_5` | +100c | 对话活跃上报 ×5（自动补足差额） |
| `Model_chat_GLM5.2` | +100c +5e | glm-5.2 真实对话一次（发一条短消息） |
| `RichMeow_Chat` | +100c +5e +UR Buddy | 桌面端对话事件链（6 事件，含成功回执） |
| `Buddy_App` | +100c +5e | Buddy 应用「发现→进入→授权」事件链 |
| `Buddy_App_QQ` | +50c +5e | 企鹅教师助手进入事件链（与上一条共用） |
| `automation_1` | +100c +5e | 定时任务创建成功事件 |
| `Library_read` | +100c +5e | 资料库阅读点击（web 域上报） |
| `template_5` | +100c +5e | 模板使用事件组 ×5 |
| `playbook_prompt` | +100c +5e | 灵感案例「做同款」发送事件 |
| `expert_5` | +100c +5e | 真实专家召唤+使用链 ×5（专家市场拉真实专家 → 真实对话 → 使用事件） |
| `Expert_team_use_3` | +100c +5e | 专家团召唤+使用链 ×3 |
| `Hp_Appearance` | +100c +5e | 主题设置 + 皮肤生效事件 |
| `Expert_lighthouse` | +100c +5e | 轻量云专家召唤+使用链（真实对话 requestId，**可免费领一个月轻量服务器**） |
| `skill_1` | +100c +5e | 真实对话 + 技能加载事件（skill_info） |

**全新账号一键全做完 ≈ +1950 credits +78 能量**，其中仅数个任务涉及真实对话（`Model_chat_GLM5.2` 一条、`expert_5`/`Expert_team_use_3`/`skill_1` 各数条 fast-model 短对话），其余全部为行为事件上报，零对话消耗。

### 不可自动的 1 个

| 任务 | 原因 |
|---|---|
| `Expert_Philanthropy` | 需真实捐款（服务端领奖时校验捐赠回执，已实测无法绕过） |

### 实现原理（简述）

任务计分走 `/v2/report` 行为上报，但**不同任务认不同客户端指纹**：CLI 指纹（`www.codebuddy.cn`）、桌面指纹（`copilot.tencent.com` + `desktopUAFor` 给出的桌面 UA + `workbuddy-desktop` 事件族）、web 指纹（`www.workbuddy.cn` + Node/axios 形态的 `/v2/report`）。网关为每类任务构造对应指纹的判据事件链（`internal/upstream/desktop.go`）；专家类任务额外要求真实专家 id 与真实对话回执（`internal/upstream/streak.go` 之外的 expert 序列）。上报 200 ≠ 计分——面板在执行后轮询任务进度，达标即自动调用 Web 域领奖接口。

> 两个曾经写错、已按真机抓包纠正的点：桌面 UA 不是硬编码版本号，而是与业务链路同源的 `desktopUAFor`（版本由配置驱动，CN 5.9.3 / GL 5.9.4）；**`/v2/report` 路径**不带 `x-client-platform`、`Origin`、`Referer` 或任何 `sec-ch-*`——官方 report 是纯 Node 形态的 11 个头，往这里补 Client Hints 反而是破功。别把两处混为一谈：`/console/account`（`profile.go`）与任务领奖端点（`tasks.go`）**确实**带 `x-client-platform: web`，那是另一条链，删了必坏。

> ⚠️ 行为事件按天幂等：重复点「一键完成」不会重复扣资源，已达标的任务自动跳过。
> 补充（issue #174）：**白名单外**的任务只要进度已达标也会入队，以「只领奖」形态执行（不跑任何动作、不消耗对话配额）——避免任务达标但没人 claim、积分卡在 2000/2100/2300 上不去。

### 🧭 任务中心（面板新视图）

「任务中心」视图把散落的任务能力收拢成一处：

- **全账号任务扫描**：一键拉取每个账号未完成且可自动化的成长任务（含小程序口径），列表一目了然
- **执行队列**：把待办按账号排队执行——账号内串行（与单任务/一键完成共用互斥锁），账号间可选并发（1-3）；执行进度实时更新到每个条目。**同一队列也可按排程每日自动跑**（`schedule.growth_hours`，默认 11 点；排程轮次账号间固定串行）
- **日志分频道**：运行日志按「任务 / 对话 / 系统」三个频道筛选——对话流量再大，任务结果也不会被冲掉；日志条目带频道徽标与时间

任务中心扫描会合并小程序口径待办（`Sequential_Tasks_*`，默认列表不可见）。accept 带登记回读：上游有过 200 但没落账的情况，未生效会再试一次。

### 连登兑换与抽奖（自动）

成长中心连登档位（连续登录 7/14/28 天）兑换后发放积分 / 能量 / 补签卡 / **抽奖次数**，抽奖次数只能从兑换获得。网关把它挂在每日签到排程末尾自动跑闭环（见[定时任务](#定时任务)）：档位解锁当天自动兑换、有抽奖次数自动抽完，全程无需人工盯。

## 🔄 版本与更新

- 版本号使用**上海时间标签**（`YYYY.M.D.HHMM`），构建时通过 `-X main.appVersion=...` 注入；更新器只比较这种格式的 Release。
- 更新检查以**本仓库**的 GitHub Release 为准（`ulquiola/workbuddy2api-panel`），不再跟随上游。
- 仓库暂无时间标签 Release 时，面板显示「仓库暂无 Release；发布后会自动检查」，不会误报上游版本或「更新失败」。
- 容器部署请拉取新镜像；原生 Linux / Windows 二进制的「一键安装」由 Release 资产提供。

## 🪟 Windows 常驻运行（开机自启 + 崩溃重启）

二进制**不自带守护**：进程退出后不会自己回来，也没有内置服务（SCM）模式。把面板当本地网关用时（上游客户端写死 `127.0.0.1:<端口>`），进程不在等于该 provider 整体不可用——客户端立刻拿到 `connection refused`，不会重试。所以「开机自启 + 退出即重启」得由系统侧提供，下面两种任选一种。

先确认三件事：`wb2api.exe` 的绝对路径（工作目录就是放 `config.json` 的地方）；配置里 `listen` 默认只绑 `127.0.0.1`，不用额外放通防火墙；手动跑一次能起（首次会自动生成含随机 `api_key` 的 `config.json`），再交给常驻项。

### 方案 A：NSSM（推荐，真正的服务）

NSSM 按「进程退出」拉起，不看退出码——面板无论因何退出都会回来，日志也一起接管：

```powershell
nssm install wb2api-panel 'C:\wb2api\wb2api.exe'
nssm set wb2api-panel AppDirectory 'C:\wb2api'
nssm set wb2api-panel Start SERVICE_AUTO_START
nssm set wb2api-panel AppExit Default Restart
nssm set wb2api-panel AppStdout 'C:\wb2api\service.log'
nssm set wb2api-panel AppStderr 'C:\wb2api\service.log'
nssm start wb2api-panel
```

WinSW 同理（`<startmode>Automatic</startmode>` + `<onfailure action="restart"/>`），不另举例。

### 方案 B：任务计划程序（系统自带，不需要装东西）

管理员 PowerShell 执行，把路径换掉：

```powershell
$exe = 'C:\wb2api\wb2api.exe'
$act = New-ScheduledTaskAction -Execute $exe -WorkingDirectory (Split-Path $exe)
$trg = New-ScheduledTaskTrigger -AtStartup
$set = New-ScheduledTaskSettingsSet -RestartCount 999 -RestartInterval (New-TimeSpan -Minutes 1) `
       -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit ([TimeSpan]::Zero)
Register-ScheduledTask -TaskName 'wb2api-panel' -Action $act -Trigger $trg -Settings $set -RunLevel Highest
Start-ScheduledTask -TaskName 'wb2api-panel'
```

两个坑：
- `-ExecutionTimeLimit 0`：默认任务跑满 3 天会被结束掉，常驻进程必须关掉这个限制；
- 这里的「失败后重启」**只对非零退出码生效**——进程自己正常退出（码 0）或被人 `taskkill` 掉，任务不会重来。所以要做到「任何原因退出都拉起」，用方案 A。

### 更新时的注意事项

「一键更新」的设计是**先停旧进程 → 换二进制 → 起新进程**，中间有几秒端口空窗，这期间的请求会失败一次（客户端一般不重试）。两件事别忘：

1. 常驻项若在旧进程退出后立刻拉起，可能抢在 updater 起新进程之前占住端口——更新完确认只有一个进程在跑：`Get-Process wb2api` 与 `netstat -ano | findstr :7863` 的 PID 要对得上；
2. 更新前后各打一次 `curl http://127.0.0.1:7863/healthz`，确认新进程确实起来了（面板右下角也会显示版本号）。

> 内置服务模式（`wb2api.exe install-service` 走 SCM）没有做：更新重启要连带重写服务控制逻辑，且 CI 只有 Linux 构建机，Windows 服务生命周期没法做回归。用上面两种效果等价。

## 🆚 与上游的差异

本分支相对 [当前上游](https://github.com/misakano7545/workbuddy2api-panel) 的增量（已逐项核对；上游已有的能力不再重复列出）：

| 能力 | 说明 |
|---|---|
| **工具调用合并** | `mergeAdjacentToolCalls`：出站把背靠背的 `assistant.tool_calls` 合并为一条，修上游 deepseek 系 `11148`（tool_call_sequence_broken） |
| **服务端备注名** | `state.json` 的 `alias` + `POST /panel/api/accounts/{uid}/alias`；账号池就地编辑，所有账号视图统一显示；替代旧的 localStorage 备注（跨设备保留、昵称同步不覆盖） |
| **更新检查指向本仓库** | `appRepository` 与 updater 兜底仓库改为 `ulquiola/workbuddy2api-panel`；仓库暂无 Release 时按提示展示 |

> 上游已有的 Web 管理面板、请求指标、浏览器 OAuth、在线配置、积分任务、模型繁忙度、联网检索、成长任务队列、余额趋势、补领清扫、保留积分、积分历史、模型能力透出、安全加固、原生工具标记还原、零输出守卫、整体不可用模型提升、嵌入端点、补签卡上限等能力，均直接采用当前上游实现，不在上表重复。
