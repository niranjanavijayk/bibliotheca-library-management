package service

import (
	"context"
	"strings"

	"bibliotheca/internal/apperror"
	"bibliotheca/internal/models"
	"bibliotheca/internal/repository"
)

// BookService manages the catalogue.
type BookService interface {
	Create(ctx context.Context, in models.BookInput) (*models.Book, error)
	Update(ctx context.Context, id int64, in models.BookInput) (*models.Book, error)
	Delete(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (*models.Book, error)
	List(ctx context.Context, f models.BookFilter) ([]models.Book, int, error)
	Categories(ctx context.Context) ([]models.CategoryCount, error)
}

type bookService struct {
	books repository.BookRepository
	clock Clock
}

func NewBookService(books repository.BookRepository, clock Clock) BookService {
	return &bookService{books: books, clock: clock}
}

func defaultCover(isbn string) string {
	return "https://covers.openlibrary.org/b/isbn/" + isbn + "-L.jpg?default=false"
}

func (s *bookService) validate(ctx context.Context, in *models.BookInput, excludeID int64) error {
	fe := fieldErrors{}
	in.Title = fe.required("title", "Title", in.Title, 255)
	in.Author = fe.required("author", "Author", in.Author, 255)
	in.Category = titleCase(fe.required("category", "Category", in.Category, 100))
	in.ISBN = normalizeISBN(in.ISBN)
	switch {
	case in.ISBN == "":
		fe.add("isbn", "ISBN is required")
	case !validISBN(in.ISBN):
		fe.add("isbn", "ISBN must be 10 or 13 digits")
	}
	in.Description = strings.TrimSpace(in.Description)
	in.CoverURL = strings.TrimSpace(in.CoverURL)
	if in.CoverURL != "" && !strings.HasPrefix(in.CoverURL, "http://") && !strings.HasPrefix(in.CoverURL, "https://") {
		fe.add("coverUrl", "Cover URL must start with http:// or https://")
	}
	if !validYear(in.PublishedYear, s.clock()) {
		fe.add("publishedYear", "Enter a valid publication year")
	}
	if in.TotalCopies < 1 || in.TotalCopies > 1000 {
		fe.add("totalCopies", "Copies must be between 1 and 1000")
	}
	if err := fe.err(); err != nil {
		return err
	}
	exists, err := s.books.ISBNExists(ctx, in.ISBN, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return apperror.Validation(map[string]string{"isbn": "A book with this ISBN already exists"})
	}
	return nil
}

func (s *bookService) Create(ctx context.Context, in models.BookInput) (*models.Book, error) {
	if err := s.validate(ctx, &in, 0); err != nil {
		return nil, err
	}
	b := &models.Book{Title: in.Title, Author: in.Author, Category: in.Category, ISBN: in.ISBN,
		Description: in.Description, CoverURL: in.CoverURL, PublishedYear: in.PublishedYear,
		TotalCopies: in.TotalCopies, AvailableCopies: in.TotalCopies}
	if b.CoverURL == "" {
		b.CoverURL = defaultCover(b.ISBN)
	}
	if err := s.books.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *bookService) Update(ctx context.Context, id int64, in models.BookInput) (*models.Book, error) {
	cur, err := s.books.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.validate(ctx, &in, id); err != nil {
		return nil, err
	}
	issued := cur.TotalCopies - cur.AvailableCopies
	if in.TotalCopies < issued {
		return nil, apperror.Validation(map[string]string{
			"totalCopies": "Cannot be lower than the copies currently issued"})
	}
	b := &models.Book{ID: id, Title: in.Title, Author: in.Author, Category: in.Category, ISBN: in.ISBN,
		Description: in.Description, CoverURL: in.CoverURL, PublishedYear: in.PublishedYear, TotalCopies: in.TotalCopies}
	if b.CoverURL == "" {
		b.CoverURL = defaultCover(b.ISBN)
	}
	if err := s.books.Update(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *bookService) Delete(ctx context.Context, id int64) error {
	b, err := s.books.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if b.TotalCopies-b.AvailableCopies > 0 {
		return apperror.Conflict("This book has copies currently issued. Return them before deleting.")
	}
	return s.books.Delete(ctx, id)
}

func (s *bookService) Get(ctx context.Context, id int64) (*models.Book, error) {
	return s.books.GetByID(ctx, id)
}

func (s *bookService) List(ctx context.Context, f models.BookFilter) ([]models.Book, int, error) {
	f.Query = strings.TrimSpace(f.Query)
	return s.books.List(ctx, f)
}

func (s *bookService) Categories(ctx context.Context) ([]models.CategoryCount, error) {
	return s.books.Categories(ctx)
}
