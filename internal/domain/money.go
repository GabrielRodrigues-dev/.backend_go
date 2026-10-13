package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvalidAmountFormat = errors.New("formato de valor invalido, use XX.YY")
	ErrNegativeAmount      = errors.New("valor não pode ser negativo nas entradas")
	ErrInvalidCurrency     = errors.New("moeda invalida")
	ErrCurrencyMismatch    = errors.New("moedas incompativeis para operacao")
)

// Ninguém de fora do pacote 'domain' pode alterar 'cents' ou 'currency' diretamente, garantindo a imutabilidade.
type Money struct {
	cents    int64
	currency string
}

// Retornamos dois valores: o Money criado e um erro (se houver).
func NewMoneyFromString(amount string, currency string) (Money, error) {
	if len(currency) != 3 {
		return Money{}, ErrInvalidCurrency
	}

	// Rejeita valores negativos diretamente na string.
	if strings.HasPrefix(amount, "-") {
		return Money{}, ErrNegativeAmount
	}

	// Escala fixa de duas casas, ex: "25.00"
	parts := strings.Split(amount, ".")
	if len(parts) != 2 || len(parts[1]) != 2 {
		return Money{}, ErrInvalidAmountFormat
	}

	// Convertemos para int64
	cleanAmount := parts[0] + parts[1]
	cents, err := strconv.ParseInt(cleanAmount, 10, 64)
	if err != nil {
		return Money{}, ErrInvalidAmountFormat
	}

	return Money{
		cents:    cents,
		currency: currency,
	}, nil
}

// Amount Retorna o valor de volta para string no formato "XX.YY"
func (m Money) Amount() string {
	return fmt.Sprintf("%d.%02d", m.cents/100, m.cents%100)
}

// Currency retorna a moeda
func (m Money) Currency() string {
	return m.currency
}
