package handlers

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"fly-print-cloud/api/internal/database"
	"fly-print-cloud/api/internal/models"
	"fly-print-cloud/api/internal/websocket"
)

type PrintJobHandler struct {
	printJobRepo *database.PrintJobRepository
	printerRepo  *database.PrinterRepository
	wsManager    *websocket.ConnectionManager
}

func NewPrintJobHandler(printJobRepo *database.PrintJobRepository, printerRepo *database.PrinterRepository, wsManager *websocket.ConnectionManager) *PrintJobHandler {
	return &PrintJobHandler{
		printJobRepo: printJobRepo,
		printerRepo:  printerRepo,
		wsManager:    wsManager,
	}
}

// CreatePrintJobRequest 创建打印任务请求
type CreatePrintJobRequest struct {
	Name         string `json:"name"`                         // 可选，不提供时自动生成
	PrinterID    string `json:"printer_id" binding:"required"`
	FilePath     string `json:"file_path"`                    // 本地文件路径
	FileURL      string `json:"file_url"`                     // 文件URL
	FileSize     int64  `json:"file_size"`                    // 可选
	PageCount    int    `json:"page_count"`                   // 可选
	Copies       int    `json:"copies" binding:"omitempty,min=1"` // 可选，默认1
	PaperSize    string `json:"paper_size"`
	ColorMode    string `json:"color_mode"`
	DuplexMode   string `json:"duplex_mode"`
	MaxRetries   int    `json:"max_retries"`                  // 可选，默认3
}

// UpdatePrintJobRequest 更新打印任务请求
type UpdatePrintJobRequest struct {
	Name         *string `json:"name,omitempty"`
	Status       *string `json:"status,omitempty"`
	FilePath     *string `json:"file_path,omitempty"`
	FileSize     *int64  `json:"file_size,omitempty"`
	PageCount    *int    `json:"page_count,omitempty"`
	Copies       *int    `json:"copies,omitempty"`
	PaperSize    *string `json:"paper_size,omitempty"`
	ColorMode    *string `json:"color_mode,omitempty"`
	DuplexMode   *string `json:"duplex_mode,omitempty"`
	ErrorMessage *string `json:"error_message,omitempty"`
	RetryCount   *int    `json:"retry_count,omitempty"`
	MaxRetries   *int    `json:"max_retries,omitempty"`
}

// CreatePrintJob 创建打印任务
func (h *PrintJobHandler) CreatePrintJob(c *gin.Context) {
	var req CreatePrintJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 验证文件路径或URL至少有一个
	if req.FilePath == "" && req.FileURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "必须提供file_path或file_url"})
		return
	}

	// 从OAuth2认证中获取用户信息
	userID, exists := c.Get("external_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	userName, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 自动生成任务名称
	jobName := req.Name
	if jobName == "" {
		if req.FileURL != "" {
			// 从URL提取文件名
			parts := strings.Split(req.FileURL, "/")
			filename := parts[len(parts)-1]
			if filename != "" {
				jobName = filename
			} else {
				jobName = fmt.Sprintf("打印任务_%s", time.Now().Format("20060102_150405"))
			}
		} else if req.FilePath != "" {
			// 从文件路径提取文件名
			jobName = filepath.Base(req.FilePath)
		} else {
			jobName = fmt.Sprintf("打印任务_%s", time.Now().Format("20060102_150405"))
		}
	}

	job := &models.PrintJob{
		Name:         jobName,
		Status:       "pending",
		PrinterID:    req.PrinterID,
		UserID:       userID.(string),
		UserName:     userName.(string),
		FilePath:     req.FilePath,
		FileURL:      req.FileURL,
		FileSize:     req.FileSize,
		PageCount:    req.PageCount,
		Copies:       req.Copies,
		PaperSize:    req.PaperSize,
		ColorMode:    req.ColorMode,
		DuplexMode:   req.DuplexMode,
		RetryCount:   0,
		MaxRetries:   req.MaxRetries,
	}

	// 设置默认值
	if job.Copies == 0 {
		job.Copies = 1
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = 3
	}

	err := h.printJobRepo.CreatePrintJob(job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建打印任务失败"})
		return
	}

	// 获取打印机信息以确定Edge Node
	printer, err := h.printerRepo.GetPrinterByID(job.PrinterID)
	if err != nil {
		log.Printf("Failed to get printer %s: %v", job.PrinterID, err)
		c.JSON(http.StatusCreated, job) // 任务已创建，但分发失败
		return
	}

	if printer == nil {
		log.Printf("Printer %s not found", job.PrinterID)
		c.JSON(http.StatusCreated, job) // 任务已创建，但分发失败
		return
	}

	// 分发任务到Edge Node
	err = h.wsManager.DispatchPrintJob(printer.EdgeNodeID, job)
	if err != nil {
		log.Printf("Failed to dispatch print job %s to node %s: %v", job.ID, printer.EdgeNodeID, err)
		// 任务已创建，但分发失败，保持pending状态
	} else {
		log.Printf("Print job %s dispatched to node %s", job.ID, printer.EdgeNodeID)
		// 更新任务状态为已分发
		job.Status = "dispatched"
		if updateErr := h.printJobRepo.UpdatePrintJob(job); updateErr != nil {
			log.Printf("Failed to update job status to dispatched: %v", updateErr)
		}
	}

	c.JSON(http.StatusCreated, job)
}

// GetPrintJob 获取打印任务详情
func (h *PrintJobHandler) GetPrintJob(c *gin.Context) {
	id := c.Param("id")

	job, err := h.printJobRepo.GetPrintJobByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取打印任务失败"})
		return
	}

	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "打印任务不存在"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListPrintJobs 获取打印任务列表
func (h *PrintJobHandler) ListPrintJobs(c *gin.Context) {
	// 分页参数
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// 过滤参数
	status := c.Query("status")
	printerID := c.Query("printer_id")
	userID := c.Query("user_id")

	jobs, err := h.printJobRepo.ListPrintJobs(limit, offset, status, printerID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取打印任务列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jobs": jobs,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// UpdatePrintJob 更新打印任务
func (h *PrintJobHandler) UpdatePrintJob(c *gin.Context) {
	id := c.Param("id")

	// 获取现有任务
	job, err := h.printJobRepo.GetPrintJobByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取打印任务失败"})
		return
	}

	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "打印任务不存在"})
		return
	}

	var req UpdatePrintJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 更新字段
	if req.Name != nil {
		job.Name = *req.Name
	}
	if req.Status != nil {
		job.Status = *req.Status
		// 状态变更时设置时间
		if *req.Status == "printing" && job.StartTime.IsZero() {
			job.StartTime = time.Now()
		}
		if (*req.Status == "completed" || *req.Status == "failed" || *req.Status == "cancelled") && job.EndTime.IsZero() {
			job.EndTime = time.Now()
		}
	}
	if req.FilePath != nil {
		job.FilePath = *req.FilePath
	}
	if req.FileSize != nil {
		job.FileSize = *req.FileSize
	}
	if req.PageCount != nil {
		job.PageCount = *req.PageCount
	}
	if req.Copies != nil {
		job.Copies = *req.Copies
	}
	if req.PaperSize != nil {
		job.PaperSize = *req.PaperSize
	}
	if req.ColorMode != nil {
		job.ColorMode = *req.ColorMode
	}
	if req.DuplexMode != nil {
		job.DuplexMode = *req.DuplexMode
	}
	if req.ErrorMessage != nil {
		job.ErrorMessage = *req.ErrorMessage
	}
	if req.RetryCount != nil {
		job.RetryCount = *req.RetryCount
	}
	if req.MaxRetries != nil {
		job.MaxRetries = *req.MaxRetries
	}

	err = h.printJobRepo.UpdatePrintJob(job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新打印任务失败"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// DeletePrintJob 删除打印任务
func (h *PrintJobHandler) DeletePrintJob(c *gin.Context) {
	id := c.Param("id")

	// 检查任务是否存在
	job, err := h.printJobRepo.GetPrintJobByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取打印任务失败"})
		return
	}

	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "打印任务不存在"})
		return
	}

	err = h.printJobRepo.DeletePrintJob(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除打印任务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "打印任务删除成功"})
}

// CancelPrintJob 取消打印任务
func (h *PrintJobHandler) CancelPrintJob(c *gin.Context) {
	id := c.Param("id")

	job, err := h.printJobRepo.GetPrintJobByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取打印任务失败"})
		return
	}

	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "打印任务不存在"})
		return
	}

	// 只有pending和printing状态的任务可以取消
	if job.Status != "pending" && job.Status != "printing" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务状态不允许取消"})
		return
	}

	job.Status = "cancelled"
	if job.EndTime.IsZero() {
		job.EndTime = time.Now()
	}

	err = h.printJobRepo.UpdatePrintJob(job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消打印任务失败"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// RetryPrintJob 重试打印任务
func (h *PrintJobHandler) RetryPrintJob(c *gin.Context) {
	id := c.Param("id")

	job, err := h.printJobRepo.GetPrintJobByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取打印任务失败"})
		return
	}

	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "打印任务不存在"})
		return
	}

	// 只有failed状态的任务可以重试
	if job.Status != "failed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务状态不允许重试"})
		return
	}

	// 检查重试次数
	if job.RetryCount >= job.MaxRetries {
		c.JSON(http.StatusBadRequest, gin.H{"error": "已达到最大重试次数"})
		return
	}

	job.Status = "pending"
	job.RetryCount++
	job.ErrorMessage = ""
	job.StartTime = time.Time{}
	job.EndTime = time.Time{}

	err = h.printJobRepo.UpdatePrintJob(job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重试打印任务失败"})
		return
	}

	c.JSON(http.StatusOK, job)
}
