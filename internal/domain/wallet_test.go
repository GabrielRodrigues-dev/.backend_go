package domain

import (
	"errors"
	"testing"
)

func TestWallet_Debit_InsufficientFunds(t *testing.T) {
	// Arrange
	initialBalance, _ := NewMoneyFromString("50.00", "BRL")
	wallet := NewWallet("wallet-1", "player-1", initialBalance)

	betAmount, _ := NewMoneyFromString("80.00", "BRL")

	// Act
	err := wallet.Debit(betAmount)

	// Assert
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Errorf("esperava erro de saldo insuficiente, recebeu: %v", err)
	}

	// A versão não deve mudar se a operação falhou.
	if wallet.Version() != 1 {
		t.Errorf("versão da carteira deveria ser 1, mas é %d", wallet.Version())
	}
}

func TestWallet_Credit_Success(t *testing.T) {
	initialBalance, _ := NewMoneyFromString("100.00", "BRL")
	wallet := NewWallet("wallet-1", "player-1", initialBalance)

	winAmount, _ := NewMoneyFromString("25.50", "BRL")

	err := wallet.Credit(winAmount)

	if err != nil {
		t.Fatalf("não esperava erro no crédito, recebeu: %v", err)
	}

	if wallet.Balance().Amount() != "125.50" {
		t.Errorf("esperava saldo de 125.50, recebeu %s", wallet.Balance().Amount())
	}

	if wallet.Version() != 2 {
		t.Errorf("esperava versão 2, recebeu %d", wallet.Version())
	}
}
