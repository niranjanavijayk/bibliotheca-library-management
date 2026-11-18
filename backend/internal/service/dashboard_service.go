package service

import (
	"context"
	"sync"

	"bibliotheca/internal/models"
	"bibliotheca/internal/repository"
)

// DashboardService aggregates library statistics.
type DashboardService interface {
	Get(ctx context.Context) (*models.Dashboard, error)
}

type dashboardService struct {
	store repository.Store
	loans LoanService
	clock Clock
}

func NewDashboardService(store repository.Store, loans LoanService, clock Clock) DashboardService {
	return &dashboardService{store: store, loans: loans, clock: clock}
}

// Get runs the independent queries concurrently, joined with a WaitGroup and
// guarded by a Mutex that protects the shared result and first-error slot.
func (s *dashboardService) Get(ctx context.Context) (*models.Dashboard, error) {
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		firstErr error
		d        = &models.Dashboard{}
	)
	run := func(fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}
	now := s.clock()

	run(func() error {
		st, err := s.store.Books().Stats(ctx)
		mu.Lock()
		d.TotalBooks, d.TotalCopies, d.AvailableCopies = st.Titles, st.Copies, st.Available
		mu.Unlock()
		return err
	})
	run(func() error {
		n, err := s.store.Members().Count(ctx)
		mu.Lock()
		d.TotalMembers = n
		mu.Unlock()
		return err
	})
	run(func() error {
		n, err := s.store.Loans().CountActive(ctx)
		mu.Lock()
		d.IssuedBooks = n
		mu.Unlock()
		return err
	})
	run(func() error {
		n, err := s.store.Loans().CountOverdue(ctx, now)
		mu.Lock()
		d.OverdueBooks = n
		mu.Unlock()
		return err
	})
	run(func() error {
		c, err := s.store.Books().Categories(ctx)
		mu.Lock()
		d.Categories = c
		mu.Unlock()
		return err
	})
	run(func() error {
		l, err := s.loans.List(ctx, models.LoanFilter{Limit: 6})
		mu.Lock()
		d.RecentLoans = l
		mu.Unlock()
		return err
	})

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	if d.Categories == nil {
		d.Categories = []models.CategoryCount{}
	}
	if d.RecentLoans == nil {
		d.RecentLoans = []models.Loan{}
	}
	return d, nil
}
