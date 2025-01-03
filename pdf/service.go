package pdf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bxcodec/go-clean-arch/domain"
	"github.com/google/uuid"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type PdfRepository interface {
	Store(ctx context.Context, p *domain.Pdf) error
	StoreAll(ctx context.Context, p []*domain.Pdf) error
	Fetch(ctx context.Context, num int64) ([]domain.Pdf, error)
	Update(ctx context.Context, p *domain.Pdf) (err error)
	GetByFileName(ctx context.Context, fileName string) (res domain.Pdf, err error)
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

func (p *Service) Fetch(ctx context.Context, num int64) ([]domain.Pdf, error) {

	res, err := p.pdfRepo.Fetch(ctx, num)
	if err != nil {
		return nil, err
	}
	return res, nil

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
		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			return domain.Pdf{}, fmt.Errorf("file does not exist: %s", srcDir)
		}
		fullPaths = append(fullPaths, srcDir)
	}
	now := time.Now()
	var pdf domain.Pdf
	pdf.CreatedAt = now
	pdf.UpdatedAt = now

	filename := fmt.Sprintf("%s%s%s", "output_", uuid.New().String(), ".pdf")
	desPath, _ := filepath.Abs(filepath.Join("tmp", filename))

	pdf.FilePath = desPath
	pdf.FileName = filename
	err := api.MergeCreateFile(fullPaths, pdf.FilePath, false, nil)
	if err != nil {
		return domain.Pdf{}, fmt.Errorf("failed to merge PDFs: %w", err)
	}

	err = p.pdfRepo.Store(ctx, &pdf)
	if err != nil {
		return domain.Pdf{}, fmt.Errorf("failed to upload PDFs: %w", err)
	}

	return pdf, nil
}

func (p *Service) Compress(ctx context.Context, inputFile string) (domain.Pdf, error) {
	srcDir, _ := filepath.Abs(filepath.Join("tmp", inputFile))
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return domain.Pdf{}, fmt.Errorf("input file does not exist: %v", err)
	}

	filename := fmt.Sprintf("%s%s%s", "compressed_", uuid.New().String(), ".pdf")
	desPath, _ := filepath.Abs(filepath.Join("tmp", filename))
	now := time.Now()

	var pdf domain.Pdf
	pdf.FilePath = desPath
	pdf.FileName = filename
	pdf.CreatedAt = now
	pdf.UpdatedAt = now

	// Define compression options
	conf := model.NewDefaultConfiguration()
	conf.Optimize = true
	conf.OptimizeResourceDicts = true
	conf.OptimizeDuplicateContentStreams = true

	// Compress the PDF
	err := api.OptimizeFile(srcDir, pdf.FilePath, conf)
	if err != nil {
		return domain.Pdf{}, fmt.Errorf("failed to compress PDF: %w", err)
	}

	fileInfo, err := os.Stat(pdf.FilePath)
	if err != nil {
		return domain.Pdf{}, err
	}

	pdf.FileSize = fileInfo.Size()

	err = p.pdfRepo.Store(ctx, &pdf)
	if err != nil {
		return domain.Pdf{}, fmt.Errorf("failed to compress PDFs: %w", err)
	}

	return pdf, nil
}

func (p *Service) RotatePdfPage(ctx context.Context, filename string, rotationAngle int) (domain.Pdf, error) {
	if rotationAngle != 90 && rotationAngle != 180 && rotationAngle != 270 {
		return domain.Pdf{}, fmt.Errorf("invalid rotation angle, must be 90, 180, or 270")
	}

	srcDir, _ := filepath.Abs(filepath.Join("tmp", filename))
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return domain.Pdf{}, fmt.Errorf("input file does not exist: %v", err)
	}

	desPath, _ := filepath.Abs(filepath.Join("tmp", filename))

	err := api.RotateFile(desPath, "", rotationAngle, nil, nil)
	if err != nil {
		return domain.Pdf{}, fmt.Errorf("failed to rotate pages: %w", err)
	}

	pdf, err := p.pdfRepo.GetByFileName(ctx, filename)
	if err != nil {
		return domain.Pdf{}, fmt.Errorf("failed to rotate PDFs: %w", err)
	}

	pdf.UpdatedAt = time.Now()

	err = p.pdfRepo.Update(ctx, &pdf)
	if err != nil {
		return domain.Pdf{}, fmt.Errorf("failed to rotate PDFs: %w", err)
	}

	return pdf, nil
}
