package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Response 通用API响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

// CreatedResponse 创建成功响应
func CreatedResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    http.StatusCreated,
		Message: "created successfully",
		Data:    data,
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequestResponse 请求错误响应
func BadRequestResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusBadRequest, message)
}

// ValidationErrorResponse 字段验证错误响应
func ValidationErrorResponse(c *gin.Context, err error) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errors []string
		for _, fieldError := range validationErrors {
			errors = append(errors, getFieldErrorMessage(fieldError))
		}
		ErrorResponse(c, http.StatusBadRequest, "字段验证失败: "+strings.Join(errors, ", "))
	} else {
		BadRequestResponse(c, "请求参数无效")
	}
}

// getFieldErrorMessage 获取字段错误信息
func getFieldErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " 是必填字段"
	case "email":
		return fe.Field() + " 必须是有效的邮箱地址"
	case "min":
		return fe.Field() + " 长度不能少于 " + fe.Param() + " 个字符"
	case "max":
		return fe.Field() + " 长度不能超过 " + fe.Param() + " 个字符"
	default:
		return fe.Field() + " 格式不正确"
	}
}

// UnauthorizedResponse 未授权响应
func UnauthorizedResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, message)
}

// ForbiddenResponse 禁止访问响应
func ForbiddenResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, message)
}

// NotFoundResponse 未找到响应
func NotFoundResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, message)
}

// InternalErrorResponse 内部错误响应
func InternalErrorResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, message)
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    PaginatedData          `json:"data"`
}

// PaginatedData 分页数据
type PaginatedData struct {
	Items      interface{} `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// PaginatedSuccessResponse 分页成功响应
func PaginatedSuccessResponse(c *gin.Context, items interface{}, total, page, pageSize int) {
	totalPages := (total + pageSize - 1) / pageSize
	
	c.JSON(http.StatusOK, PaginatedResponse{
		Code:    http.StatusOK,
		Message: "success",
		Data: PaginatedData{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}