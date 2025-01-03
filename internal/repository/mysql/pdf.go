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

func (m *PdfRepository) Store(ctx context.Context, p *domain.Pdf) (err error) {
	query := `INSERT INTO pdf (file_name, file_path, created_at, updated_at, file_size) 
	VALUES (?,?,?,?,?)`
	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	res, err := stmt.ExecContext(ctx, p.FileName, p.FilePath, p.CreatedAt, p.UpdatedAt, p.FileSize)
	if err != nil {
		return
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return
	}
	p.ID = lastID
	return
}

func (m *PdfRepository) Update(ctx context.Context, p *domain.Pdf) (err error) {
	query := `UPDATE pdf set file_path=?, file_name=?, file_size=?, updated_at=? WHERE id = ?`

	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	res, err := stmt.ExecContext(ctx, p.FilePath, p.FileName, p.FileSize, p.UpdatedAt, p.ID)
	if err != nil {
		return
	}
	affect, err := res.RowsAffected()
	if err != nil {
		return
	}
	if affect != 1 {
		err = fmt.Errorf("weird  Behavior. Total Affected: %d", affect)
		return
	}

	return
}

func (m *PdfRepository) GetByFileName(ctx context.Context, fileName string) (res domain.Pdf, err error) {
	query := `SELECT id,file_path,file_name, file_size, updated_at, created_at FROM pdf WHERE file_name = ?`

	list, err := m.fetch(ctx, query, fileName)
	if err != nil {
		return
	}

	if len(list) > 0 {
		res = list[0]
	} else {
		return res, domain.ErrNotFound
	}
	return
}
