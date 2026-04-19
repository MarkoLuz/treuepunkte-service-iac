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

func NewLoyaltyService(db *sql.DB) *LoyaltyService {
	return &LoyaltyService{
		DB:   db,
		Repo: storage.NewRepository(db),
	}
}

func (s *LoyaltyService) Accrue(ctx context.Context, customerID, orderID string, points int, idemKey string) (bool, error) {
	if customerID == "" {
		return false, fmt.Errorf("%w: customer_id is required", domain.ErrInvalidInput)
	}
	if orderID == "" {
		return false, fmt.Errorf("%w: order_id is required", domain.ErrInvalidInput)
	}
	if points <= 0 {
		return false, fmt.Errorf("%w: points must be greater than 0", domain.ErrInvalidInput)
	}

	return s.Repo.AccruePoints(ctx, customerID, orderID, points, idemKey)
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
