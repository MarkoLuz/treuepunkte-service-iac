package service

import (
	"context"
	"testing"
)

func TestRedeem_EmptyCustomerID(t *testing.T) {
	s := &LoyaltyService{}

	_, err := s.Redeem(context.Background(), "", "ref-1", 100, "")

	if err == nil {
		t.Fatal("expected error for empty customer_id, got nil")
	}
}