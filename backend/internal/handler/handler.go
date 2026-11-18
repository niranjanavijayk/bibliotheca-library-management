package handler

import (
	"net/http"

	"bibliotheca/internal/models"
	"bibliotheca/internal/service"
)

// API wires services to HTTP routes.
type API struct {
	books     service.BookService
	members   service.MemberService
	loans     service.LoanService
	dashboard service.DashboardService
}

func New(b service.BookService, m service.MemberService, l service.LoanService, d service.DashboardService) *API {
	return &API{books: b, members: m, loans: l, dashboard: d}
}

// Routes registers every endpoint using Go 1.22 method+pattern routing.
func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) { ok(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /api/dashboard", a.getDashboard)

	mux.HandleFunc("GET /api/books", a.listBooks)
	mux.HandleFunc("POST /api/books", a.createBook)
	mux.HandleFunc("GET /api/books/{id}", a.getBook)
	mux.HandleFunc("PUT /api/books/{id}", a.updateBook)
	mux.HandleFunc("DELETE /api/books/{id}", a.deleteBook)
	mux.HandleFunc("GET /api/books/{id}/history", a.bookHistory)
	mux.HandleFunc("GET /api/categories", a.categories)

	mux.HandleFunc("GET /api/members", a.listMembers)
	mux.HandleFunc("POST /api/members", a.createMember)
	mux.HandleFunc("GET /api/members/{id}", a.getMember)
	mux.HandleFunc("PUT /api/members/{id}", a.updateMember)
	mux.HandleFunc("DELETE /api/members/{id}", a.deleteMember)
	mux.HandleFunc("GET /api/members/{id}/history", a.memberHistory)

	mux.HandleFunc("GET /api/loans", a.listLoans)
	mux.HandleFunc("POST /api/loans", a.issueLoan)
	mux.HandleFunc("GET /api/loans/{id}", a.getLoan)
	mux.HandleFunc("POST /api/loans/{id}/return", a.returnLoan)
	mux.HandleFunc("POST /api/loans/{id}/renew", a.renewLoan)
	return mux
}

func (a *API) getDashboard(w http.ResponseWriter, r *http.Request) {
	d, err := a.dashboard.Get(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	ok(w, 200, d)
}

// ---- Books ----

func (a *API) listBooks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := models.BookFilter{
		Query: q.Get("q"), Category: q.Get("category"), Available: boolParam(r, "available"),
		SortBy: q.Get("sort"), Order: q.Get("order"),
		Page: intParam(r, "page", 1, 1, 1_000_000), PageSize: intParam(r, "pageSize", 12, 1, 100),
	}
	books, total, err := a.books.List(r.Context(), f)
	if err != nil {
		fail(w, err)
		return
	}
	okList(w, books, total, f.Page, f.PageSize)
}

func (a *API) createBook(w http.ResponseWriter, r *http.Request) {
	var in models.BookInput
	if err := decode(r, &in); err != nil {
		fail(w, err)
		return
	}
	b, err := a.books.Create(r.Context(), in)
	if err != nil {
		fail(w, err)
		return
	}
	ok(w, 201, b)
}

func (a *API) getBook(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		fail(w, err)
		return
	}
	b, err := a.books.Get(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	ok(w, 200, b)
}

func (a *API) updateBook(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		fail(w, err)
		return
	}
	var in models.BookInput
	if err := decode(r, &in); err != nil {
		fail(w, err)
		return
	}
	b, err := a.books.Update(r.Context(), id, in)
	if err != nil {
		fail(w, err)
		return
	}
	ok(w, 200, b)
}

func (a *API) deleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		fail(w, err)
		return
	}
	if err := a.books.Delete(r.Context(), id); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) bookHistory(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		fail(w, err)
		return
	}
	h, err := a.loans.BookHistory(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	ok(w, 200, h)
}

func (a *API) categories(w http.ResponseWriter, r *http.Request) {
	c, err := a.books.Categories(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	ok(w, 200, c)
}

// ---- Members ----

func (a *API) listMembers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := models.MemberFilter{
		Query: q.Get("q"), Status: q.Get("status"), SortBy: q.Get("sort"), Order: q.Get("order"),
		Page: intParam(r, "page", 1, 1, 1_000_000), PageSize: intParam(r, "pageSize", 12, 1, 100),
	}
	ms, total, err := a.members.List(r.Context(), f)
	if err != nil {
		fail(w, err)
		return
	}
	okList(w, ms, total, f.Page, f.PageSize)
}

func (a *API) createMember(w http.ResponseWriter, r *http.Request) {
	var in models.MemberInput
	if err := decode(r, &in); err != nil {
		fail(w, err)
		return
	}
	m, err := a.members.Create(r.Context(), in)
	if err != nil {
		fail(w, err)
		return
	}
	ok(w, 201, m)
}

func (a *API) getMember(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		fail(w, err)
		return
	}
	m, err := a.members.Get(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	ok(w, 200, m)
}

func (a *API) updateMember(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		fail(w, err)
		return
	}
	var in models.MemberInput
	if err := decode(r, &in); err != nil {
		fail(w, err)
		return
	}
	m, err := a.members.Update(r.Context(), id, in)
	if err != nil {
		fail(w, err)
		return
	}
	ok(w, 200, m)
}

func (a *API) deleteMember(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		fail(w, err)
		return
	}
	if err := a.members.Delete(r.Context(), id); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) memberHistory(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		fail(w, err)
		return
	}
	h, err := a.loans.MemberHistory(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	ok(w, 200, h)
}

// ---- Loans ----

func (a *API) listLoans(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	loans, err := a.loans.List(r.Context(), models.LoanFilter{
		Status: q.Get("status"), Query: q.Get("q"), Limit: intParam(r, "limit", 0, 0, 500),
	})
	if err != nil {
		fail(w, err)
		return
	}
	ok(w, 200, loans)
}

func (a *API) issueLoan(w http.ResponseWriter, r *http.Request) {
	var in models.IssueInput
	if err := decode(r, &in); err != nil {
		fail(w, err)
		return
	}
	l, err := a.loans.Issue(r.Context(), in)
	if err != nil {
		fail(w, err)
		return
	}
	ok(w, 201, l)
}

func (a *API) getLoan(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		fail(w, err)
		return
	}
	l, err := a.loans.Get(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	ok(w, 200, l)
}

func (a *API) returnLoan(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		fail(w, err)
		return
	}
	l, err := a.loans.Return(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	ok(w, 200, l)
}

func (a *API) renewLoan(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		fail(w, err)
		return
	}
	l, err := a.loans.Renew(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	ok(w, 200, l)
}
