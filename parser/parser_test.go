package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccount(t *testing.T) {
	raw := "2016-07-07 open Assets:Bank:CMB:1111:Deposit CNY"
	parsed, ok := lexAccount(raw)
	require.True(t, ok, "must parse account")
	t.Log(parsed)
}
