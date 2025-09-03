package database

import (
	"database/sql"
	"fmt"
	"time"

	"fly-print-cloud/api/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	db *DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser 创建用户
func (r *UserRepository) CreateUser(user *models.User) error {
	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO users (username, email, password_hash, role, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err = r.db.QueryRow(query, user.Username, user.Email, string(hashedPassword), user.Role, user.Status).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// 清空密码哈希值，不应该返回
	user.PasswordHash = ""
	return nil
}

// GetUserByID 根据ID获取用户
func (r *UserRepository) GetUserByID(id string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, role, status, last_login, created_at, updated_at
		FROM users WHERE id = $1 AND status = 'active'`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role, &user.Status,
		&user.LastLogin, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByUsername 根据用户名获取用户（包含密码哈希，用于登录验证）
func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password_hash, role, status, last_login, created_at, updated_at
		FROM users WHERE username = $1 AND status = 'active'`

	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role,
		&user.Status, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpdateUser 更新用户信息
func (r *UserRepository) UpdateUser(user *models.User) error {
	query := `
		UPDATE users
		SET username = $2, email = $3, role = $4, status = $5
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRow(query, user.ID, user.Username, user.Email, user.Role, user.Status).
		Scan(&user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdatePassword 更新用户密码
func (r *UserRepository) UpdatePassword(userID, newPassword string) error {
	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `UPDATE users SET password_hash = $2 WHERE id = $1`
	_, err = r.db.Exec(query, userID, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// UpdateLastLogin 更新最后登录时间
func (r *UserRepository) UpdateLastLogin(userID string) error {
	query := `UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// DeleteUser 删除用户（软删除，设置状态为inactive）
func (r *UserRepository) DeleteUser(userID string) error {
	query := `UPDATE users SET status = 'inactive' WHERE id = $1`
	result, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// ListUsers 获取用户列表
func (r *UserRepository) ListUsers(offset, limit int) ([]*models.User, int, error) {
	var users []*models.User
	var total int

	// 获取总数
	countQuery := `SELECT COUNT(*) FROM users WHERE status = 'active'`
	err := r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 获取用户列表
	query := `
		SELECT id, username, email, role, status, last_login, created_at, updated_at
		FROM users 
		WHERE status = 'active'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Role, &user.Status,
			&user.LastLogin, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return users, total, nil
}

// VerifyPassword 验证密码
func (r *UserRepository) VerifyPassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// EmailExists 检查邮箱是否已存在
func (r *UserRepository) EmailExists(email string, excludeUserID ...string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE email = $1 AND status = 'active'`
	args := []interface{}{email}

	if len(excludeUserID) > 0 && excludeUserID[0] != "" {
		query += ` AND id != $2`
		args = append(args, excludeUserID[0])
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}

// UsernameExists 检查用户名是否已存在
func (r *UserRepository) UsernameExists(username string, excludeUserID ...string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE username = $1 AND status = 'active'`
	args := []interface{}{username}

	if len(excludeUserID) > 0 && excludeUserID[0] != "" {
		query += ` AND id != $2`
		args = append(args, excludeUserID[0])
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return count > 0, nil
}

// GetUserByExternalID 通过外部ID获取用户
func (r *UserRepository) GetUserByExternalID(externalID string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, external_id, role, status, last_login, created_at, updated_at
		FROM users 
		WHERE external_id = $1`
	
	var user models.User
	var externalIDPtr sql.NullString
	
	err := r.db.QueryRow(query, externalID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&externalIDPtr,
		&user.Role,
		&user.Status,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	if externalIDPtr.Valid {
		user.ExternalID = &externalIDPtr.String
	}
	
	return &user, nil
}

// CreateUserFromOAuth2 从OAuth2信息创建用户
func (r *UserRepository) CreateUserFromOAuth2(externalID, username, email string) (*models.User, error) {
	query := `
		INSERT INTO users (username, email, external_id, password_hash, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`
	
	user := &models.User{
		Username:     username,
		Email:        email,
		ExternalID:   &externalID,
		PasswordHash: "oauth2_user", // OAuth2 用户的占位符密码哈希
		Role:         "admin",
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	err := r.db.QueryRow(query,
		user.Username,
		user.Email,
		user.ExternalID,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}