package repository

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"bibliotheca/internal/apperror"
	"bibliotheca/internal/models"
)

type memData struct {
	mu       sync.RWMutex
	books    map[int64]*models.Book
	members  map[int64]*models.Member
	loans    map[int64]*models.Loan
	nextBook int64
	nextMem  int64
	nextLoan int64
}

// MemoryStore is a thread-safe in-memory Store used by unit tests and demo mode.
type MemoryStore struct{ d *memData }

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{d: &memData{
		books: map[int64]*models.Book{}, members: map[int64]*models.Member{}, loans: map[int64]*models.Loan{},
		nextBook: 1, nextMem: 1, nextLoan: 1,
	}}
}

func (s *MemoryStore) Books() BookRepository     { return &memBooks{s.d} }
func (s *MemoryStore) Members() MemberRepository { return &memMembers{s.d} }
func (s *MemoryStore) Loans() LoanRepository     { return &memLoans{s.d} }

func paginate(total, page, size int) (int, int) {
	if size <= 0 {
		size = 12
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	return start, end
}

// ---------------- Books ----------------

type memBooks struct{ d *memData }

func (r *memBooks) Create(_ context.Context, b *models.Book) error {
	r.d.mu.Lock()
	defer r.d.mu.Unlock()
	b.ID = r.d.nextBook
	r.d.nextBook++
	b.CreatedAt = time.Now()
	b.Normalize()
	c := *b
	r.d.books[b.ID] = &c
	return nil
}

func (r *memBooks) Update(_ context.Context, b *models.Book) error {
	r.d.mu.Lock()
	defer r.d.mu.Unlock()
	cur, ok := r.d.books[b.ID]
	if !ok {
		return apperror.NotFound("Book")
	}
	avail := cur.AvailableCopies + (b.TotalCopies - cur.TotalCopies)
	if avail < 0 {
		return apperror.Conflict("Total copies cannot be lower than the number of copies currently issued")
	}
	b.AvailableCopies = avail
	b.CreatedAt = cur.CreatedAt
	b.Normalize()
	c := *b
	r.d.books[b.ID] = &c
	return nil
}

func (r *memBooks) Delete(_ context.Context, id int64) error {
	r.d.mu.Lock()
	defer r.d.mu.Unlock()
	if _, ok := r.d.books[id]; !ok {
		return apperror.NotFound("Book")
	}
	delete(r.d.books, id)
	for lid, l := range r.d.loans {
		if l.BookID == id {
			delete(r.d.loans, lid)
		}
	}
	return nil
}

func (r *memBooks) GetByID(_ context.Context, id int64) (*models.Book, error) {
	r.d.mu.RLock()
	defer r.d.mu.RUnlock()
	b, ok := r.d.books[id]
	if !ok {
		return nil, apperror.NotFound("Book")
	}
	c := *b
	return &c, nil
}

func (r *memBooks) List(_ context.Context, f models.BookFilter) ([]models.Book, int, error) {
	r.d.mu.RLock()
	defer r.d.mu.RUnlock()
	q := strings.ToLower(f.Query)
	out := []models.Book{}
	for _, b := range r.d.books {
		if q != "" && !strings.Contains(strings.ToLower(b.Title+" "+b.Author+" "+b.ISBN), q) {
			continue
		}
		if f.Category != "" && !strings.EqualFold(b.Category, f.Category) {
			continue
		}
		if f.Available != nil && (b.AvailableCopies > 0) != *f.Available {
			continue
		}
		out = append(out, *b)
	}
	less := func(i, j int) bool { return out[i].ID < out[j].ID }
	switch f.SortBy {
	case "title":
		less = func(i, j int) bool { return strings.ToLower(out[i].Title) < strings.ToLower(out[j].Title) }
	case "author":
		less = func(i, j int) bool { return strings.ToLower(out[i].Author) < strings.ToLower(out[j].Author) }
	case "category":
		less = func(i, j int) bool { return strings.ToLower(out[i].Category) < strings.ToLower(out[j].Category) }
	case "year":
		less = func(i, j int) bool { return out[i].PublishedYear < out[j].PublishedYear }
	case "copies":
		less = func(i, j int) bool { return out[i].AvailableCopies < out[j].AvailableCopies }
	case "added":
		less = func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) }
	}
	if strings.EqualFold(f.Order, "desc") {
		base := less
		less = func(i, j int) bool { return base(j, i) }
	}
	sort.SliceStable(out, less)
	total := len(out)
	s, e := paginate(total, f.Page, f.PageSize)
	return out[s:e], total, nil
}

func (r *memBooks) ISBNExists(_ context.Context, isbn string, excludeID int64) (bool, error) {
	r.d.mu.RLock()
	defer r.d.mu.RUnlock()
	for _, b := range r.d.books {
		if b.ISBN == isbn && b.ID != excludeID {
			return true, nil
		}
	}
	return false, nil
}

func (r *memBooks) Stats(_ context.Context) (models.BookStats, error) {
	r.d.mu.RLock()
	defer r.d.mu.RUnlock()
	st := models.BookStats{Titles: len(r.d.books)}
	for _, b := range r.d.books {
		st.Copies += b.TotalCopies
		st.Available += b.AvailableCopies
	}
	return st, nil
}

func (r *memBooks) Categories(_ context.Context) ([]models.CategoryCount, error) {
	r.d.mu.RLock()
	defer r.d.mu.RUnlock()
	counts := map[string]int{}
	for _, b := range r.d.books {
		counts[b.Category]++
	}
	out := make([]models.CategoryCount, 0, len(counts))
	for k, v := range counts {
		out = append(out, models.CategoryCount{Name: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// ---------------- Members ----------------

type memMembers struct{ d *memData }

func (r *memMembers) active(id int64) int {
	n := 0
	for _, l := range r.d.loans {
		if l.MemberID == id && l.ReturnedAt == nil {
			n++
		}
	}
	return n
}

func (r *memMembers) Create(_ context.Context, m *models.Member) error {
	r.d.mu.Lock()
	defer r.d.mu.Unlock()
	m.ID = r.d.nextMem
	r.d.nextMem++
	if m.JoinedAt.IsZero() {
		m.JoinedAt = time.Now()
	}
	c := *m
	r.d.members[m.ID] = &c
	return nil
}

func (r *memMembers) Update(_ context.Context, m *models.Member) error {
	r.d.mu.Lock()
	defer r.d.mu.Unlock()
	cur, ok := r.d.members[m.ID]
	if !ok {
		return apperror.NotFound("Member")
	}
	m.JoinedAt = cur.JoinedAt
	c := *m
	r.d.members[m.ID] = &c
	return nil
}

func (r *memMembers) Delete(_ context.Context, id int64) error {
	r.d.mu.Lock()
	defer r.d.mu.Unlock()
	if _, ok := r.d.members[id]; !ok {
		return apperror.NotFound("Member")
	}
	delete(r.d.members, id)
	for lid, l := range r.d.loans {
		if l.MemberID == id {
			delete(r.d.loans, lid)
		}
	}
	return nil
}

func (r *memMembers) GetByID(_ context.Context, id int64) (*models.Member, error) {
	r.d.mu.RLock()
	defer r.d.mu.RUnlock()
	m, ok := r.d.members[id]
	if !ok {
		return nil, apperror.NotFound("Member")
	}
	c := *m
	c.ActiveLoans = r.active(id)
	return &c, nil
}

func (r *memMembers) List(_ context.Context, f models.MemberFilter) ([]models.Member, int, error) {
	r.d.mu.RLock()
	defer r.d.mu.RUnlock()
	q := strings.ToLower(f.Query)
	out := []models.Member{}
	for _, m := range r.d.members {
		if q != "" && !strings.Contains(strings.ToLower(m.Name+" "+m.Email+" "+m.Phone), q) {
			continue
		}
		if f.Status != "" && m.Status != f.Status {
			continue
		}
		c := *m
		c.ActiveLoans = r.active(m.ID)
		out = append(out, c)
	}
	less := func(i, j int) bool { return out[i].ID < out[j].ID }
	switch f.SortBy {
	case "name":
		less = func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) }
	case "joined":
		less = func(i, j int) bool { return out[i].JoinedAt.Before(out[j].JoinedAt) }
	case "loans":
		less = func(i, j int) bool { return out[i].ActiveLoans < out[j].ActiveLoans }
	}
	if strings.EqualFold(f.Order, "desc") {
		base := less
		less = func(i, j int) bool { return base(j, i) }
	}
	sort.SliceStable(out, less)
	total := len(out)
	s, e := paginate(total, f.Page, f.PageSize)
	return out[s:e], total, nil
}

func (r *memMembers) EmailExists(_ context.Context, email string, excludeID int64) (bool, error) {
	r.d.mu.RLock()
	defer r.d.mu.RUnlock()
	for _, m := range r.d.members {
		if strings.EqualFold(m.Email, email) && m.ID != excludeID {
			return true, nil
		}
	}
	return false, nil
}

func (r *memMembers) Count(_ context.Context) (int, error) {
	r.d.mu.RLock()
	defer r.d.mu.RUnlock()
	return len(r.d.members), nil
}

// ---------------- Loans ----------------

type memLoans struct{ d *memData }

func (r *memLoans) hydrate(l *models.Loan) models.Loan {
	c := *l
	if b, ok := r.d.books[l.BookID]; ok {
		c.BookTitle, c.BookAuthor, c.BookISBN, c.BookCover = b.Title, b.Author, b.ISBN, b.CoverURL
	}
	if m, ok := r.d.members[l.MemberID]; ok {
		c.MemberName = m.Name
	}
	return c
}

func (r *memLoans) Issue(_ context.Context, p models.IssueParams) (*models.Loan, error) {
	r.d.mu.Lock()
	defer r.d.mu.Unlock()
	b, ok := r.d.books[p.BookID]
	if !ok {
		return nil, apperror.NotFound("Book")
	}
	if _, ok := r.d.members[p.MemberID]; !ok {
		return nil, apperror.NotFound("Member")
	}
	if b.AvailableCopies <= 0 {
		return nil, apperror.Unavailable("No copies of \"" + b.Title + "\" are currently available")
	}
	active := 0
	for _, l := range r.d.loans {
		if l.MemberID == p.MemberID && l.ReturnedAt == nil {
			active++
			if l.BookID == p.BookID {
				return nil, apperror.Conflict("This member already has this book on loan")
			}
		}
	}
	if p.MaxActive > 0 && active >= p.MaxActive {
		return nil, apperror.Rule("Member has reached the maximum number of active loans")
	}
	b.AvailableCopies--
	b.Normalize()
	l := &models.Loan{ID: r.d.nextLoan, BookID: p.BookID, MemberID: p.MemberID, IssuedAt: p.IssuedAt, DueDate: p.DueDate}
	r.d.nextLoan++
	r.d.loans[l.ID] = l
	c := r.hydrate(l)
	return &c, nil
}

func (r *memLoans) Return(_ context.Context, id int64, at time.Time, finePerDay float64) (*models.Loan, error) {
	r.d.mu.Lock()
	defer r.d.mu.Unlock()
	l, ok := r.d.loans[id]
	if !ok {
		return nil, apperror.NotFound("Loan")
	}
	if l.ReturnedAt != nil {
		return nil, apperror.Conflict("This loan has already been returned")
	}
	l.ReturnedAt = &at
	l.Fine = models.CalcFine(l.DueDate, at, finePerDay)
	if b, ok := r.d.books[l.BookID]; ok {
		b.AvailableCopies++
		b.Normalize()
	}
	c := r.hydrate(l)
	return &c, nil
}

func (r *memLoans) Renew(_ context.Context, id int64, newDue time.Time, maxRenewals int) (*models.Loan, error) {
	r.d.mu.Lock()
	defer r.d.mu.Unlock()
	l, ok := r.d.loans[id]
	if !ok {
		return nil, apperror.NotFound("Loan")
	}
	if l.ReturnedAt != nil {
		return nil, apperror.Conflict("This loan has already been returned")
	}
	if l.Renewals >= maxRenewals {
		return nil, apperror.Rule("Maximum number of renewals reached")
	}
	l.DueDate = newDue
	l.Renewals++
	c := r.hydrate(l)
	return &c, nil
}

func (r *memLoans) GetByID(_ context.Context, id int64) (*models.Loan, error) {
	r.d.mu.RLock()
	defer r.d.mu.RUnlock()
	l, ok := r.d.loans[id]
	if !ok {
		return nil, apperror.NotFound("Loan")
	}
	c := r.hydrate(l)
	return &c, nil
}

func (r *memLoans) List(_ context.Context, f models.LoanFilter) ([]models.Loan, error) {
	r.d.mu.RLock()
	defer r.d.mu.RUnlock()
	q := strings.ToLower(f.Query)
	out := []models.Loan{}
	for _, l := range r.d.loans {
		if f.BookID != 0 && l.BookID != f.BookID {
			continue
		}
		if f.MemberID != 0 && l.MemberID != f.MemberID {
			continue
		}
		switch f.Status {
		case "active":
			if l.ReturnedAt != nil {
				continue
			}
		case "overdue":
			if l.ReturnedAt != nil || !f.Now.After(l.DueDate) {
				continue
			}
		case "returned":
			if l.ReturnedAt == nil {
				continue
			}
		}
		c := r.hydrate(l)
		if q != "" && !strings.Contains(strings.ToLower(c.BookTitle+" "+c.MemberName), q) {
			continue
		}
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].IssuedAt.After(out[j].IssuedAt) || (out[i].IssuedAt.Equal(out[j].IssuedAt) && out[i].ID > out[j].ID)
	})
	if f.Limit > 0 && len(out) > f.Limit {
		out = out[:f.Limit]
	}
	return out, nil
}

func (r *memLoans) CountActive(_ context.Context) (int, error) {
	r.d.mu.RLock()
	defer r.d.mu.RUnlock()
	n := 0
	for _, l := range r.d.loans {
		if l.ReturnedAt == nil {
			n++
		}
	}
	return n, nil
}

func (r *memLoans) CountOverdue(_ context.Context, now time.Time) (int, error) {
	r.d.mu.RLock()
	defer r.d.mu.RUnlock()
	n := 0
	for _, l := range r.d.loans {
		if l.ReturnedAt == nil && now.After(l.DueDate) {
			n++
		}
	}
	return n, nil
}
