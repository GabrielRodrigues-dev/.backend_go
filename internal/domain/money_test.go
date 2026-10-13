package domain

import (
	"errors"
	"testing"
)

func TestNewMoneyFromString_Success(t *testing.T) {
	// Arrange (Preparação)
	amountStr := "25.00"
	currency := "BRL"

	// Act (Ação)
	money, err := NewMoneyFromString(amountStr, currency)

	// Assert (Verificação)
	if err != nil {
		// t.Fatalf para a execução do teste imediatamente e mostra o erro
		t.Fatalf("esperava sucesso, mas recebeu erro: %v", err)
	}

	if money.Amount() != "25.00" {
		t.Errorf("esperava valor 25.00, recebeu %s", money.Amount())
	}

	if money.Currency() != "BRL" {
		t.Errorf("esperava moeda BRL, recebeu %s", money.Currency())
	}
}

func TestNewMoneyFromString_InvalidFormat(t *testing.T) {
	// Vamos testar as regras de bloqueio
	invalidInputs := []string{"25", "25.0", "25.000", "abc", "-25.00"}

	for _, input := range invalidInputs {
		_, err := NewMoneyFromString(input, "BRL")
		if err == nil {
			t.Errorf("esperava um erro para a entrada %s, mas passou", input)
		}
	}
}

func TestNewMoneyFromString_NegativeNotAllowed(t *testing.T) {
	_, err := NewMoneyFromString("-10.00", "BRL")

	// Verifica se o erro retornado é exatamente o ErrNegativeAmount usando errors.Is
	if !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("esperava ErrNegativeAmount, recebeu: %v", err)
	}
}
