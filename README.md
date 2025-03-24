# Trae-DS-Go-Backend

一个基于 Go 语言开发的 AI 聊天后端系统，支持与本地大语言模型通信，提供流式响应和聊天历史记录管理功能。

## 项目简介

本项目是一个轻量级的 AI 聊天后端服务，基于 Gin 框架开发，支持用户注册登录、与本地部署的大语言模型（如 Deepseek）进行对话，并提供完整的聊天历史记录管理功能。系统支持流式响应，提供更好的用户体验。

## 功能特点

-   **用户认证系统**：支持用户注册、登录，使用 JWT 进行身份验证
-   **大语言模型集成**：默认集成本地部署的 Deepseek 模型
-   **流式响应**：支持流式输出 AI 回复，提供更好的用户体验
-   **聊天历史管理**：自动保存聊天记录，支持查询、删除操作
-   **可配置性**：通过环境变量灵活配置服务参数
-   **安全性**：密码加密存储，输入数据验证和清理

## 技术架构

### 核心技术栈

-   **Go 语言**：主要开发语言
-   **Gin**：Web 框架
-   **GORM**：ORM 库，用于数据库操作
-   **SQLite**：轻量级数据库
-   **JWT**：用户认证
-   **HTTP/SSE**：流式数据传输

### 项目结构

```
.
├── config/         # 配置相关代码
│   ├── llm.go      # LLM模型配置和基础请求
│   └── stream.go   # 流式响应处理
├── controllers/    # 控制器
│   ├── auth.go     # 认证相关
│   ├── chat.go     # 聊天功能
│   └── history.go  # 历史记录管理
├── middleware/     # 中间件
│   └── jwt.go      # JWT认证
├── models/         # 数据模型
│   ├── chat_history.go  # 聊天历史记录
│   ├── setup.go    # 数据库设置
│   └── user.go     # 用户模型
├── utils/          # 工具函数
├── .env            # 环境变量配置
├── go.mod          # Go模块定义
├── go.sum          # 依赖校验
└── main.go         # 主程序入口
```

## 安装与部署

### 前置条件

-   Go 1.21 或更高版本
-   本地部署的 Deepseek 模型或其他兼容的 LLM 服务

### 安装步骤

1. 克隆仓库

```bash
git clone https://github.com/yourusername/trae-ds-go-backend.git
cd trae-ds-go-backend
```

2. 安装依赖

```bash
go mod download
```

3. 配置环境变量

创建或修改`.env`文件：

```
# 服务器配置
PORT=8080
GIN_MODE=debug  # 生产环境请设置为release

# JWT配置
JWT_SECRET=your_jwt_secret_key_change_this_in_production

# 数据库配置
DB_PATH=data.db

# LLM模型配置
LLM_API_URL=http://localhost:11434/api/chat/
```

4. 运行服务

```bash
go run main.go
```

服务将在`http://localhost:8080`启动（或根据您在.env 中配置的端口）。

## API 接口文档

### 认证接口

#### 用户注册

```
POST /api/register
```

请求体：

```json
{
    "username": "用户名",
    "password": "密码",
    "email": "邮箱地址"
}
```

响应：

```json
{
    "token": "JWT令牌",
    "user": {
        "id": 1,
        "username": "用户名",
        "email": "邮箱地址"
    }
}
```

#### 用户登录

```
POST /api/login
```

请求体：

```json
{
    "username": "用户名",
    "password": "密码"
}
```

响应：

```json
{
    "token": "JWT令牌",
    "user": {
        "id": 1,
        "username": "用户名",
        "email": "邮箱地址"
    }
}
```

### 模型接口

#### 获取可用模型列表

```
GET /api/models
```

响应：

```json
{
    "models": ["deepseek-r1:7b"]
}
```

### 聊天接口

#### 流式聊天

```
POST /api/stream-chat
```

请求头：

```
Authorization: Bearer <JWT令牌>
```

请求体：

```json
{
    "messages": [{ "role": "user", "content": "你好，请介绍一下自己" }],
    "model": "deepseek-r1:7b",
    "options": {
        "temperature": 0.7
    },
    "history_id": "可选的历史记录ID"
}
```

响应：

服务器发送的是 Server-Sent Events (SSE)格式的流式数据，每个事件包含模型生成的部分响应。

### 聊天历史记录接口

#### 保存聊天历史

```
POST /api/chat-history
```

请求头：

```
Authorization: Bearer <JWT令牌>
```

请求体：

```json
{
    "model": "deepseek-r1:7b",
    "messages": [
        { "role": "user", "content": "你好" },
        { "role": "assistant", "content": "你好！有什么我可以帮助你的吗？" }
    ]
}
```

#### 获取用户的聊天历史列表

```
GET /api/chat-histories
```

请求头：

```
Authorization: Bearer <JWT令牌>
```

#### 获取特定聊天历史详情

```
GET /api/chat-history/:history_id
```

请求头：

```
Authorization: Bearer <JWT令牌>
```

#### 删除聊天历史

```
DELETE /api/chat-history/:id
```

请求头：

```
Authorization: Bearer <JWT令牌>
```

## 配置说明

### 环境变量

| 变量名      | 说明              | 默认值                          |
| ----------- | ----------------- | ------------------------------- |
| PORT        | 服务器端口        | 8080                            |
| GIN_MODE    | Gin 运行模式      | debug                           |
| JWT_SECRET  | JWT 密钥          | -                               |
| DB_PATH     | SQLite 数据库路径 | data.db                         |
| LLM_API_URL | LLM 模型 API 地址 | http://localhost:11434/api/chat |

### LLM 模型配置

默认配置：

```go
var DefaultLLMConfig = LLMConfig{
	APIURL:     "http://localhost:11434/api/chat", // 默认本地deepseek模型API地址
	MaxTokens:  2048,                             // 默认最大生成token数
	Temperature: 0.7,                             // 默认温度参数
	Timeout:     time.Second * 120,               // 默认超时时间
}
```

## 开发与扩展

### 配合前端页面推荐

可配合 github 项目 trae-ds-page-assist 使用

### 自定义响应处理

可以通过修改`controllers/chat.go`中的`ResponseCollector`来自定义响应处理逻辑。

## 许可证

[MIT License](LICENSE)
