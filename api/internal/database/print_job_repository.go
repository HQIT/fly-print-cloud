package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"fly-print-cloud/api/internal/models"
)

type PrintJobRepository struct {
	db *DB
}

func NewPrintJobRepository(db *DB) *PrintJobRepository {
	return &PrintJobRepository{db: db}
}

// CreatePrintJob 创建打印任务
func (r *PrintJobRepository) CreatePrintJob(job *models.PrintJob) error {
	query := `
		INSERT INTO print_jobs (
			id, name, status, priority, printer_id, 
			user_id, user_name, file_path, file_size, page_count, 
			copies, paper_size, color_mode, duplex_mode, 
			start_time, end_time, error_message, retry_count, 
			max_retries, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 
			$12, $13, $14, $15, $16, $17, $18, $19, $20, $21
		)`

	now := time.Now()
	job.ID = uuid.New().String()
	job.CreatedAt = now
	job.UpdatedAt = now

	_, err := r.db.DB.Exec(query,
		job.ID, job.Name, job.Status, job.Priority, job.PrinterID,
		job.UserID, job.UserName, job.FilePath, job.FileSize, job.PageCount,
		job.Copies, job.PaperSize, job.ColorMode, job.DuplexMode,
		job.StartTime, job.EndTime, job.ErrorMessage, job.RetryCount,
		job.MaxRetries, job.CreatedAt, job.UpdatedAt,
	)

	return err
}

// GetPrintJobByID 根据ID获取打印任务
func (r *PrintJobRepository) GetPrintJobByID(id string) (*models.PrintJob, error) {
	query := `
		SELECT id, name, status, priority, printer_id, 
			   user_id, user_name, file_path, file_size, page_count, 
			   copies, paper_size, color_mode, duplex_mode, 
			   start_time, end_time, error_message, retry_count, 
			   max_retries, created_at, updated_at
		FROM print_jobs WHERE id = $1`

	job := &models.PrintJob{}
	err := r.db.DB.QueryRow(query, id).Scan(
		&job.ID, &job.Name, &job.Status, &job.Priority, &job.PrinterID,
		&job.UserID, &job.UserName, &job.FilePath, &job.FileSize, &job.PageCount,
		&job.Copies, &job.PaperSize, &job.ColorMode, &job.DuplexMode,
		&job.StartTime, &job.EndTime, &job.ErrorMessage, &job.RetryCount,
		&job.MaxRetries, &job.CreatedAt, &job.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return job, nil
}

// ListPrintJobs 获取打印任务列表
func (r *PrintJobRepository) ListPrintJobs(limit, offset int, status, printerID, userID string) ([]*models.PrintJob, error) {
	query := `
		SELECT id, name, status, priority, printer_id, 
			   user_id, user_name, file_path, file_size, page_count, 
			   copies, paper_size, color_mode, duplex_mode, 
			   start_time, end_time, error_message, retry_count, 
			   max_retries, created_at, updated_at
		FROM print_jobs WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if printerID != "" {
		query += fmt.Sprintf(" AND printer_id = $%d", argIndex)
		args = append(args, printerID)
		argIndex++
	}

	if userID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
		argIndex++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
	}

	rows, err := r.db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.PrintJob
	for rows.Next() {
		job := &models.PrintJob{}
		err := rows.Scan(
			&job.ID, &job.Name, &job.Status, &job.Priority, &job.PrinterID,
			&job.UserID, &job.UserName, &job.FilePath, &job.FileSize, &job.PageCount,
			&job.Copies, &job.PaperSize, &job.ColorMode, &job.DuplexMode,
			&job.StartTime, &job.EndTime, &job.ErrorMessage, &job.RetryCount,
			&job.MaxRetries, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// UpdatePrintJob 更新打印任务
func (r *PrintJobRepository) UpdatePrintJob(job *models.PrintJob) error {
	query := `
		UPDATE print_jobs SET 
			name = $2, status = $3, priority = $4, file_path = $5, 
			file_size = $6, page_count = $7, copies = $8, paper_size = $9, 
			color_mode = $10, duplex_mode = $11, start_time = $12, 
			end_time = $13, error_message = $14, retry_count = $15, 
			max_retries = $16, updated_at = $17
		WHERE id = $1`

	job.UpdatedAt = time.Now()

	_, err := r.db.DB.Exec(query,
		job.ID, job.Name, job.Status, job.Priority, job.FilePath,
		job.FileSize, job.PageCount, job.Copies, job.PaperSize,
		job.ColorMode, job.DuplexMode, job.StartTime,
		job.EndTime, job.ErrorMessage, job.RetryCount,
		job.MaxRetries, job.UpdatedAt,
	)

	return err
}

// DeletePrintJob 删除打印任务
func (r *PrintJobRepository) DeletePrintJob(id string) error {
	query := `DELETE FROM print_jobs WHERE id = $1`
	_, err := r.db.DB.Exec(query, id)
	return err
}

// GetPrintJobsByPrinterID 根据打印机ID获取任务列表
func (r *PrintJobRepository) GetPrintJobsByPrinterID(printerID string, limit, offset int) ([]*models.PrintJob, error) {
	return r.ListPrintJobs(limit, offset, "", printerID, "")
}

// GetPrintJobsByUserID 根据用户ID获取任务列表
func (r *PrintJobRepository) GetPrintJobsByUserID(userID string, limit, offset int) ([]*models.PrintJob, error) {
	return r.ListPrintJobs(limit, offset, "", "", userID)
}

// GetEdgeNodeIDByPrintJob 根据打印任务获取对应的 Edge Node ID
func (r *PrintJobRepository) GetEdgeNodeIDByPrintJob(jobID string) (string, error) {
	query := `
		SELECT p.edge_node_id 
		FROM print_jobs pj 
		JOIN printers p ON pj.printer_id = p.id 
		WHERE pj.id = $1`

	var edgeNodeID string
	err := r.db.DB.QueryRow(query, jobID).Scan(&edgeNodeID)
	if err != nil {
		return "", err
	}

	return edgeNodeID, nil
}
