package httpx

import "database/sql"

type AccrueRequest struct {
	CustomerID     string `json:"customer_id"`
	OrderID        string `json:"order_id"`
	Points         int    `json:"points"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type ConfirmRequest struct {
	CustomerID     string `json:"customer_id"`
	OrderID        string `json:"order_id"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type RevokeRequest struct {
	CustomerID     string `json:"customer_id"`
	OrderID        string `json:"order_id"`
	ReturnID       string `json:"return_id"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type RedeemRequest struct {
	CustomerID     string `json:"customer_id"`
	Reference      string `json:"reference"`
	Points         int    `json:"points"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type RestoreRequest struct {
	CustomerID     string `json:"customer_id"`
	Reference      string `json:"reference"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type BalanceResponse struct {
	CustomerID    string `json:"customer_id"`
	ActivePoints  int    `json:"active_points"`
	PendingPoints int    `json:"pending_points"`
}

type TransactionResponse struct {
	ID         uint64  `json:"id"`
	CustomerID string  `json:"customer_id"`
	OrderID    *string `json:"order_id"`
	Reference  *string `json:"reference"`
	ReturnID   *string `json:"return_id"`
	Kind       string  `json:"kind"`
	Status     string  `json:"status"`
	Points     int     `json:"points"`
	OccurredAt string  `json:"occurred_at"`
}

func stringPtrFromNullString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}