package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"bibliotheca/internal/apperror"
	"bibliotheca/internal/models"
)

// PostgresStore implements Store on PostgreSQL.
type PostgresStore struct{ db *sql.DB }

func NewPostgresStore(db *sql.DB) *PostgresStore { return &PostgresStore{db: db} }

func (s *PostgresStore) Books() BookRepository     { return &pgBooks{s.db} }
func (s *PostgresStore) Members() MemberRepository { return &pgMembers{s.db} }
func (s *PostgresStore) Loans() LoanRepository     { return &pgLoans{s.db} }

func isUnique(err error) bool {
	var pe *pq.Error
	return errors.As(err, &pe) && pe.Code == "23505"
}

func orderDir(o string) string {
	if strings.EqualFold(o, "desc") {
		return "DESC"
	}
	return "ASC"
}

func pageArgs(page, size int) (limit, offset int) {
	if size <= 0 {
		size = 12
	}
	if page <= 0 {
		page = 1
	}
	return size, (page - 1) * size
}

// ---------------- Books ----------------

type pgBooks struct{ db *sql.DB }

const bookCols = `id, title, author, category, isbn, description, cover_url, published_year, total_copies, available_copies, created_at`

func scanBook(sc interface{ Scan(...any) error }) (*models.Book, error) {
	var b models.Book
	if err := sc.Scan(&b.ID, &b.Title, &b.Author, &b.Category, &b.ISBN, &b.Description, &b.CoverURL,
		&b.PublishedYear, &b.TotalCopies, &b.AvailableCopies, &b.CreatedAt); err != nil {
		return nil, err
	}
	b.Normalize()
	return &b, nil
}

func (r *pgBooks) Create(ctx context.Context, b *models.Book) error {
	err := r.db.QueryRowContext(ctx, `INSERT INTO books
		(title, author, category, isbn, description, cover_url, published_year, total_copies, available_copies)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8) RETURNING id, created_at`,
		b.Title, b.Author, b.Category, b.ISBN, b.Description, b.CoverURL, b.PublishedYear, b.TotalCopies,
	).Scan(&b.ID, &b.CreatedAt)
	if isUnique(err) {
		return apperror.Conflict("A book with this ISBN already exists")
	}
	if err != nil {
		return err
	}
	b.AvailableCopies = b.TotalCopies
	b.Normalize()
	return nil
}

func (r *pgBooks) Update(ctx context.Context, b *models.Book) error {
	row := r.db.QueryRowContext(ctx, `UPDATE books SET
		title=$2, author=$3, category=$4, isbn=$5, description=$6, cover_url=$7, published_year=$8,
		available_copies = available_copies + ($9 - total_copies), total_copies=$9
		WHERE id=$1 AND available_copies + ($9 - total_copies) >= 0
		RETURNING `+bookCols,
		b.ID, b.Title, b.Author, b.Category, b.ISBN, b.Description, b.CoverURL, b.PublishedYear, b.TotalCopies)
	nb, err := scanBook(row)
	if isUnique(err) {
		return apperror.Conflict("A book with this ISBN already exists")
	}
	if errors.Is(err, sql.ErrNoRows) {
		if _, gerr := r.GetByID(ctx, b.ID); gerr != nil {
			return gerr
		}
		return apperror.Conflict("Total copies cannot be lower than the number of copies currently issued")
	}
	if err != nil {
		return err
	}
	*b = *nb
	return nil
}

func (r *pgBooks) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM books WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return apperror.NotFound("Book")
	}
	return nil
}

func (r *pgBooks) GetByID(ctx context.Context, id int64) (*models.Book, error) {
	b, err := scanBook(r.db.QueryRowContext(ctx, `SELECT `+bookCols+` FROM books WHERE id=$1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NotFound("Book")
	}
	return b, err
}

var bookSort = map[string]string{
	"title": "LOWER(title)", "author": "LOWER(author)", "category": "LOWER(category)",
	"year": "published_year", "copies": "available_copies", "added": "created_at",
}

func (r *pgBooks) List(ctx context.Context, f models.BookFilter) ([]models.Book, int, error) {
	where := []string{"1=1"}
	args := []any{}
	if f.Query != "" {
		args = append(args, "%"+f.Query+"%")
		n := len(args)
		where = append(where, fmt.Sprintf("(title ILIKE $%d OR author ILIKE $%d OR isbn ILIKE $%d)", n, n, n))
	}
	if f.Category != "" {
		args = append(args, f.Category)
		where = append(where, fmt.Sprintf("LOWER(category) = LOWER($%d)", len(args)))
	}
	if f.Available != nil {
		if *f.Available {
			where = append(where, "available_copies > 0")
		} else {
			where = append(where, "available_copies = 0")
		}
	}
	w := strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM books WHERE `+w, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	col, ok := bookSort[f.SortBy]
	if !ok {
		col = "id"
	}
	limit, offset := pageArgs(f.Page, f.PageSize)
	args = append(args, limit, offset)
	q := fmt.Sprintf(`SELECT %s FROM books WHERE %s ORDER BY %s %s, id ASC LIMIT $%d OFFSET $%d`,
		bookCols, w, col, orderDir(f.Order), len(args)-1, len(args))
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []models.Book{}
	for rows.Next() {
		b, err := scanBook(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *b)
	}
	return out, total, rows.Err()
}

func (r *pgBooks) ISBNExists(ctx context.Context, isbn string, excludeID int64) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM books WHERE isbn=$1 AND id<>$2)`, isbn, excludeID).Scan(&ok)
	return ok, err
}

func (r *pgBooks) Stats(ctx context.Context) (models.BookStats, error) {
	var st models.BookStats
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(total_copies),0), COALESCE(SUM(available_copies),0) FROM books`,
	).Scan(&st.Titles, &st.Copies, &st.Available)
	return st, err
}

func (r *pgBooks) Categories(ctx context.Context) ([]models.CategoryCount, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT category, COUNT(*) FROM books GROUP BY category ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.CategoryCount{}
	for rows.Next() {
		var c models.CategoryCount
		if err := rows.Scan(&c.Name, &c.Count); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ---------------- Members ----------------

type pgMembers struct{ db *sql.DB }

const memberCols = `m.id, m.name, m.email, m.phone, m.address, m.membership_type, m.status, m.joined_at, m.expires_at,
	(SELECT COUNT(*) FROM loans l WHERE l.member_id = m.id AND l.returned_at IS NULL)`

func scanMember(sc interface{ Scan(...any) error }) (*models.Member, error) {
	var m models.Member
	err := sc.Scan(&m.ID, &m.Name, &m.Email, &m.Phone, &m.Address, &m.MembershipType, &m.Status,
		&m.JoinedAt, &m.ExpiresAt, &m.ActiveLoans)
	return &m, err
}

func (r *pgMembers) Create(ctx context.Context, m *models.Member) error {
	err := r.db.QueryRowContext(ctx, `INSERT INTO members
		(name, email, phone, address, membership_type, status, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, joined_at`,
		m.Name, m.Email, m.Phone, m.Address, m.MembershipType, m.Status, m.ExpiresAt,
	).Scan(&m.ID, &m.JoinedAt)
	if isUnique(err) {
		return apperror.Conflict("A member with this email already exists")
	}
	return err
}

func (r *pgMembers) Update(ctx context.Context, m *models.Member) error {
	res, err := r.db.ExecContext(ctx, `UPDATE members SET name=$2, email=$3, phone=$4, address=$5,
		membership_type=$6, status=$7, expires_at=$8 WHERE id=$1`,
		m.ID, m.Name, m.Email, m.Phone, m.Address, m.MembershipType, m.Status, m.ExpiresAt)
	if isUnique(err) {
		return apperror.Conflict("A member with this email already exists")
	}
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return apperror.NotFound("Member")
	}
	return nil
}

func (r *pgMembers) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM members WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return apperror.NotFound("Member")
	}
	return nil
}

func (r *pgMembers) GetByID(ctx context.Context, id int64) (*models.Member, error) {
	m, err := scanMember(r.db.QueryRowContext(ctx, `SELECT `+memberCols+` FROM members m WHERE m.id=$1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NotFound("Member")
	}
	return m, err
}

var memberSort = map[string]string{"name": "LOWER(m.name)", "joined": "m.joined_at", "loans": "7"}

func (r *pgMembers) List(ctx context.Context, f models.MemberFilter) ([]models.Member, int, error) {
	where := []string{"1=1"}
	args := []any{}
	if f.Query != "" {
		args = append(args, "%"+f.Query+"%")
		n := len(args)
		where = append(where, fmt.Sprintf("(m.name ILIKE $%d OR m.email ILIKE $%d OR m.phone ILIKE $%d)", n, n, n))
	}
	if f.Status != "" {
		args = append(args, f.Status)
		where = append(where, fmt.Sprintf("m.status = $%d", len(args)))
	}
	w := strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM members m WHERE `+w, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	col, ok := memberSort[f.SortBy]
	if !ok {
		col = "m.id"
	}
	if col == "7" {
		col = "(SELECT COUNT(*) FROM loans l WHERE l.member_id = m.id AND l.returned_at IS NULL)"
	}
	limit, offset := pageArgs(f.Page, f.PageSize)
	args = append(args, limit, offset)
	q := fmt.Sprintf(`SELECT %s FROM members m WHERE %s ORDER BY %s %s, m.id ASC LIMIT $%d OFFSET $%d`,
		memberCols, w, col, orderDir(f.Order), len(args)-1, len(args))
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []models.Member{}
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *m)
	}
	return out, total, rows.Err()
}

func (r *pgMembers) EmailExists(ctx context.Context, email string, excludeID int64) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM members WHERE LOWER(email)=LOWER($1) AND id<>$2)`,
		email, excludeID).Scan(&ok)
	return ok, err
}

func (r *pgMembers) Count(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM members`).Scan(&n)
	return n, err
}

// ---------------- Loans ----------------

type pgLoans struct{ db *sql.DB }

const loanSelect = `SELECT l.id, l.book_id, l.member_id, l.issued_at, l.due_date, l.returned_at, l.fine, l.renewals,
	b.title, b.author, b.isbn, b.cover_url, m.name
	FROM loans l JOIN books b ON b.id = l.book_id JOIN members m ON m.id = l.member_id`

func scanLoan(sc interface{ Scan(...any) error }) (*models.Loan, error) {
	var l models.Loan
	var ret sql.NullTime
	err := sc.Scan(&l.ID, &l.BookID, &l.MemberID, &l.IssuedAt, &l.DueDate, &ret, &l.Fine, &l.Renewals,
		&l.BookTitle, &l.BookAuthor, &l.BookISBN, &l.BookCover, &l.MemberName)
	if ret.Valid {
		t := ret.Time
		l.ReturnedAt = &t
	}
	return &l, err
}

func (r *pgLoans) GetByID(ctx context.Context, id int64) (*models.Loan, error) {
	l, err := scanLoan(r.db.QueryRowContext(ctx, loanSelect+` WHERE l.id=$1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NotFound("Loan")
	}
	return l, err
}

func (r *pgLoans) Issue(ctx context.Context, p models.IssueParams) (*models.Loan, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var title string
	var avail int
	err = tx.QueryRowContext(ctx, `SELECT title, available_copies FROM books WHERE id=$1 FOR UPDATE`, p.BookID).Scan(&title, &avail)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NotFound("Book")
	}
	if err != nil {
		return nil, err
	}
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM members WHERE id=$1)`, p.MemberID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, apperror.NotFound("Member")
	}
	if avail <= 0 {
		return nil, apperror.Unavailable("No copies of \"" + title + "\" are currently available")
	}
	var dup bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM loans WHERE book_id=$1 AND member_id=$2 AND returned_at IS NULL)`,
		p.BookID, p.MemberID).Scan(&dup); err != nil {
		return nil, err
	}
	if dup {
		return nil, apperror.Conflict("This member already has this book on loan")
	}
	if p.MaxActive > 0 {
		var active int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM loans WHERE member_id=$1 AND returned_at IS NULL`, p.MemberID).Scan(&active); err != nil {
			return nil, err
		}
		if active >= p.MaxActive {
			return nil, apperror.Rule("Member has reached the maximum number of active loans")
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE books SET available_copies = available_copies - 1 WHERE id=$1`, p.BookID); err != nil {
		return nil, err
	}
	var id int64
	if err := tx.QueryRowContext(ctx, `INSERT INTO loans (book_id, member_id, issued_at, due_date) VALUES ($1,$2,$3,$4) RETURNING id`,
		p.BookID, p.MemberID, p.IssuedAt, p.DueDate).Scan(&id); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *pgLoans) Return(ctx context.Context, id int64, at time.Time, finePerDay float64) (*models.Loan, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var bookID int64
	var due time.Time
	var ret sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT book_id, due_date, returned_at FROM loans WHERE id=$1 FOR UPDATE`, id).Scan(&bookID, &due, &ret)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NotFound("Loan")
	}
	if err != nil {
		return nil, err
	}
	if ret.Valid {
		return nil, apperror.Conflict("This loan has already been returned")
	}
	fine := models.CalcFine(due, at, finePerDay)
	if _, err := tx.ExecContext(ctx, `UPDATE loans SET returned_at=$2, fine=$3 WHERE id=$1`, id, at, fine); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE books SET available_copies = available_copies + 1 WHERE id=$1`, bookID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *pgLoans) Renew(ctx context.Context, id int64, newDue time.Time, maxRenewals int) (*models.Loan, error) {
	res, err := r.db.ExecContext(ctx, `UPDATE loans SET due_date=$2, renewals = renewals + 1
		WHERE id=$1 AND returned_at IS NULL AND renewals < $3`, id, newDue, maxRenewals)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		l, gerr := r.GetByID(ctx, id)
		if gerr != nil {
			return nil, gerr
		}
		if l.ReturnedAt != nil {
			return nil, apperror.Conflict("This loan has already been returned")
		}
		return nil, apperror.Rule("Maximum number of renewals reached")
	}
	return r.GetByID(ctx, id)
}

func (r *pgLoans) List(ctx context.Context, f models.LoanFilter) ([]models.Loan, error) {
	where := []string{"1=1"}
	args := []any{}
	if f.BookID != 0 {
		args = append(args, f.BookID)
		where = append(where, fmt.Sprintf("l.book_id=$%d", len(args)))
	}
	if f.MemberID != 0 {
		args = append(args, f.MemberID)
		where = append(where, fmt.Sprintf("l.member_id=$%d", len(args)))
	}
	switch f.Status {
	case "active":
		where = append(where, "l.returned_at IS NULL")
	case "overdue":
		args = append(args, f.Now)
		where = append(where, fmt.Sprintf("l.returned_at IS NULL AND l.due_date < $%d", len(args)))
	case "returned":
		where = append(where, "l.returned_at IS NOT NULL")
	}
	if f.Query != "" {
		args = append(args, "%"+f.Query+"%")
		n := len(args)
		where = append(where, fmt.Sprintf("(b.title ILIKE $%d OR m.name ILIKE $%d)", n, n))
	}
	q := loanSelect + " WHERE " + strings.Join(where, " AND ") + " ORDER BY l.issued_at DESC, l.id DESC"
	if f.Limit > 0 {
		args = append(args, f.Limit)
		q += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Loan{}
	for rows.Next() {
		l, err := scanLoan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *l)
	}
	return out, rows.Err()
}

func (r *pgLoans) CountActive(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM loans WHERE returned_at IS NULL`).Scan(&n)
	return n, err
}

func (r *pgLoans) CountOverdue(ctx context.Context, now time.Time) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM loans WHERE returned_at IS NULL AND due_date < $1`, now).Scan(&n)
	return n, err
}
