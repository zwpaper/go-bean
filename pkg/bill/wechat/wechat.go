package wechat

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	ibill "github.com/zwpaper/go-bean/pkg/bill"
)

type bill struct {
	start        time.Time
	end          time.Time
	transactions []ibill.Transaction
}

func (b bill) Transactions() ([]ibill.Transaction, error) {
	return b.transactions, nil
}

func (b bill) Range() (time.Time, time.Time) {
	return b.start, b.end
}

func New(f io.Reader) (*bill, error) {
	b := bill{
		transactions: make([]ibill.Transaction, 0),
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

	return &b, nil
}

func (b *bill) parseRange(fields []string) error {
	return nil
}

func (b *bill) parseTransaction(fields []string, hi map[string]int) (ibill.Transaction, error) {
	t := ibill.Transaction{
		Raw: strings.Join(fields, ","),
	}
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
		t.Title = fields[index]
	} else {
		return t, fmt.Errorf("title not found, in bill")
	}
	if index, ok = getIndex("交易对方"); ok {
		if fields[index] == "/" {
			t.Payee = "零钱"
		}
		t.Payee = fields[index]
	} else {
		return t, fmt.Errorf("payee not found, in bill")
	}
	if index, ok = getIndex("支付方式"); ok {
		t.Payer = fields[index]
	} else {
		return t, fmt.Errorf("payer not found, in bill")
	}
	if index, ok = getIndex("交易时间"); ok {
		var err error
		t.At, err = time.Parse("2006-01-02 15:04:05", fields[index])
		if err != nil {
			return t, fmt.Errorf("transaction time err: %w", err)
		}
	} else {
		return t, fmt.Errorf("transaction time not found, in bill")
	}
	if index, ok = getIndex("收/支"); ok {
		switch fields[index] {
		case "支出":
			t.Kind = ibill.TransKindPay
		case "收入":
			t.Kind = ibill.TransKindIncome
		case "转账":
			t.Kind = ibill.TransKindTransfer
		case "微信红包-退款":
			t.Kind = ibill.TransKindRefund
		default:
			return t, fmt.Errorf("%s not recognize", fields[index])
		}
	} else {
		return t, fmt.Errorf("transaction time not found, in bill")
	}
	if index, ok = getIndex("金额(元)"); ok {
		// TODO: currency
		switch ([]rune(fields[index]))[0] {
		case '¥':
			t.Currency = ibill.CurrencyCNY
		}
		var err error
		t.Amount, err = strconv.ParseFloat(string([]rune(fields[index])[1:]), 64)
		if err != nil {
			return t, fmt.Errorf("amount err: %w", err)
		}
	} else {
		return t, fmt.Errorf("amount time not found, in bill")
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
