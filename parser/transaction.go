package parser

import (
	"fmt"
	"strconv"
	"time"
)

type Heading struct {
	Payee  string
	Name   string
	Status string
	At     time.Time
	Tags   []string
	Body   []Body
}

var availableStatus = map[string]struct{}{
	"*": struct{}{},
	"!": struct{}{},
}

func (t Heading) String() string {
	s := fmt.Sprintf("%s %s %s %s", t.At.Format("2006-01-02"), t.Status, t.Payee, t.Name)
	for _, d := range t.Body {
		s += fmt.Sprintf("\n%s", d)
	}
	return s
}

func (t Heading) node() {}

type Body struct {
	Account  string
	Number   float64
	Currency string
}

func (d Body) String() string {
	return fmt.Sprintf("  %s %+f %s", d.Account, d.Number, d.Currency)
}

func (d Body) node() {}

func (d *Document) parseHeading(i int, parentStop stopFn) (int, Node, error) {
	current := i
	m := d.tokens[current].matches
	if len(m) < 4 {
		return i, nil, fmt.Errorf("can not parse transaction, format error: %+v", m[1:])
	}
	open, err := time.Parse("2006-01-02", m[1])
	if err != nil {
		return i, nil, fmt.Errorf("can not parse transaction, time error: %s(%v)", m[1], m)
	}

	if _, ok := availableStatus[m[2]]; !ok {
		return i, nil, fmt.Errorf("can not parse transaction, not supported status: %s(%v)", m[2], m[0])
	}
	tran := Heading{
		Payee: m[3],
		Name:  m[4],
		// tranStatus type checked in lex
		Status: m[2],
		Tags:   make([]string, len(m)-5),
		At:     open,
	}
	for it, t := range m[5:] {
		tran.Tags[it] = t
	}

	for current < len(d.tokens)-1 && d.tokens[current+1].kind == tokenBody {
		c, n, err := d.parseOne(current+1, parentStop)
		if err != nil {
			return current - i + 1, nil, err
		}
		current += c
		tran.Body = append(tran.Body, n.(Body))
	}

	return current - i + 1, tran, nil
}

func (d *Document) parseBody(i int, parentStop stopFn) (int, Node, error) {
	m := d.tokens[i].matches
	if len(m) != 2 && len(m) != 4 {
		return i, nil, fmt.Errorf("can not parse transaction detail, format error: %v", m)
	}

	a, ok := d.NamedNodes[m[1]]
	if !ok {
		return 0, nil, fmt.Errorf("detail account %s not found", m[1])
	}
	detail := Body{
		Account: a.(Account).Name,
	}

	if len(m) == 4 {
		detail.Currency = m[3]
		n, err := strconv.ParseFloat(m[2], 64)
		if err != nil {
			return 0, nil, fmt.Errorf("detail number error: %s(%+v)", m[2], m[1:])
		}
		detail.Number = n
	}

	return 1, detail, nil
}
