package middleware

import (
	"log"
	"strings"

	"fly-print-cloud/api/internal/auth"
	"fly-print-cloud/api/internal/database"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware(authService *auth.Service, userRepo *database.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从多个地方获取令牌
		token := getTokenFromRequest(c)
		if token == "" {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "缺少访问令牌",
			})
			c.Abort()
			return
		}

		// 验证令牌
		claims, err := authService.ValidateToken(token)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			c.JSON(401, gin.H{
				"code":    401,
				"message": "无效的访问令牌",
			})
			c.Abort()
			return
		}

		// 验证用户是否仍然存在且状态为活跃
		user, err := userRepo.GetUserByID(claims.UserID)
		if err != nil {
			log.Printf("User not found for token: %s", claims.UserID)
			c.JSON(401, gin.H{
				"code":    401,
				"message": "用户不存在",
			})
			c.Abort()
			return
		}

		if user.Status != "active" {
			log.Printf("Inactive user attempted access: %s", user.Username)
			c.JSON(401, gin.H{
				"code":    401,
				"message": "账户已被禁用",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("user", user)

		c.Next()
	}
}

// RequireRole 角色权限中间件
func RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "未找到用户角色信息",
			})
			c.Abort()
			return
		}

		role := userRole.(string)
		
		// 检查用户角色是否在允许的角色列表中
		hasPermission := false
		for _, requiredRole := range requiredRoles {
			if role == requiredRole {
				hasPermission = true
				break
			}
		}

		// admin角色拥有所有权限
		if role == "admin" {
			hasPermission = true
		}

		if !hasPermission {
			log.Printf("User %s with role %s attempted to access resource requiring roles: %v", 
				c.GetString("username"), role, requiredRoles)
			c.JSON(403, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth 可选认证中间件（不强制要求认证）
func OptionalAuth(authService *auth.Service, userRepo *database.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := getTokenFromRequest(c)
		if token == "" {
			// 没有令牌时继续执行，但不设置用户信息
			c.Next()
			return
		}

		// 尝试验证令牌
		claims, err := authService.ValidateToken(token)
		if err != nil {
			// 令牌无效时也继续执行，但不设置用户信息
			c.Next()
			return
		}

		// 验证用户
		user, err := userRepo.GetUserByID(claims.UserID)
		if err != nil || user.Status != "active" {
			// 用户不存在或已禁用时继续执行，但不设置用户信息
			c.Next()
			return
		}

		// 设置用户信息到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("user", user)

		c.Next()
	}
}

// getTokenFromRequest 从请求中获取令牌
func getTokenFromRequest(c *gin.Context) string {
	// 1. 尝试从Authorization Header获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// 支持 "Bearer <token>" 格式
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
		// 也支持直接传递令牌
		return authHeader
	}

	// 2. 尝试从Cookie获取
	if token, err := c.Cookie("auth_token"); err == nil && token != "" {
		return token
	}

	// 3. 尝试从Query参数获取（不推荐，但有时需要）
	if token := c.Query("token"); token != "" {
		return token
	}

	return ""
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		log.Printf("[%s] %s %s %d %s %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
		)
		return ""
	})
}