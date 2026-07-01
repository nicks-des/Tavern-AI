# Tavern AI

AI 角色扮演聊天平台 —— 酒馆 AI 的复刻项目。

## 架构

```
Web 前端 (React) ──→ Go 后端 ──→ SQLite
       :3000             :8081       tavern.db
                           │
                           └──→ DeepSeek / OpenAI API
```

## 技术栈

| 层 | 技术 |
|---|------|
| 前端 | React 18 + TypeScript + Vite + Tailwind CSS + Zustand |
| 后端 | Go 1.23 + net/http |
| 数据库 | SQLite (WAL 模式 + modernc 纯 Go 驱动) |
| AI | DeepSeek / OpenAI 兼容 API (SSE 流式) |

## 快速开始

### 1. 启动后端

```bash
cd backend

# 编译
go build -o server.exe ./cmd/server/

# 启动 (mock 模式，无需 AI key)
TAVERN_DATA_DIR=./data ./server.exe

# 启动 (使用 DeepSeek)
set OPENAI_API_KEY=sk-xxx
set OPENAI_BASE_URL=https://api.deepseek.com
TAVERN_DATA_DIR=./data server.exe
```

### 2. 启动前端

```bash
cd frontend
npm install
npx vite --port 3000
```

### 3. 打开浏览器

访问 http://localhost:3000

## API 接口

| 方法 | 端点 | 说明 |
|------|------|------|
| POST | /api/characters | 创建角色 |
| GET | /api/characters | 角色列表 |
| GET | /api/characters/{id} | 角色详情 |
| PUT | /api/characters/{id} | 更新角色 |
| DELETE | /api/characters/{id} | 删除角色 |
| POST | /api/sessions | 创建会话 |
| GET | /api/sessions | 会话列表 |
| GET | /api/sessions/{id} | 会话+消息 |
| DELETE | /api/sessions/{id} | 删除会话 |
| POST | /api/sessions/{id}/chat | SSE 流式对话 |

### 对话请求示例

```bash
# 创建角色
curl -X POST http://localhost:8081/api/characters \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","description":"一个友好的AI"}'

# 创建会话
curl -X POST http://localhost:8081/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"characterId":"<char_id>","title":"聊天"}'

# SSE 流式对话
curl -N -X POST http://localhost:8081/api/sessions/<session_id>/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"你好！"}'
```

## 项目结构

```
tavern-ai/
├── backend/
│   ├── cmd/server/main.go          # 入口
│   └── internal/
│       ├── config/config.go        # 配置 (环境变量)
│       ├── database/sqlite.go      # SQLite 连接 + 建表
│       ├── models/models.go        # 数据模型
│       ├── repository/
│       │   ├── character.go        # 角色 CRUD
│       │   └── session.go          # 会话/消息 CRUD
│       ├── handlers/
│       │   ├── character.go        # 角色 API
│       │   ├── session.go          # 会话 API
│       │   └── chat.go             # 聊天 + SSE
│       └── llm/openai.go           # LLM 适配器
├── frontend/
│   └── src/
│       ├── api/                    # API 调用层
│       ├── components/             # React 组件
│       ├── hooks/useChat.ts        # 聊天 SSE hook
│       ├── store/                  # Zustand 状态管理
│       └── types/                  # TypeScript 类型
└── README.md
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| TAVERN_DATA_DIR | ./data | SQLite 数据库目录 |
| TAVERN_HTTP_PORT | 8081 | 后端端口 |
| OPENAI_API_KEY | (空) | API 密钥，不设则 mock 模式 |
| OPENAI_BASE_URL | https://api.openai.com/v1 | API 地址 |
| LLM_MODEL | deepseek-chat | 模型名称 |

## 路线图

- [x] MVP v0.1 — 角色管理 + 单角色对话 + AI 接入
- [ ] 房间系统 + 世界书 (Lorebook)
- [ ] C++ 网络层 (Boost.Asio + WebSocket)
- [ ] 多 LLM 适配器
- [ ] 向量记忆 (RAG)
- [ ] 多角色群聊
- [ ] 插件系统
