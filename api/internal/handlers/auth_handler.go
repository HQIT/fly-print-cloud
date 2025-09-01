package handlers

import (
	"log"

	"fly-print-cloud/api/internal/auth"
	"fly-print-cloud/api/internal/database"
	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userRepo   *database.UserRepository
	authService *auth.Service
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(userRepo *database.UserRepository, authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		authService: authService,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	User         UserInfo  `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "请求参数无效")
		return
	}

	// 获取用户
	user, err := h.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		log.Printf("Login failed - user not found: %s, error: %v", req.Username, err)
		UnauthorizedResponse(c, "用户名或密码错误")
		return
	}

	// 验证密码
	if !h.userRepo.VerifyPassword(user, req.Password) {
		log.Printf("Login failed - invalid password for user: %s", req.Username)
		UnauthorizedResponse(c, "用户名或密码错误")
		return
	}

	// 检查用户状态
	if user.Status != "active" {
		log.Printf("Login failed - user inactive: %s", req.Username)
		UnauthorizedResponse(c, "账户已被禁用")
		return
	}

	// 生成令牌
	accessToken, err := h.authService.GenerateToken(user)
	if err != nil {
		log.Printf("Failed to generate access token for user %s: %v", user.Username, err)
		InternalErrorResponse(c, "生成访问令牌失败")
		return
	}

	refreshToken, err := h.authService.GenerateRefreshToken(user)
	if err != nil {
		log.Printf("Failed to generate refresh token for user %s: %v", user.Username, err)
		InternalErrorResponse(c, "生成刷新令牌失败")
		return
	}

	// 更新最后登录时间
	if err := h.userRepo.UpdateLastLogin(user.ID); err != nil {
		log.Printf("Failed to update last login for user %s: %v", user.Username, err)
		// 不影响登录流程，只记录日志
	}

	// 设置Cookie
	c.SetCookie("auth_token", accessToken, 24*3600, "/", "", false, true) // HttpOnly
	c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/", "", false, true)

	// 返回响应
	response := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}

	log.Printf("User %s logged in successfully", user.Username)
	SuccessResponse(c, response)
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 清除Cookie
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	SuccessResponse(c, gin.H{"message": "登出成功"})
}

// RefreshRequest 刷新令牌请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh 刷新访问令牌
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	
	// 首先尝试从请求体获取刷新令牌
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果请求体中没有，尝试从Cookie获取
		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			BadRequestResponse(c, "缺少刷新令牌")
			return
		}
		req.RefreshToken = refreshToken
	}

	// 验证刷新令牌并提取用户信息
	claims, err := h.authService.ValidateToken(req.RefreshToken)
	if err != nil {
		UnauthorizedResponse(c, "无效的刷新令牌")
		return
	}

	// 获取用户信息
	user, err := h.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		UnauthorizedResponse(c, "用户不存在")
		return
	}

	// 生成新的令牌
	newAccessToken, newRefreshToken, err := h.authService.RefreshToken(req.RefreshToken, user)
	if err != nil {
		log.Printf("Failed to refresh token for user %s: %v", user.Username, err)
		UnauthorizedResponse(c, "令牌刷新失败")
		return
	}

	// 更新Cookie
	c.SetCookie("auth_token", newAccessToken, 24*3600, "/", "", false, true)
	c.SetCookie("refresh_token", newRefreshToken, 7*24*3600, "/", "", false, true)

	// 返回新令牌
	response := LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}

	SuccessResponse(c, response)
}

// Me 获取当前用户信息
func (h *AuthHandler) Me(c *gin.Context) {
	// 从上下文中获取用户信息（由认证中间件设置）
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "未找到用户信息")
		return
	}

	// 获取用户详细信息
	user, err := h.userRepo.GetUserByID(userID.(string))
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