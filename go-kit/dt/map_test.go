package dt_test

import (
	"encoding/json"
	"testing"
	"github.com/duongtuttbn/shared-resource/go-kit/dt"

	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	var m dt.Map
	m.Add("foo", "bar")
	require.True(t, m.Contains("foo"))
	bytes, err := json.Marshal(m)
	require.NoError(t, err)
	require.Equal(t, `{"foo":"bar"}`, string(bytes))
}
