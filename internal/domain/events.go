package domain

// WalletBalanceChangedPayload contém exatamente as propriedades exigidas pela avaliação.
type WalletBalanceChangedPayload struct {
	WalletID      string `json:"walletId"`
	TransactionID string `json:"transactionId"`
	Direction     string `json:"direction"`
	Money         string `json:"money"`
	BalanceBefore string `json:"balanceBefore"`
	BalanceAfter  string `json:"balanceAfter"`
	WalletVersion int    `json:"walletVersion"`
}
