package models

import (
	"time"
)

// EdgeNode Edge节点
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
	
	// 网络信息
	IPAddress      *string   `json:"ip_address,omitempty"`      // IP地址
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

// Printer 打印机
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
	IPAddress     *string `json:"ip_address"`        // IP地址 (修复：改为指针类型以处理NULL)
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

// PrinterCapabilities 打印机能力
type PrinterCapabilities struct {
	PaperSizes   []string `json:"paper_sizes"`     // 支持的纸张尺寸
	ColorSupport bool     `json:"color_support"`   // 是否支持彩色
	DuplexSupport bool    `json:"duplex_support"`  // 是否支持双面
	Resolution   string   `json:"resolution"`      // 分辨率
	PrintSpeed   string   `json:"print_speed"`     // 打印速度
	MediaTypes   []string `json:"media_types"`     // 支持的介质类型
}

// PrintJob 打印任务
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
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	ErrorMessage string    `json:"error_message"`
	
	// 重试信息
	RetryCount   int       `json:"retry_count"`
	MaxRetries   int       `json:"max_retries"`
	
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// User 用户
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
