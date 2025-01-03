package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bxcodec/go-clean-arch/domain"
	"github.com/sirupsen/logrus"
)

type PdfRepository struct {
	Conn *sql.DB
}

func NewPdfRepository(conn *sql.DB) *PdfRepository {
	return &PdfRepository{conn}
}

func (m *PdfRepository) fetch(ctx context.Context, query string, args ...interface{}) (result []domain.Pdf, err error) {
	rows, err := m.Conn.QueryContext(ctx, query, args...)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	defer func() {
		errRow := rows.Close()
		if errRow != nil {
			logrus.Error(errRow)
		}
	}()

	result = make([]domain.Pdf, 0)
	for rows.Next() {
		t := domain.Pdf{}
		err = rows.Scan(
			&t.ID,
			&t.FilePath,
			&t.FileName,
			&t.FileSize,
			&t.UpdatedAt,
			&t.CreatedAt,
		)

		if err != nil {
			logrus.Error(err)
			return nil, err
		}
		result = append(result, t)
	}

	return result, nil
}

func (m *PdfRepository) Fetch(ctx context.Context, num int64) (res []domain.Pdf, err error) {
	query := `SELECT id,file_path,file_name, file_size, updated_at, created_at
  						FROM pdf ORDER BY created_at LIMIT ? `

	res, err = m.fetch(ctx, query, num)
	if err != nil {
		return nil, err
	}

	return
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
