package domain

import (
	"database/sql"
	"time"
)

type LedgerKind string
type LedgerStatus string

const (
	KindAccrue  LedgerKind = "accrue"
	KindConfirm LedgerKind = "confirm"
	KindRevoke  LedgerKind = "revoke"
	KindRedeem  LedgerKind = "redeem"
	KindRestore LedgerKind = "restore"
)

const (
	StatusPending LedgerStatus = "pending"
	StatusActive  LedgerStatus = "active"
)

type Balance struct {
	CustomerID    string
	ActivePoints  int
	PendingPoints int
}

type Transaction struct {
	ID         uint64
	CustomerID string
	OrderID    sql.NullString
	Reference  sql.NullString
	ReturnID   sql.NullString
	Kind       LedgerKind
	Status     LedgerStatus
	Points     int
	OccurredAt time.Time
}