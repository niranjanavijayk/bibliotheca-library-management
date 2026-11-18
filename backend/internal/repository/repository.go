// Package repository defines persistence contracts and their implementations.
package repository

import (
	"context"
	"time"

	"bibliotheca/internal/models"
)

// BookRepository persists books.
type BookRepository interface {
	Create(ctx context.Context, b *models.Book) error
	Update(ctx context.Context, b *models.Book) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*models.Book, error)
	List(ctx context.Context, f models.BookFilter) ([]models.Book, int, error)
	ISBNExists(ctx context.Context, isbn string, excludeID int64) (bool, error)
	Stats(ctx context.Context) (models.BookStats, error)
	Categories(ctx context.Context) ([]models.CategoryCount, error)
}

// MemberRepository persists members.
type MemberRepository interface {
	Create(ctx context.Context, m *models.Member) error
	Update(ctx context.Context, m *models.Member) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*models.Member, error)
	List(ctx context.Context, f models.MemberFilter) ([]models.Member, int, error)
	EmailExists(ctx context.Context, email string, excludeID int64) (bool, error)
	Count(ctx context.Context) (int, error)
}

// LoanRepository persists loans. Issue and Return are atomic with respect to book availability.
type LoanRepository interface {
	Issue(ctx context.Context, p models.IssueParams) (*models.Loan, error)
	Return(ctx context.Context, loanID int64, at time.Time, finePerDay float64) (*models.Loan, error)
	Renew(ctx context.Context, loanID int64, newDue time.Time, maxRenewals int) (*models.Loan, error)
	GetByID(ctx context.Context, id int64) (*models.Loan, error)
	List(ctx context.Context, f models.LoanFilter) ([]models.Loan, error)
	CountActive(ctx context.Context) (int, error)
	CountOverdue(ctx context.Context, now time.Time) (int, error)
}

// Store groups all repositories.
type Store interface {
	Books() BookRepository
	Members() MemberRepository
	Loans() LoanRepository
}
