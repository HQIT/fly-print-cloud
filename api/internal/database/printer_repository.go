package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"fly-print-cloud/api/internal/models"
)

type PrinterRepository struct {
	db *DB
}

func NewPrinterRepository(db *DB) *PrinterRepository {
	return &PrinterRepository{db: db}
}

// CreatePrinter 创建打印机
func (r *PrinterRepository) CreatePrinter(printer *models.Printer) error {
	// 将 Capabilities 结构体转换为 JSON
	capabilitiesJSON, err := json.Marshal(printer.Capabilities)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	query := `
		INSERT INTO printers (id, name, display_name, model, serial_number, status, firmware_version, 
		                     port_info, ip_address, mac_address, network_config,
		                     latitude, longitude, location, capabilities, edge_node_id, queue_length)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING created_at, updated_at`
	
	err = r.db.QueryRow(query,
		printer.ID, printer.Name, printer.DisplayName, printer.Model, printer.SerialNumber, printer.Status,
		printer.FirmwareVersion, printer.PortInfo, printer.IPAddress, printer.MACAddress,
		printer.NetworkConfig, printer.Latitude, printer.Longitude, printer.Location,
		capabilitiesJSON, printer.EdgeNodeID, printer.QueueLength,
	).Scan(&printer.CreatedAt, &printer.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create printer: %w", err)
	}
	return nil
}

// GetPrinterByNameAndEdgeNode 根据名称和边缘节点ID获取打印机
func (r *PrinterRepository) GetPrinterByNameAndEdgeNode(name, edgeNodeID string) (*models.Printer, error) {
	query := `
		SELECT id, name, display_name, model, serial_number, status, enabled, firmware_version, port_info,
		       ip_address, mac_address, network_config, latitude, longitude, location,
		       capabilities, edge_node_id, queue_length, created_at, updated_at
		FROM printers 
		WHERE name = $1 AND edge_node_id = $2`
	
	var printer models.Printer
	var capabilitiesJSON []byte
	var firmwareVersion, portInfo sql.NullString
	
	var displayName sql.NullString
	err := r.db.QueryRow(query, name, edgeNodeID).Scan(
		&printer.ID, &printer.Name, &displayName, &printer.Model, &printer.SerialNumber, &printer.Status, &printer.Enabled,
		&firmwareVersion, &portInfo, &printer.IPAddress, &printer.MACAddress,
		&printer.NetworkConfig, &printer.Latitude, &printer.Longitude, &printer.Location,
		&capabilitiesJSON, &printer.EdgeNodeID, &printer.QueueLength,
		&printer.CreatedAt, &printer.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get printer by name and edge node: %w", err)
	}
	
	// 处理可能为 NULL 的字段
	if firmwareVersion.Valid {
		printer.FirmwareVersion = firmwareVersion.String
	}
	if portInfo.Valid {
		printer.PortInfo = portInfo.String
	}
	if displayName.Valid {
		printer.DisplayName = displayName.String
	}
	
	// 解析 capabilities JSON
	if len(capabilitiesJSON) > 0 {
		if err := json.Unmarshal(capabilitiesJSON, &printer.Capabilities); err != nil {
			return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
		}
	}
	
	return &printer, nil
}

// GetPrinterByID 根据ID获取打印机
func (r *PrinterRepository) GetPrinterByID(printerID string) (*models.Printer, error) {
	query := `
		SELECT id, name, display_name, model, serial_number, status, enabled, firmware_version, port_info,
		       ip_address, mac_address, network_config, latitude, longitude, location,
		       capabilities, edge_node_id, queue_length, created_at, updated_at
		FROM printers WHERE id = $1`
	
	printer := &models.Printer{}
	var ipAddress sql.NullString
	var firmwareVersion sql.NullString
	var displayName sql.NullString
	var capabilitiesJSON []byte
	
	err := r.db.QueryRow(query, printerID).Scan(
		&printer.ID, &printer.Name, &displayName, &printer.Model, &printer.SerialNumber, &printer.Status, &printer.Enabled,
		&firmwareVersion, &printer.PortInfo, &ipAddress, &printer.MACAddress,
		&printer.NetworkConfig, &printer.Latitude, &printer.Longitude, &printer.Location,
		&capabilitiesJSON, &printer.EdgeNodeID, &printer.QueueLength,
		&printer.CreatedAt, &printer.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("printer not found")
		}
		return nil, fmt.Errorf("failed to get printer: %w", err)
	}
	
	// 处理可空字段
	if ipAddress.Valid {
		printer.IPAddress = &ipAddress.String
	}
	if firmwareVersion.Valid {
		printer.FirmwareVersion = firmwareVersion.String
	}
	if displayName.Valid {
		printer.DisplayName = displayName.String
	}
	
	// 解析 JSON capabilities
	if err := json.Unmarshal(capabilitiesJSON, &printer.Capabilities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
	}
	
	return printer, nil
}

// ListPrinters 获取打印机列表
func (r *PrinterRepository) ListPrinters(page, pageSize int) ([]*models.Printer, int, error) {
	offset := (page - 1) * pageSize
	
	// 获取总数
	var total int
	countQuery := `SELECT COUNT(*) FROM printers`
	err := r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get printer count: %w", err)
	}
	
	// 获取分页数据
	query := `
		SELECT id, name, display_name, model, serial_number, status, enabled, firmware_version, port_info,
		       ip_address, mac_address, network_config, latitude, longitude, location,
		       capabilities, edge_node_id, queue_length, created_at, updated_at
		FROM printers 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`
	
	rows, err := r.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list printers: %w", err)
	}
	defer rows.Close()
	
	var printers []*models.Printer
	for rows.Next() {
		printer := &models.Printer{}
		var ipAddress sql.NullString
		var firmwareVersion sql.NullString
		var displayName sql.NullString
		var capabilitiesJSON []byte
		
		err := rows.Scan(
			&printer.ID, &printer.Name, &displayName, &printer.Model, &printer.SerialNumber, &printer.Status, &printer.Enabled,
			&firmwareVersion, &printer.PortInfo, &ipAddress, &printer.MACAddress,
			&printer.NetworkConfig, &printer.Latitude, &printer.Longitude, &printer.Location,
			&capabilitiesJSON, &printer.EdgeNodeID, &printer.QueueLength,
			&printer.CreatedAt, &printer.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan printer: %w", err)
		}
		
		// 处理可空字段
		if ipAddress.Valid {
			printer.IPAddress = &ipAddress.String
		}
		if firmwareVersion.Valid {
			printer.FirmwareVersion = firmwareVersion.String
		}
		if displayName.Valid {
			printer.DisplayName = displayName.String
		}
		
		// 解析 JSON capabilities
		if err := json.Unmarshal(capabilitiesJSON, &printer.Capabilities); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal capabilities: %w", err)
		}
		
		printers = append(printers, printer)
	}
	
	return printers, total, nil
}

// CountPrintersByEdgeNode 统计边缘节点的打印机数量
func (r *PrinterRepository) CountPrintersByEdgeNode(edgeNodeID string) (int, error) {
	query := `SELECT COUNT(*) FROM printers WHERE edge_node_id = $1`
	
	var count int
	err := r.db.QueryRow(query, edgeNodeID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count printers by edge node: %w", err)
	}
	
	return count, nil
}

// ListPrintersByEdgeNode 根据 Edge Node ID 获取打印机列表
func (r *PrinterRepository) ListPrintersByEdgeNode(edgeNodeID string) ([]*models.Printer, error) {
	query := `
		SELECT id, name, display_name, model, serial_number, status, enabled, firmware_version, port_info,
		       ip_address, mac_address, network_config, latitude, longitude, location,
		       capabilities, edge_node_id, queue_length, created_at, updated_at
		FROM printers 
		WHERE edge_node_id = $1
		ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query, edgeNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list printers: %w", err)
	}
	defer rows.Close()
	
	var printers []*models.Printer
	for rows.Next() {
		printer := &models.Printer{}
		var ipAddress sql.NullString
		var firmwareVersion sql.NullString
		var displayName sql.NullString
		var capabilitiesJSON []byte
		
		err := rows.Scan(
			&printer.ID, &printer.Name, &displayName, &printer.Model, &printer.SerialNumber, &printer.Status, &printer.Enabled,
			&firmwareVersion, &printer.PortInfo, &ipAddress, &printer.MACAddress,
			&printer.NetworkConfig, &printer.Latitude, &printer.Longitude, &printer.Location,
			&capabilitiesJSON, &printer.EdgeNodeID, &printer.QueueLength,
			&printer.CreatedAt, &printer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan printer: %w", err)
		}
		
		// 处理可空字段
		if ipAddress.Valid {
			printer.IPAddress = &ipAddress.String
		}
		if firmwareVersion.Valid {
			printer.FirmwareVersion = firmwareVersion.String
		}
		if displayName.Valid {
			printer.DisplayName = displayName.String
		}
		
		// 解析 JSON capabilities
		if err := json.Unmarshal(capabilitiesJSON, &printer.Capabilities); err != nil {
			return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
		}
		
		printers = append(printers, printer)
	}
	
	return printers, nil
}

// UpdatePrinter 更新打印机
func (r *PrinterRepository) UpdatePrinter(printer *models.Printer) error {
	// 将 Capabilities 结构体转换为 JSON
	capabilitiesJSON, err := json.Marshal(printer.Capabilities)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	query := `
		UPDATE printers 
		SET name = $2, display_name = $3, model = $4, serial_number = $5, status = $6, enabled = $7,
		    firmware_version = $8, port_info = $9, ip_address = $10, mac_address = $11, network_config = $12,
		    latitude = $13, longitude = $14, location = $15, capabilities = $16,
		    queue_length = $17, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`
	
	err = r.db.QueryRow(query,
		printer.ID, printer.Name, printer.DisplayName, printer.Model, printer.SerialNumber, printer.Status, printer.Enabled,
		printer.FirmwareVersion, printer.PortInfo, printer.IPAddress, printer.MACAddress,
		printer.NetworkConfig, printer.Latitude, printer.Longitude, printer.Location,
		capabilitiesJSON, printer.QueueLength,
	).Scan(&printer.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("printer not found")
		}
		return fmt.Errorf("failed to update printer: %w", err)
	}
	return nil
}

// DeletePrinter 删除打印机
func (r *PrinterRepository) DeletePrinter(printerID string) error {
	query := `DELETE FROM printers WHERE id = $1`
	
	result, err := r.db.Exec(query, printerID)
	if err != nil {
		return fmt.Errorf("failed to delete printer: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("printer not found")
	}
	
	return nil
}

// UpsertPrinter 插入或更新打印机（基于 name + edge_node_id 的唯一性）
func (r *PrinterRepository) UpsertPrinter(printer *models.Printer) error {
	// 将 Capabilities 结构体转换为 JSON
	capabilitiesJSON, err := json.Marshal(printer.Capabilities)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	query := `
		INSERT INTO printers (
			id, name, model, serial_number, status, firmware_version, port_info,
			ip_address, mac_address, network_config, latitude, longitude, location,
			capabilities, edge_node_id, queue_length, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
		ON CONFLICT (name, edge_node_id) 
		DO UPDATE SET
			model = EXCLUDED.model,
			serial_number = EXCLUDED.serial_number,
			status = EXCLUDED.status,
			firmware_version = EXCLUDED.firmware_version,
			port_info = EXCLUDED.port_info,
			ip_address = EXCLUDED.ip_address,
			mac_address = EXCLUDED.mac_address,
			network_config = EXCLUDED.network_config,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			location = EXCLUDED.location,
			capabilities = EXCLUDED.capabilities,
			queue_length = EXCLUDED.queue_length,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id`

	var returnedID string
	err = r.db.QueryRow(
		query,
		printer.ID, printer.Name, printer.Model, printer.SerialNumber,
		printer.Status, printer.FirmwareVersion, printer.PortInfo,
		printer.IPAddress, printer.MACAddress, printer.NetworkConfig,
		printer.Latitude, printer.Longitude, printer.Location,
		capabilitiesJSON, printer.EdgeNodeID, printer.QueueLength,
		time.Now(), time.Now(),
	).Scan(&returnedID)

	if err != nil {
		return fmt.Errorf("failed to upsert printer: %w", err)
	}

	// 更新返回的 ID（如果是更新操作，ID 可能不同）
	printer.ID = returnedID
	return nil
}

// DisablePrintersByEdgeNode 禁用指定Edge Node下的所有打印机
func (r *PrinterRepository) DisablePrintersByEdgeNode(edgeNodeID string) error {
	query := `UPDATE printers SET enabled = false WHERE edge_node_id = $1`
	_, err := r.db.DB.Exec(query, edgeNodeID)
	return err
}

// EnablePrintersByEdgeNode 启用指定Edge Node下的所有打印机
func (r *PrinterRepository) EnablePrintersByEdgeNode(edgeNodeID string) error {
	query := `UPDATE printers SET enabled = true WHERE edge_node_id = $1`
	_, err := r.db.DB.Exec(query, edgeNodeID)
	return err
}

