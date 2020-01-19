package importcmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/zwpaper/go-bean"
	"github.com/zwpaper/go-bean/pkg/billparser/wechat"
	"github.com/zwpaper/go-bean/pkg/config"
)

type importer struct {
	accounts map[string]account
	payee    map[string]payee
	bill     io.ReadCloser
	billKind string
}

type account struct {
	Name  string
	alias map[string]struct{}
}

type payee struct {
	Name     string
	accounts []string
}

func NewCommand(c *config.Config) *cobra.Command {
	imp := &importer{
		accounts: make(map[string]account),
		payee:    make(map[string]payee),
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
				logrus.Error("parse error: ", err)
				os.Exit(1)
			}
			err = imp.parseBill()
			if err != nil {
				logrus.Error("parse error: ", err)
				os.Exit(1)
			}
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

func (i importer) parseBean(b *bean.Bean) error {
	// parse account
	for _, a := range b.AllAccounts() {
		created := account{
			Name: a.Name,
		}
		alias := make(map[string]struct{})
		for _, n := range a.Notes {
			nl := strings.Split(n, ":")
			if len(nl) != 3 ||
				strings.ToLower(nl[0]) != "alias" ||
				strings.ToLower(nl[1]) != i.billKind {
				continue
			}
			alias[nl[2]] = struct{}{}
		}
		created.alias = alias
		i.accounts[created.Name] = created
	}

	// parse payee
	for _, t := range b.AllTransactions() {
		created := payee{
			Name:     t.Payee,
			accounts: make([]string, 0),
		}
		for _, d := range t.Details {
			created.accounts = append(created.accounts, d.Account.Name)
		}
		i.payee[created.Name] = created
	}

	return nil
}

var kind string

type billKind string

const (
	billKindWechat = "wechat"
	// aliPay billKind = "alipay"
)

func (i importer) parseBill() error {
	switch i.billKind {
	case billKindWechat:
		p, err := wechat.New(i.bill)
		if err != nil {
			return fmt.Errorf("parser wechat bill error: %w", err)
		}
		fmt.Println(p.Transactions())
	default:
		return fmt.Errorf("bill kind %s not supported", i.billKind)
	}
	return nil
}
