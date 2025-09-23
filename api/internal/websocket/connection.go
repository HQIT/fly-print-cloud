package websocket

import (
	"encoding/json"
	"log"
	"time"

	"fly-print-cloud/api/internal/database"
	"github.com/gorilla/websocket"
)

const (
	// 写入等待时间
	writeWait = 10 * time.Second

	// Pong 等待时间
	pongWait = 60 * time.Second

	// Ping 发送间隔
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512
)

// Connection 表示单个 WebSocket 连接
type Connection struct {
	NodeID       string
	Conn         *websocket.Conn
	Send         chan []byte
	Manager      *ConnectionManager
	PrinterRepo  *database.PrinterRepository
	EdgeNodeRepo *database.EdgeNodeRepository
}

// NewConnection 创建新连接
func NewConnection(nodeID string, conn *websocket.Conn, manager *ConnectionManager, printerRepo *database.PrinterRepository, edgeNodeRepo *database.EdgeNodeRepository) *Connection {
	return &Connection{
		NodeID:       nodeID,
		Conn:         conn,
		Send:         make(chan []byte, 256),
		Manager:      manager,
		PrinterRepo:  printerRepo,
		EdgeNodeRepo: edgeNodeRepo,
	}
}

// ReadPump 处理从客户端读取消息
func (c *Connection) ReadPump() {
	defer func() {
		c.Manager.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for node %s: %v", c.NodeID, err)
			}
			break
		}

		log.Printf("WebSocket received raw message from node %s: %s", c.NodeID, string(messageBytes))

		// 解析消息
		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Failed to parse message from node %s: %v", c.NodeID, err)
			continue
		}

		log.Printf("WebSocket parsed message from node %s: type=%s", c.NodeID, msg.Type)

		// 处理消息
		c.handleMessage(&msg)
	}
}

// WritePump 处理向客户端发送消息
func (c *Connection) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送队列中的其他消息
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理接收到的消息
func (c *Connection) handleMessage(msg *Message) {
	log.Printf("Received message from node %s: type=%s", c.NodeID, msg.Type)

	switch msg.Type {
	case MsgTypeHeartbeat:
		c.handleHeartbeat(msg)
	case MsgTypePrinterStatus:
		c.handlePrinterStatus(msg)
	case MsgTypeJobUpdate:
		c.handleJobUpdate(msg)
	default:
		log.Printf("Unknown message type: %s from node %s", msg.Type, c.NodeID)
	}
}

// handleHeartbeat 处理心跳消息
func (c *Connection) handleHeartbeat(msg *Message) {
	log.Printf("Processing heartbeat from node %s", c.NodeID)
	
	// 更新 Edge Node 的最后心跳时间和状态
	if err := c.EdgeNodeRepo.UpdateHeartbeat(c.NodeID); err != nil {
		log.Printf("Failed to update heartbeat for node %s: %v", c.NodeID, err)
		return
	}
	
	// 解析心跳数据（可选）
	if msg.Data != nil {
		var heartbeatData HeartbeatData
		dataBytes, err := json.Marshal(msg.Data)
		if err == nil {
			if err := json.Unmarshal(dataBytes, &heartbeatData); err == nil {
				log.Printf("Heartbeat data from node %s: CPU=%.2f%%, Memory=%.2f%%, Disk=%.2f%%", 
					c.NodeID, heartbeatData.SystemInfo.CPUUsage, 
					heartbeatData.SystemInfo.MemoryUsage, heartbeatData.SystemInfo.DiskUsage)
			}
		}
	}
	
	log.Printf("Successfully processed heartbeat from node %s", c.NodeID)
}

// handlePrinterStatus 处理打印机状态消息
func (c *Connection) handlePrinterStatus(msg *Message) {
	log.Printf("Processing printer status update from node %s", c.NodeID)
	
	// 解析打印机状态数据
	var statusData PrinterStatusData
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		log.Printf("Failed to marshal printer status data from node %s: %v", c.NodeID, err)
		return
	}
	
	if err := json.Unmarshal(dataBytes, &statusData); err != nil {
		log.Printf("Failed to parse printer status data from node %s: %v", c.NodeID, err)
		return
	}
	
	log.Printf("Printer status data: printer_id=%s, status=%s, queue_length=%d", 
		statusData.PrinterID, statusData.Status, statusData.QueueLength)
	
	// 使用消息中的node_id而不是连接时的NodeID，因为可能不匹配
	messageNodeID := msg.NodeID
	if messageNodeID == "" {
		messageNodeID = c.NodeID // 如果消息中没有node_id，使用连接时的ID
	}
	
	// 通过名称和边缘节点ID查找打印机
	printer, err := c.PrinterRepo.GetPrinterByNameAndEdgeNode(statusData.PrinterID, messageNodeID)
	if err != nil {
		log.Printf("Printer %s not found for node %s (connection: %s): %v", statusData.PrinterID, messageNodeID, c.NodeID, err)
		return
	}
	
	// 直接使用客户端状态（统一标准）
	printer.Status = statusData.Status
	printer.QueueLength = statusData.QueueLength
	
	if err := c.PrinterRepo.UpdatePrinter(printer); err != nil {
		log.Printf("Failed to update printer %s status: %v", statusData.PrinterID, err)
		return
	}
	
	log.Printf("Successfully updated printer %s status to %s (queue: %d)", 
		statusData.PrinterID, statusData.Status, statusData.QueueLength)
}

// handleJobUpdate 处理任务状态更新
func (c *Connection) handleJobUpdate(msg *Message) {
	// TODO: 更新打印任务状态到数据库
	log.Printf("Job update from node %s", c.NodeID)
}


// SendCommand 发送指令到 Edge Node
func (c *Connection) SendCommand(cmd *Command) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	select {
	case c.Send <- data:
		return nil
	default:
		close(c.Send)
		return err
	}
}
