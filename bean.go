package bean

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zwpaper/go-bean/parser"
)

type Bean struct {
	// TODO use Accounts
	Accounts     []Account
	Transactions []Transaction
}

type Account struct {
	Name     string
	Labels   Label
	Currency string
	Children []Account
	Notes    []string
}

type Transaction struct {
	Payee   string
	Name    string
	Status  tranStatus
	At      time.Time
	Tags    []string
	Details []Detail
}

type tranStatus string

const (
	confirm tranStatus = "*"
	wait    tranStatus = "!"
)

type Detail struct {
	Account  Account
	Number   float64
	Currency string
}

// Accounts tree accounts
// TODO:
type Accounts []Account

// custom type for account notes, may convenience for later usage
type Label map[string]string

func Count(f string) (*Bean, error) {
	file, err := os.Open(f)
	if err != nil {
		return nil, fmt.Errorf("open file error: %w", err)
	}

	parser := parser.New()
	doc, err := parser.Parse(file, f)
	if err != nil {
		return nil, fmt.Errorf("parse %s error: %w", f, err)
	}

	return buildBean(doc)
}

func buildBean(doc *parser.Document) (*Bean, error) {
	bean := Bean{
		Accounts:     make([]Account, 0),
		Transactions: make([]Transaction, 0),
	}
	addedAccount := map[string]int{}
	for _, a := range doc.Nodes {
		switch node := a.(type) {
		case parser.Account:
			bean.Accounts = append(bean.Accounts, Account{
				Name:     node.Name,
				Currency: node.Currency,
			})
			addedAccount[node.Name] = len(bean.Accounts) - 1
		case parser.Note:
			index, ok := addedAccount[node.AccountName]
			if !ok {
				return nil, fmt.Errorf("account %s not found", node.AccountName)
			}
			account := bean.Accounts[index]
			if account.Notes == nil {
				account.Notes = make([]string, 0)
			}
			account.Notes = append(account.Notes, node.Note)
			bean.Accounts[index] = account
		case parser.Heading:
			t := Transaction{
				Payee:  node.Payee,
				Name:   node.Name,
				Status: tranStatus(node.Status),
				At:     node.At,
				Tags:   node.Tags,
			}
			for _, n := range node.Body {
				index, ok := addedAccount[n.Account]
				if !ok {
					return nil, fmt.Errorf("account %s not found", n.Account)
				}
				t.Details = append(t.Details, Detail{
					Account:  bean.Accounts[index],
					Number:   n.Number,
					Currency: n.Currency,
				})
			}
			bean.Transactions = append(bean.Transactions, t)
		}
	}

	for _, a := range bean.Accounts {
		logrus.Info(a)
	}

	for _, a := range bean.Transactions {
		logrus.Info(a)
	}

	return &bean, nil
}
