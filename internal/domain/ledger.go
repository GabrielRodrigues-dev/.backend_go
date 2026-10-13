package domain

import "time"

// LedgerEntry representa o histórico imutável de uma movimentação
type LedgerEntry struct {
	id            string
	walletId      string
	transactionId string
	direction     string // "DEBIT" ou "CREDIT"
	amount        Money
	balanceBefore Money
	balanceAfter  Money
	createdAt     time.Time
}

func NewLedgerEntry(id, walletId, transactionId, direction string, amount, before, after Money) LedgerEntry {
	return LedgerEntry{
		id:            id,
		walletId:      walletId,
		transactionId: transactionId,
		direction:     direction,
		amount:        amount,
		balanceBefore: before,
		balanceAfter:  after,
		createdAt:     time.Now().UTC(),
	}
}

// Getters
func (l LedgerEntry) ID() string            { return l.id }
func (l LedgerEntry) WalletID() string      { return l.walletId }
func (l LedgerEntry) TransactionID() string { return l.transactionId }
func (l LedgerEntry) Direction() string     { return l.direction }
func (l LedgerEntry) Amount() Money         { return l.amount }
func (l LedgerEntry) BalanceBefore() Money  { return l.balanceBefore }
func (l LedgerEntry) BalanceAfter() Money   { return l.balanceAfter }
func (l LedgerEntry) CreatedAt() time.Time  { return l.createdAt }
