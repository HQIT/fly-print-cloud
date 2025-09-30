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
			// 去掉查询参数
			if idx := strings.Index(filename, "?"); idx != -1 {
				filename = filename[:idx]
			}
			// 限制长度，避免超过数据库字段限制
			if filename != "" && len(filename) <= 150 {
				jobName = filename
			} else if filename != "" {
				// 如果文件名太长，截取前150个字符
				jobName = filename[:150]
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
		RetryCount:   0,  // 保留字段但不使用
		MaxRetries:   req.MaxRetries,
	}

	// 设置默认值
	if job.Copies == 0 {
		job.Copies = 1
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = 3
	}

	// 获取打印机信息进行能力校验
	printer, err := h.printerRepo.GetPrinterByID(job.PrinterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取打印机信息失败"})
		return
	}

	if printer == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "打印机不存在"})
		return
	}

	// 校验打印机能力
	if err := h.validatePrintJobCapabilities(job, printer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.printJobRepo.CreatePrintJob(job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建打印任务失败"})
		return
	}

	// 打印机信息已在上面获取并校验过

	// 分发任务到Edge Node
	err = h.wsManager.DispatchPrintJob(printer.EdgeNodeID, job, printer.Name)
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
	// 支持两种分页参数格式
	var limit, offset int
	
	// 优先使用 page/pageSize 参数（前端使用）
	pageStr := c.Query("page")
	pageSizeStr := c.Query("pageSize")
	page_sizeStr := c.Query("page_size") // 兼容下划线格式
	
	if pageStr != "" && (pageSizeStr != "" || page_sizeStr != "") {
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		
		pageSize := 20 // 默认值
		if pageSizeStr != "" {
			pageSize, _ = strconv.Atoi(pageSizeStr)
		} else if page_sizeStr != "" {
			pageSize, _ = strconv.Atoi(page_sizeStr)
		}
		
		if pageSize < 1 {
			pageSize = 20
		}
		if pageSize > 100 {
			pageSize = 100 // 限制最大页面大小
		}
		
		limit = pageSize
		offset = (page - 1) * pageSize
	} else {
		// fallback 到 limit/offset 参数
		limitStr := c.DefaultQuery("limit", "20")
		offsetStr := c.DefaultQuery("offset", "0")
		
		limit, _ = strconv.Atoi(limitStr)
		offset, _ = strconv.Atoi(offsetStr)
	}

	// 过滤参数
	status := c.Query("status")
	printerID := c.Query("printer_id")
	userID := c.Query("user_id")

	jobs, total, err := h.printJobRepo.ListPrintJobsWithTotal(limit, offset, status, printerID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取打印任务列表失败"})
		return
	}

	// 计算当前页码
	currentPage := (offset / limit) + 1
	if limit == 0 {
		currentPage = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"jobs": jobs,
		"pagination": gin.H{
			"page":      currentPage,
			"pageSize":  limit,
			"total":     total,
			"limit":     limit,
			"offset":    offset,
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

// ReprintRequest 重新打印请求
type ReprintRequest struct {
	PrinterID  string `json:"printer_id" binding:"required"`
	Copies     int    `json:"copies" binding:"omitempty,min=1,max=99"`
	PaperSize  string `json:"paper_size"`
	ColorMode  string `json:"color_mode"`
	DuplexMode string `json:"duplex_mode"`
}

// ReprintJob 重新打印任务（基于原任务创建新任务）
func (h *PrintJobHandler) ReprintJob(c *gin.Context) {
	id := c.Param("id")

	// 获取原任务
	originalJob, err := h.printJobRepo.GetPrintJobByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取原任务失败"})
		return
	}

	if originalJob == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "原任务不存在"})
		return
	}

	// 只有已完成、失败、取消的任务可以重新打印
	if originalJob.Status != "completed" && originalJob.Status != "failed" && originalJob.Status != "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务状态不允许重新打印"})
		return
	}

	// 解析重新打印参数
	var req ReprintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
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

	// 创建新任务（基于原任务和新参数）
	newJob := &models.PrintJob{
		Name:         fmt.Sprintf("重打-%s", originalJob.Name),
		Status:       "pending",
		PrinterID:    req.PrinterID,  // 使用请求中的打印机ID
		UserID:       userID.(string),
		UserName:     userName.(string),
		FilePath:     originalJob.FilePath,  // 文件信息保持不变
		FileURL:      originalJob.FileURL,
		FileSize:     originalJob.FileSize,
		PageCount:    originalJob.PageCount,
		Copies:       req.Copies,     // 使用请求中的份数
		PaperSize:    req.PaperSize,  // 使用请求中的纸张大小
		ColorMode:    req.ColorMode,  // 使用请求中的颜色模式
		DuplexMode:   req.DuplexMode, // 使用请求中的双面模式
		RetryCount:   0,  // 新任务重置为0
		MaxRetries:   3,  // 新任务使用默认值
	}

	// 设置默认值
	if newJob.Copies == 0 {
		newJob.Copies = 1
	}

	// 获取打印机信息进行能力校验
	printer, err := h.printerRepo.GetPrinterByID(newJob.PrinterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取打印机信息失败"})
		return
	}

	if printer == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "打印机不存在"})
		return
	}

	// 校验打印机能力
	if err := h.validatePrintJobCapabilities(newJob, printer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.printJobRepo.CreatePrintJob(newJob)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建重新打印任务失败"})
		return
	}

	// 打印机信息已在上面获取并校验过

	// 分发任务到Edge Node
	err = h.wsManager.DispatchPrintJob(printer.EdgeNodeID, newJob, printer.Name)
	if err != nil {
		log.Printf("Failed to dispatch print job %s to node %s: %v", newJob.ID, printer.EdgeNodeID, err)
		// 任务已创建，但分发失败，保持pending状态
	} else {
		log.Printf("Print job %s dispatched to node %s", newJob.ID, printer.EdgeNodeID)
		// 更新任务状态为已分发
		newJob.Status = "dispatched"
		if updateErr := h.printJobRepo.UpdatePrintJob(newJob); updateErr != nil {
			log.Printf("Failed to update job status to dispatched: %v", updateErr)
		}
	}

	c.JSON(http.StatusCreated, newJob)
}

// validatePrintJobCapabilities 校验打印任务参数是否符合打印机能力
func (h *PrintJobHandler) validatePrintJobCapabilities(job *models.PrintJob, printer *models.Printer) error {
	// 校验颜色模式
	if job.ColorMode == "color" && !printer.Capabilities.ColorSupport {
		return fmt.Errorf("打印机 %s 不支持彩色打印", printer.Name)
	}

	// 校验双面模式
	if job.DuplexMode == "duplex" && !printer.Capabilities.DuplexSupport {
		return fmt.Errorf("打印机 %s 不支持双面打印", printer.Name)
	}

	// 校验纸张大小
	if job.PaperSize != "" && len(printer.Capabilities.PaperSizes) > 0 {
		supportedSize := false
		for _, size := range printer.Capabilities.PaperSizes {
			if size == job.PaperSize {
				supportedSize = true
				break
			}
		}
		if !supportedSize {
			return fmt.Errorf("打印机 %s 不支持纸张大小 %s，支持的大小：%v", 
				printer.Name, job.PaperSize, printer.Capabilities.PaperSizes)
		}
	}

	// 校验份数（一般限制）
	if job.Copies <= 0 {
		return fmt.Errorf("打印份数必须大于0")
	}
	if job.Copies > 99 {
		return fmt.Errorf("打印份数不能超过99份")
	}

	return nil
}
