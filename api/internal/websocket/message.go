package websocket

import "time"


// 基础消息格式
type Message struct {
	Type      string      `json:"type"`
	NodeID    string      `json:"node_id"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// 上行消息类型
const (
	MsgTypeHeartbeat     = "edge_heartbeat"
	MsgTypePrinterStatus = "printer_status"
	MsgTypeJobUpdate     = "job_update"
)

// 下行指令类型
const (
	CmdTypePrintJob     = "print_job"
	CmdTypeConfigUpdate = "config_update"
	CmdTypeReportStatus = "report_status"
)

// 指令消息格式
type Command struct {
	Type      string      `json:"type"`
	CommandID string      `json:"command_id"`
	Timestamp time.Time   `json:"timestamp"`
	Target    string      `json:"target"`    // edge_node_id 或 printer_id
	Data      interface{} `json:"data"`
}

// 指令确认响应
type CommandAck struct {
	Type      string    `json:"type"`
	CommandID string    `json:"command_id"`
	NodeID    string    `json:"node_id"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`  // accepted/rejected/processing
	Message   string    `json:"message"`
}

// 心跳数据
type HeartbeatData struct {
	SystemInfo SystemInfo `json:"system_info"`
}

type SystemInfo struct {
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    float64 `json:"memory_usage"`
	DiskUsage      float64 `json:"disk_usage"`
	NetworkQuality string  `json:"network_quality"`
	Latency        int     `json:"latency"`
}

// 打印机状态数据
type PrinterStatusData struct {
	PrinterID   string            `json:"printer_id"`
	Status      string            `json:"status"`
	QueueLength int               `json:"queue_length"`
	ErrorCode   *string           `json:"error_code"`
	Supplies    map[string]interface{} `json:"supplies"`
}

// 任务状态更新数据
type JobUpdateData struct {
	JobID        string  `json:"job_id"`
	Status       string  `json:"status"`
	Progress     int     `json:"progress"`
	ErrorMessage *string `json:"error_message"`
}

// 打印任务分发数据
type PrintJobData struct {
	JobID       string `json:"job_id"`
	Name        string `json:"name"`
	PrinterID   string `json:"printer_id"`
	PrinterName string `json:"printer_name"`
	FilePath    string `json:"file_path,omitempty"`
	FileURL     string `json:"file_url,omitempty"`
	FileSize    int64  `json:"file_size"`
	PageCount   int    `json:"page_count"`
	Copies      int    `json:"copies"`
	PaperSize   string `json:"paper_size"`
	ColorMode   string `json:"color_mode"`
	DuplexMode  string `json:"duplex_mode"`
	MaxRetries  int    `json:"max_retries"`
}
