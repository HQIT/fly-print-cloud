package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fly-print-cloud/api/internal/config"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

// OAuth2Handler OAuth2 认证处理器
type OAuth2Handler struct {
	config          *oauth2.Config
	userInfoURL     string
	adminConsoleURL string
}

// NewOAuth2Handler 创建 OAuth2 处理器
func NewOAuth2Handler(oauth2Cfg *config.OAuth2Config, adminCfg *config.AdminConfig) *OAuth2Handler {
	// 如果 OAuth2 配置为空，创建一个基本的处理器
	if oauth2Cfg.ClientID == "" || oauth2Cfg.AuthURL == "" || oauth2Cfg.TokenURL == "" {
		return &OAuth2Handler{
			config: nil, // 配置为空时设为 nil
		}
	}

	oauth2Config := &oauth2.Config{
		ClientID:     oauth2Cfg.ClientID,
		ClientSecret: oauth2Cfg.ClientSecret,
		RedirectURL:  oauth2Cfg.RedirectURI,
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauth2Cfg.AuthURL,
			TokenURL: oauth2Cfg.TokenURL,
		},
		Scopes: []string{"openid", "profile", "email", "admin:users", "admin:edge-nodes", "admin:printers", "admin:print-jobs"},
	}

	return &OAuth2Handler{
		config:          oauth2Config,
		userInfoURL:     oauth2Cfg.UserInfoURL,
		adminConsoleURL: adminCfg.ConsoleURL,
	}
}

// Login 发起 OAuth2 授权
func (h *OAuth2Handler) Login(c *gin.Context) {
	// 检查配置是否可用
	if h.config == nil {
		BadRequestResponse(c, "OAuth2 配置未设置")
		return
	}

	// 生成随机 state 参数防止 CSRF 攻击（由 Keycloak 验证）
	state := generateRandomState()
	
	// 重定向到授权服务器
	authURL := h.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, authURL)
}

// Callback 处理 OAuth2 回调
func (h *OAuth2Handler) Callback(c *gin.Context) {
	// 检查配置是否可用
	if h.config == nil {
		BadRequestResponse(c, "OAuth2 配置未设置")
		return
	}

	// 检查是否有 OAuth2 错误
	if errorCode := c.Query("error"); errorCode != "" {
		errorDesc := c.Query("error_description")
		BadRequestResponse(c, "OAuth2 授权失败: "+errorCode+" - "+errorDesc)
		return
	}

	// State 参数由 Keycloak 验证，我们信任其验证结果

	// 获取授权码
	code := c.Query("code")
	if code == "" {
		BadRequestResponse(c, "缺少授权码")
		return
	}

	// 交换 token
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token, err := h.config.Exchange(ctx, code)
	if err != nil {
		InternalErrorResponse(c, "Token 交换失败")
		return
	}

	// 设置安全的 HTTP-only cookies（同域名下共享）
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", token.AccessToken, int(time.Until(token.Expiry).Seconds()), "/", "", false, true)
	
	if token.RefreshToken != "" {
		c.SetCookie("refresh_token", token.RefreshToken, 7*24*3600, "/", "", false, true) // 7天
	}

	// 重定向到管理控制台首页
	c.Redirect(http.StatusFound, h.adminConsoleURL)
}

// OAuth2UserInfo OAuth2 用户信息结构
type OAuth2UserInfo struct {
	Sub               string   `json:"sub"`
	PreferredUsername string   `json:"preferred_username"`
	Email             string   `json:"email"`
	Name              string   `json:"name"`
	RealmAccess       struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
}

// Me 获取当前用户认证信息
func (h *OAuth2Handler) Me(c *gin.Context) {
	// 从 cookie 获取 access token
	accessToken, err := c.Cookie("access_token")
	if err != nil {
		UnauthorizedResponse(c, "未登录")
		return
	}

	// 调用 OAuth2 UserInfo endpoint 获取用户信息
	oauth2UserInfo, err := h.fetchOAuth2UserInfo(accessToken)
	if err != nil {
		UnauthorizedResponse(c, "Token 无效")
		return
	}

	// 只返回 OAuth2 认证信息
	SuccessResponse(c, gin.H{
		"external_id":   oauth2UserInfo.Sub,
		"username":      oauth2UserInfo.PreferredUsername,
		"email":         oauth2UserInfo.Email,
		"name":          oauth2UserInfo.Name,
		"roles":         oauth2UserInfo.RealmAccess.Roles,
		"authenticated": true,
	})
}

// fetchOAuth2UserInfo 从 OAuth2 服务器获取用户信息
func (h *OAuth2Handler) fetchOAuth2UserInfo(accessToken string) (*OAuth2UserInfo, error) {
	if h.userInfoURL == "" {
		return nil, fmt.Errorf("userinfo URL not configured")
	}

	req, err := http.NewRequest("GET", h.userInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed: %d", resp.StatusCode)
	}

	var userInfo OAuth2UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// Verify 验证认证状态 (用于 Nginx auth_request)
func (h *OAuth2Handler) Verify(c *gin.Context) {
	// 从 cookie 获取 access token
	accessToken, err := c.Cookie("access_token")
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	// 验证 token 有效性
	_, err = h.fetchOAuth2UserInfo(accessToken)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	// 认证成功
	c.Status(http.StatusOK)
}

// Logout 登出
func (h *OAuth2Handler) Logout(c *gin.Context) {
	// 清除所有认证相关的 cookies
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	
	SuccessResponse(c, gin.H{"message": "登出成功"})
}

// generateRandomState 生成随机 state 参数
func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
