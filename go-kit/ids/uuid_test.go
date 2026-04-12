package ids_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/duongtuttbn/shared-resource/go-kit/ids"
)

func TestNewUUID(t *testing.T) {
	require.Equal(t, 36, len(ids.NewUUID()))
}

func TestNewUUID7(t *testing.T) {
	require.Equal(t, 36, len(ids.NewUUID7()))

	for i := 0; i < 2000; i++ {
		require.Less(t, ids.NewUUID7(), ids.NewUUID7())
	}
}
