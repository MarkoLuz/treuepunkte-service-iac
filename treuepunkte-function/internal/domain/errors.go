package domain

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")

	ErrNotFound       = errors.New("not found")
	ErrAccrueNotFound = errors.New("accrue transaction not found")
	ErrRedeemNotFound = errors.New("redeem transaction not found")

	ErrConflict                 = errors.New("conflict")
	ErrInsufficientActivePoints = errors.New("not enough active points")
	ErrTransactionNotPending    = errors.New("transaction not pending")
)
