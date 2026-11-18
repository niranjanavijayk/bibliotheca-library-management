package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"bibliotheca/internal/handler"
	"bibliotheca/internal/repository"
	"bibliotheca/internal/service"
)

func newServer() http.Handler {
	store := repository.NewMemoryStore()
	clock := service.Clock(time.Now)
	loans := service.NewLoanService(store.Loans(), store.Books(), store.Members(), service.DefaultConfig(), clock)
	api := handler.New(service.NewBookService(store.Books(), clock), service.NewMemberService(store.Members(), clock),
		loans, service.NewDashboardService(store, loans, clock))
	return api.Routes()
}

type resp struct {
	Data  json.RawMessage `json:"data"`
	Meta  map[string]int  `json:"meta"`
	Error *struct {
		Code   string            `json:"code"`
		Fields map[string]string `json:"fields"`
	} `json:"error"`
}

func call(t *testing.T, h http.Handler, method, path string, body any) (int, resp) {
	t.Helper()
	var rdr *bytes.Reader
	switch v := body.(type) {
	case nil:
		rdr = bytes.NewReader(nil)
	case string:
		rdr = bytes.NewReader([]byte(v))
	default:
		b, _ := json.Marshal(v)
		rdr = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rdr)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var r resp
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
			t.Fatalf("bad json: %v: %s", err, rec.Body.String())
		}
	}
	return rec.Code, r
}

func id(t *testing.T, r resp) int {
	t.Helper()
	var o struct{ ID int }
	if err := json.Unmarshal(r.Data, &o); err != nil || o.ID == 0 {
		t.Fatalf("no id in %s", r.Data)
	}
	return o.ID
}

func TestHealth(t *testing.T) {
	if c, _ := call(t, newServer(), "GET", "/api/health", nil); c != 200 {
		t.Fatal(c)
	}
}

func TestBookCRUDAndValidation(t *testing.T) {
	h := newServer()
	c, r := call(t, h, "POST", "/api/books", map[string]any{})
	if c != 422 || r.Error == nil || r.Error.Fields["title"] == "" {
		t.Fatalf("expected 422 with field errors, got %d %+v", c, r.Error)
	}
	if c, _ = call(t, h, "POST", "/api/books", "{not json"); c != 400 {
		t.Fatalf("expected 400 for bad json, got %d", c)
	}
	book := map[string]any{"title": "Dune", "author": "Frank Herbert", "category": "Sci-Fi", "isbn": "9780441172719", "totalCopies": 2, "publishedYear": 1965}
	c, r = call(t, h, "POST", "/api/books", book)
	if c != 201 {
		t.Fatalf("create: %d", c)
	}
	bid := id(t, r)
	if c, r = call(t, h, "POST", "/api/books", book); c != 422 || r.Error.Fields["isbn"] == "" {
		t.Fatalf("duplicate isbn: %d", c)
	}
	book["title"] = "Dune (Updated)"
	if c, _ = call(t, h, "PUT", "/api/books/"+itoa(bid), book); c != 200 {
		t.Fatalf("update: %d", c)
	}
	if c, _ = call(t, h, "GET", "/api/books/"+itoa(bid), nil); c != 200 {
		t.Fatalf("get: %d", c)
	}
	if c, _ = call(t, h, "GET", "/api/books/9999", nil); c != 404 {
		t.Fatalf("missing: %d", c)
	}
	if c, _ = call(t, h, "GET", "/api/books/abc", nil); c != 400 {
		t.Fatalf("bad id: %d", c)
	}
	if c, _ = call(t, h, "DELETE", "/api/books/"+itoa(bid), nil); c != 204 {
		t.Fatalf("delete: %d", c)
	}
}

func itoa(i int) string { return strconv.Itoa(i) }

func TestSearchFilterSortList(t *testing.T) {
	h := newServer()
	for _, b := range []map[string]any{
		{"title": "Beta", "author": "A", "category": "Fiction", "isbn": "1111111111", "totalCopies": 1},
		{"title": "Alpha", "author": "B", "category": "Fiction", "isbn": "2222222222", "totalCopies": 1},
		{"title": "Gamma", "author": "C", "category": "History", "isbn": "3333333333", "totalCopies": 1},
	} {
		call(t, h, "POST", "/api/books", b)
	}
	_, r := call(t, h, "GET", "/api/books?category=fiction&sort=title&order=asc", nil)
	var books []struct{ Title string }
	_ = json.Unmarshal(r.Data, &books)
	if len(books) != 2 || books[0].Title != "Alpha" || r.Meta["total"] != 2 {
		t.Fatalf("filter+sort wrong: %+v meta=%v", books, r.Meta)
	}
	_, r = call(t, h, "GET", "/api/books?q=gam", nil)
	_ = json.Unmarshal(r.Data, &books)
	if len(books) != 1 || books[0].Title != "Gamma" {
		t.Fatalf("search wrong: %+v", books)
	}
	_, r = call(t, h, "GET", "/api/books?available=false", nil)
	if r.Meta["total"] != 0 {
		t.Fatalf("available=false should be empty")
	}
}

func TestIssueReturnFlowAndDashboard(t *testing.T) {
	h := newServer()
	_, rb := call(t, h, "POST", "/api/books", map[string]any{"title": "Dune", "author": "FH", "category": "Sci-Fi", "isbn": "9780441172719", "totalCopies": 1})
	_, rm := call(t, h, "POST", "/api/members", map[string]any{"name": "Anu", "email": "anu@x.com", "phone": "9847011223"})
	_, rm2 := call(t, h, "POST", "/api/members", map[string]any{"name": "Bob", "email": "bob@x.com", "phone": "9847011224"})
	bid, mid, mid2 := id(t, rb), id(t, rm), id(t, rm2)

	c, rl := call(t, h, "POST", "/api/loans", map[string]any{"bookId": bid, "memberId": mid, "days": 7})
	if c != 201 {
		t.Fatalf("issue: %d %+v", c, rl.Error)
	}
	lid := id(t, rl)
	c, rl2 := call(t, h, "POST", "/api/loans", map[string]any{"bookId": bid, "memberId": mid2})
	if c != 409 || rl2.Error.Code != "BOOK_UNAVAILABLE" {
		t.Fatalf("expected 409 unavailable, got %d", c)
	}
	_, rd := call(t, h, "GET", "/api/dashboard", nil)
	var d struct{ TotalBooks, TotalMembers, IssuedBooks, OverdueBooks int }
	_ = json.Unmarshal(rd.Data, &d)
	if d.TotalBooks != 1 || d.TotalMembers != 2 || d.IssuedBooks != 1 {
		t.Fatalf("dashboard: %+v", d)
	}
	if c, _ = call(t, h, "DELETE", "/api/members/"+itoa(mid), nil); c != 409 {
		t.Fatalf("delete member with loan: %d", c)
	}
	if c, _ = call(t, h, "POST", "/api/loans/"+itoa(lid)+"/return", nil); c != 200 {
		t.Fatalf("return: %d", c)
	}
	if c, _ = call(t, h, "POST", "/api/loans/"+itoa(lid)+"/return", nil); c != 409 {
		t.Fatalf("double return: %d", c)
	}
	_, rh := call(t, h, "GET", "/api/books/"+itoa(bid)+"/history", nil)
	var hist []map[string]any
	_ = json.Unmarshal(rh.Data, &hist)
	if len(hist) != 1 || hist[0]["status"] != "returned" {
		t.Fatalf("history: %v", hist)
	}
	if c, _ = call(t, h, "GET", "/api/loans?status=bogus", nil); c != 400 {
		t.Fatalf("bad status filter: %d", c)
	}
}
