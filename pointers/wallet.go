package wallet

import (
	"errors"
	"fmt"
)

// Stringer interface
type Stringer interface {
	String() string
}

// Bitcoin ...
type Bitcoin int

func (b Bitcoin) String() string {
	return fmt.Sprintf("%d BTC", b)
}

// Wallet ...
type Wallet struct {
	balance Bitcoin
}

// Deposit пополняет баланс
func (w *Wallet) Deposit(amount Bitcoin) {
	w.balance += amount
}

// ErrInsufficientFunds ...
var ErrInsufficientFunds = errors.New("cannot withdraw, insufficient funds")

// Withdraw ...
func (w *Wallet) Withdraw(amount Bitcoin) error {
	if amount > w.balance {
		return ErrInsufficientFunds
	}
	w.balance -= amount
	return nil
}

// Balance возвращает текущий баланс
func (w *Wallet) Balance() Bitcoin {
	return w.balance
}
