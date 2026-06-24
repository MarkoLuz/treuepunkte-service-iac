package service

import (
	"context"
	"database/sql"
	"fmt"
	"treuepunkte/internal/domain"
	"treuepunkte/internal/storage"
)

type LoyaltyService struct {
	DB   *sql.DB
	Repo *storage.Repository
}

type AccrueInput struct {
	CustomerID        string
	OrderID           string
	Home24MerchCents int
	MiraklMerchCents int
	ShippingCents    int
	Currency         string
	IdempotencyKey   string
}

func NewLoyaltyService(db *sql.DB) *LoyaltyService {
	return &LoyaltyService{
		DB:   db,
		Repo: storage.NewRepository(db),
	}
}

func (s *LoyaltyService) Accrue(ctx context.Context, input AccrueInput) (bool, error) {
	if input.CustomerID == "" {
		return false, fmt.Errorf("%w: customer_id is required", domain.ErrInvalidInput)
	}
	if input.OrderID == "" {
		return false, fmt.Errorf("%w: order_id is required", domain.ErrInvalidInput)
	}
	if input.Currency != "EUR" {
		return false, fmt.Errorf("%w: currency must be EUR", domain.ErrInvalidInput)
	}
	if input.Home24MerchCents < 0 {
		return false, fmt.Errorf("%w: home24_merch_cents must not be negative", domain.ErrInvalidInput)
	}
	if input.MiraklMerchCents < 0 {
		return false, fmt.Errorf("%w: mirakl_merch_cents must not be negative", domain.ErrInvalidInput)
	}
	if input.ShippingCents < 0 {
		return false, fmt.Errorf("%w: shipping_cents must not be negative", domain.ErrInvalidInput)
	}

	orderTotalCents := input.Home24MerchCents + input.MiraklMerchCents
	if orderTotalCents <= 0 {
		return false, fmt.Errorf("%w: merchandise total must be greater than 0", domain.ErrInvalidInput)
	}

	home24Points := input.Home24MerchCents * 10 / 100
	miraklPoints := input.MiraklMerchCents * 5 / 100
	points := home24Points + miraklPoints

	if points <= 0 {
		return false, fmt.Errorf("%w: calculated points must be greater than 0", domain.ErrInvalidInput)
	}

	return s.Repo.AccruePoints(
		ctx,
		input.CustomerID,
		input.OrderID,
		points,
		input.Home24MerchCents,
		input.MiraklMerchCents,
		orderTotalCents,
		input.ShippingCents,
		input.Currency,
		input.IdempotencyKey,
	)
}

func (s *LoyaltyService) Confirm(ctx context.Context, customerID, orderID string, idemKey string) (bool, error) {
	if customerID == "" {
		return false, fmt.Errorf("%w: customer_id is required", domain.ErrInvalidInput)
	}
	if orderID == "" {
		return false, fmt.Errorf("%w: order_id is required", domain.ErrInvalidInput)
	}

	return s.Repo.ConfirmAccrue(ctx, customerID, orderID, idemKey)
}

func (s *LoyaltyService) Revoke(ctx context.Context, customerID, orderID, returnID, idemKey string) (bool, error) {
	if customerID == "" {
		return false, fmt.Errorf("%w: customer_id is required", domain.ErrInvalidInput)
	}
	if orderID == "" {
		return false, fmt.Errorf("%w: order_id is required", domain.ErrInvalidInput)
	}
	if returnID == "" {
		return false, fmt.Errorf("%w: return_id is required", domain.ErrInvalidInput)
	}

	return s.Repo.RevokePoints(ctx, customerID, orderID, returnID, idemKey)
}

func (s *LoyaltyService) Redeem(ctx context.Context, customerID, reference string, points int, idemKey string) (bool, error) {
	if customerID == "" {
		return false, fmt.Errorf("%w: customer_id is required", domain.ErrInvalidInput)
	}
	if reference == "" {
		return false, fmt.Errorf("%w: reference is required", domain.ErrInvalidInput)
	}
	if points <= 0 {
		return false, fmt.Errorf("%w: points must be greater than 0", domain.ErrInvalidInput)
	}

	return s.Repo.RedeemPoints(ctx, customerID, reference, points, idemKey)
}

func (s *LoyaltyService) Restore(ctx context.Context, customerID, reference string, idemKey string) (bool, error) {
	if customerID == "" {
		return false, fmt.Errorf("%w: customer_id is required", domain.ErrInvalidInput)
	}
	if reference == "" {
		return false, fmt.Errorf("%w: reference is required", domain.ErrInvalidInput)
	}

	return s.Repo.RestorePoints(ctx, customerID, reference, idemKey)
}

func (s *LoyaltyService) GetBalance(ctx context.Context, customerID string) (domain.Balance, error) {
	if customerID == "" {
		return domain.Balance{}, fmt.Errorf("%w: customer_id is required", domain.ErrInvalidInput)
	}

	return s.Repo.GetBalance(ctx, customerID)
}

func (s *LoyaltyService) GetTransactions(ctx context.Context, customerID string) ([]domain.Transaction, error) {
	if customerID == "" {
		return nil, fmt.Errorf("%w: customer_id is required", domain.ErrInvalidInput)
	}

	return s.Repo.GetTransactions(ctx, customerID)
}