package wechat

import (
	"bufio"
	"io"
	"time"
)

type bill struct {
	start        time.Time
	end          time.Time
	transactions []transaction
}

type transaction struct {
	at time.Time
}

func (b bill) Transactions() ([]transaction, error) {
	return b.transactions, nil
}

func (b bill) Range() (time.Time, time.Time) {
	return b.start, b.end
}

func New(f io.Reader) (*bill, error) {
	b := bill{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		err := b.parseLine(scanner.Text())
		if err != nil {
			return nil, err
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &b, nil
}

func (b *bill) parseLine(line string) error {

}

func (b *bill) parseRange(line string) error {
	return nil
}

func (b *bill) parseTransaction(line string) error {
	return nil
}

func (b *bill) parseHeader(line string) error {
	return nil
}
