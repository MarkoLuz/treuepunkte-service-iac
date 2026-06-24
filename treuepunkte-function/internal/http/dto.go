package httpx

import "database/sql"

type AccrueRequest struct {
	CustomerID       string `json:"customer_id"`
	OrderID          string `json:"order_id"`
	Home24MerchCents int    `json:"home24_merch_cents"`
	MiraklMerchCents int    `json:"mirakl_merch_cents"`
	ShippingCents    int    `json:"shipping_cents"`
	Currency         string `json:"currency"`
	IdempotencyKey   string `json:"idempotency_key,omitempty"`
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
	ID               uint64  `json:"id"`
	CustomerID       string  `json:"customer_id"`
	OrderID          *string `json:"order_id"`
	Reference        *string `json:"reference"`
	ReturnID         *string `json:"return_id"`
	Kind             string  `json:"kind"`
	Status           string  `json:"status"`
	Points           int     `json:"points"`
	Home24MerchCents *int64  `json:"home24_merch_cents"`
	MiraklMerchCents *int64  `json:"mirakl_merch_cents"`
	OrderTotalCents  *int64  `json:"order_total_cents"`
	ShippingCents    *int64  `json:"shipping_cents"`
	Currency         *string `json:"currency"`
	OccurredAt       string  `json:"occurred_at"`
}

func stringPtrFromNullString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func int64PtrFromNullInt64(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}
