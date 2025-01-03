package pdf

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/bxcodec/go-clean-arch/domain"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type PdfRepository interface {
	StoreAll(ctx context.Context, p []*domain.Pdf) error
}

type Service struct {
	pdfRepo PdfRepository
}

// NewService will create a new article service object
func NewService(p PdfRepository) *Service {
	return &Service{
		pdfRepo: p,
	}
}

func (p *Service) Upload(ctx context.Context, pdfs []*domain.Pdf) (string, error) {
	// fmt.Println("call service")
	err := p.pdfRepo.StoreAll(ctx, pdfs)
	if err != nil {
		return "", fmt.Errorf("failed to upload PDFs: %w", err)
	}
	return "", nil
}
func (p *Service) Merge(ctx context.Context, inputFiles []string) (domain.Pdf, error) {
	var fullPaths []string
	for _, file := range inputFiles {
		srcDir, _ := filepath.Abs(filepath.Join("tmp", file))
		fullPaths = append(fullPaths, srcDir)
	}
	var pdf domain.Pdf
	err := api.MergeCreateFile(fullPaths, pdf.FilePath, false, nil)
	if err != nil {
		return domain.Pdf{}, fmt.Errorf("failed to merge PDFs: %w", err)
	}
	return pdf, nil
}
