package httpx

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"treuepunkte/internal/service"
)

type Handlers struct {
	Loyalty *service.LoyaltyService
}

func resolveIdempotencyKey(r *http.Request, bodyValue string) string {
	if bodyValue != "" {
		return bodyValue
	}
	return r.Header.Get("Idempotency-Key")
}

func (h *Handlers) PostAccrue(w http.ResponseWriter, r *http.Request) {
	var in AccrueRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	idemKey := resolveIdempotencyKey(r, in.IdempotencyKey)

	created, err := h.Loyalty.Accrue(r.Context(), in.CustomerID, in.OrderID, in.Points, idemKey)
	if err != nil {
		writeMappedError(w, err)
		return
	}

	writeSuccess(w, created)
}

func (h *Handlers) PostConfirm(w http.ResponseWriter, r *http.Request) {
	var in ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	idemKey := resolveIdempotencyKey(r, in.IdempotencyKey)

	created, err := h.Loyalty.Confirm(r.Context(), in.CustomerID, in.OrderID, idemKey)
	if err != nil {
		writeMappedError(w, err)
		return
	}

	writeSuccess(w, created)
}

func (h *Handlers) PostRevoke(w http.ResponseWriter, r *http.Request) {
	var in RevokeRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	idemKey := resolveIdempotencyKey(r, in.IdempotencyKey)

	created, err := h.Loyalty.Revoke(r.Context(), in.CustomerID, in.OrderID, in.ReturnID, idemKey)
	if err != nil {
		writeMappedError(w, err)
		return
	}

	writeSuccess(w, created)
}

func (h *Handlers) PostRedeem(w http.ResponseWriter, r *http.Request) {
	var in RedeemRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	idemKey := resolveIdempotencyKey(r, in.IdempotencyKey)

	created, err := h.Loyalty.Redeem(r.Context(), in.CustomerID, in.Reference, in.Points, idemKey)
	if err != nil {
		writeMappedError(w, err)
		return
	}

	writeSuccess(w, created)
}

func (h *Handlers) PostRestore(w http.ResponseWriter, r *http.Request) {
	var in RestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	idemKey := resolveIdempotencyKey(r, in.IdempotencyKey)

	created, err := h.Loyalty.Restore(r.Context(), in.CustomerID, in.Reference, idemKey)
	if err != nil {
		writeMappedError(w, err)
		return
	}

	writeSuccess(w, created)
}

func (h *Handlers) GetCustomerRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/customers/")
	parts := strings.Split(path, "/")

	if len(parts) == 2 && parts[0] != "" && parts[1] == "balance" {
		h.GetBalance(w, r, parts[0])
		return
	}

	if len(parts) == 2 && parts[0] != "" && parts[1] == "transactions" {
		h.GetTransactions(w, r, parts[0])
		return
	}

	writeError(w, http.StatusNotFound, "endpoint not found")
}

func (h *Handlers) GetBalance(w http.ResponseWriter, r *http.Request, customerID string) {
	balance, err := h.Loyalty.GetBalance(r.Context(), customerID)
	if err != nil {
		writeMappedError(w, err)
		return
	}

	resp := BalanceResponse{
		CustomerID:    balance.CustomerID,
		ActivePoints:  balance.ActivePoints,
		PendingPoints: balance.PendingPoints,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) GetTransactions(w http.ResponseWriter, r *http.Request, customerID string) {
	transactions, err := h.Loyalty.GetTransactions(r.Context(), customerID)
	if err != nil {
		writeMappedError(w, err)
		return
	}

	resp := make([]TransactionResponse, 0, len(transactions))
	for _, tx := range transactions {
		resp = append(resp, TransactionResponse{
			ID:         tx.ID,
			CustomerID: tx.CustomerID,
			OrderID:    stringPtrFromNullString(tx.OrderID),
			Reference:  stringPtrFromNullString(tx.Reference),
			ReturnID:   stringPtrFromNullString(tx.ReturnID),
			Kind:       string(tx.Kind),
			Status:     string(tx.Status),
			Points:     tx.Points,
			OccurredAt: tx.OccurredAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
