package database

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"os"

	"fly-print-cloud/api/internal/config"
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
	createDefault := os.Getenv("CREATE_DEFAULT_ADMIN")
	if createDefault != "true" {
		log.Println("No admin users found, but CREATE_DEFAULT_ADMIN is not set to 'true'")
		log.Println("To create a default admin, set CREATE_DEFAULT_ADMIN=true and restart")
		return nil
	}

	// 从环境变量获取管理员密码，如果没有则使用随机密码
	adminPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
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
	if os.Getenv("DEFAULT_ADMIN_PASSWORD") == "" {
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