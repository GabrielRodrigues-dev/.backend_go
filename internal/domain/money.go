package domain

import (
	"errors"
	"fmt"
	"math"
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

// ZeroMoney cria um valor zerado para uma moeda específica.
func ZeroMoney(currency string) Money {
	return Money{cents: 0, currency: currency}
}

// Soma de dois valores.
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrCurrencyMismatch
	}
	// Prevenção de Overflow
	if (other.cents > 0 && m.cents > math.MaxInt64-other.cents) ||
		(other.cents < 0 && m.cents < math.MinInt64-other.cents) {
		return Money{}, errors.New("overflow na soma")
	}
	return Money{cents: m.cents + other.cents, currency: m.currency}, nil
}

// Subtração de dois valores.
func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrCurrencyMismatch
	}
	// Prevenção de Overflow
	if (other.cents < 0 && m.cents > math.MaxInt64+other.cents) ||
		(other.cents > 0 && m.cents < math.MinInt64+other.cents) {
		return Money{}, errors.New("overflow na subtracao")
	}
	return Money{cents: m.cents - other.cents, currency: m.currency}, nil
}

// Negate inverte o sinal do valor.
func (m Money) Negate() (Money, error) {
	// Prevenção de Overflow
	if m.cents == math.MinInt64 {
		return Money{}, errors.New("overflow na negacao")
	}
	return Money{cents: -m.cents, currency: m.currency}, nil
}

// Métodos auxiliares de comparação
func (m Money) IsNegative() bool {
	return m.cents < 0
}

func (m Money) Cents() int64 {
	return m.cents
}
