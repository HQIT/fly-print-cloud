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



// UpdatePrinterRequest 更新打印机请求
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

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	printers, total, err := h.printerRepo.ListPrinters(page, pageSize)
	if err != nil {
		log.Printf("Failed to list printers: %v", err)
		InternalErrorResponse(c, "获取打印机列表失败")
		return
	}

	totalPages := (total + pageSize - 1) / pageSize
	response := gin.H{
		"items":       printers,
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

	SuccessResponse(c, printer)
}

// UpdatePrinter 更新打印机（管理员）
func (h *PrinterHandler) UpdatePrinter(c *gin.Context) {
	printerID := c.Param("id")
	if printerID == "" {
		BadRequestResponse(c, "打印机 ID 不能为空")
		return
	}

	var req UpdatePrinterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "请求参数无效")
		return
	}

	// 检查打印机是否存在
	printer, err := h.printerRepo.GetPrinterByID(printerID)
	if err != nil {
		NotFoundResponse(c, "打印机不存在")
		return
	}

	// 更新打印机信息
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

	if err := h.printerRepo.CreatePrinter(printer); err != nil {
		log.Printf("Failed to register printer by edge node %s: %v", edgeNodeID, err)
		InternalErrorResponse(c, "注册打印机失败")
		return
	}

	log.Printf("Printer %s registered by edge node %s", printer.Name, edgeNodeID)
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
