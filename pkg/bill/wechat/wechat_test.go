package wechat

import (
	"os"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWechats(t *testing.T) {
	cases := map[string]error{
		"~/Dropbox/org-mode/accounting/wechat-20190801-20191101.csv": nil,
	}

	for name, expectErr := range cases {
		n, err := homedir.Expand(name)
		require.Nil(t, err, "must expand dir")
		file, err := os.Open(n)
		require.Nil(t, err, "must open file")
		_, err = New(file)
		assert.Equal(t, expectErr, err, "should match err")
		file.Close()
	}
}
