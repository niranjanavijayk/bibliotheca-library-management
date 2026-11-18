package service

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"bibliotheca/internal/apperror"
)

var (
	emailRe = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)
	phoneRe = regexp.MustCompile(`^\+?[0-9][0-9 \-]{6,18}[0-9]$`)
	isbnRe  = regexp.MustCompile(`^(\d{13}|\d{9}[\dX])$`)
)

type fieldErrors map[string]string

func (f fieldErrors) add(field, msg string) {
	if _, ok := f[field]; !ok {
		f[field] = msg
	}
}

func (f fieldErrors) err() error {
	if len(f) == 0 {
		return nil
	}
	return apperror.Validation(f)
}

func (f fieldErrors) required(field, label, value string, max int) string {
	v := strings.TrimSpace(value)
	if v == "" {
		f.add(field, label+" is required")
	} else if len([]rune(v)) > max {
		f.add(field, fmt.Sprintf("%s must be at most %d characters", label, max))
	}
	return v
}

// normalizeISBN strips separators and upper-cases the check digit.
func normalizeISBN(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.NewReplacer("-", "", " ", "").Replace(s)
	return s
}

func validISBN(s string) bool { return isbnRe.MatchString(s) }

func validYear(y int, now time.Time) bool { return y == 0 || (y >= 1000 && y <= now.Year()+1) }

func titleCase(s string) string {
	r := []rune(strings.ToLower(strings.TrimSpace(s)))
	up := true
	for i, c := range r {
		if up && c != ' ' && c != '-' {
			r[i] = []rune(strings.ToUpper(string(c)))[0]
			up = false
		}
		if c == ' ' || c == '-' {
			up = true
		}
	}
	return strings.Join(strings.Fields(string(r)), " ")
}
