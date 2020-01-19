package bean

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBeancount(t *testing.T) {
	raw := "test/example.bean"
	b, err := Count(raw)
	require.Nil(t, err, "must count")
	require.NotNil(t, b, "must get bean")
}
