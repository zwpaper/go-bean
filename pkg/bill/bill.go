package bill

import "time"

const (
	TransKindPay      = "pay"
	TransKindTransfer = "transfer"
	TransKindIncome   = "income"
	TransKindRefund   = "refund"
)

const (
	CurrencyCNY = "CNY"
	CurrencyUSD = "USD"
)

type Parser interface {
	Transactions() ([]Transaction, error)
	Range() (time.Time, time.Time)
}

type Transaction struct {
	At       time.Time
	Payer    string
	Payee    string
	Title    string
	Amount   float64
	Kind     string
	Currency string

	Raw string
}
