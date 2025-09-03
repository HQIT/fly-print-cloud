package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"fly-print-cloud/api/internal/database"
	"fly-print-cloud/api/internal/models"
)

type PrintJobHandler struct {
	printJobRepo *database.PrintJobRepository
}

func NewPrintJobHandler(printJobRepo *database.PrintJobRepository) *PrintJobHandler {
	return &PrintJobHandler{
		printJobRepo: printJobRepo,
	}
}

// CreatePrintJobRequest 创建打印任务请求
type CreatePrintJobRequest struct {
	Name         string `json:"name" binding:"required"`
	PrinterID    string `json:"printer_id" binding:"required"`
	FilePath     string `json:"file_path" binding:"required"`
	FileSize     int64  `json:"file_size" binding:"required"`
	PageCount    int    `json:"page_count" binding:"required"`
	Copies       int    `json:"copies" binding:"min=1"`
	PaperSize    string `json:"paper_size"`
	ColorMode    string `json:"color_mode"`
	DuplexMode   string `json:"duplex_mode"`
	Priority     int    `json:"priority" binding:"min=1,max=10"`
	MaxRetries   int    `json:"max_retries"`
}

// UpdatePrintJobRequest 更新打印任务请求
type UpdatePrintJobRequest struct {
	Name         *string `json:"name,omitempty"`
	Status       *string `json:"status,omitempty"`
	Priority     *int    `json:"priority,omitempty"`
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

	// 从OAuth2认证中获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	userName, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	job := &models.PrintJob{
		Name:         req.Name,
		Status:       "pending",
		Priority:     req.Priority,
		PrinterID:    req.PrinterID,
		UserID:       userID.(string),
		UserName:     userName.(string),
		FilePath:     req.FilePath,
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
	if job.Priority == 0 {
		job.Priority = 5
	}
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
	if req.Priority != nil {
		job.Priority = *req.Priority
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
