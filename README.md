# Fly Print Cloud - 云打印管理系统

## 系统概述

Fly Print Cloud 是一个云打印管理后台系统，用于管理和监控分布式的打印资源。系统包含管理后台、API服务、边缘端连接器等核心组件，支持多Edge Node和多打印机的统一管理。

## 系统架构

```
fly-print-cloud/
├── admin-console/          # 管理后台 (React + TypeScript)
├── api-server/            # API服务 (Go + Gin/Echo)
├── edge-connector/        # 边缘端连接器 (Go WebSocket)
├── shared/                # 共享类型定义和工具
├── docker-compose.yml     # 开发环境
└── README.md             # 项目说明
```

## 核心功能

### 1. Admin Console (管理后台)
- **Edge Node 管理**：注册、状态监控、连接管理、设备信息
- **打印机管理**：注册、状态监控、配置管理、硬件信息
- **打印任务管理**：创建、监控、队列管理、优先级控制
- **打印历史**：任务记录、统计报表、审计日志
- **系统监控**：实时状态、连接状态、健康检查、性能指标
- **用户管理**：内置用户管理，支持本地登录
- **权限控制**：基于角色的权限管理

### 2. API Server (Headless API)
- 打印机资源管理 API
- 打印任务 API
- 边缘端通信 API
- 认证和授权 (Keycloak集成)

### 3. Edge Connector (边缘端连接)
- WebSocket 连接管理
- 状态上报接收
- 任务下发
- 心跳检测

## 技术选型

### 后端
- **语言**: Go
- **Web框架**: Gin/Echo
- **WebSocket**: Gorilla WebSocket
- **数据库**: PostgreSQL + Redis
- **消息队列**: Redis Streams
- **认证**: 标准OAuth 2.0/OpenID Connect (支持Keycloak等)

### 前端
- **框架**: React + TypeScript
- **UI组件**: Ant Design/Element Plus
- **状态管理**: Redux/Zustand
- **WebSocket**: 原生WebSocket API

## 权限控制

### 认证方式
- **内置用户管理**: 支持用户名密码登录，适合简单部署
- **OAuth 2.0**: 支持外部OAuth Provider (Keycloak、Auth0等)
- **并行支持**: 两种认证方式可以同时使用

### 内置用户管理
- 用户名/密码登录
- 支持用户创建、修改、删除
- 密码加密存储
- 登录历史记录

### RBAC角色设计
- **admin**: 所有权限
- **operator**: 查看状态 + 创建任务
- **viewer**: 只读权限
- **edge-manager**: Edge Node管理权限
- **print-manager**: 打印任务管理权限

## 部署架构

### 网络拓扑
```
公网/上级网络
┌─────────────────────────────────────────────────────────────┐
│                    Fly Print Cloud                         │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │   Admin Console │    │   API Server    │                │
│  │   (React SPA)   │◄──►│   (Go HTTP)     │                │
│  └─────────────────┘    └─────────────────┘                │
│           │                       │                        │
│           │                       ▼                        │
│           │              ┌─────────────────┐                │
│           │              │  Edge Connector │                │
│           │              │  (Go WebSocket) │                │
│           │              └─────────────────┘                │
│           │                       │                        │
│           │                       ▼                        │
│           │              ┌─────────────────┐                │
│           │              │   PostgreSQL    │                │
│           │              │   + Redis       │                │
│           │              └─────────────────┘                │
└─────────────────────────────────────────────────────────────┘
                                    │
                                    │ WebSocket 长连接
                                    ▼
子网/局域网
┌─────────────────────────────────────────────────────────────┐
│                    Fly Print Edge                          │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │   Edge Service  │    │  Printer Driver │                │
│  │  (Go WebSocket) │◄──►│   Interface     │                │
│  └─────────────────┘    └─────────────────┘                │
│           │                       │                        │
│           │                       ▼                        │
│           │              ┌─────────────────┐                │
│           │              │   Printers      │                │
│           │              │   (USB/Network) │                │
│           │              └─────────────────┘                │
└─────────────────────────────────────────────────────────────┘
```

### 通信特点
- **上行**: Edge → Cloud (状态上报、心跳)
- **下行**: Cloud → Edge (指令下发、任务推送)
- **连接**: Edge主动建立WebSocket长连接
- **实时性**: 支持实时指令下发和状态上报

## 快速开始

### 环境要求
- Go 1.21+
- PostgreSQL 14+
- Redis 6+
- Node.js 18+
- Docker & Docker Compose

### 本地开发
```bash
# 克隆项目
git clone <repository-url>
cd fly-print-cloud

# 启动依赖服务
docker-compose up -d

# 启动后端服务
cd api-server
go run main.go

# 启动前端服务
cd admin-console
npm install
npm run dev
```