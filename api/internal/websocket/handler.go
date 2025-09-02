package websocket

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: 生产环境需要更严格的 Origin 检查
		return true
	},
}

// WebSocketHandler WebSocket 处理器
type WebSocketHandler struct {
	manager *ConnectionManager
}

// NewWebSocketHandler 创建 WebSocket 处理器
func NewWebSocketHandler(manager *ConnectionManager) *WebSocketHandler {
	return &WebSocketHandler{
		manager: manager,
	}
}

// HandleConnection 处理 WebSocket 连接升级
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	// 从 Authorization header 获取 token 并验证
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
		return
	}

	// TODO: 验证 token 并提取 node_id
	// 临时使用测试逻辑
	nodeID := h.extractNodeIDFromToken(token)
	if nodeID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token or missing node_id"})
		return
	}

	// 升级 HTTP 连接到 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection for node %s: %v", nodeID, err)
		return
	}

	// 创建连接对象
	connection := NewConnection(nodeID, conn, h.manager)

	// 注册连接
	h.manager.register <- connection

	// 启动读写协程
	go connection.WritePump()
	go connection.ReadPump()

	log.Printf("WebSocket connection established for Edge Node: %s", nodeID)
}

// extractNodeIDFromToken 从 token 中提取 node_id
// TODO: 集成真实的 OAuth2 token 验证
func (h *WebSocketHandler) extractNodeIDFromToken(token string) string {
	// 临时测试逻辑：从预定义 token 映射中获取
	testTokens := map[string]string{
		"test-edge-token-with-sufficient-length": "edge-001",
		"test-edge-token-node-002":               "edge-002",
	}

	if nodeID, exists := testTokens[token]; exists {
		return nodeID
	}

	return ""
}
