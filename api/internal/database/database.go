package database

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"

	"fly-print-cloud/api/internal/config"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
)

// DB 数据库实例
type DB struct {
	*sql.DB
}

// New 创建数据库连接
func New(cfg *config.DatabaseConfig) (*DB, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return &DB{db}, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	return db.DB.Close()
}

// InitTables 初始化数据库表
func (db *DB) InitTables() error {
	// 创建用户表
	userTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		role VARCHAR(20) NOT NULL DEFAULT 'viewer',
		status VARCHAR(20) NOT NULL DEFAULT 'active',
		last_login TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(userTableSQL); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// 创建更新时间触发器
	updateTriggerSQL := `
	CREATE OR REPLACE FUNCTION update_updated_at_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ language 'plpgsql';

	DROP TRIGGER IF EXISTS update_users_updated_at ON users;
	CREATE TRIGGER update_users_updated_at
		BEFORE UPDATE ON users
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column();`

	if _, err := db.Exec(updateTriggerSQL); err != nil {
		return fmt.Errorf("failed to create update trigger: %w", err)
	}

	// 创建 Edge Node 表
	edgeNodeTableSQL := `
	CREATE TABLE IF NOT EXISTS edge_nodes (
		id VARCHAR(100) PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'offline',
		version VARCHAR(50),
		last_heartbeat TIMESTAMP,
		deleted_at TIMESTAMP,
		
		-- 位置信息
		location VARCHAR(255),
		latitude DECIMAL(10, 8),
		longitude DECIMAL(11, 8),
		
		-- 网络信息
		ip_address INET,
		mac_address VARCHAR(17),
		network_interface VARCHAR(50),
		
		-- 系统信息
		os_version VARCHAR(100),
		cpu_info TEXT,
		memory_info TEXT,
		disk_info TEXT,
		
		-- 连接信息
		connection_quality VARCHAR(20),
		latency INTEGER,
		
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(edgeNodeTableSQL); err != nil {
		return fmt.Errorf("failed to create edge_nodes table: %w", err)
	}

	// 创建 Edge Node 更新时间触发器
	edgeNodeTriggerSQL := `
	DROP TRIGGER IF EXISTS update_edge_nodes_updated_at ON edge_nodes;
	CREATE TRIGGER update_edge_nodes_updated_at
		BEFORE UPDATE ON edge_nodes
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column();`

	if _, err := db.Exec(edgeNodeTriggerSQL); err != nil {
		return fmt.Errorf("failed to create edge_nodes update trigger: %w", err)
	}

	// 创建打印机表
	printerTableSQL := `
	CREATE TABLE IF NOT EXISTS printers (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(100) NOT NULL,
		model VARCHAR(100),
		serial_number VARCHAR(100),
		status VARCHAR(20) NOT NULL DEFAULT 'offline',
		
		-- 硬件信息
		firmware_version VARCHAR(50),
		port_info VARCHAR(100),
		
		-- 网络信息
		ip_address INET,
		mac_address VARCHAR(17),
		network_config TEXT,
		
		-- 地理位置信息
		latitude DECIMAL(10, 8),
		longitude DECIMAL(11, 8),
		location VARCHAR(255),
		
		-- 能力信息 (JSON 格式)
		capabilities JSONB,
		
		-- 关联信息
		edge_node_id VARCHAR(100) REFERENCES edge_nodes(id) ON DELETE CASCADE,
		queue_length INTEGER DEFAULT 0,
		
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(printerTableSQL); err != nil {
		return fmt.Errorf("failed to create printers table: %w", err)
	}

	// 创建打印机更新时间触发器
	printerTriggerSQL := `
	DROP TRIGGER IF EXISTS update_printers_updated_at ON printers;
	CREATE TRIGGER update_printers_updated_at
		BEFORE UPDATE ON printers
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column();`

	if _, err := db.Exec(printerTriggerSQL); err != nil {
		return fmt.Errorf("failed to create printers update trigger: %w", err)
	}

	// 创建打印任务表
	printJobTableSQL := `
	CREATE TABLE IF NOT EXISTS print_jobs (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(200) NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'pending',
		
		-- 关联信息
		printer_id UUID REFERENCES printers(id) ON DELETE CASCADE,
		user_id UUID REFERENCES users(id) ON DELETE SET NULL,
		user_name VARCHAR(100),
		
		-- 任务信息
		file_path VARCHAR(500),
		file_url VARCHAR(1000),
		file_size BIGINT,
		page_count INTEGER,
		copies INTEGER DEFAULT 1,
		
		-- 打印设置
		paper_size VARCHAR(20),
		color_mode VARCHAR(20),
		duplex_mode VARCHAR(20),
		
		-- 执行信息
		start_time TIMESTAMP,
		end_time TIMESTAMP,
		error_message TEXT,
		
		-- 重试信息
		retry_count INTEGER DEFAULT 0,
		max_retries INTEGER DEFAULT 3,
		
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(printJobTableSQL); err != nil {
		return fmt.Errorf("failed to create print_jobs table: %w", err)
	}

	// 创建打印任务更新时间触发器
	printJobTriggerSQL := `
	DROP TRIGGER IF EXISTS update_print_jobs_updated_at ON print_jobs;
	CREATE TRIGGER update_print_jobs_updated_at
		BEFORE UPDATE ON print_jobs
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column();`

	if _, err := db.Exec(printJobTriggerSQL); err != nil {
		return fmt.Errorf("failed to create print_jobs update trigger: %w", err)
	}

	// 创建索引
	indexesSQL := []string{
		"CREATE INDEX IF NOT EXISTS idx_edge_nodes_status ON edge_nodes(status);",
		"CREATE INDEX IF NOT EXISTS idx_edge_nodes_last_heartbeat ON edge_nodes(last_heartbeat);",
		"CREATE INDEX IF NOT EXISTS idx_printers_edge_node_id ON printers(edge_node_id);",
		"CREATE INDEX IF NOT EXISTS idx_printers_status ON printers(status);",
		"CREATE INDEX IF NOT EXISTS idx_print_jobs_status ON print_jobs(status);",
		"CREATE INDEX IF NOT EXISTS idx_print_jobs_printer_id ON print_jobs(printer_id);",
		"CREATE INDEX IF NOT EXISTS idx_print_jobs_user_id ON print_jobs(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_print_jobs_created_at ON print_jobs(created_at);",
	}

	for _, indexSQL := range indexesSQL {
		if _, err := db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	log.Println("Database tables initialized successfully")
	return nil
}

// CreateDefaultAdmin 创建默认管理员账户
func (db *DB) CreateDefaultAdmin() error {
	// 检查是否已存在管理员账户
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check admin users: %w", err)
	}

	if count > 0 {
		log.Println("Admin user already exists, skipping creation")
		return nil
	}

	// 只在环境变量允许时创建默认管理员
	createDefault := viper.GetString("create_default_admin")
	if createDefault != "true" {
		log.Println("No admin users found, but CREATE_DEFAULT_ADMIN is not set to 'true'")
		log.Println("To create a default admin, set CREATE_DEFAULT_ADMIN=true and restart")
		return nil
	}

	// 从环境变量获取管理员密码，如果没有则使用随机密码
	adminPassword := viper.GetString("default_admin_password")
	if adminPassword == "" {
		adminPassword = generateRandomPassword(16)
		log.Printf("Generated random admin password: %s", adminPassword)
		log.Println("IMPORTANT: Save this password immediately! It will not be shown again.")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	defaultAdminSQL := `
	INSERT INTO users (username, email, password_hash, role, status)
	VALUES ('admin', 'admin@flyprint.local', $1, 'admin', 'active')`

	if _, err := db.Exec(defaultAdminSQL, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to create default admin: %w", err)
	}

	log.Println("Default admin user created successfully (username: admin)")
	if viper.GetString("default_admin_password") == "" {
		log.Println("=====================================")
		log.Println("🔑 IMPORTANT: ADMIN CREDENTIALS")
		log.Println("=====================================")
		log.Printf("Username: admin")
		log.Printf("Password: %s", adminPassword)
		log.Println("=====================================")
		log.Println("⚠️  SAVE THIS PASSWORD IMMEDIATELY!")
		log.Println("⚠️  This password will NOT be shown again!")
		log.Println("⚠️  Change it after first login for security!")
		log.Println("=====================================")
	} else {
		log.Println("Using custom admin password from environment variable")
	}
	return nil
}

// generateRandomPassword 生成随机密码
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	for i := range password {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[num.Int64()]
	}
	return string(password)
}