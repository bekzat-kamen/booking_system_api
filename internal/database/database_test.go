package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCloseNilDB(t *testing.T) {
	err := Close(nil)
	require.NoError(t, err)
}

func TestCloseNilRedisClient(t *testing.T) {
	err := CloseRedis(nil)
	require.NoError(t, err)
}
