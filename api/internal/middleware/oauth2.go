package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"fly-print-cloud/api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// OAuth2TokenInfo OAuth2 token 信息
type OAuth2TokenInfo struct {
	Sub               string   `json:"sub"`
	PreferredUsername string   `json:"preferred_username"`
	Email             string   `json:"email"`
	Groups            []string `json:"groups,omitempty"`           // OIDC 标准 groups claim
	Roles             []string `json:"roles,omitempty"`            // 常见 roles claim
	Scope             string   `json:"scope,omitempty"`            // OAuth2 标准 scope
	RealmAccess       struct {
		Roles []string `json:"roles"`
	} `json:"realm_access,omitempty"`                              // Keycloak realm roles
	ResourceAccess    map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access,omitempty"`                           // Keycloak client roles
}

// OAuth2ResourceServer OAuth2 资源服务器中间件（AND逻辑）
// 验证 Bearer token 和 scope 权限，需要拥有所有指定权限
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

		// 验证 token 有效性
		tokenInfo, err := validateOAuth2Token(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "invalid_token",
				"error_description": err.Error(),
			})
			c.Abort()
			return
		}

		// 提取标准化角色
		userRoles := extractStandardRoles(tokenInfo)
		
		// 验证权限
		if !validateScopes(userRoles, requiredScopes) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":             "insufficient_scope",
				"error_description": "token does not have required scopes",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到 context 中
		c.Set("oauth2_token", token)
		c.Set("external_id", tokenInfo.Sub)
		c.Set("username", tokenInfo.PreferredUsername)
		c.Set("email", tokenInfo.Email)
		c.Set("roles", userRoles)
		
		c.Next()
	}
}

// ValidateOAuth2Token 验证 OAuth2 token（导出方法）
func ValidateOAuth2Token(token string) (*OAuth2TokenInfo, error) {
	return validateOAuth2Token(token)
}

// validateOAuth2Token 验证 OAuth2 token 有效性（内部方法）
func validateOAuth2Token(token string) (*OAuth2TokenInfo, error) {
	// 首先尝试解析 JWT token（用于 Client Credentials Flow）
	if tokenInfo, err := parseJWTToken(token); err == nil {
		return tokenInfo, nil
	}
	
	// 如果 JWT 解析失败，回退到 UserInfo 端点验证（用于 Authorization Code Flow）
	return validateTokenViaUserInfo(token)
}

// parseJWTToken 解析 JWT token（不验证签名，仅提取 claims）
func parseJWTToken(tokenString string) (*OAuth2TokenInfo, error) {
	// 解析 JWT token（跳过签名验证）
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT claims")
	}

	tokenInfo := &OAuth2TokenInfo{}
	
	// 提取标准 claims
	if sub, ok := claims["sub"].(string); ok {
		tokenInfo.Sub = sub
	}
	if preferredUsername, ok := claims["preferred_username"].(string); ok {
		tokenInfo.PreferredUsername = preferredUsername
	}
	if email, ok := claims["email"].(string); ok {
		tokenInfo.Email = email
	}
	if scope, ok := claims["scope"].(string); ok {
		tokenInfo.Scope = scope
	}

	// 提取 realm_access roles
	if realmAccess, ok := claims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := realmAccess["roles"].([]interface{}); ok {
			for _, role := range roles {
				if roleStr, ok := role.(string); ok {
					tokenInfo.RealmAccess.Roles = append(tokenInfo.RealmAccess.Roles, roleStr)
				}
			}
		}
	}

	// 提取 resource_access roles
	if resourceAccess, ok := claims["resource_access"].(map[string]interface{}); ok {
		tokenInfo.ResourceAccess = make(map[string]struct {
			Roles []string `json:"roles"`
		})
		for client, access := range resourceAccess {
			if accessMap, ok := access.(map[string]interface{}); ok {
				if roles, ok := accessMap["roles"].([]interface{}); ok {
					var roleStrings []string
					for _, role := range roles {
						if roleStr, ok := role.(string); ok {
							roleStrings = append(roleStrings, roleStr)
						}
					}
					tokenInfo.ResourceAccess[client] = struct {
						Roles []string `json:"roles"`
					}{Roles: roleStrings}
				}
			}
		}
	}

	return tokenInfo, nil
}

// validateTokenViaUserInfo 通过 UserInfo 端点验证 token
func validateTokenViaUserInfo(token string) (*OAuth2TokenInfo, error) {
	// 从配置中获取 UserInfo URL
	userInfoURL := config.GetOAuth2UserInfoURL()
	if userInfoURL == "" {
		return nil, fmt.Errorf("OAuth2 UserInfo URL not configured")
	}

	// 创建请求
	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid token: userinfo returned %d", resp.StatusCode)
	}

	// 解析响应
	var tokenInfo OAuth2TokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo: %w", err)
	}

	return &tokenInfo, nil
}

// extractStandardRoles 从多个标准位置提取用户角色
func extractStandardRoles(tokenInfo *OAuth2TokenInfo) []string {
	var allRoles []string
	
	// 1. OIDC 标准 groups claim
	allRoles = append(allRoles, tokenInfo.Groups...)
	
	// 2. 常见 roles claim  
	allRoles = append(allRoles, tokenInfo.Roles...)
	
	// 3. Keycloak realm roles
	allRoles = append(allRoles, tokenInfo.RealmAccess.Roles...)
	
	// 4. Keycloak client roles (从所有客户端)
	for _, clientAccess := range tokenInfo.ResourceAccess {
		allRoles = append(allRoles, clientAccess.Roles...)
	}
	
	// 5. OAuth2 scope 转换为角色
	if tokenInfo.Scope != "" {
		scopeRoles := strings.Fields(tokenInfo.Scope)
		allRoles = append(allRoles, scopeRoles...)
	}
	
	// 去重
	return removeDuplicates(allRoles)
}

// HasRequiredScope 检查是否有必需的 scope（导出方法）
func HasRequiredScope(tokenInfo *OAuth2TokenInfo, requiredScope string) bool {
	// 从 scope 字符串中提取权限列表
	scopes := strings.Fields(tokenInfo.Scope)
	
	// 检查是否包含所需的 scope
	for _, scope := range scopes {
		if scope == requiredScope {
			return true
		}
	}
	
	// 检查 realm roles（某些情况下 scope 可能存储在 roles 中）
	for _, role := range tokenInfo.RealmAccess.Roles {
		if role == requiredScope {
			return true
		}
	}
	
	return false
}

// validateScopes 验证用户角色是否满足权限要求（内部方法）
func validateScopes(userRoles []string, requiredScopes []string) bool {
	// 如果没有要求特定权限，只要有任何角色就允许
	if len(requiredScopes) == 0 {
		return len(userRoles) > 0
	}
	
	// admin 角色拥有所有权限
	if contains(userRoles, "admin") {
		return true
	}
	
	// 检查是否包含所有必需的权限（AND逻辑）
	for _, requiredScope := range requiredScopes {
		if !contains(userRoles, requiredScope) {
			return false
		}
	}
	return true
}


// removeDuplicates 去除重复的角色
func removeDuplicates(roles []string) []string {
	keys := make(map[string]bool)
	var result []string
	
	for _, role := range roles {
		if role != "" && !keys[role] {
			keys[role] = true
			result = append(result, role)
		}
	}
	return result
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
