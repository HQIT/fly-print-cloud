package handlers

import (
	"log"
	"strconv"

	"fly-print-cloud/api/internal/database"
	"fly-print-cloud/api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler 用户管理处理器
type UserHandler struct {
	userRepo *database.UserRepository
}

// NewUserHandler 创建用户管理处理器
func NewUserHandler(userRepo *database.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=admin operator viewer"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role" binding:"required,oneof=admin operator viewer"`
	Status   string `json:"status" binding:"required,oneof=active inactive"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ListUsers 获取用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// 查询用户列表
	users, total, err := h.userRepo.ListUsers(offset, pageSize)
	if err != nil {
		log.Printf("Failed to list users: %v", err)
		InternalErrorResponse(c, "获取用户列表失败")
		return
	}

	// 转换为响应格式
	userInfos := make([]UserInfo, len(users))
	for i, user := range users {
		userInfos[i] = UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		}
	}

	PaginatedSuccessResponse(c, userInfos, total, page, pageSize)
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "请求参数无效")
		return
	}

	// 检查用户名是否已存在
	exists, err := h.userRepo.UsernameExists(req.Username)
	if err != nil {
		log.Printf("Failed to check username existence: %v", err)
		InternalErrorResponse(c, "检查用户名失败")
		return
	}
	if exists {
		BadRequestResponse(c, "用户名已存在")
		return
	}

	// 检查邮箱是否已存在
	exists, err = h.userRepo.EmailExists(req.Email)
	if err != nil {
		log.Printf("Failed to check email existence: %v", err)
		InternalErrorResponse(c, "检查邮箱失败")
		return
	}
	if exists {
		BadRequestResponse(c, "邮箱已存在")
		return
	}

	// 创建用户
	user := &models.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.Password, // 在repository中会被加密
		Role:         req.Role,
		Status:       "active",
	}

	if err := h.userRepo.CreateUser(user); err != nil {
		log.Printf("Failed to create user: %v", err)
		InternalErrorResponse(c, "创建用户失败")
		return
	}

	// 返回用户信息（不包含密码）
	userInfo := UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	log.Printf("User %s created successfully by admin", user.Username)
	CreatedResponse(c, userInfo)
}

// GetUser 获取用户详情
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		BadRequestResponse(c, "用户ID不能为空")
		return
	}

	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		NotFoundResponse(c, "用户不存在")
		return
	}

	userInfo := UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	SuccessResponse(c, userInfo)
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		BadRequestResponse(c, "用户ID不能为空")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "请求参数无效")
		return
	}

	// 检查用户是否存在
	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		NotFoundResponse(c, "用户不存在")
		return
	}

	// 检查用户名是否已被其他用户使用
	exists, err := h.userRepo.UsernameExists(req.Username, userID)
	if err != nil {
		log.Printf("Failed to check username existence: %v", err)
		InternalErrorResponse(c, "检查用户名失败")
		return
	}
	if exists {
		BadRequestResponse(c, "用户名已被其他用户使用")
		return
	}

	// 检查邮箱是否已被其他用户使用
	exists, err = h.userRepo.EmailExists(req.Email, userID)
	if err != nil {
		log.Printf("Failed to check email existence: %v", err)
		InternalErrorResponse(c, "检查邮箱失败")
		return
	}
	if exists {
		BadRequestResponse(c, "邮箱已被其他用户使用")
		return
	}

	// 更新用户信息
	user.Username = req.Username
	user.Email = req.Email
	user.Role = req.Role
	user.Status = req.Status

	if err := h.userRepo.UpdateUser(user); err != nil {
		log.Printf("Failed to update user %s: %v", userID, err)
		InternalErrorResponse(c, "更新用户失败")
		return
	}

	userInfo := UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	log.Printf("User %s updated successfully", user.Username)
	SuccessResponse(c, userInfo)
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		BadRequestResponse(c, "用户ID不能为空")
		return
	}

	// 获取当前操作用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "未找到当前用户信息")
		return
	}

	// 不能删除自己
	if userID == currentUserID.(string) {
		BadRequestResponse(c, "不能删除自己")
		return
	}

	// 检查用户是否存在
	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		NotFoundResponse(c, "用户不存在")
		return
	}

	// 删除用户（软删除）
	if err := h.userRepo.DeleteUser(userID); err != nil {
		log.Printf("Failed to delete user %s: %v", userID, err)
		InternalErrorResponse(c, "删除用户失败")
		return
	}

	log.Printf("User %s deleted successfully", user.Username)
	SuccessResponse(c, gin.H{"message": "用户删除成功"})
}

// ChangePassword 修改用户密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		BadRequestResponse(c, "用户ID不能为空")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "请求参数无效")
		return
	}

	// 检查用户是否存在
	_, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		NotFoundResponse(c, "用户不存在")
		return
	}

	// 更新密码
	if err := h.userRepo.UpdatePassword(userID, req.NewPassword); err != nil {
		log.Printf("Failed to change password for user %s: %v", userID, err)
		InternalErrorResponse(c, "修改密码失败")
		return
	}

	log.Printf("Password changed successfully for user %s", userID)
	SuccessResponse(c, gin.H{"message": "密码修改成功"})
}