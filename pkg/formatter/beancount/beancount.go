package beancount

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/zwpaper/go-bean/pkg/bill"
)

type formatter struct {
	payeeTable   map[string][]string
	accountTable map[string]string
}

func New() *formatter {
	return &formatter{
		payeeTable:   make(map[string][]string),
		accountTable: make(map[string]string),
	}
}

func (f *formatter) AddPayee(name string, payee []string) error {
	f.payeeTable[name] = payee

	return nil
}

func (f *formatter) AddAccount(name string, account string) error {
	f.accountTable[name] = account

	return nil
}

const transactionTpl = `;; {{.Raw}}
{{.At}} ! "{{.Payee}}" "{{.Title}}"
{{- with $item := . -}}
  {{- with $ps := $item.PayeeAccounts -}}
	{{- range $i, $el := $ps}}
  {{$el}} {{ if eq (len $ps) 1 }} +{{$item.Amount}} {{end}}{{$item.Currency}}
{{- end -}}
{{- end -}}
{{- with $as := $item.Accounts -}}
{{- range $el := $as}}
  {{$el}} {{ if eq (len $as) 1 }} -{{$item.Amount}} {{end}}{{$item.Currency}}
{{- end -}}
{{- end -}}
{{- end}}
`

type transactionItem struct {
	Raw           string
	At            string
	Payee         string
	Title         string
	PayeeAccounts []string
	Accounts      []string
	Amount        float64
	Currency      string
}

// CreateTransaction always return a unconfirmed transaction
func (f formatter) CreateTransaction(t bill.Transaction, ins, outs []string) (string, error) {
	item := transactionItem{
		Raw:           t.Raw,
		At:            t.At.Format("2006-01-02"),
		Payee:         t.Payee,
		Title:         t.Title,
		PayeeAccounts: ins,
		Accounts:      outs,
		Amount:        t.Amount,
		Currency:      t.Currency,
	}
	if item.PayeeAccounts == nil || len(item.PayeeAccounts) == 0 {
		item.PayeeAccounts = []string{"TODO"}
	}
	if item.Accounts == nil || len(item.Accounts) == 0 {
		item.Accounts = []string{"TODO"}
	}

	var tplBytes bytes.Buffer
	tpl, err := template.New("transaction").Parse(transactionTpl)
	if err != nil {
		return "", fmt.Errorf("template error, should be a bug: %w", err)
	}

	err = tpl.Execute(&tplBytes, item)
	if err != nil {
		return "", fmt.Errorf("template exec error, should be a bug: %w", err)
	}

	return tplBytes.String(), nil
}
