package importcmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/zwpaper/go-bean"
	"github.com/zwpaper/go-bean/pkg/bill"
	"github.com/zwpaper/go-bean/pkg/bill/wechat"
	"github.com/zwpaper/go-bean/pkg/config"
	"github.com/zwpaper/go-bean/pkg/formatter/beancount"
)

type importer struct {
	accountAlias  map[string]string
	payeeAccounts map[string][]string
	bill          io.ReadCloser
	billKind      string
}

func NewCommand(c *config.Config) *cobra.Command {
	imp := &importer{
		accountAlias:  make(map[string]string),
		payeeAccounts: make(map[string][]string),
	}
	command := &cobra.Command{
		Use:   "import",
		Short: "import transactions to beancount",
		Long:  ``,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := supportedBillKind[kind]; !ok {
				return fmt.Errorf("bill type not supported: %s", kind)
			}
			imp.billKind = kind
			if len(args) == 0 {
				return fmt.Errorf("no bill not found, should use bill file as args")
			}
			file, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("can not open bill file: %w", err)
			}
			imp.bill = file
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			bean, err := bean.Count(c.BeanFile)
			if err != nil {
				fmt.Println("can not parse beancount file, ", err)
				os.Exit(1)
			}
			err = imp.parseBean(bean)
			if err != nil {
				logrus.Error("parse bean error: ", err)
				os.Exit(1)
			}
			bp, err := imp.parseBill()
			if err != nil {
				logrus.Error("parse bill error: ", err)
				os.Exit(1)
			}
			ts, err := bp.Transactions()
			if err != nil {
				logrus.Error("transactions error: ", err)
				os.Exit(1)
			}
			tally, err := imp.createTally(ts)
			if err != nil {
				logrus.Error("crate tally error: ", err)
				os.Exit(1)
			}
			fmt.Println(tally)
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			if imp.bill != nil {
				imp.bill.Close()
			}
		},
	}

	err := build(imp, command)
	if err != nil {
		logrus.Error("config file or args error: ", err)
		os.Exit(1)
	}

	return command
}

var supportedBillKind = map[string]struct{}{
	billKindWechat: struct{}{},
}

func build(l *importer, cmd *cobra.Command) error {
	cmd.Flags().StringVarP(&kind, "bill-type", "t", "wechat",
		"bill type: wechat")

	return nil
}

// update importer directly
func (i importer) parseBean(b *bean.Bean) error {
	// parse account
	for _, a := range b.AllAccounts() {
		for _, n := range a.Notes {
			nl := strings.Split(n, ":")
			if len(nl) != 3 ||
				strings.ToLower(nl[0]) != "alias" ||
				strings.ToLower(nl[1]) != i.billKind {
				continue
			}
			i.accountAlias[nl[2]] = a.Name
		}
	}

	// parse payee
	payeeMap := map[string]map[string]struct{}{}
	for _, t := range b.AllTransactions() {
		as, ok := payeeMap[t.Payee]
		if !ok {
			as = make(map[string]struct{})
		}
		for _, d := range t.Details {
			// only add + account as beans go into this payee
			if d.Number > 0 {
				as[d.Account.Name] = struct{}{}
			}
		}
		payeeMap[t.Payee] = as
	}

	for n, am := range payeeMap {
		as := make([]string, 0)
		for a := range am {
			as = append(as, a)
		}
		i.payeeAccounts[n] = as
	}

	return nil
}

var kind string

type billKind string

const (
	billKindWechat = "wechat"
	// aliPay billKind = "alipay"
)

func (i importer) parseBill() (bill.Parser, error) {
	switch i.billKind {
	case billKindWechat:
		p, err := wechat.New(i.bill)
		if err != nil {
			return nil, fmt.Errorf("parser wechat bill error: %w", err)
		}
		return p, nil
	default:
		return nil, fmt.Errorf("bill kind %s not supported", i.billKind)
	}
}

func (i importer) createTally(ts []bill.Transaction) (string, error) {
	beanformat := beancount.New()

	var tally string
	for _, t := range ts {
		var accs []string
		var accountPayee []string
		if as, ok := i.accountAlias[t.Payee]; ok {
			accountPayee = []string{as}
		} else {
			// check nil in CreateTransaction
			accountPayee, _ = i.payeeAccounts[t.Payee]
		}
		acc, ok := i.accountAlias[t.Payer]
		if ok {
			accs = []string{acc}
		}

		s, err := beanformat.CreateTransaction(t, accountPayee, accs)
		if err != nil {
			return "", err
		}
		tally += fmt.Sprintf("\n%s", s)
	}

	return tally, nil
}
