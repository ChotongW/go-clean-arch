package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/bxcodec/go-clean-arch/domain"
)

//go:generate mockery --name ArticleService
type PdfService interface {
	Upload(ctx context.Context, p []*domain.Pdf) (string, error)
	Merge(ctx context.Context, p []string) (domain.Pdf, error)
	Fetch(ctx context.Context, n int64) ([]domain.Pdf, error)
	Compress(ctx context.Context, filename string) (domain.Pdf, error)
	RotatePdfPage(ctx context.Context, filename string, rotationAngle int) (domain.Pdf, error)
	SplitPdf(ctx context.Context, filename string, pagesPerSplit int) ([]domain.Pdf, error)
}

type PdfHandler struct {
	Service PdfService
}

func NewPdfHandler(e *echo.Echo, svc PdfService) {
	handler := &PdfHandler{
		Service: svc,
	}
	e.POST("/pdf/upload", handler.Upload)
	e.POST("/pdf/merge", handler.Merge)
	e.GET("/pdf/fetch", handler.Fetch)
	e.POST("/pdf/compress/:file_name", handler.Compress)
	e.PUT("/pdf/edit", handler.EditPdf)
	e.POST("/pdf/split/:file_name", handler.SplitPdf)
	// e.DELETE("/articles/:id", handler.)
}

// Fetch godoc
// @Summary Fetch a list of PDFs
// @Description Fetches a list of PDFs based on the provided query parameter `num`
// @Tags PDFs
// @Accept json
// @Produce json
// @Param num query int false "Number of PDFs to fetch" default(10)
// @Success 200 {array} domain.Pdf "List of PDFs"
// @Failure 400 {object} ResponseError "Bad request"
// @Failure 500 {object} ResponseError "Internal server error"
// @Router /pdf/fetch [get]
func (p *PdfHandler) Fetch(c echo.Context) (err error) {
	numS := c.QueryParam("num")
	num, err := strconv.Atoi(numS)
	if err != nil || num == 0 {
		num = defaultNum
	}
	ctx := c.Request().Context()

	listAr, err := p.Service.Fetch(ctx, int64(num))
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, listAr)
}

// @Summary Upload PDF files
// @Description Upload one or more PDF files and save them to the server
// @Tags PDFs
// @Accept multipart/form-data
// @Produce json
// @Param pdfs formData file true "PDF files to be uploaded"
// @Success 201 {array} domain.Pdf "List of uploaded PDFs"
// @Failure 400 {object} map[string]string "Invalid request, error details"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /pdf/upload [post]
func (p *PdfHandler) Upload(c echo.Context) (err error) {
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to parse form"})
	}

	files := form.File["pdfs"]
	if len(files) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "at least one PDF file is required"})
	}

	for _, file := range files {
		if filepath.Ext(file.Filename) != ".pdf" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "only PDF files are allowed"})
		}
	}

	var pdfs []*domain.Pdf
	now := time.Now()
	for _, file := range files {
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
		destPath, _ := filepath.Abs(filepath.Join("tmp", filename))
		out, err := os.Create(destPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create file"})
		}
		defer out.Close()

		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to open uploaded file"})
		}
		defer src.Close()

		if _, err := io.Copy(out, src); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to copy file"})
		}
		pdf := &domain.Pdf{
			FilePath:  destPath,
			FileName:  filename,
			FileSize:  file.Size,
			CreatedAt: now,
			UpdatedAt: now,
		}

		pdfs = append(pdfs, pdf)
	}

	ctx := c.Request().Context()
	_, err = p.Service.Upload(ctx, pdfs)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, pdfs)
}

type MergeRequest struct {
	FileNames []string `json:"fileNames"`
}

// @Summary Merge PDF files
// @Description Merge the provided list of PDF files into a single file
// @Tags PDFs
// @Accept json
// @Produce json
// @Param request body MergeRequest true "List of file names to be merged"
// @Success 201 {object} domain.Pdf "The merged PDF file"
// @Failure 422 {object} string "Unprocessable entity error details"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /pdf/merge [post]
func (p *PdfHandler) Merge(c echo.Context) (err error) {
	req := new(MergeRequest)
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	ctx := c.Request().Context()
	mergedFile, err := p.Service.Merge(ctx, req.FileNames)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, mergedFile)
}

// Compress compresses a specified PDF file.
// @Summary Compress a PDF file
// @Description Compresses the PDF file specified by the file name parameter.
// @Tags PDFs
// @Accept json
// @Produce json
// @Param file_name path string true "The name of the PDF file to compress"
// @Success 201 {object} domain.Pdf "Compressed PDF file details"
// @Failure 404 {object} ResponseError "File not found"
// @Failure 500 {object} ResponseError "Internal server error"
// @Router /pdf/compress/{file_name} [post]
func (p *PdfHandler) Compress(c echo.Context) (err error) {
	filename := c.Param("file_name")
	if filename == "" {
		return c.JSON(http.StatusNotFound, domain.ErrNotFound.Error())
	}

	ctx := c.Request().Context()
	mergedFile, err := p.Service.Compress(ctx, filename)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, mergedFile)
}

type PdfEditRequest struct {
	Operation string                 `json:"operation"`
	FileName  string                 `json:"file_name"`
	Params    map[string]interface{} `json:"params"`
}

// PdfEditRequest defines the input structure for editing PDFs
// @Tags PDFs
// @Description Request body for editing a PDF file (rotate, add image, add SVG, etc.)
// @Accept json
// @Produce json
// @Param body body PdfEditRequest true "PDF Edit Request"
// @Success 200 {object} domain.Pdf "Edited PDF"
// @Failure 400 {object} ResponseError "Invalid request body or missing parameters"
// @Failure 500 {object} ResponseError "Internal server error"
// @Router /pdf/edit [put]
func (p *PdfHandler) EditPdf(c echo.Context) (err error) {
	var req PdfEditRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ResponseError{Message: "Invalid request body"})
	}
	ctx := c.Request().Context()
	var resultPdf domain.Pdf
	// Determine the operation based on the request
	switch req.Operation {
	case "rotate":
		angle, ok := req.Params["angle"].(float64)
		if !ok {
			return c.JSON(http.StatusBadRequest, ResponseError{Message: "angle is required"})
		}
		resultPdf, err = p.Service.RotatePdfPage(ctx, req.FileName, int(angle))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
		}

	case "addImage":
		// implement other operation
	case "addSvg":
		// implement other operation
	default:
		return c.JSON(http.StatusBadRequest, ResponseError{Message: "Unknown operation"})
	}

	return c.JSON(http.StatusOK, resultPdf)
}

// @Summary Split a PDF file
// @Tags PDFs
// @Description Splits the given PDF file into multiple files based on its content
// @Accept  json
// @Produce  json
// @Param file_name path string true "Name of the PDF file to split"
// @Param pages_per_split query int false "Number of pages per split" default(1)
// @Success 201 {array} domain.Pdf "Array of split PDF files"
// @Failure 400 {object} ResponseError "Bad request"
// @Failure 404 {object} ResponseError "File not found"
// @Failure 500 {object} ResponseError "Internal server error"
// @Router /pdf/split/{file_name} [post]
func (p *PdfHandler) SplitPdf(c echo.Context) (err error) {
	filename := c.Param("file_name")
	if filename == "" {
		return c.JSON(http.StatusNotFound, domain.ErrNotFound.Error())
	}

	pagesPerSplit := 1
	pagesParam := c.QueryParam("pages_per_split")
	if pagesParam != "" {
		parsedPages, err := strconv.Atoi(pagesParam)
		if err == nil {
			pagesPerSplit = parsedPages
		}
	} else {
		return c.JSON(http.StatusBadRequest, ResponseError{Message: "pages_per_split is required"})
	}

	ctx := c.Request().Context()
	splitFiles, err := p.Service.SplitPdf(ctx, filename, pagesPerSplit)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, splitFiles)
}
