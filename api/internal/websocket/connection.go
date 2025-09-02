package websocket

import (
	"encoding/json"
	"log"
	"time"

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
	NodeID  string
	Conn    *websocket.Conn
	Send    chan []byte
	Manager *ConnectionManager
}

// NewConnection 创建新连接
func NewConnection(nodeID string, conn *websocket.Conn, manager *ConnectionManager) *Connection {
	return &Connection{
		NodeID:  nodeID,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Manager: manager,
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

		// 解析消息
		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Failed to parse message from node %s: %v", c.NodeID, err)
			continue
		}

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
	// TODO: 更新 Edge Node 状态到数据库
	log.Printf("Heartbeat from node %s", c.NodeID)
}

// handlePrinterStatus 处理打印机状态消息
func (c *Connection) handlePrinterStatus(msg *Message) {
	// TODO: 更新打印机状态到数据库
	log.Printf("Printer status update from node %s", c.NodeID)
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
