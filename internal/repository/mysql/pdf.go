package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bxcodec/go-clean-arch/domain"
)

type PdfRepository struct {
	Conn *sql.DB
}

func NewPdfRepository(conn *sql.DB) *PdfRepository {
	return &PdfRepository{conn}
}
func (m *PdfRepository) StoreAll(ctx context.Context, a []*domain.Pdf) (err error) {
	query := `INSERT INTO pdf (file_name, file_path, created_at, updated_at, file_size) 
	VALUES (?,?,?,?,?)`
	tx, err := m.Conn.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	for _, pdf := range a {
		res, err := tx.ExecContext(ctx, query,
			pdf.FileName, pdf.FilePath, pdf.CreatedAt, pdf.UpdatedAt, pdf.FileSize)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("could not insert PDF: %v", err)
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			return err
		}
		pdf.ID = lastID
	}

	tx.Commit()
	return
}
