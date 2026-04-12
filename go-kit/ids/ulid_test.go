package ids_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/duongtuttbn/shared-resource/go-kit/ids"
)

func TestNewULID(t *testing.T) {
	require.Equal(t, 26, len(ids.NewULID()))

	for i := 0; i < 2000; i++ {
		require.Less(t, ids.NewULID(), ids.NewULID())
	}
}
