package formatter

import (
	"github.com/zwpaper/go-bean/pkg/bill"
)

type Formatter interface {
	AddPayee(string, []string) error
	AddAccount(string, string) error
	CreateTransaction(bill.Transaction) (string, error)
}
