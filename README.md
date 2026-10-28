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
| ⏰ **定时任务** | 签到（09/21 点，末尾自动跑**连登管家**：兑换已解锁档位 + 抽完抽奖次数）+ 活跃上报（10 点，点亮连登 / 解锁领养 + streak 自检）+ 猫猫旅行（09/21 点，独立排程）+ token 保活（22 点）+ **成长任务队列**（11 点，自动跑一遍任务中心待办），各类独立开关 |
| ⚡ **流式 + 非流式** | 出站强制 `stream:true`；SSE 帧按规范白名单重建；非流式由本地聚合为单响应 |
| 🧠 **推理模型兼容** | DeepSeek 思维链注入（`thinking.type=enabled` + 默认档）、`reasoning_content` 多轮回填、effort 档位自动对齐到模型支持的**最近档**（同距优先模型默认档，行里显示生效档位）；网关侧 `thinking.*` 档位策略（客户端优先 / 网关优先 + 作用域「只档位 / 档位与开关」，含按模型覆盖、`off` 尊重与保存后热生效） |
| 💬 **系统提示词体系** | 网关自有提示词替换客户端 system（默认 `custom`），从源头消灭 system 来源的内容误报；`passthrough` 遇拦截自动降级重试 |
| 🗑️ **指纹脱敏** | 出站请求体黑名单指纹字段清洗（可关闭），与提示词体系两层叠加 |
| 📊 **可观测** | 每请求一行表格日志（TTFB / token 速率 / uid）；`/healthz` 带 `service` 身份标识可接负载均衡 / 宿主探活 |
| 💾 **状态持久化** | 池状态本地原子落盘 + Upstash Redis 异步镜像（可选），重启择新恢复 |
| 🖥️ **Web 管理面板** | 内嵌单页面板（明暗主题），账号运维 / 模型档位查询 / 在线改配置（热生效）/ 运行日志 / 积分任务，见 [Web 管理面板](#-web-管理面板) |
| 🧩 **工具调用合并** | 出站 payload 把背靠背的 `assistant.tool_calls` 合并为一条（部分 agent 客户端回放并行调用的报文形态），修复上游 deepseek 系 `11148` |
| 🏷️ **本地备注名** | 账号池点击昵称就地编辑；备注存 `state.json`（`alias`），与上游昵称独立、昵称同步不覆盖；账号池 / 用量 / 到期提醒 / 任务中心 / 套餐统一显示 |
| 🔄 **版本与更新** | 版本号用上海时间标签；更新检查以本仓库 Release 为准；仓库暂无 Release 时显示提示而不是「更新失败」；容器部署更新镜像 |

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

任务计分走 `/v2/report` 行为上报，但**不同任务认不同客户端指纹**：CLI 指纹（`www.codebuddy.cn`）、桌面指纹（`copilot.tencent.com` + `WorkBuddy/5.5.6` UA + `workbuddy-desktop` 事件族）、web 指纹（`www.workbuddy.cn` + `x-client-platform: web`）。网关为每类任务构造对应指纹的判据事件链（`internal/upstream/desktop.go`）；专家类任务额外要求真实专家 id 与真实对话回执（`internal/upstream/streak.go` 之外的 expert 序列）。上报 200 ≠ 计分——面板在执行后轮询任务进度，达标即自动调用 Web 域领奖接口。

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

## 🆚 与上游的差异

本分支相对 [当前上游](https://github.com/misakano7545/workbuddy2api-panel) 的增量（已逐项核对；上游已有的能力不再重复列出）：

| 能力 | 说明 |
|---|---|
| **工具调用合并** | `mergeAdjacentToolCalls`：出站把背靠背的 `assistant.tool_calls` 合并为一条，修上游 deepseek 系 `11148`（tool_call_sequence_broken） |
| **服务端备注名** | `state.json` 的 `alias` + `POST /panel/api/accounts/{uid}/alias`；账号池就地编辑，所有账号视图统一显示；替代旧的 localStorage 备注（跨设备保留、昵称同步不覆盖） |
| **更新检查指向本仓库** | `appRepository` 与 updater 兜底仓库改为 `ulquiola/workbuddy2api-panel`；仓库暂无 Release 时按提示展示 |

> 上游已有的 Web 管理面板、请求指标、浏览器 OAuth、在线配置、积分任务、模型繁忙度、联网检索、成长任务队列、余额趋势、补领清扫、保留积分、积分历史、模型能力透出、安全加固、原生工具标记还原、零输出守卫、整体不可用模型提升、嵌入端点、补签卡上限等能力，均直接采用当前上游实现，不在上表重复。
