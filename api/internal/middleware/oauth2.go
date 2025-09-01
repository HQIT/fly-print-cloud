package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// OAuth2ResourceServer OAuth2 资源服务器中间件
// 验证 Bearer token 和 scope 权限
func OAuth2ResourceServer(requiredScopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "unauthorized", 
				"error_description": "missing authorization header",
			})
			c.Abort()
			return
		}

		// 检查是否为 Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "unauthorized",
				"error_description": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		// 提取 token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "unauthorized",
				"error_description": "missing access token",
			})
			c.Abort()
			return
		}

		// 验证 token 和 scope
		if !validateTokenAndScopes(token, requiredScopes) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":             "insufficient_scope",
				"error_description": "token does not have required scopes",
			})
			c.Abort()
			return
		}

		// 将 token 信息存储到 context 中
		c.Set("oauth2_token", token)
		c.Set("oauth2_scopes", extractScopesFromToken(token))
		
		c.Next()
	}
}

// validateTokenAndScopes 验证 token 是否有效且包含所需的 scope
func validateTokenAndScopes(token string, requiredScopes []string) bool {
	// TODO: 这里需要根据实际的 OAuth2 Server 实现具体的验证逻辑
	// 可能的实现方式：
	// 1. 如果是 JWT token，解析并验证签名和 scope
	// 2. 如果需要远程验证，调用 OAuth2 Server 的 introspection 端点
	// 3. 如果有本地缓存，检查缓存中的 token 信息
	
	// 暂时的简单实现：检查 token 格式和长度
	if len(token) < 10 {
		return false
	}

	// 提取 token 中的 scope（这里需要根据实际 token 格式实现）
	tokenScopes := extractScopesFromToken(token)
	
	// 检查是否包含所有必需的 scope
	for _, requiredScope := range requiredScopes {
		if !contains(tokenScopes, requiredScope) {
			return false
		}
	}

	return true
}

// extractScopesFromToken 从 token 中提取 scope 信息
func extractScopesFromToken(token string) []string {
	// TODO: 根据实际的 token 格式实现
	// 如果是 JWT，解析 payload 中的 scope 字段
	// 如果是 opaque token，可能需要调用 introspection 端点
	
	// 暂时的简单实现：假设 token 包含基本的 edge 权限
	// 实际部署时需要替换为真实的实现
	return []string{"edge:register", "edge:heartbeat"}
}

// contains 检查 slice 中是否包含指定的字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
