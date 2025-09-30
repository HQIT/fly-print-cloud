package handlers

import (
	"fly-print-cloud/api/internal/database"
	"fly-print-cloud/api/internal/models"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PrinterHandler struct {
	printerRepo  *database.PrinterRepository
	edgeNodeRepo *database.EdgeNodeRepository
}

func NewPrinterHandler(printerRepo *database.PrinterRepository, edgeNodeRepo *database.EdgeNodeRepository) *PrinterHandler {
	return &PrinterHandler{
		printerRepo:  printerRepo,
		edgeNodeRepo: edgeNodeRepo,
	}
}



// UpdatePrinterRequest 更新打印机请求（Edge Node使用）
type UpdatePrinterRequest struct {
	Name            string                        `json:"name" binding:"required,min=1,max=100"`
	Model           string                        `json:"model"`
	SerialNumber    string                        `json:"serial_number"`
	Status          string                        `json:"status" binding:"required,oneof=ready printing error offline"`
	FirmwareVersion string                        `json:"firmware_version"`
	PortInfo        string                        `json:"port_info"`
	IPAddress       *string                       `json:"ip_address"`
	MACAddress      string                        `json:"mac_address"`
	NetworkConfig   string                        `json:"network_config"`
	Latitude        *float64                      `json:"latitude"`
	Longitude       *float64                      `json:"longitude"`
	Location        string                        `json:"location"`
	Capabilities    models.PrinterCapabilities    `json:"capabilities"`
	QueueLength     int                           `json:"queue_length"`
}

// AdminUpdatePrinterRequest 管理界面更新打印机请求
type AdminUpdatePrinterRequest struct {
	DisplayName string `json:"display_name" binding:"omitempty,max=100"`
	Enabled     *bool  `json:"enabled"`  // 使用指针类型以区分未设置和false
}

// PrinterWithStatus 包含实际状态的打印机信息
type PrinterWithStatus struct {
	*models.Printer
	EdgeNodeEnabled bool   `json:"edge_node_enabled"`
	ActuallyEnabled bool   `json:"actually_enabled"`
	DisabledReason  string `json:"disabled_reason,omitempty"`
}

// NewPrinterWithStatus 创建包含实际状态的打印机信息
func NewPrinterWithStatus(printer *models.Printer, edgeNodeEnabled bool) *PrinterWithStatus {
	actuallyEnabled := printer.Enabled && edgeNodeEnabled
	
	var disabledReason string
	if !actuallyEnabled {
		if !printer.Enabled {
			disabledReason = "打印机被禁用"
		} else if !edgeNodeEnabled {
			disabledReason = "Edge Node被禁用"
		}
	}
	
	return &PrinterWithStatus{
		Printer:         printer,
		EdgeNodeEnabled: edgeNodeEnabled,
		ActuallyEnabled: actuallyEnabled,
		DisabledReason:  disabledReason,
	}
}

// Edge 注册打印机请求（简化版）
type EdgeRegisterPrinterRequest struct {
	Name            string                        `json:"name" binding:"required,min=1,max=100"`
	Model           string                        `json:"model"`
	SerialNumber    string                        `json:"serial_number"`
	FirmwareVersion string                        `json:"firmware_version"`
	PortInfo        string                        `json:"port_info"`
	IPAddress       *string                       `json:"ip_address"`
	MACAddress      string                        `json:"mac_address"`
	Capabilities    models.PrinterCapabilities    `json:"capabilities"`
}

// 管理员 API

// ListPrinters 获取所有打印机列表（管理员）
func (h *PrinterHandler) ListPrinters(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	edgeNodeID := c.Query("edge_node_id") // 支持按Edge Node筛选

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var printers []*models.Printer
	var total int
	var err error

	if edgeNodeID != "" {
		// 按Edge Node筛选
		printers, err = h.printerRepo.ListPrintersByEdgeNode(edgeNodeID)
		if err != nil {
			log.Printf("Failed to list printers by edge node: %v", err)
			InternalErrorResponse(c, "获取打印机列表失败")
			return
		}
		total = len(printers)
	} else {
		// 获取所有打印机
		printers, total, err = h.printerRepo.ListPrinters(page, pageSize)
		if err != nil {
			log.Printf("Failed to list printers: %v", err)
			InternalErrorResponse(c, "获取打印机列表失败")
			return
		}
	}

	// 获取所有相关的Edge Node状态信息
	edgeNodeStatusMap := make(map[string]bool)
	for _, printer := range printers {
		if _, exists := edgeNodeStatusMap[printer.EdgeNodeID]; !exists {
			edgeNode, err := h.edgeNodeRepo.GetEdgeNodeByID(printer.EdgeNodeID)
			if err != nil {
				log.Printf("Failed to get edge node %s: %v", printer.EdgeNodeID, err)
				edgeNodeStatusMap[printer.EdgeNodeID] = false // 默认为禁用
			} else {
				edgeNodeStatusMap[printer.EdgeNodeID] = edgeNode.Enabled
			}
		}
	}

	// 转换为包含实际状态的打印机信息
	printersWithStatus := make([]*PrinterWithStatus, len(printers))
	for i, printer := range printers {
		edgeNodeEnabled := edgeNodeStatusMap[printer.EdgeNodeID]
		printersWithStatus[i] = NewPrinterWithStatus(printer, edgeNodeEnabled)
	}

	totalPages := (total + pageSize - 1) / pageSize
	response := gin.H{
		"items":       printersWithStatus,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	}

	SuccessResponse(c, response)
}

// GetPrinter 获取打印机详情
func (h *PrinterHandler) GetPrinter(c *gin.Context) {
	printerID := c.Param("id")
	if printerID == "" {
		BadRequestResponse(c, "打印机 ID 不能为空")
		return
	}

	printer, err := h.printerRepo.GetPrinterByID(printerID)
	if err != nil {
		NotFoundResponse(c, "打印机不存在")
		return
	}

	// 获取Edge Node状态
	edgeNode, err := h.edgeNodeRepo.GetEdgeNodeByID(printer.EdgeNodeID)
	if err != nil {
		log.Printf("Failed to get edge node %s: %v", printer.EdgeNodeID, err)
		// 如果无法获取Edge Node状态，假设为禁用
		printerWithStatus := NewPrinterWithStatus(printer, false)
		SuccessResponse(c, printerWithStatus)
		return
	}

	printerWithStatus := NewPrinterWithStatus(printer, edgeNode.Enabled)
	SuccessResponse(c, printerWithStatus)
}

// UpdatePrinter 更新打印机（管理员）
func (h *PrinterHandler) UpdatePrinter(c *gin.Context) {
	printerID := c.Param("id")
	if printerID == "" {
		BadRequestResponse(c, "打印机 ID 不能为空")
		return
	}

	// 检查打印机是否存在
	printer, err := h.printerRepo.GetPrinterByID(printerID)
	if err != nil {
		NotFoundResponse(c, "打印机不存在")
		return
	}

	// 尝试解析为管理界面的简单更新请求
	var adminReq AdminUpdatePrinterRequest
	if err := c.ShouldBindJSON(&adminReq); err == nil {
		// 管理界面更新（仅更新display_name和enabled）
		if adminReq.DisplayName != "" {
			printer.DisplayName = adminReq.DisplayName
		}
		if adminReq.Enabled != nil {
			printer.Enabled = *adminReq.Enabled
		}
	} else {
		// 尝试解析为Edge Node的完整更新请求
		var req UpdatePrinterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数无效")
			return
		}

		// 更新所有打印机信息
		printer.Name = req.Name
		printer.Model = req.Model
		printer.SerialNumber = req.SerialNumber
		printer.Status = req.Status
		printer.FirmwareVersion = req.FirmwareVersion
		printer.PortInfo = req.PortInfo
		printer.IPAddress = req.IPAddress
		printer.MACAddress = req.MACAddress
		printer.NetworkConfig = req.NetworkConfig
		printer.Latitude = req.Latitude
		printer.Longitude = req.Longitude
		printer.Location = req.Location
		printer.Capabilities = req.Capabilities
		printer.QueueLength = req.QueueLength
	}

	if err := h.printerRepo.UpdatePrinter(printer); err != nil {
		log.Printf("Failed to update printer %s: %v", printerID, err)
		InternalErrorResponse(c, "更新打印机失败")
		return
	}

	log.Printf("Printer %s updated successfully", printer.Name)
	SuccessResponse(c, printer)
}

// DeletePrinter 删除打印机
func (h *PrinterHandler) DeletePrinter(c *gin.Context) {
	printerID := c.Param("id")
	if printerID == "" {
		BadRequestResponse(c, "打印机 ID 不能为空")
		return
	}

	// 检查打印机是否存在
	_, err := h.printerRepo.GetPrinterByID(printerID)
	if err != nil {
		NotFoundResponse(c, "打印机不存在")
		return
	}

	// 删除打印机
	if err := h.printerRepo.DeletePrinter(printerID); err != nil {
		log.Printf("Failed to delete printer %s: %v", printerID, err)
		InternalErrorResponse(c, "删除打印机失败")
		return
	}

	log.Printf("Printer %s deleted successfully", printerID)
	SuccessResponse(c, gin.H{"message": "打印机删除成功"})
}

// Edge Node API

// EdgeRegisterPrinter Edge Node 注册打印机
func (h *PrinterHandler) EdgeRegisterPrinter(c *gin.Context) {
	edgeNodeID := c.Param("node_id")
	if edgeNodeID == "" {
		BadRequestResponse(c, "Edge Node ID 不能为空")
		return
	}

	var req EdgeRegisterPrinterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "请求参数无效")
		return
	}

	// 验证 Edge Node 是否存在
	_, err := h.edgeNodeRepo.GetEdgeNodeByID(edgeNodeID)
	if err != nil {
		BadRequestResponse(c, "Edge Node 不存在")
		return
	}

	printer := &models.Printer{
		ID:              uuid.New().String(),
		Name:            req.Name,
		Model:           req.Model,
		SerialNumber:    req.SerialNumber,
		Status:          "offline", // 默认状态
		FirmwareVersion: req.FirmwareVersion,
		PortInfo:        req.PortInfo,
		IPAddress:       req.IPAddress,
		MACAddress:      req.MACAddress,
		NetworkConfig:   "",
		Capabilities:    req.Capabilities,
		EdgeNodeID:      edgeNodeID,
		QueueLength:     0,
	}

	if err := h.printerRepo.UpsertPrinter(printer); err != nil {
		log.Printf("Failed to register/update printer by edge node %s: %v", edgeNodeID, err)
		InternalErrorResponse(c, "注册打印机失败")
		return
	}

	log.Printf("Printer %s registered/updated by edge node %s", printer.Name, edgeNodeID)
	CreatedResponse(c, printer)
}

// EdgeListPrinters Edge Node 获取自己的打印机列表
func (h *PrinterHandler) EdgeListPrinters(c *gin.Context) {
	edgeNodeID := c.Param("node_id")
	if edgeNodeID == "" {
		BadRequestResponse(c, "Edge Node ID 不能为空")
		return
	}

	// 验证 Edge Node 是否存在
	_, err := h.edgeNodeRepo.GetEdgeNodeByID(edgeNodeID)
	if err != nil {
		BadRequestResponse(c, "Edge Node 不存在")
		return
	}

	printers, err := h.printerRepo.ListPrintersByEdgeNode(edgeNodeID)
	if err != nil {
		log.Printf("Failed to list printers for edge node %s: %v", edgeNodeID, err)
		InternalErrorResponse(c, "获取打印机列表失败")
		return
	}

	SuccessResponse(c, gin.H{"items": printers})
}
