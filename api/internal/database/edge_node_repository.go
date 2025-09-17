package database

import (
	"database/sql"
	"fmt"
	"log"

	"fly-print-cloud/api/internal/models"
)

// EdgeNodeRepository Edge Node 数据访问层
type EdgeNodeRepository struct {
	db *DB
}

// NewEdgeNodeRepository 创建 Edge Node 数据访问层
func NewEdgeNodeRepository(db *DB) *EdgeNodeRepository {
	return &EdgeNodeRepository{db: db}
}

// CreateEdgeNode 创建 Edge Node
func (r *EdgeNodeRepository) CreateEdgeNode(node *models.EdgeNode) error {
	query := `
		INSERT INTO edge_nodes (
			id, name, status, version, last_heartbeat,
			location, latitude, longitude,
			ip_address, mac_address, network_interface,
			os_version, cpu_info, memory_info, disk_info,
			connection_quality, latency
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13, $14, $15,
			$16, $17
		)`

	_, err := r.db.Exec(query,
		node.ID, node.Name, node.Status, node.Version, node.LastHeartbeat,
		node.Location, node.Latitude, node.Longitude,
		node.IPAddress, node.MACAddress, node.NetworkInterface,
		node.OSVersion, node.CPUInfo, node.MemoryInfo, node.DiskInfo,
		node.ConnectionQuality, node.Latency,
	)

	if err != nil {
		return fmt.Errorf("failed to create edge node: %w", err)
	}

	return nil
}

// UpsertEdgeNode 创建或更新 Edge Node（用于注册）
func (r *EdgeNodeRepository) UpsertEdgeNode(node *models.EdgeNode) error {
	query := `
		INSERT INTO edge_nodes (
			id, name, status, version, last_heartbeat,
			location, latitude, longitude,
			ip_address, mac_address, network_interface,
			os_version, cpu_info, memory_info, disk_info,
			connection_quality, latency
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13, $14, $15,
			$16, $17
		)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			status = EXCLUDED.status,
			version = EXCLUDED.version,
			last_heartbeat = EXCLUDED.last_heartbeat,
			location = EXCLUDED.location,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			ip_address = EXCLUDED.ip_address,
			mac_address = EXCLUDED.mac_address,
			network_interface = EXCLUDED.network_interface,
			os_version = EXCLUDED.os_version,
			cpu_info = EXCLUDED.cpu_info,
			memory_info = EXCLUDED.memory_info,
			disk_info = EXCLUDED.disk_info,
			connection_quality = EXCLUDED.connection_quality,
			latency = EXCLUDED.latency,
			deleted_at = NULL`

	_, err := r.db.Exec(query,
		node.ID, node.Name, node.Status, node.Version, node.LastHeartbeat,
		node.Location, node.Latitude, node.Longitude,
		node.IPAddress, node.MACAddress, node.NetworkInterface,
		node.OSVersion, node.CPUInfo, node.MemoryInfo, node.DiskInfo,
		node.ConnectionQuality, node.Latency,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert edge node: %w", err)
	}

	return nil
}

// GetEdgeNodeByID 根据ID获取 Edge Node
func (r *EdgeNodeRepository) GetEdgeNodeByID(id string) (*models.EdgeNode, error) {
	node := &models.EdgeNode{}
	query := `
		SELECT id, name, status, version, last_heartbeat,
			   location, latitude, longitude,
			   ip_address, mac_address, network_interface,
			   os_version, cpu_info, memory_info, disk_info,
			   connection_quality, latency,
			   created_at, updated_at, deleted_at
		FROM edge_nodes WHERE id = $1 AND deleted_at IS NULL`

	var lastHeartbeat sql.NullTime
	var latitude, longitude sql.NullFloat64
	var location, ipAddress, macAddress, networkInterface sql.NullString
	var osVersion, cpuInfo, memoryInfo, diskInfo sql.NullString
	var connectionQuality sql.NullString
	var latency sql.NullInt32
	var version sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&node.ID, &node.Name, &node.Status, &version, &lastHeartbeat,
		&location, &latitude, &longitude,
		&ipAddress, &macAddress, &networkInterface,
		&osVersion, &cpuInfo, &memoryInfo, &diskInfo,
		&connectionQuality, &latency,
		&node.CreatedAt, &node.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("edge node not found")
		}
		return nil, fmt.Errorf("failed to get edge node: %w", err)
	}

	// 处理可为空的字段
	if version.Valid {
		node.Version = version.String
	}
	if lastHeartbeat.Valid {
		node.LastHeartbeat = lastHeartbeat.Time
	}
	if location.Valid {
		node.Location = location.String
	}
	if latitude.Valid {
		lat := latitude.Float64
		node.Latitude = &lat
	}
	if longitude.Valid {
		lng := longitude.Float64
		node.Longitude = &lng
	}
	if ipAddress.Valid {
		ip := ipAddress.String
		node.IPAddress = &ip
	}
	if macAddress.Valid {
		node.MACAddress = macAddress.String
	}
	if networkInterface.Valid {
		node.NetworkInterface = networkInterface.String
	}
	if osVersion.Valid {
		node.OSVersion = osVersion.String
	}
	if cpuInfo.Valid {
		node.CPUInfo = cpuInfo.String
	}
	if memoryInfo.Valid {
		node.MemoryInfo = memoryInfo.String
	}
	if diskInfo.Valid {
		node.DiskInfo = diskInfo.String
	}
	if connectionQuality.Valid {
		node.ConnectionQuality = connectionQuality.String
	}
	if latency.Valid {
		node.Latency = int(latency.Int32)
	}
	if deletedAt.Valid {
		node.DeletedAt = &deletedAt.Time
	}

	return node, nil
}

// UpdateEdgeNode 更新 Edge Node
func (r *EdgeNodeRepository) UpdateEdgeNode(node *models.EdgeNode) error {
	query := `
		UPDATE edge_nodes SET
			name = $2, status = $3, version = $4, last_heartbeat = $5,
			location = $6, latitude = $7, longitude = $8,
			ip_address = $9, mac_address = $10, network_interface = $11,
			os_version = $12, cpu_info = $13, memory_info = $14, disk_info = $15,
			connection_quality = $16, latency = $17
		WHERE id = $1`

	_, err := r.db.Exec(query,
		node.ID, node.Name, node.Status, node.Version, node.LastHeartbeat,
		node.Location, node.Latitude, node.Longitude,
		node.IPAddress, node.MACAddress, node.NetworkInterface,
		node.OSVersion, node.CPUInfo, node.MemoryInfo, node.DiskInfo,
		node.ConnectionQuality, node.Latency,
	)

	if err != nil {
		return fmt.Errorf("failed to update edge node: %w", err)
	}

	return nil
}

// DeleteEdgeNode 删除 Edge Node（软删除）
func (r *EdgeNodeRepository) DeleteEdgeNode(id string) error {
	query := `UPDATE edge_nodes SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`
	
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete edge node: %w", err)
	}

	return nil
}

// HardDeleteEdgeNode 硬删除 Edge Node（彻底删除）
func (r *EdgeNodeRepository) HardDeleteEdgeNode(id string) error {
	query := `DELETE FROM edge_nodes WHERE id = $1`
	
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete edge node: %w", err)
	}

	return nil
}

// ListEdgeNodes 获取 Edge Node 列表
func (r *EdgeNodeRepository) ListEdgeNodes(offset, limit int, status string) ([]*models.EdgeNode, int, error) {
	log.Printf("🔍 [DB DEBUG] ListEdgeNodes: offset=%d, limit=%d, status='%s'", offset, limit, status)
	var nodes []*models.EdgeNode
	
	// 构建查询条件
	whereClause := "WHERE deleted_at IS NULL"
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM edge_nodes %s", whereClause)
	log.Printf("📊 [DB DEBUG] Count query: %s, args: %v", countQuery, args)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		log.Printf("❌ [DB DEBUG] Count query failed: %v", err)
		return nil, 0, fmt.Errorf("failed to count edge nodes: %w", err)
	}
	log.Printf("📊 [DB DEBUG] Total count: %d", total)

	// 查询数据
	query := fmt.Sprintf(`
		SELECT id, name, status, version, last_heartbeat,
			   location, latitude, longitude,
			   ip_address, mac_address, network_interface,
			   os_version, cpu_info, memory_info, disk_info,
			   connection_quality, latency,
			   created_at, updated_at, deleted_at
		FROM edge_nodes %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)
	log.Printf("📊 [DB DEBUG] Data query: %s", query)
	log.Printf("📊 [DB DEBUG] Query args: %v", args)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		log.Printf("❌ [DB DEBUG] Data query failed: %v", err)
		return nil, 0, fmt.Errorf("failed to query edge nodes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		node := &models.EdgeNode{}
		var lastHeartbeat sql.NullTime
		var latitude, longitude sql.NullFloat64
		var location, ipAddress, macAddress, networkInterface sql.NullString
		var osVersion, cpuInfo, memoryInfo, diskInfo sql.NullString
		var connectionQuality sql.NullString
		var latency sql.NullInt32
		var version sql.NullString

		var deletedAt sql.NullTime
		err := rows.Scan(
			&node.ID, &node.Name, &node.Status, &version, &lastHeartbeat,
			&location, &latitude, &longitude,
			&ipAddress, &macAddress, &networkInterface,
			&osVersion, &cpuInfo, &memoryInfo, &diskInfo,
			&connectionQuality, &latency,
			&node.CreatedAt, &node.UpdatedAt, &deletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan edge node: %w", err)
		}

		// 处理可为空的字段
		if version.Valid {
			node.Version = version.String
		}
		if lastHeartbeat.Valid {
			node.LastHeartbeat = lastHeartbeat.Time
		}
		if location.Valid {
			node.Location = location.String
		}
		if latitude.Valid {
			lat := latitude.Float64
			node.Latitude = &lat
		}
		if longitude.Valid {
			lng := longitude.Float64
			node.Longitude = &lng
		}
		if ipAddress.Valid {
			ip := ipAddress.String
			node.IPAddress = &ip
		}
		if macAddress.Valid {
			node.MACAddress = macAddress.String
		}
		if networkInterface.Valid {
			node.NetworkInterface = networkInterface.String
		}
		if osVersion.Valid {
			node.OSVersion = osVersion.String
		}
		if cpuInfo.Valid {
			node.CPUInfo = cpuInfo.String
		}
		if memoryInfo.Valid {
			node.MemoryInfo = memoryInfo.String
		}
		if diskInfo.Valid {
			node.DiskInfo = diskInfo.String
		}
		if connectionQuality.Valid {
			node.ConnectionQuality = connectionQuality.String
		}
		if latency.Valid {
			node.Latency = int(latency.Int32)
		}
		if deletedAt.Valid {
			node.DeletedAt = &deletedAt.Time
		}

		nodes = append(nodes, node)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return nodes, total, nil
}

// UpdateHeartbeat 更新心跳时间
func (r *EdgeNodeRepository) UpdateHeartbeat(id string) error {
	query := `UPDATE edge_nodes SET last_heartbeat = CURRENT_TIMESTAMP WHERE id = $1`
	
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	return nil
}

// UpdateStatus 更新状态
func (r *EdgeNodeRepository) UpdateStatus(id, status string) error {
	query := `UPDATE edge_nodes SET status = $2 WHERE id = $1`
	
	_, err := r.db.Exec(query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}
