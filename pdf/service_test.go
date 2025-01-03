package pdf_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bxcodec/go-clean-arch/article"
	"github.com/bxcodec/go-clean-arch/domain"
	"github.com/bxcodec/go-clean-arch/pdf"
	"github.com/bxcodec/go-clean-arch/pdf/mocks"
)

func TestFetchPdf(t *testing.T) {
	mockPdfRepo := new(mocks.PdfRepository)
	mockPdf := domain.Pdf{
		FilePath: "Hello",
		FileName: "Content",
		FileSize: 64,
	}

	mockListPdf := make([]domain.Pdf, 0)
	mockListPdf = append(mockListPdf, mockPdf)

	t.Run("success", func(t *testing.T) {
		mockPdfRepo.On("Fetch", mock.Anything,
			mock.AnythingOfType("int64")).Return(mockListPdf, "next-cursor", nil).Once()

		u := pdf.NewService(mockPdfRepo)
		num := int64(1)
		list, err := u.Fetch(context.TODO(), num)
		assert.NoError(t, err)
		assert.Len(t, list, len(mockListPdf))

		mockPdfRepo.AssertExpectations(t)
	})

	t.Run("error-failed", func(t *testing.T) {
		mockPdfRepo.On("Fetch", mock.Anything,
			mock.AnythingOfType("int64")).Return(nil, "", errors.New("Unexpexted Error")).Once()

		u := pdf.NewService(mockPdfRepo)
		num := int64(1)
		list, err := u.Fetch(context.TODO(), num)

		assert.Error(t, err)
		assert.Len(t, list, 0)
		mockPdfRepo.AssertExpectations(t)
	})
}

// func TestStore(t *testing.T) {
// 	mockPdfRepo := new(mocks.PdfRepository)
// 	mockPdf := domain.Pdf{
// 		FilePath: "Hello",
// 		FileName: "Content",
// 		FileSize: 64,
// 	}

// 	t.Run("success", func(t *testing.T) {
// 		tempmockPdf := mockPdf
// 		tempmockPdf.ID = 0
// 		mockPdfRepo.On("GetByTitle", mock.Anything, mock.AnythingOfType("string")).Return(domain.Article{}, domain.ErrNotFound).Once()
// 		mockPdfRepo.On("Store", mock.Anything, mock.AnythingOfType("*domain.Article")).Return(nil).Once()

// 		u := pdf.NewService(mockPdfRepo)

// 		res, err := u.Compress(context.TODO(), &tempmockPdf)

// 		assert.NoError(t, err)
// 		assert.Equal(t, mockPdf.Title, tempmockPdf.Title)
// 		mockPdfRepo.AssertExpectations(t)
// 	})
// 	t.Run("existing-title", func(t *testing.T) {
// 		existingArticle := mockPdf
// 		mockPdfRepo.On("GetByTitle", mock.Anything, mock.AnythingOfType("string")).Return(existingArticle, nil).Once()
// 		mockAuthor := domain.Author{
// 			ID:   1,
// 			Name: "Iman Tumorang",
// 		}
// 		mockAuthorrepo.On("GetByID", mock.Anything, mock.AnythingOfType("int64")).Return(mockAuthor, nil)

// 		u := article.NewService(mockPdfRepo, mockAuthorrepo)

// 		err := u.Store(context.TODO(), &mockPdf)

// 		assert.Error(t, err)
// 		mockPdfRepo.AssertExpectations(t)
// 		mockAuthorrepo.AssertExpectations(t)
// 	})
// }

// func TestUpdate(t *testing.T) {
// 	mockArticleRepo := new(mocks.PdfRepository)
// 	mockArticle := domain.Article{
// 		Title:   "Hello",
// 		Content: "Content",
// 		ID:      23,
// 	}

// 	t.Run("success", func(t *testing.T) {
// 		mockArticleRepo.On("Update", mock.Anything, &mockArticle).Once().Return(nil)

// 		mockAuthorrepo := new(mocks.AuthorRepository)
// 		u := article.NewService(mockArticleRepo, mockAuthorrepo)

// 		err := u.Update(context.TODO(), &mockArticle)
// 		assert.NoError(t, err)
// 		mockArticleRepo.AssertExpectations(t)
// 	})
// }
