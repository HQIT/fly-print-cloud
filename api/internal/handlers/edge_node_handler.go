package handlers

import (
	"log"
	"strconv"
	"time"

	"fly-print-cloud/api/internal/database"
	"fly-print-cloud/api/internal/models"
	"github.com/gin-gonic/gin"
)

// EdgeNodeHandler Edge Node 管理处理器
type EdgeNodeHandler struct {
	edgeNodeRepo *database.EdgeNodeRepository
	printerRepo  *database.PrinterRepository
}

// NewEdgeNodeHandler 创建 Edge Node 管理处理器
func NewEdgeNodeHandler(edgeNodeRepo *database.EdgeNodeRepository, printerRepo *database.PrinterRepository) *EdgeNodeHandler {
	return &EdgeNodeHandler{
		edgeNodeRepo: edgeNodeRepo,
		printerRepo:  printerRepo,
	}
}

// RegisterEdgeNodeRequest Edge Node 注册请求（按照README规划）
type RegisterEdgeNodeRequest struct {
	NodeID string `json:"node_id" binding:"required,min=1,max=100"`
	Name   string `json:"name" binding:"required,min=1,max=100"`
}

// UpdateEdgeNodeRequest Edge Node 更新请求
type UpdateEdgeNodeRequest struct {
	Name              string   `json:"name" binding:"required,min=1,max=100"`
	Status            string   `json:"status" binding:"omitempty,oneof=online offline maintenance"`
	Enabled           *bool    `json:"enabled"`  // 使用指针类型以区分未设置和false
	Version           string   `json:"version"`
	Location          string   `json:"location"`
	Latitude          *float64 `json:"latitude"`
	Longitude         *float64 `json:"longitude"`
	IPAddress         *string  `json:"ip_address"`
	MACAddress        string   `json:"mac_address"`
	NetworkInterface  string   `json:"network_interface"`
	OSVersion         string   `json:"os_version"`
	CPUInfo           string   `json:"cpu_info"`
	MemoryInfo        string   `json:"memory_info"`
	DiskInfo          string   `json:"disk_info"`
	ConnectionQuality string   `json:"connection_quality"`
	Latency           int      `json:"latency"`
}


// EdgeNodeInfo Edge Node 信息响应
type EdgeNodeInfo struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Status            string    `json:"status"`
	Enabled           bool      `json:"enabled"`
	Version           string    `json:"version"`
	LastHeartbeat     time.Time `json:"last_heartbeat"`
	Location          string    `json:"location"`
	Latitude          *float64  `json:"latitude"`
	Longitude         *float64  `json:"longitude"`
	IPAddress         *string   `json:"ip_address,omitempty"`
	MACAddress        string    `json:"mac_address"`
	NetworkInterface  string    `json:"network_interface"`
	OSVersion         string    `json:"os_version"`
	CPUInfo           string    `json:"cpu_info"`
	MemoryInfo        string    `json:"memory_info"`
	DiskInfo          string    `json:"disk_info"`
	ConnectionQuality string    `json:"connection_quality"`
	Latency           int       `json:"latency"`
	PrinterCount      int       `json:"printer_count"`    // 管理的打印机数量
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// RegisterEdgeNode 注册 Edge Node
func (h *EdgeNodeHandler) RegisterEdgeNode(c *gin.Context) {
	var req RegisterEdgeNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err)
		return
	}

	// 创建 Edge Node（按照README规划，只设置基本信息）
	node := &models.EdgeNode{
		ID:            req.NodeID, // 使用客户端提供的 node_id
		Name:          req.Name,
		Status:        "online", // 注册时默认为在线状态
		LastHeartbeat: time.Now(),
	}

	if err := h.edgeNodeRepo.UpsertEdgeNode(node); err != nil {
		log.Printf("Failed to register edge node: %v", err)
		InternalErrorResponse(c, "注册 Edge Node 失败")
		return
	}

	// 返回节点信息
	nodeInfo := EdgeNodeInfo{
		ID:                node.ID,
		Name:              node.Name,
		Status:            node.Status,
		Enabled:           node.Enabled,
		Version:           node.Version,
		LastHeartbeat:     node.LastHeartbeat,
		Location:          node.Location,
		Latitude:          node.Latitude,
		Longitude:         node.Longitude,
		IPAddress:         node.IPAddress,
		MACAddress:        node.MACAddress,
		NetworkInterface:  node.NetworkInterface,
		OSVersion:         node.OSVersion,
		CPUInfo:           node.CPUInfo,
		MemoryInfo:        node.MemoryInfo,
		DiskInfo:          node.DiskInfo,
		ConnectionQuality: node.ConnectionQuality,
		Latency:           node.Latency,
		CreatedAt:         node.CreatedAt,
		UpdatedAt:         node.UpdatedAt,
	}

	log.Printf("Edge Node %s registered successfully", node.Name)
	CreatedResponse(c, nodeInfo)
}

// ListEdgeNodes 获取 Edge Node 列表
func (h *EdgeNodeHandler) ListEdgeNodes(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// 可选：检查并更新超时的节点状态（3分钟超时）
	if updatedCount, err := h.edgeNodeRepo.CheckAndUpdateOfflineNodes(3); err != nil {
		log.Printf("⚠️ [DEBUG] Failed to check offline nodes: %v", err)
	} else if updatedCount > 0 {
		log.Printf("📱 [DEBUG] Updated %d nodes to offline status", updatedCount)
	}

	// 查询 Edge Node 列表
	log.Printf("🔍 [DEBUG] 查询Edge Nodes: offset=%d, pageSize=%d, status='%s'", offset, pageSize, status)
	nodes, total, err := h.edgeNodeRepo.ListEdgeNodes(offset, pageSize, status)
	if err != nil {
		log.Printf("❌ [DEBUG] Failed to list edge nodes: %v", err)
		InternalErrorResponse(c, "获取 Edge Node 列表失败")
		return
	}
	log.Printf("📊 [DEBUG] 查询结果: 找到 %d 个节点，总数 %d", len(nodes), total)

	// 转换为响应格式
	nodeInfos := make([]EdgeNodeInfo, len(nodes))
	for i, node := range nodes {
		// 获取该边缘节点管理的打印机数量
		printerCount, err := h.printerRepo.CountPrintersByEdgeNode(node.ID)
		if err != nil {
			log.Printf("⚠️ [DEBUG] Failed to get printer count for edge node %s: %v", node.ID, err)
			printerCount = 0 // 如果查询失败，设置为0
		}
		
		nodeInfos[i] = EdgeNodeInfo{
			ID:                node.ID,
			Name:              node.Name,
			Status:            node.Status,
			Enabled:           node.Enabled,
			Version:           node.Version,
			LastHeartbeat:     node.LastHeartbeat,
			Location:          node.Location,
			Latitude:          node.Latitude,
			Longitude:         node.Longitude,
			IPAddress:         node.IPAddress,
			MACAddress:        node.MACAddress,
			NetworkInterface:  node.NetworkInterface,
			OSVersion:         node.OSVersion,
			CPUInfo:           node.CPUInfo,
			MemoryInfo:        node.MemoryInfo,
			DiskInfo:          node.DiskInfo,
			ConnectionQuality: node.ConnectionQuality,
			Latency:           node.Latency,
			PrinterCount:      printerCount,
			CreatedAt:         node.CreatedAt,
			UpdatedAt:         node.UpdatedAt,
		}
	}

	PaginatedSuccessResponse(c, nodeInfos, total, page, pageSize)
}

// GetEdgeNode 获取 Edge Node 详情
func (h *EdgeNodeHandler) GetEdgeNode(c *gin.Context) {
	nodeID := c.Param("id")
	if nodeID == "" {
		BadRequestResponse(c, "Edge Node ID 不能为空")
		return
	}

	node, err := h.edgeNodeRepo.GetEdgeNodeByID(nodeID)
	if err != nil {
		NotFoundResponse(c, "Edge Node 不存在")
		return
	}

	// 获取该边缘节点管理的打印机数量
	printerCount, err := h.printerRepo.CountPrintersByEdgeNode(node.ID)
	if err != nil {
		log.Printf("⚠️ [DEBUG] Failed to get printer count for edge node %s: %v", node.ID, err)
		printerCount = 0 // 如果查询失败，设置为0
	}

	nodeInfo := EdgeNodeInfo{
		ID:                node.ID,
		Name:              node.Name,
		Status:            node.Status,
		Enabled:           node.Enabled,
		Version:           node.Version,
		LastHeartbeat:     node.LastHeartbeat,
		Location:          node.Location,
		Latitude:          node.Latitude,
		Longitude:         node.Longitude,
		IPAddress:         node.IPAddress,
		MACAddress:        node.MACAddress,
		NetworkInterface:  node.NetworkInterface,
		OSVersion:         node.OSVersion,
		CPUInfo:           node.CPUInfo,
		MemoryInfo:        node.MemoryInfo,
		DiskInfo:          node.DiskInfo,
		ConnectionQuality: node.ConnectionQuality,
		Latency:           node.Latency,
		PrinterCount:      printerCount,
		CreatedAt:         node.CreatedAt,
		UpdatedAt:         node.UpdatedAt,
	}

	SuccessResponse(c, nodeInfo)
}

// UpdateEdgeNode 更新 Edge Node
func (h *EdgeNodeHandler) UpdateEdgeNode(c *gin.Context) {
	nodeID := c.Param("id")
	if nodeID == "" {
		BadRequestResponse(c, "Edge Node ID 不能为空")
		return
	}

	var req UpdateEdgeNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err)
		return
	}

	// 检查节点是否存在
	node, err := h.edgeNodeRepo.GetEdgeNodeByID(nodeID)
	if err != nil {
		NotFoundResponse(c, "Edge Node 不存在")
		return
	}

	// 更新节点信息
	node.Name = req.Name
	
	// 只有当Status字段不为空时才更新
	if req.Status != "" {
		node.Status = req.Status
	}
	
	// 处理Enabled字段更新（逻辑级联，不修改printer的enable状态）
	if req.Enabled != nil {
		node.Enabled = *req.Enabled
	}
	
	node.Version = req.Version
	node.Location = req.Location
	node.Latitude = req.Latitude
	node.Longitude = req.Longitude
	node.IPAddress = req.IPAddress
	node.MACAddress = req.MACAddress
	node.NetworkInterface = req.NetworkInterface
	node.OSVersion = req.OSVersion
	node.CPUInfo = req.CPUInfo
	node.MemoryInfo = req.MemoryInfo
	node.DiskInfo = req.DiskInfo
	node.ConnectionQuality = req.ConnectionQuality
	node.Latency = req.Latency

	if err := h.edgeNodeRepo.UpdateEdgeNode(node); err != nil {
		log.Printf("Failed to update edge node %s: %v", nodeID, err)
		InternalErrorResponse(c, "更新 Edge Node 失败")
		return
	}

	nodeInfo := EdgeNodeInfo{
		ID:                node.ID,
		Name:              node.Name,
		Status:            node.Status,
		Enabled:           node.Enabled,
		Version:           node.Version,
		LastHeartbeat:     node.LastHeartbeat,
		Location:          node.Location,
		Latitude:          node.Latitude,
		Longitude:         node.Longitude,
		IPAddress:         node.IPAddress,
		MACAddress:        node.MACAddress,
		NetworkInterface:  node.NetworkInterface,
		OSVersion:         node.OSVersion,
		CPUInfo:           node.CPUInfo,
		MemoryInfo:        node.MemoryInfo,
		DiskInfo:          node.DiskInfo,
		ConnectionQuality: node.ConnectionQuality,
		Latency:           node.Latency,
		CreatedAt:         node.CreatedAt,
		UpdatedAt:         node.UpdatedAt,
	}

	log.Printf("Edge Node %s updated successfully", node.Name)
	SuccessResponse(c, nodeInfo)
}

// DeleteEdgeNode 删除 Edge Node
func (h *EdgeNodeHandler) DeleteEdgeNode(c *gin.Context) {
	nodeID := c.Param("id")
	if nodeID == "" {
		BadRequestResponse(c, "Edge Node ID 不能为空")
		return
	}

	// 检查节点是否存在
	_, err := h.edgeNodeRepo.GetEdgeNodeByID(nodeID)
	if err != nil {
		NotFoundResponse(c, "Edge Node 不存在")
		return
	}

	// 删除节点（软删除）
	if err := h.edgeNodeRepo.DeleteEdgeNode(nodeID); err != nil {
		log.Printf("Failed to delete edge node %s: %v", nodeID, err)
		InternalErrorResponse(c, "删除 Edge Node 失败")
		return
	}

	log.Printf("Edge Node %s deleted successfully", nodeID)
	SuccessResponse(c, gin.H{"message": "Edge Node 删除成功"})
}

// HeartbeatRequest 心跳请求
type HeartbeatRequest struct {
	NodeID string `json:"node_id" binding:"required"`
}

// Heartbeat Edge Node 心跳
func (h *EdgeNodeHandler) Heartbeat(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err)
		return
	}

	// 更新心跳时间
	if err := h.edgeNodeRepo.UpdateHeartbeat(req.NodeID); err != nil {
		log.Printf("Failed to update heartbeat for edge node %s: %v", req.NodeID, err)
		InternalErrorResponse(c, "更新心跳失败")
		return
	}

	// 更新状态为在线
	if err := h.edgeNodeRepo.UpdateStatus(req.NodeID, "online"); err != nil {
		log.Printf("Failed to update status for edge node %s: %v", req.NodeID, err)
		InternalErrorResponse(c, "更新状态失败")
		return
	}

	SuccessResponse(c, gin.H{"message": "心跳更新成功"})
}
