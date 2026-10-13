package domain

import (
	"errors"
	"time"
)

var (
	ErrInsufficientFunds = errors.New("saldo insuficiente na carteira")
)

// Wallet representa a carteira do jogador.
type Wallet struct {
	id        string
	playerId  string
	balance   Money
	version   int
	createdAt time.Time
	updatedAt time.Time
}

// NewWallet é o construtor para criar uma nova carteira.
func NewWallet(id string, playerId string, initialBalance Money) Wallet {
	now := time.Now().UTC()
	return Wallet{
		id:        id,
		playerId:  playerId,
		balance:   initialBalance,
		version:   1, // A versão inicial é sempre 1.
		createdAt: now,
		updatedAt: now,
	}
}

// Credit adiciona fundos.
func (w *Wallet) Credit(amount Money) error {
	if amount.IsNegative() {
		return ErrNegativeAmount
	}

	newBalance, err := w.balance.Add(amount)
	if err != nil {
		return err // Pode ser erro de moeda incompatível ou overflow
	}

	w.balance = newBalance
	w.version++ // Incrementa apenas em mudança de saldo.
	w.updatedAt = time.Now().UTC()
	return nil
}

// Debit remove fundos, bloqueando caso o saldo fique negativo.
func (w *Wallet) Debit(amount Money) error {
	if amount.IsNegative() {
		return ErrNegativeAmount
	}

	newBalance, err := w.balance.Subtract(amount)
	if err != nil {
		return err
	}

	// O saldo não pode ficar abaixo de zero.
	if newBalance.IsNegative() {
		return ErrInsufficientFunds
	}

	w.balance = newBalance
	w.version++
	w.updatedAt = time.Now().UTC()
	return nil
}

// RestoreWallet recria a entidade a partir dos dados do banco (Reidratação).
func RestoreWallet(id, playerId string, cents int64, currency string, version int, createdAt, updatedAt time.Time) Wallet {
	return Wallet{
		id:        id,
		playerId:  playerId,
		balance:   Money{cents: cents, currency: currency},
		version:   version,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// Getters para expor os dados para o banco de dados futuramente
func (w Wallet) ID() string           { return w.id }
func (w Wallet) PlayerID() string     { return w.playerId }
func (w Wallet) Balance() Money       { return w.balance }
func (w Wallet) Version() int         { return w.version }
func (w Wallet) CreatedAt() time.Time { return w.createdAt }
func (w Wallet) UpdatedAt() time.Time { return w.updatedAt }
