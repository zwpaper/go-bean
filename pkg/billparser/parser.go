package billparser

import "time"

type TransKind string

const (
	Pay      TransKind = "pay"
	Transfer TransKind = "transfer"
	Income   TransKind = "income"
	Refund   TransKind = "refund"
)

type Parser interface {
	Transactions() ([]Transaction, error)
	Range() (time.Time, time.Time)
}

type Transaction interface {
	Payer() string
	Payee() string
	Amount() float64
	Kind() string
	At() time.Time

	Raw() string
}
