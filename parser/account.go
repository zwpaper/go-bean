package parser

import (
	"fmt"
	"strings"
	"time"
)

// Account empty currency means default currency
type Account struct {
	Name     string
	Currency string
	OpenAt   time.Time
}

type Note struct {
	AccountName string
	Note        string
	At          time.Time
}

func (a Account) String() string {
	return fmt.Sprintf("%s open %s %s", a.OpenAt.Format("2006-01-02"), a.Name, a.Currency)
}

func (a Account) node() {}

const (
	accountOpen = "open"
	accountNote = "note"
)

func (n Note) String() string {
	return fmt.Sprintf(`%s note %s "%s"`, n.At.Format("2006-01-02"), n.AccountName, n.Note)
}

func (n Note) node() {}

func (d *Document) parseAccount(i int, parentStop stopFn) (int, Node, error) {
	m := d.tokens[i].matches
	if len(m) < 4 {
		return i, nil, fmt.Errorf("can not parse account, format error: %v", m)
	}
	open, err := time.Parse("2006-01-02", m[1])
	if err != nil {
		return i, nil, fmt.Errorf("can not parse account, time error: %v", m)
	}
	switch m[2] {
	case accountOpen:
		account := Account{
			Name: m[3],
		}
		if len(m) >= 5 {
			account.Currency = m[4]
		}
		account.OpenAt = open
		d.NamedNodes[account.Name] = account
		return 1, account, nil
	case accountNote:
		for i := range d.Nodes {
			a, ok := d.Nodes[i].(Account)
			if !ok {
				continue
			}
			if strings.HasPrefix(a.Name, m[3]) {
				return 1, Note{
					AccountName: m[3],
					Note:        strings.Trim(m[4], " \""),
					At:          open,
				}, nil
			}
		}
		return 1, nil, fmt.Errorf("account %s not found for note %s", m[3], m[0])
	default:
		return 0, nil, fmt.Errorf("not supported account kind: %s(%v)", m[2], m)
	}
}
