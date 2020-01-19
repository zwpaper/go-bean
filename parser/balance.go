package parser

import (
	"fmt"
	"strconv"
	"time"
)

type Balance struct {
	Account  Account
	Number   float64
	Currency string
	At       time.Time
}

func (b Balance) String() string {
	return fmt.Sprintf("%s balance %s %+f %s", b.At.Format("2006-01-02"), b.Account.Name, b.Number, b.Currency)
}

func (b Balance) node() {}

func (d *Document) parseBalance(i int, parentStop stopFn) (int, Node, error) {
	m := d.tokens[i].matches
	if len(m) < 6 {
		return i, nil, fmt.Errorf("can not parse balance, format error: %v", m)
	}

	open, err := time.Parse("2006-01-02", m[1])
	if err != nil {
		return i, nil, fmt.Errorf("can not parse balance, time error: %v", m)
	}
	n, err := strconv.ParseFloat(m[4], 64)
	if err != nil {
		return i, nil, fmt.Errorf("can not parse balance, number error: %s(%v)", m[4], m)
	}

	a, ok := d.NamedNodes[m[3]]
	if !ok {
		return i, nil, fmt.Errorf("can not parse balance, account not found: %s", m[3])
	}
	b := Balance{
		Account:  a.(Account),
		Number:   n,
		Currency: m[4],
		At:       open,
	}
	return 1, b, nil
}

type Pad struct {
	AccountIn  Account
	AccountOut Account
	At         time.Time
}

func (p Pad) String() string {
	return fmt.Sprintf("%s pad %s %s", p.At.Format("2006-01-02"), p.AccountIn.Name, p.AccountOut.Name)
}

func (p Pad) node() {}

func (d *Document) parsePad(i int, parentStop stopFn) (int, Node, error) {
	m := d.tokens[i].matches
	if len(m) < 5 {
		return i, nil, fmt.Errorf("can not parse pad, format error: %v", m)
	}

	open, err := time.Parse("2006-01-02", m[1])
	if err != nil {
		return i, nil, fmt.Errorf("can not parse pad, time error: %s(%v)", m[1], m)
	}

	a, ok := d.NamedNodes[m[3]]
	if !ok {
		return i, nil, fmt.Errorf("can not parse pad, account not found: %s", m[3])
	}
	b, ok := d.NamedNodes[m[4]]
	if !ok {
		return i, nil, fmt.Errorf("can not parse pad, account not found: %s", m[4])
	}

	p := Pad{
		AccountIn:  a.(Account),
		AccountOut: b.(Account),
		At:         open,
	}
	return 1, p, nil
}
