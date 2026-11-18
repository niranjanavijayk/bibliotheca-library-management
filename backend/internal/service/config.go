// Package service contains the business logic. Handlers stay thin and delegate here.
package service

import "time"

// Config holds the library's lending rules.
type Config struct {
	LoanDays    int     // default loan period
	MaxLoanDays int     // upper bound a librarian may choose
	MaxLoans    int     // max concurrent loans per member
	MaxRenewals int     // max renewals per loan
	RenewDays   int     // days added per renewal
	FinePerDay  float64 // late fee per overdue day
}

func DefaultConfig() Config {
	return Config{LoanDays: 14, MaxLoanDays: 60, MaxLoans: 5, MaxRenewals: 2, RenewDays: 7, FinePerDay: 2}
}

// Clock allows time to be controlled in tests.
type Clock func() time.Time
