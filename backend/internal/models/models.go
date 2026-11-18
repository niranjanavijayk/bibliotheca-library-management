// Package models holds the domain entities and input DTOs.
package models

import (
	"math"
	"time"
)

const (
	StatusIssued   = "issued"
	StatusOverdue  = "overdue"
	StatusReturned = "returned"

	MemberActive    = "active"
	MemberSuspended = "suspended"
)

var MembershipTypes = []string{"Standard", "Premium", "Student"}

type Book struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Author          string    `json:"author"`
	Category        string    `json:"category"`
	ISBN            string    `json:"isbn"`
	Description     string    `json:"description"`
	CoverURL        string    `json:"coverUrl"`
	PublishedYear   int       `json:"publishedYear"`
	TotalCopies     int       `json:"totalCopies"`
	AvailableCopies int       `json:"availableCopies"`
	Available       bool      `json:"available"`
	CreatedAt       time.Time `json:"createdAt"`
}

// Normalize derives computed fields.
func (b *Book) Normalize() { b.Available = b.AvailableCopies > 0 }

type Member struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	Address        string    `json:"address"`
	MembershipType string    `json:"membershipType"`
	Status         string    `json:"status"`
	JoinedAt       time.Time `json:"joinedAt"`
	ExpiresAt      time.Time `json:"expiresAt"`
	ActiveLoans    int       `json:"activeLoans"`
}

type Loan struct {
	ID          int64      `json:"id"`
	BookID      int64      `json:"bookId"`
	MemberID    int64      `json:"memberId"`
	BookTitle   string     `json:"bookTitle"`
	BookAuthor  string     `json:"bookAuthor"`
	BookISBN    string     `json:"bookIsbn"`
	BookCover   string     `json:"bookCover"`
	MemberName  string     `json:"memberName"`
	IssuedAt    time.Time  `json:"issuedAt"`
	DueDate     time.Time  `json:"dueDate"`
	ReturnedAt  *time.Time `json:"returnedAt"`
	Fine        float64    `json:"fine"`
	Renewals    int        `json:"renewals"`
	Status      string     `json:"status"`
	DaysOverdue int        `json:"daysOverdue"`
}

// OverdueDays returns whole days (rounded up) between due and at; 0 if not late.
func OverdueDays(due, at time.Time) int {
	if !at.After(due) {
		return 0
	}
	return int(math.Ceil(at.Sub(due).Hours() / 24))
}

// CalcFine computes the late fee for a loan.
func CalcFine(due, at time.Time, perDay float64) float64 {
	return float64(OverdueDays(due, at)) * perDay
}

// Refresh recomputes status, overdue days and the projected fine at time now.
func (l *Loan) Refresh(now time.Time, finePerDay float64) {
	switch {
	case l.ReturnedAt != nil:
		l.Status = StatusReturned
		l.DaysOverdue = OverdueDays(l.DueDate, *l.ReturnedAt)
	case now.After(l.DueDate):
		l.Status = StatusOverdue
		l.DaysOverdue = OverdueDays(l.DueDate, now)
		l.Fine = CalcFine(l.DueDate, now, finePerDay)
	default:
		l.Status = StatusIssued
		l.DaysOverdue = 0
	}
}

type CategoryCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type BookStats struct {
	Titles    int
	Copies    int
	Available int
}

type Dashboard struct {
	TotalBooks      int             `json:"totalBooks"`
	TotalCopies     int             `json:"totalCopies"`
	AvailableCopies int             `json:"availableCopies"`
	TotalMembers    int             `json:"totalMembers"`
	IssuedBooks     int             `json:"issuedBooks"`
	OverdueBooks    int             `json:"overdueBooks"`
	Categories      []CategoryCount `json:"categories"`
	RecentLoans     []Loan          `json:"recentLoans"`
}

// ---- Inputs ----

type BookInput struct {
	Title         string `json:"title"`
	Author        string `json:"author"`
	Category      string `json:"category"`
	ISBN          string `json:"isbn"`
	Description   string `json:"description"`
	CoverURL      string `json:"coverUrl"`
	PublishedYear int    `json:"publishedYear"`
	TotalCopies   int    `json:"totalCopies"`
}

type MemberInput struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Address        string `json:"address"`
	MembershipType string `json:"membershipType"`
	Status         string `json:"status"`
}

type IssueInput struct {
	BookID   int64 `json:"bookId"`
	MemberID int64 `json:"memberId"`
	Days     int   `json:"days"`
}

// ---- Filters ----

type BookFilter struct {
	Query     string
	Category  string
	Available *bool
	SortBy    string
	Order     string
	Page      int
	PageSize  int
}

type MemberFilter struct {
	Query    string
	Status   string
	SortBy   string
	Order    string
	Page     int
	PageSize int
}

type LoanFilter struct {
	Status   string // all | active | overdue | returned
	BookID   int64
	MemberID int64
	Query    string
	Limit    int
	Now      time.Time
}

type IssueParams struct {
	BookID    int64
	MemberID  int64
	IssuedAt  time.Time
	DueDate   time.Time
	MaxActive int
}
