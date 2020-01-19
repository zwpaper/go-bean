package bill

import "time"

const (
	TransKindPay      = "pay"
	TransKindTransfer = "transfer"
	TransKindIncome   = "income"
	TransKindRefund   = "refund"
)

type Parser interface {
	Transactions() ([]Transaction, error)
	Range() (time.Time, time.Time)
}

type Transaction interface {
	Payer() string
	Payee() string
	Title() string
	Amount() float64
	Kind() string
	At() time.Time

	Raw() string
}
