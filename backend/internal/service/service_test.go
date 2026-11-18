package service_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"bibliotheca/internal/apperror"
	"bibliotheca/internal/models"
	"bibliotheca/internal/repository"
	"bibliotheca/internal/service"
)

type env struct {
	books   service.BookService
	members service.MemberService
	loans   service.LoanService
	dash    service.DashboardService
	now     *time.Time
}

func newEnv() *env {
	store := repository.NewMemoryStore()
	now := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	cfg := service.DefaultConfig()
	loans := service.NewLoanService(store.Loans(), store.Books(), store.Members(), cfg, clock)
	return &env{
		books:   service.NewBookService(store.Books(), clock),
		members: service.NewMemberService(store.Members(), clock),
		loans:   loans,
		dash:    service.NewDashboardService(store, loans, clock),
		now:     &now,
	}
}

func bookIn(isbn string, copies int) models.BookInput {
	return models.BookInput{Title: "Go in Action", Author: "Kennedy", Category: "technology", ISBN: isbn, PublishedYear: 2015, TotalCopies: copies}
}

func memberIn(email string) models.MemberInput {
	return models.MemberInput{Name: "Test User", Email: email, Phone: "+91 98470 11223", MembershipType: "Standard"}
}

func code(err error) string {
	var ae *apperror.Error
	if errors.As(err, &ae) {
		return ae.Code
	}
	return ""
}

func mustBook(t *testing.T, e *env, isbn string, copies int) *models.Book {
	t.Helper()
	b, err := e.books.Create(context.Background(), bookIn(isbn, copies))
	if err != nil {
		t.Fatalf("create book: %v", err)
	}
	return b
}

func mustMember(t *testing.T, e *env, email string) *models.Member {
	t.Helper()
	m, err := e.members.Create(context.Background(), memberIn(email))
	if err != nil {
		t.Fatalf("create member: %v", err)
	}
	return m
}

// ---------- Books ----------

func TestBookValidation(t *testing.T) {
	e := newEnv()
	_, err := e.books.Create(context.Background(), models.BookInput{})
	var ae *apperror.Error
	if !errors.As(err, &ae) || ae.Code != "VALIDATION_FAILED" {
		t.Fatalf("expected validation error, got %v", err)
	}
	for _, f := range []string{"title", "author", "category", "isbn", "totalCopies"} {
		if ae.Fields[f] == "" {
			t.Errorf("expected field error for %s", f)
		}
	}
}

func TestBookCreateNormalisesAndDefaultsCover(t *testing.T) {
	e := newEnv()
	b := mustBook(t, e, "978-0-13-419044-0", 3)
	if b.ISBN != "9780134190440" {
		t.Errorf("isbn not normalised: %s", b.ISBN)
	}
	if b.Category != "Technology" {
		t.Errorf("category not title-cased: %s", b.Category)
	}
	if b.CoverURL == "" || b.AvailableCopies != 3 || !b.Available {
		t.Errorf("unexpected defaults: %+v", b)
	}
}

func TestBookDuplicateISBN(t *testing.T) {
	e := newEnv()
	mustBook(t, e, "9780134190440", 1)
	_, err := e.books.Create(context.Background(), bookIn("978-0-13-419044-0", 1))
	if code(err) != "VALIDATION_FAILED" {
		t.Fatalf("expected duplicate ISBN rejection, got %v", err)
	}
}

func TestBookUpdateCannotDropBelowIssued(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	b := mustBook(t, e, "9780134190440", 3)
	m1, m2 := mustMember(t, e, "a@x.com"), mustMember(t, e, "b@x.com")
	for _, m := range []*models.Member{m1, m2} {
		if _, err := e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: m.ID}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := e.books.Update(ctx, b.ID, bookIn("9780134190440", 1)); err == nil {
		t.Fatal("expected error reducing copies below issued")
	}
	up, err := e.books.Update(ctx, b.ID, bookIn("9780134190440", 5))
	if err != nil || up.AvailableCopies != 3 {
		t.Fatalf("expected 3 available after growing to 5 with 2 issued, got %+v %v", up, err)
	}
}

func TestBookDeleteBlockedWhileIssued(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	b, m := mustBook(t, e, "9780134190440", 1), mustMember(t, e, "a@x.com")
	l, _ := e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: m.ID})
	if code(e.books.Delete(ctx, b.ID)) != "CONFLICT" {
		t.Fatal("expected conflict deleting an issued book")
	}
	_, _ = e.loans.Return(ctx, l.ID)
	if err := e.books.Delete(ctx, b.ID); err != nil {
		t.Fatalf("delete after return: %v", err)
	}
	if _, err := e.books.Get(ctx, b.ID); code(err) != "NOT_FOUND" {
		t.Fatal("expected not found after delete")
	}
}

func TestBookSearchFilterSort(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	for i, t0 := range []struct{ title, author, cat, isbn string }{
		{"Zebra Tales", "Ann", "Fiction", "1111111111"},
		{"Alpha Code", "Bob", "Technology", "2222222222"},
		{"Middle Earth", "Cid", "Fantasy", "3333333333"},
	} {
		_, err := e.books.Create(ctx, models.BookInput{Title: t0.title, Author: t0.author, Category: t0.cat, ISBN: t0.isbn, TotalCopies: 1 + i})
		if err != nil {
			t.Fatal(err)
		}
	}
	got, total, _ := e.books.List(ctx, models.BookFilter{SortBy: "title", Order: "asc"})
	if total != 3 || got[0].Title != "Alpha Code" || got[2].Title != "Zebra Tales" {
		t.Errorf("sort asc wrong: %+v", got)
	}
	got, _, _ = e.books.List(ctx, models.BookFilter{SortBy: "title", Order: "desc"})
	if got[0].Title != "Zebra Tales" {
		t.Errorf("sort desc wrong")
	}
	got, total, _ = e.books.List(ctx, models.BookFilter{Query: "bob"})
	if total != 1 || got[0].Title != "Alpha Code" {
		t.Errorf("search by author failed: %+v", got)
	}
	got, total, _ = e.books.List(ctx, models.BookFilter{Category: "fantasy"})
	if total != 1 || got[0].Title != "Middle Earth" {
		t.Errorf("category filter failed")
	}
	got, total, _ = e.books.List(ctx, models.BookFilter{PageSize: 2, Page: 2, SortBy: "title"})
	if total != 3 || len(got) != 1 {
		t.Errorf("pagination wrong: total=%d len=%d", total, len(got))
	}
}

// ---------- Members ----------

func TestMemberValidation(t *testing.T) {
	e := newEnv()
	in := models.MemberInput{Name: "X", Email: "not-an-email", Phone: "abc"}
	_, err := e.members.Create(context.Background(), in)
	var ae *apperror.Error
	if !errors.As(err, &ae) || ae.Fields["email"] == "" || ae.Fields["phone"] == "" {
		t.Fatalf("expected email and phone errors, got %v", err)
	}
}

func TestMemberDuplicateEmailCaseInsensitive(t *testing.T) {
	e := newEnv()
	mustMember(t, e, "dup@x.com")
	if _, err := e.members.Create(context.Background(), memberIn("DUP@x.com")); code(err) != "VALIDATION_FAILED" {
		t.Fatalf("expected duplicate email rejection, got %v", err)
	}
}

func TestMemberDeleteBlockedWithActiveLoan(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	b, m := mustBook(t, e, "9780134190440", 1), mustMember(t, e, "a@x.com")
	l, _ := e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: m.ID})
	if code(e.members.Delete(ctx, m.ID)) != "CONFLICT" {
		t.Fatal("expected conflict")
	}
	_, _ = e.loans.Return(ctx, l.ID)
	if err := e.members.Delete(ctx, m.ID); err != nil {
		t.Fatal(err)
	}
}

// ---------- Loans ----------

func TestIssueAndReturnManageAvailability(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	b, m := mustBook(t, e, "9780134190440", 1), mustMember(t, e, "a@x.com")
	l, err := e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: m.ID, Days: 7})
	if err != nil {
		t.Fatal(err)
	}
	if want := e.now.AddDate(0, 0, 7); !l.DueDate.Equal(want) {
		t.Errorf("due date %v want %v", l.DueDate, want)
	}
	got, _ := e.books.Get(ctx, b.ID)
	if got.Available || got.AvailableCopies != 0 {
		t.Fatal("book should be unavailable")
	}
	m2 := mustMember(t, e, "b@x.com")
	if _, err := e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: m2.ID}); code(err) != "BOOK_UNAVAILABLE" {
		t.Fatalf("expected BOOK_UNAVAILABLE, got %v", err)
	}
	ret, err := e.loans.Return(ctx, l.ID)
	if err != nil || ret.Status != models.StatusReturned {
		t.Fatalf("return failed: %v %+v", err, ret)
	}
	got, _ = e.books.Get(ctx, b.ID)
	if !got.Available || got.AvailableCopies != 1 {
		t.Fatal("book should be available again")
	}
	if _, err := e.loans.Return(ctx, l.ID); code(err) != "CONFLICT" {
		t.Fatalf("double return must conflict, got %v", err)
	}
}

func TestIssueRules(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	b := mustBook(t, e, "9780134190440", 5)
	m := mustMember(t, e, "a@x.com")
	if _, err := e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: m.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: m.ID}); code(err) != "CONFLICT" {
		t.Errorf("same member+book twice should conflict, got %v", err)
	}
	if _, err := e.loans.Issue(ctx, models.IssueInput{BookID: 999, MemberID: m.ID}); code(err) != "NOT_FOUND" {
		t.Errorf("unknown book should be NOT_FOUND, got %v", err)
	}
	if _, err := e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: m.ID, Days: 400}); code(err) != "VALIDATION_FAILED" {
		t.Errorf("bad days should fail validation")
	}
	in := memberIn("sus@x.com")
	in.Status = models.MemberSuspended
	sus, _ := e.members.Create(ctx, in)
	if _, err := e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: sus.ID}); code(err) != "RULE_VIOLATION" {
		t.Errorf("suspended member should be blocked")
	}
}

func TestMaxLoansPerMember(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	m := mustMember(t, e, "a@x.com")
	for i := 0; i < 5; i++ {
		b := mustBook(t, e, fmt.Sprintf("97801341900%02d", i), 1)
		if _, err := e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: m.ID}); err != nil {
			t.Fatalf("loan %d: %v", i, err)
		}
	}
	b := mustBook(t, e, "9780134199999", 1)
	if _, err := e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: m.ID}); code(err) != "RULE_VIOLATION" {
		t.Fatalf("6th loan should be refused, got %v", err)
	}
}

func TestOverdueAndFine(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	b, m := mustBook(t, e, "9780134190440", 1), mustMember(t, e, "a@x.com")
	l, _ := e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: m.ID, Days: 14})
	*e.now = e.now.AddDate(0, 0, 17) // 3 days late

	over, _ := e.loans.List(ctx, models.LoanFilter{Status: "overdue"})
	if len(over) != 1 || over[0].Status != models.StatusOverdue || over[0].DaysOverdue != 3 || over[0].Fine != 6 {
		t.Fatalf("overdue list wrong: %+v", over)
	}
	if _, err := e.loans.Renew(ctx, l.ID); code(err) != "RULE_VIOLATION" {
		t.Errorf("overdue loans must not be renewable, got %v", err)
	}
	ret, _ := e.loans.Return(ctx, l.ID)
	if ret.Fine != 6 || ret.DaysOverdue != 3 {
		t.Errorf("final fine wrong: %+v", ret)
	}
	if over, _ = e.loans.List(ctx, models.LoanFilter{Status: "overdue"}); len(over) != 0 {
		t.Error("returned loan must not be overdue")
	}
}

func TestRenewLimit(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	b, m := mustBook(t, e, "9780134190440", 1), mustMember(t, e, "a@x.com")
	l, _ := e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: m.ID})
	due := l.DueDate
	for i := 0; i < 2; i++ {
		r, err := e.loans.Renew(ctx, l.ID)
		if err != nil {
			t.Fatal(err)
		}
		due = due.AddDate(0, 0, 7)
		if !r.DueDate.Equal(due) {
			t.Errorf("due date after renew %d: %v want %v", i+1, r.DueDate, due)
		}
	}
	if _, err := e.loans.Renew(ctx, l.ID); code(err) != "RULE_VIOLATION" {
		t.Fatalf("third renewal should be refused, got %v", err)
	}
}

func TestHistory(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	b, m := mustBook(t, e, "9780134190440", 1), mustMember(t, e, "a@x.com")
	l, _ := e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: m.ID})
	_, _ = e.loans.Return(ctx, l.ID)
	_, _ = e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: m.ID})
	bh, err := e.loans.BookHistory(ctx, b.ID)
	mh, err2 := e.loans.MemberHistory(ctx, m.ID)
	if err != nil || err2 != nil || len(bh) != 2 || len(mh) != 2 {
		t.Fatalf("history wrong: %v %v %d %d", err, err2, len(bh), len(mh))
	}
	if _, err := e.loans.BookHistory(ctx, 777); code(err) != "NOT_FOUND" {
		t.Error("history for unknown book should be NOT_FOUND")
	}
}

// Concurrency: many members race for the single copy of one book; exactly one wins.
func TestConcurrentIssueNeverOversells(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	b := mustBook(t, e, "9780134190440", 1)
	const n = 40
	ids := make([]int64, n)
	for i := range ids {
		ids[i] = mustMember(t, e, fmt.Sprintf("u%d@x.com", i)).ID
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	success, unavailable := 0, 0
	for _, id := range ids {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			_, err := e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: id})
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				success++
			} else if code(err) == "BOOK_UNAVAILABLE" {
				unavailable++
			}
		}(id)
	}
	wg.Wait()
	if success != 1 || unavailable != n-1 {
		t.Fatalf("expected 1 success and %d unavailable, got %d/%d", n-1, success, unavailable)
	}
	got, _ := e.books.Get(ctx, b.ID)
	if got.AvailableCopies != 0 {
		t.Fatalf("available copies = %d, want 0", got.AvailableCopies)
	}
}

// ---------- Dashboard ----------

func TestDashboardReflectsChanges(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	d, err := e.dash.Get(ctx)
	if err != nil || d.TotalBooks != 0 || d.TotalMembers != 0 || d.IssuedBooks != 0 {
		t.Fatalf("empty dashboard wrong: %+v %v", d, err)
	}
	b, m := mustBook(t, e, "9780134190440", 3), mustMember(t, e, "a@x.com")
	l, _ := e.loans.Issue(ctx, models.IssueInput{BookID: b.ID, MemberID: m.ID, Days: 1})
	d, _ = e.dash.Get(ctx)
	if d.TotalBooks != 1 || d.TotalCopies != 3 || d.AvailableCopies != 2 || d.TotalMembers != 1 || d.IssuedBooks != 1 || d.OverdueBooks != 0 {
		t.Fatalf("after issue: %+v", d)
	}
	*e.now = e.now.AddDate(0, 0, 3)
	d, _ = e.dash.Get(ctx)
	if d.OverdueBooks != 1 {
		t.Fatalf("expected 1 overdue, got %d", d.OverdueBooks)
	}
	_, _ = e.loans.Return(ctx, l.ID)
	d, _ = e.dash.Get(ctx)
	if d.IssuedBooks != 0 || d.OverdueBooks != 0 || d.AvailableCopies != 3 || len(d.RecentLoans) != 1 {
		t.Fatalf("after return: %+v", d)
	}
}
