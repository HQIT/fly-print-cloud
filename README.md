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

## 数据模型

### Edge Node
```go
type EdgeNode struct {
    ID              string    `json:"id"`
    Name            string    `json:"name"`
    Status          string    `json:"status"` // online/offline
    Version         string    `json:"version"`
    LastHeartbeat   time.Time `json:"last_heartbeat"`
    
    // 位置信息
    Location        string    `json:"location"`        // 地理位置描述
    Latitude        *float64  `json:"latitude,omitempty"`      // 纬度
    Longitude       *float64  `json:"longitude,omitempty"`     // 经度
    //Floor           string    `json:"floor"`           // 楼层
    //Room            string    `json:"room"`            // 房间
    
    // 网络信息
    IPAddress      string    `json:"ip_address"`      // IP地址
    MACAddress     string    `json:"mac_address"`     // MAC地址
    NetworkInterface string  `json:"network_interface"` // 网络接口
    
    // 系统信息
    OSVersion      string    `json:"os_version"`      // 操作系统版本
    CPUInfo        string    `json:"cpu_info"`        // CPU信息
    MemoryInfo     string    `json:"memory_info"`     // 内存信息
    DiskInfo       string    `json:"disk_info"`       // 磁盘信息
    
    // 连接信息
    ConnectionQuality string `json:"connection_quality"` // 连接质量
    Latency         int     `json:"latency"`         // 延迟(ms)
    
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
```

### Printer
```go
type Printer struct {
    ID           string   `json:"id"`
    Name         string   `json:"name"`
    Model        string   `json:"model"`
    SerialNumber string   `json:"serial_number"`    // 序列号
    Status       string   `json:"status"`           // ready/printing/error/offline
    
    // 硬件信息
    FirmwareVersion string `json:"firmware_version"` // 固件版本
    PortInfo       string `json:"port_info"`        // 端口信息
    
    // 网络信息
    IPAddress     string `json:"ip_address"`        // IP地址
    MACAddress    string `json:"mac_address"`       // MAC地址
    NetworkConfig string `json:"network_config"`    // 网络配置
    
    // 地理位置信息 (可选)
    Latitude     *float64 `json:"latitude,omitempty"`      // 纬度
    Longitude    *float64 `json:"longitude,omitempty"`     // 经度
    Location     string   `json:"location,omitempty"`      // 位置描述
    
    // 能力信息
    Capabilities  PrinterCapabilities `json:"capabilities"`
    
    // 关联信息
    EdgeNodeID   string `json:"edge_node_id"`       // 关联Edge Node
    QueueLength  int    `json:"queue_length"`       // 队列长度
    
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

type PrinterCapabilities struct {
    PaperSizes   []string `json:"paper_sizes"`     // 支持的纸张尺寸
    ColorSupport bool     `json:"color_support"`   // 是否支持彩色
    DuplexSupport bool    `json:"duplex_support"`  // 是否支持双面
    Resolution   string   `json:"resolution"`      // 分辨率
    PrintSpeed   string   `json:"print_speed"`     // 打印速度
    MediaTypes   []string `json:"media_types"`     // 支持的介质类型
}
```

### Print Job
```go
type PrintJob struct {
    ID           string    `json:"id"`
    Name         string    `json:"name"`
    Status       string    `json:"status"`        // pending/printing/completed/failed/cancelled
    Priority     int       `json:"priority"`      // 优先级 1-10
    
    // 关联信息
    PrinterID    string    `json:"printer_id"`
    EdgeNodeID   string    `json:"edge_node_id"`
    UserID       string    `json:"user_id"`       // 提交用户
    UserName     string    `json:"user_name"`     // 提交用户名
    
    // 任务信息
    FilePath     string    `json:"file_path"`     // 文件路径
    FileSize     int64     `json:"file_size"`     // 文件大小
    PageCount    int       `json:"page_count"`    // 页数
    Copies       int       `json:"copies"`        // 份数
    
    // 打印设置
    PaperSize    string    `json:"paper_size"`
    ColorMode    string    `json:"color_mode"`    // color/grayscale
    DuplexMode   string    `json:"duplex_mode"`   // single/duplex
    
    // 执行信息
    //Progress     int       `json:"progress"`      // 进度 0-100
    StartTime    time.Time `json:"start_time"`
    EndTime      time.Time `json:"end_time"`
    ErrorMessage string    `json:"error_message"`
    
    // 重试信息
    RetryCount   int       `json:"retry_count"`
    MaxRetries   int       `json:"max_retries"`
    
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

### User
```go
type User struct {
    ID           string    `json:"id"`
    Username     string    `json:"username"`     // 用户名
    Email        string    `json:"email"`        // 邮箱
    PasswordHash string    `json:"-"`            // 密码哈希 (不返回)
    Role         string    `json:"role"`         // 角色: admin/operator/viewer
    Status       string    `json:"status"`       // 状态: active/inactive
    LastLogin    time.Time `json:"last_login"`   // 最后登录时间
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

## 通信协议

### 下行指令策略

**设计原则：**
- Edge端主动建立WebSocket长连接
- Cloud通过WebSocket推送下行指令
- 支持指令确认和重试机制
- 指令系统具备良好的扩展性

**下行指令类型：**
- 打印任务下发
- 打印机配置更新
- 立即上报状态指令
```
POST /api/edge/register
Headers: Authorization: Bearer {pre-allocated-token}
Body: {
  "node_id": "edge-001",
  "name": "办公室打印服务器"
}
```

### 2. Printer 注册
```
POST /api/printer/register
Headers: Authorization: Bearer {edge-node-token}
Body: {
  "printer_id": "hp-001",
  "name": "HP-LaserJet-001",
  "model": "HP LaserJet Pro",
  "edge_node_id": "edge-001",
  "capabilities": ["color", "duplex", "a4"]
}
```

### 3. 状态上报 (WebSocket)

**设计原则：**
- EdgeNode状态：低频上报（30秒间隔）
- Printer状态：高频上报（状态变化时立即上报，单台）
- Job状态：实时上报（任务状态变化时）

**EdgeNode 状态上报 (低频)：**
```json
{
  "type": "edge_heartbeat",
  "node_id": "edge-001",
  "timestamp": "2024-01-01T10:00:00Z",
  "system_info": {
    "cpu_usage": 45.2,
    "memory_usage": 67.8,
    "disk_usage": 23.1,
    "network_quality": "good",
    "latency": 15
  }
}
```

**Printer 状态上报 (高频，单台)：**
```json
{
  "type": "printer_status",
  "node_id": "edge-001",
  "printer_id": "hp-001",
  "timestamp": "2024-01-01T10:00:00Z",
  "status": "ready",
  "queue_length": 2,
  "error_code": null,
  "supplies": {
    "toner": 85,
    "paper": "sufficient"
  }
}
```



**Job 状态更新：**
```json
{
  "type": "job_update",
  "node_id": "edge-001",
  "timestamp": "2024-01-01T10:00:00Z",
  "job_id": "job-123",
  "status": "printing",
  "progress": 65,
  "error_message": null
}
```

### 4. 下行指令 (Cloud → Edge)

**指令格式：**
```json
{
  "type": "指令类型",
  "command_id": "唯一指令ID",
  "timestamp": "指令时间戳",
  "target": "目标对象",              // edge_node_id 或 printer_id
  "data": {                        // 指令数据
    // 根据指令类型定义
  }
}
```

**打印任务指令：**
```json
{
  "type": "print_job",
  "command_id": "cmd-001",
  "timestamp": "2024-01-01T10:00:00Z",
  "target": "hp-001",
  "data": {
    "job_id": "job-123",
    "file_url": "https://cloud.example.com/files/job-123.pdf",
    "copies": 2,
    "paper_size": "A4",
    "color_mode": "color",
    "duplex_mode": "duplex"
  }
}
```

**配置更新指令：**
```json
{
  "type": "config_update",
  "command_id": "cmd-002",
  "timestamp": "2024-01-01T10:00:00Z",
  "target": "hp-001",
  "data": {
    "default_paper_size": "A4",
    "idle_timeout": 300
  }
}
```

**状态上报指令：**
```json
{
  "type": "report_status",
  "command_id": "cmd-003",
  "timestamp": "2024-01-01T10:00:00Z",
  "target": "edge-001",
  "data": {
    "scope": "all"
  }
}
```

**指令确认响应：**
```json
{
  "type": "command_ack",
  "command_id": "cmd-001",
  "node_id": "edge-001",
  "timestamp": "2024-01-01T10:00:01Z",
  "status": "accepted", // accepted/rejected/processing
  "message": "任务已接收，开始下载文件"
}
```

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

## 开发计划

### Phase 1: 基础架构
- [x] Go项目结构搭建
- [x] 数据库设计和迁移
- [x] 基础API框架
- [x] 内置用户管理系统
- [ ] OAuth集成 (可选)

### Phase 2: 核心功能
- [ ] Edge Node管理API (开发中，待测试)
- [ ] 打印机管理API
- [ ] 打印任务管理API
- [ ] WebSocket连接管理
- [ ] 分离式状态上报和心跳
- [ ] 任务队列和优先级管理

### Phase 3: 管理界面
- [x] React前端框架 (基础结构完成，构建遇到模块解析问题)
- [ ] 用户登录和管理页面
- [ ] Edge Node管理页面
- [ ] 打印机管理页面
- [ ] 打印任务管理页面
- [ ] 打印历史页面
- [ ] 实时状态监控

### Phase 4: 高级功能
- [ ] 打印任务管理
- [ ] 系统监控和告警
- [ ] 日志和审计
- [ ] 性能优化

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

## 贡献指南

欢迎提交Issue和Pull Request来改进这个项目。

## 许可证

[待定]

## 当前进度

**Phase 1**: 90% 完成 (基础架构完成，包括用户管理、认证系统，OAuth2集成待完成)
**Phase 2**: 15% 完成 (Edge Node管理API开发中)
**Phase 3**: 10% 完成 (React框架基础结构完成)
**Phase 4**: 0% 完成
