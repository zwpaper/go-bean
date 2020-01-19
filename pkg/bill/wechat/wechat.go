package wechat

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type bill struct {
	start        time.Time
	end          time.Time
	transactions []transaction
}

type transaction struct {
	at     time.Time
	payer  string
	payee  string
	title  string
	amount float64
	kind   string

	raw string
}

func (b bill) Transactions() ([]transaction, error) {
	return b.transactions, nil
}

func (b bill) Range() (time.Time, time.Time) {
	return b.start, b.end
}

func New(f io.Reader) (*bill, error) {
	b := bill{
		transactions: make([]transaction, 0),
	}

	header := false
	detail := false
	headerIndex := map[string]int{}
	var err error
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ",")
		if header {
			headerIndex, err = b.parseHeader(fields)
			if err != nil {
				return nil, err
			}
			header = false
			detail = true
			continue
		}
		if detail {
			t, err := b.parseTransaction(fields, headerIndex)
			if err != nil {
				return nil, err
			}
			b.transactions = append(b.transactions, t)
			continue
		}

		// split will return at least one element
		switch strings.Trim(fields[0], " -") {
		case "微信支付账单明细列表":
			header = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	logrus.Error(b)

	return &b, nil
}

func (b *bill) parseLine(fields []string) error {
	return nil
}

func (b *bill) parseRange(fields []string) error {
	return nil
}

func (b *bill) parseTransaction(fields []string, hi map[string]int) (transaction, error) {
	t := transaction{}
	var ok bool
	var index int
	getIndex := func(name string) (int, bool) {
		if i, ok := hi[name]; !ok || len(fields) < i {
			return 0, false
		} else {
			return i, true
		}
	}

	if index, ok = getIndex("商品"); ok {
		t.title = fields[index]
	} else {
		return t, fmt.Errorf("title not found, in bill")
	}

	return t, nil
}

func (b *bill) parseHeader(fields []string) (map[string]int, error) {
	hi := map[string]int{}
	for i, f := range fields {
		hi[f] = i
	}
	return hi, nil
}
