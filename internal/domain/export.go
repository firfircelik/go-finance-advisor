package domain

import (
	"time"
)

// ExportFormat represents the format for data export
type ExportFormat string

const (
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatJSON ExportFormat = "json"
	ExportFormatPDF  ExportFormat = "pdf"
)

// ExportRequest represents a request to export data
type ExportRequest struct {
	UserID    uint         `json:"user_id"`
	DataType  string       `json:"data_type"` // "transactions", "budgets", "reports", "all"
	Format    ExportFormat `json:"format"`
	StartDate *time.Time   `json:"start_date,omitempty"`
	EndDate   *time.Time   `json:"end_date,omitempty"`
	Filters   ExportFilter `json:"filters,omitempty"`
}

// ExportFilter represents filters for data export
type ExportFilter struct {
	Categories   []string `json:"categories,omitempty"`
	Transactions struct {
		Types     []string  `json:"types,omitempty"` // "income", "expense"
		MinAmount *float64  `json:"min_amount,omitempty"`
		MaxAmount *float64  `json:"max_amount,omitempty"`
		DateRange DateRange `json:"date_range,omitempty"`
	} `json:"transactions,omitempty"`
	Budgets struct {
		Periods  []string `json:"periods,omitempty"` // "monthly", "yearly"
		IsActive *bool    `json:"is_active,omitempty"`
	} `json:"budgets,omitempty"`
	Reports struct {
		Types []string `json:"types,omitempty"` // "monthly", "quarterly", "yearly"
		Year  *int     `json:"year,omitempty"`
		Month *int     `json:"month,omitempty"`
	} `json:"reports,omitempty"`
}

// DateRange represents a date range filter
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ExportResponse represents the response from an export operation
type ExportResponse struct {
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	ExportedAt  time.Time `json:"exported_at"`
	RecordCount int       `json:"record_count"`
	Format      string    `json:"format"`
}

// ExportData represents all user financial data for export
type ExportData struct {
	Transactions []Transaction  `json:"transactions"`
	Budgets      []Budget       `json:"budgets"`
	Categories   []Category     `json:"categories"`
	ExportedAt   time.Time      `json:"exported_at"`
	Format       ExportFormat   `json:"format"`
	Metadata     ExportMetadata `json:"metadata"`
}

// ExportMetadata contains metadata about the export
type ExportMetadata struct {
	UserID           uint      `json:"user_id"`
	ExportType       string    `json:"export_type"`
	TotalRecords     int       `json:"total_records"`
	TransactionCount int       `json:"transaction_count"`
	BudgetCount      int       `json:"budget_count"`
	CategoryCount    int       `json:"category_count"`
	DateRange        DateRange `json:"date_range,omitempty"`
	FiltersApplied   bool      `json:"filters_applied"`
}

// ExportJob represents an export job for async processing
type ExportJob struct {
	ID          uint          `json:"id" gorm:"primaryKey"`
	UserID      uint          `json:"user_id" gorm:"not null;index"`
	JobID       string        `json:"job_id" gorm:"uniqueIndex;not null"`
	DataType    string        `json:"data_type" gorm:"not null"`
	Format      ExportFormat  `json:"format" gorm:"not null"`
	Status      ExportStatus  `json:"status" gorm:"not null;default:'pending'"`
	Filename    string        `json:"filename"`
	FileSize    int64         `json:"file_size"`
	RecordCount int           `json:"record_count"`
	Progress    int           `json:"progress" gorm:"default:0"` // 0-100
	ErrorMsg    string        `json:"error_message,omitempty"`
	StartedAt   *time.Time    `json:"started_at"`
	CompletedAt *time.Time    `json:"completed_at"`
	ExpiresAt   time.Time     `json:"expires_at"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Request     ExportRequest `json:"request" gorm:"type:json"`
}

// ExportStatus represents the status of an export job
type ExportStatus string

const (
	ExportStatusPending    ExportStatus = "pending"
	ExportStatusProcessing ExportStatus = "processing"
	ExportStatusCompleted  ExportStatus = "completed"
	ExportStatusFailed     ExportStatus = "failed"
	ExportStatusExpired    ExportStatus = "expired"
)

// IsValidFormat checks if the export format is valid
func (f ExportFormat) IsValid() bool {
	switch f {
	case ExportFormatCSV, ExportFormatJSON, ExportFormatPDF:
		return true
	default:
		return false
	}
}

// GetContentType returns the MIME type for the export format
func (f ExportFormat) GetContentType() string {
	switch f {
	case ExportFormatCSV:
		return "text/csv"
	case ExportFormatJSON:
		return "application/json"
	case ExportFormatPDF:
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}

// GetFileExtension returns the file extension for the export format
func (f ExportFormat) GetFileExtension() string {
	switch f {
	case ExportFormatCSV:
		return ".csv"
	case ExportFormatJSON:
		return ".json"
	case ExportFormatPDF:
		return ".pdf"
	default:
		return ".bin"
	}
}

// IsCompleted checks if the export job is completed
func (j *ExportJob) IsCompleted() bool {
	return j.Status == ExportStatusCompleted
}

// IsFailed checks if the export job has failed
func (j *ExportJob) IsFailed() bool {
	return j.Status == ExportStatusFailed
}

// IsExpired checks if the export job has expired
func (j *ExportJob) IsExpired() bool {
	return j.Status == ExportStatusExpired || time.Now().After(j.ExpiresAt)
}

// CanDownload checks if the export file can be downloaded
func (j *ExportJob) CanDownload() bool {
	return j.IsCompleted() && !j.IsExpired() && j.Filename != ""
}

// UpdateProgress updates the job progress
func (j *ExportJob) UpdateProgress(progress int) {
	if progress < 0 {
		progress = 0
	} else if progress > 100 {
		progress = 100
	}
	j.Progress = progress
	j.UpdatedAt = time.Now()
}

// MarkAsStarted marks the job as started
func (j *ExportJob) MarkAsStarted() {
	now := time.Now()
	j.Status = ExportStatusProcessing
	j.StartedAt = &now
	j.UpdatedAt = now
}

// MarkAsCompleted marks the job as completed
func (j *ExportJob) MarkAsCompleted(filename string, fileSize int64, recordCount int) {
	now := time.Now()
	j.Status = ExportStatusCompleted
	j.Filename = filename
	j.FileSize = fileSize
	j.RecordCount = recordCount
	j.Progress = 100
	j.CompletedAt = &now
	j.UpdatedAt = now
}

// MarkAsFailed marks the job as failed
func (j *ExportJob) MarkAsFailed(errorMsg string) {
	j.Status = ExportStatusFailed
	j.ErrorMsg = errorMsg
	j.UpdatedAt = time.Now()
}
