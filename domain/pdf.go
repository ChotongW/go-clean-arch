package domain

import (
	"time"
)

// Article is representing the Article data struct
type Pdf struct {
	FilePath  string    `json:"file_path" validate:"required"`
	FileName  string    `json:"file_name" validate:"required"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
	ID        int64     `json:"id"`
	FileSize  int64     `json:"file_size" validate:"required"`
}
