package service

import (
	"context"
	"strings"
	"sync"

	"bibliotheca/internal/apperror"
	"bibliotheca/internal/models"
	"bibliotheca/internal/repository"
)

// LoanService issues, renews and returns books.
type LoanService interface {
	Issue(ctx context.Context, in models.IssueInput) (*models.Loan, error)
	Return(ctx context.Context, loanID int64) (*models.Loan, error)
	Renew(ctx context.Context, loanID int64) (*models.Loan, error)
	Get(ctx context.Context, id int64) (*models.Loan, error)
	List(ctx context.Context, f models.LoanFilter) ([]models.Loan, error)
	BookHistory(ctx context.Context, bookID int64) ([]models.Loan, error)
	MemberHistory(ctx context.Context, memberID int64) ([]models.Loan, error)
}

type loanService struct {
	mu      sync.Mutex // serialises lending operations inside this process
	loans   repository.LoanRepository
	books   repository.BookRepository
	members repository.MemberRepository
	cfg     Config
	clock   Clock
}

func NewLoanService(loans repository.LoanRepository, books repository.BookRepository,
	members repository.MemberRepository, cfg Config, clock Clock) LoanService {
	return &loanService{loans: loans, books: books, members: members, cfg: cfg, clock: clock}
}

func (s *loanService) refresh(l *models.Loan) { l.Refresh(s.clock(), s.cfg.FinePerDay) }

func (s *loanService) Issue(ctx context.Context, in models.IssueInput) (*models.Loan, error) {
	fe := fieldErrors{}
	if in.BookID <= 0 {
		fe.add("bookId", "Select a book")
	}
	if in.MemberID <= 0 {
		fe.add("memberId", "Select a member")
	}
	if in.Days == 0 {
		in.Days = s.cfg.LoanDays
	}
	if in.Days < 1 || in.Days > s.cfg.MaxLoanDays {
		fe.add("days", "Loan period must be between 1 and 60 days")
	}
	if err := fe.err(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	member, err := s.members.GetByID(ctx, in.MemberID)
	if err != nil {
		return nil, err
	}
	now := s.clock()
	if member.Status != models.MemberActive {
		return nil, apperror.Rule("This member's account is suspended and cannot borrow books")
	}
	if member.ExpiresAt.Before(now) {
		return nil, apperror.Rule("This member's membership has expired")
	}
	if _, err := s.books.GetByID(ctx, in.BookID); err != nil {
		return nil, err
	}
	loan, err := s.loans.Issue(ctx, models.IssueParams{
		BookID: in.BookID, MemberID: in.MemberID, IssuedAt: now,
		DueDate: now.AddDate(0, 0, in.Days), MaxActive: s.cfg.MaxLoans,
	})
	if err != nil {
		return nil, err
	}
	s.refresh(loan)
	return loan, nil
}

func (s *loanService) Return(ctx context.Context, loanID int64) (*models.Loan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	loan, err := s.loans.Return(ctx, loanID, s.clock(), s.cfg.FinePerDay)
	if err != nil {
		return nil, err
	}
	s.refresh(loan)
	return loan, nil
}

func (s *loanService) Renew(ctx context.Context, loanID int64) (*models.Loan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur, err := s.loans.GetByID(ctx, loanID)
	if err != nil {
		return nil, err
	}
	if cur.ReturnedAt != nil {
		return nil, apperror.Conflict("This loan has already been returned")
	}
	if s.clock().After(cur.DueDate) {
		return nil, apperror.Rule("Overdue loans cannot be renewed. Please return the book.")
	}
	loan, err := s.loans.Renew(ctx, loanID, cur.DueDate.AddDate(0, 0, s.cfg.RenewDays), s.cfg.MaxRenewals)
	if err != nil {
		return nil, err
	}
	s.refresh(loan)
	return loan, nil
}

func (s *loanService) Get(ctx context.Context, id int64) (*models.Loan, error) {
	l, err := s.loans.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.refresh(l)
	return l, nil
}

func (s *loanService) List(ctx context.Context, f models.LoanFilter) ([]models.Loan, error) {
	f.Now = s.clock()
	f.Query = strings.TrimSpace(f.Query)
	switch f.Status {
	case "", "all", "active", "overdue", "returned":
	case "issued":
		f.Status = "active"
	default:
		return nil, apperror.BadRequest("Unknown status filter")
	}
	loans, err := s.loans.List(ctx, f)
	if err != nil {
		return nil, err
	}
	for i := range loans {
		s.refresh(&loans[i])
	}
	return loans, nil
}

func (s *loanService) BookHistory(ctx context.Context, bookID int64) ([]models.Loan, error) {
	if _, err := s.books.GetByID(ctx, bookID); err != nil {
		return nil, err
	}
	return s.List(ctx, models.LoanFilter{BookID: bookID})
}

func (s *loanService) MemberHistory(ctx context.Context, memberID int64) ([]models.Loan, error) {
	if _, err := s.members.GetByID(ctx, memberID); err != nil {
		return nil, err
	}
	return s.List(ctx, models.LoanFilter{MemberID: memberID})
}
