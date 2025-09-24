package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"fly-print-cloud/api/internal/models"
)

// ConnectionManager 管理所有 WebSocket 连接
type ConnectionManager struct {
	connections map[string]*Connection // node_id -> connection
	broadcast   chan []byte           // 广播消息通道
	register    chan *Connection      // 新连接注册
	unregister  chan *Connection      // 连接断开
	mutex       sync.RWMutex         // 并发安全
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*Connection),
		broadcast:   make(chan []byte),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
	}
}

// Run 启动连接管理器
func (m *ConnectionManager) Run() {
	for {
		select {
		case conn := <-m.register:
			m.registerConnection(conn)

		case conn := <-m.unregister:
			m.unregisterConnection(conn)

		case message := <-m.broadcast:
			m.broadcastMessage(message)
		}
	}
}

// registerConnection 注册新连接
func (m *ConnectionManager) registerConnection(conn *Connection) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 如果已有连接，先关闭旧连接
	if existingConn, exists := m.connections[conn.NodeID]; exists {
		log.Printf("Replacing existing connection for node %s", conn.NodeID)
		close(existingConn.Send)
	}

	m.connections[conn.NodeID] = conn
	log.Printf("Edge Node %s connected, total connections: %d", conn.NodeID, len(m.connections))
}

// unregisterConnection 注销连接
func (m *ConnectionManager) unregisterConnection(conn *Connection) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.connections[conn.NodeID]; exists {
		delete(m.connections, conn.NodeID)
		// 安全关闭channel，避免重复关闭
		select {
		case <-conn.Send:
			// channel已经关闭
		default:
			close(conn.Send)
		}
		log.Printf("Edge Node %s disconnected, total connections: %d", conn.NodeID, len(m.connections))
	}
}

// broadcastMessage 广播消息到所有连接
func (m *ConnectionManager) broadcastMessage(message []byte) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for nodeID, conn := range m.connections {
		select {
		case conn.Send <- message:
		default:
			log.Printf("Failed to send broadcast message to node %s, closing connection", nodeID)
			close(conn.Send)
			delete(m.connections, nodeID)
		}
	}
}

// SendToNode 发送消息到指定节点
func (m *ConnectionManager) SendToNode(nodeID string, message []byte) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	conn, exists := m.connections[nodeID]
	if !exists {
		return ErrNodeNotConnected
	}

	select {
	case conn.Send <- message:
		return nil
	default:
		return ErrConnectionClosed
	}
}

// GetConnectedNodes 获取已连接的节点列表
func (m *ConnectionManager) GetConnectedNodes() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	nodes := make([]string, 0, len(m.connections))
	for nodeID := range m.connections {
		nodes = append(nodes, nodeID)
	}
	return nodes
}

// IsNodeConnected 检查节点是否在线
func (m *ConnectionManager) IsNodeConnected(nodeID string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, exists := m.connections[nodeID]
	return exists
}

// GetConnectionCount 获取连接数量
func (m *ConnectionManager) GetConnectionCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return len(m.connections)
}


// DispatchPrintJob 分发打印任务到指定Edge Node
func (m *ConnectionManager) DispatchPrintJob(nodeID string, job *models.PrintJob, printerName string) error {
	// 构造打印任务数据
	printJobData := PrintJobData{
		JobID:       job.ID,
		Name:        job.Name,
		PrinterID:   job.PrinterID,
		PrinterName: printerName,
		FilePath:    job.FilePath,
		FileURL:     job.FileURL,
		FileSize:    job.FileSize,
		PageCount:   job.PageCount,
		Copies:      job.Copies,
		PaperSize:   job.PaperSize,
		ColorMode:   job.ColorMode,
		DuplexMode:  job.DuplexMode,
		MaxRetries:  job.MaxRetries,
	}

	// 构造指令消息
	command := Command{
		Type:      CmdTypePrintJob,
		CommandID: job.ID, // 使用job ID作为command ID
		Timestamp: time.Now(),
		Target:    nodeID,
		Data:      printJobData,
	}

	// 序列化消息
	message, err := json.Marshal(command)
	if err != nil {
		return err
	}

	// 发送到指定节点
	return m.SendToNode(nodeID, message)
}
