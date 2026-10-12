# MCP 多服务智能问答 Demo

一个基于 **MCP（Model Context Protocol）** 的智能体项目：客户端同时连接多个 MCP Server（本地 stdio + Streamable HTTP），把它们的工具统一转成 OpenAI Function Calling 格式交给大模型，由模型决定调用哪个工具，实现「用自然语言驱动 MCP 工具」的多轮问答。

---

## 目录

- [项目结构](#项目结构)
- [架构说明](#架构说明)
- [环境准备](#环境准备)
- [配置大模型](#配置大模型)
- [启动服务](#启动服务)
- [使用示例](#使用示例)
- [常见问题](#常见问题)
- [相关文档](#相关文档)

---

## 项目结构

```
tmcp/
├── mcp_client.py        # MCP 客户端 + Agent（主程序入口）
├── weather_mcp.py       # 天气 MCP Server（Streamable HTTP，端口 5001）
├── time_mcp.py          # 时间 MCP Server（Streamable HTTP，端口 5002）
├── bilibili_mcp.py      # B 站搜索 MCP Server（stdio，由客户端自动拉起）
├── requirements.txt     # 依赖清单
└── venv/                # 虚拟环境
```

### 四个脚本的角色

| 文件 | 角色 | 传输协议 | 是否需要手动启动 |
|------|------|----------|------------------|
| `mcp_client.py` | MCP 客户端 + LLM Agent | — | **是**（主程序） |
| `weather_mcp.py` | 天气 MCP Server | Streamable HTTP `:5001` | **是**（需先启动） |
| `time_mcp.py` | 时间 MCP Server | Streamable HTTP `:5002` | **是**（需先启动） |
| `bilibili_mcp.py` | B 站搜索 MCP Server | stdio | 否（客户端自动以子进程拉起） |

### 提供的工具

**`weather_mcp.py`**（模拟数据，非真实天气 API）

| 工具 | 说明 |
|------|------|
| `get_location_id(location_name)` | 根据城市名获取城市 ID |
| `get_weather(location_id, date)` | 根据城市 ID 和日期获取天气 |
| `get_current_utc_datetime()` | 获取当前 UTC 时间 |

**`time_mcp.py`**（基于系统时钟与 `zoneinfo`）

| 工具 | 说明 |
|------|------|
| `get_current_time(timezone)` | 获取指定时区当前时间，默认 `Asia/Shanghai` |
| `get_current_utc_datetime()` | 获取当前 UTC 时间 |
| `get_timestamp()` | 获取 Unix 时间戳（秒级 + 毫秒级） |
| `convert_timezone(datetime_str, from_timezone, to_timezone)` | 跨时区时间换算 |
| `list_common_timezones()` | 列出常用 IANA 时区名称 |

**`bilibili_mcp.py`**（调用 B 站公开搜索接口）

| 工具 | 说明 |
|------|------|
| `search_videos(keyword, page, page_size)` | 搜索 B 站视频 |
| `search_articles(keyword, page, page_size)` | 搜索 B 站文章 |

> 说明：`weather_mcp.py` 返回的天气、温度、湿度均由 `random` 随机生成，仅用于演示 MCP 调用链路。

---

## 架构说明

```
                        ┌──────────────────┐
                        │   mcp_client.py  │
                        │  MCPClient +     │
                        │  MCPAgent        │
                        └────────┬─────────┘
                                 │
        ┌────────────┬───────────┼───────────┬────────────┐
        │            │           │           │            │
   stdio 子进程   HTTP :5001  HTTP :5002        OpenAI 兼容 API
        │            │           │                        │
        ▼            ▼           ▼                        ▼
 bilibili_mcp.py  weather_mcp.py  time_mcp.py         大模型服务
 （B 站搜索）      （天气模拟）    （时间查询）      （决策调哪个工具）
```

运行闭环：

```
用户输入
  → Agent 收集所有 MCP 工具，转成 OpenAI tools 格式
  → 调用 LLM
  → LLM 返回 tool_calls（结构化调用意图）
  → Agent 路由到对应 MCP Server 执行
  → 结果以 role:"tool" 回填 history
  → 再次调用 LLM
  → 直到 LLM 不再请求工具 → 输出最终答案
```

---

## 环境准备

### 1. 创建并激活虚拟环境

项目已自带 `venv/`，如需重建：

```bash
cd /Users/ned/Desktop/tmcp
python3 -m venv venv
```

### 2. 安装依赖

```bash
venv/bin/pip install -r requirements.txt
```

依赖清单：

```
mcp[cli]<2
fastmcp
bilibili-api-python
uvicorn
fastapi
python-dotenv
openai
requests
```

---

## 配置大模型

`mcp_client.py` 通过 **OpenAI 兼容接口** 调用大模型，配置位于 `MCPAgent.__init__`：

```python
self.model = OpenAI(
    base_url='http://localhost:11434/v1',   # 本地 Ollama
    api_key='ollama',                        # Ollama 不校验，占位即可
)
```

以及模型名（在 `process_query` 中，共两处）：

```python
model='qwen3:14b'
```

### 方案 A：使用本地 Ollama（当前配置）

```bash
# 1. 启动 Ollama 服务
ollama serve

# 2. 拉取一个支持 Function Calling 的模型
ollama pull qwen3:14b

# 3. 确认本地已有的模型名
ollama list
```

`mcp_client.py` 里两处 `model=` 的值**必须与 `ollama list` 输出的名字完全一致**，否则会报 404。

> **注意**：`qwen3` 系列没有 `0.5b` 规格，最小为 `0.6b`（`0.5b` 属于 `qwen2.5` 系列）。写错规格会报 `model not found`。

### 方案 B：使用远程 API（如 ModelScope）

```python
self.model = OpenAI(
    base_url='https://api-inference.modelscope.cn/v1',
    api_key=os.getenv('MODELSCOPE_API_KEY'),
)
```

对应设置环境变量后再启动：

```bash
export MODELSCOPE_API_KEY='你的 token'
```

> **安全建议**：不要把 API Key 硬编码在源码中，改为从环境变量读取，避免提交到 git 时泄露。

### 模型要求

无论用哪种方案，模型**必须支持 Function Calling / Tool Use**，否则 Agent 无法调用 MCP 工具，只会返回纯文本。

---

## 启动服务

需要**三个终端**。

### 终端 1：启动天气 MCP Server

```bash
cd /Users/ned/Desktop/tmcp
venv/bin/python3 weather_mcp.py
```

成功标志：

```
天气MCP服务器已启动
INFO:     Started server process [xxxxx]
INFO:     Application startup complete.
INFO:     Uvicorn running on http://127.0.0.1:5001
```

### 终端 2：启动时间 MCP Server

```bash
cd /Users/ned/Desktop/tmcp
venv/bin/python3 time_mcp.py
```

成功标志：

```
时间MCP服务器已启动
INFO:     Application startup complete.
INFO:     Uvicorn running on http://127.0.0.1:5002
```

### 终端 3：启动客户端（主程序）

```bash
cd /Users/ned/Desktop/tmcp
venv/bin/python3 mcp_client.py
```

客户端会自动完成：

1. 以子进程方式拉起 `bilibili_mcp.py`（stdio，**无需手动启动**）
2. 连接 `http://localhost:5001/mcp`（天气 Server）
3. 连接 `http://localhost:5002/mcp`（时间 Server）
4. 打印各 Server 的工具列表与连接结果汇总
5. 进入 `User:` 交互提示符

### 启动顺序

```
① 终端 1  →  weather_mcp.py    监听 127.0.0.1:5001
② 终端 2  →  time_mcp.py       监听 127.0.0.1:5002
                    ↓
③ 终端 3  →  mcp_client.py     连接 5001 + 5002 + 拉起 bilibili 子进程
                    ↓
④ 输入问题开始对话，输入 quit 退出
```

> **顺序不能颠倒**：必须先启动两个 HTTP Server，否则客户端连接会失败。
>
> 单个 Server 连接失败**不会导致客户端崩溃**，只会打印失败原因并跳过；若全部失败则提示后退出。

---

## 使用示例

```
User: 帮我查一下北京明天的天气

Assitant: 北京明天天气晴，气温 23°C，湿度 45%。
```

```
User: B站搜一下 Python 教程

Assitant: 找到以下视频：
1. 【Python 全栈教程】...
2. ...
```

退出：

```
User: quit
```

---

## 常见问题

### 1. 端口被占用

```
ERROR: [Errno 48] error while attempting to bind on address ('127.0.0.1', 5001): address already in use
```

说明 5001 已被占用（通常是上一次启动的进程没退干净）：

```bash
# 查看占用进程
lsof -nP -iTCP:5001 -sTCP:LISTEN

# 结束该进程（破坏性操作，确认 PID 无误后执行）
kill <PID>
```

### 2. 客户端连接 5001 / 5002 失败

典型报错：

```
httpx.ConnectError: All connection attempts failed
```

原因是对应的 MCP Server 没有启动。先确认端口监听状态：

```bash
lsof -nP -iTCP:5001,5002 -sTCP:LISTEN
```

无输出说明 Server 未运行，回到「启动服务」章节先启动 `weather_mcp.py` 和 `time_mcp.py`。

> 该报错后面通常会跟随大量 `anyio` / `cancel scope` 堆栈，属于 MCP SDK 关闭传输时的已知噪音，**真正的原因是第一行的 `ConnectError`**。

### 3. 模型返回 403（远程 API）

```
Error: Error code: 403 - ... please make sure your associated Aliyun account is real-name verified
```

这是账号未实名认证导致，不是代码问题。前往对应平台的账号设置页完成实名认证即可。

在此之前，MCP 连接和工具列表仍能正常打印，只有提问那一步会失败。

### 4. 模型不存在 / 404

`base_url` 与 `model` 不匹配。检查：

- 用 Ollama → `model` 必须是 `ollama list` 中存在的名字
- 用远程 API → `model` 必须是该平台的合法 Model ID

### 5. 模型不调用工具，只回复纯文本

说明所用模型不支持 Function Calling，或工具描述不够清晰。更换支持 Tool Use 的模型。

### 6. B 站搜索失败

`bilibili_mcp.py` 调用的是 B 站公开接口，可能受网络、风控或接口变更影响。错误会以 `{"error": "搜索失败: ..."}` 形式返回。

---

## 相关文档

完整的学习文档位于 [`agent/`](./agent/) 目录，按学习顺序编号：

| 文档 | 内容 |
|------|------|
| [`agent/README.md`](./agent/README.md) | 学习路线索引与概念速查 |
| [`agent/01-agent-basics.md`](./agent/01-agent-basics.md) | Agent 内部如何工作、如何调用 LLM |
| [`agent/02-agent-tech-routes.md`](./agent/02-agent-tech-routes.md) | 五条主流技术路线与取舍依据 |
| [`agent/03-mcp-protocol.md`](./agent/03-mcp-protocol.md) | MCP 协议：角色、传输方式、生命周期 |
| [`agent/04-llm-and-token.md`](./agent/04-llm-and-token.md) | Ollama 本地搭建、Token 与上下文机制 |
| [`agent/05-context-optimization.md`](./agent/05-context-optimization.md) | 上下文预算管理与优化实践 |
| [`agent/06-mcpagent-implementation.md`](./agent/06-mcpagent-implementation.md) | 本项目 `MCPAgent` 代码级拆解 |

推荐阅读路径：

- **理解原理**：01 → 02 → 06
- **动手运行**：04 → 本文件 → 06
- **排查故障**：04 → 05 → 06
