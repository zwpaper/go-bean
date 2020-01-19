package parser

import (
	"fmt"
)

const (
	optionKind     = "option"
	optionTitle    = "title"
	optionCurrency = "operating_currency"
)

type Option struct {
	Kind  string
	Value string
}

func (o Option) String() string {
	return fmt.Sprintf("%s %s %s", optionKind, o.Kind, o.Value)
}

func (o Option) node() {}

func (d *Document) parseOption(i int, parentStop stopFn) (int, Node, error) {
	m := d.tokens[i].matches
	if len(m) < 3 {
		return 0, nil, fmt.Errorf("can not parse option, format error: %v", m)
	}
	option := Option{
		Kind:  m[1],
		Value: m[2],
	}
	return 1, option, nil
}
