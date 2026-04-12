package s3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetectExtensions(t *testing.T) {
	exts, err := DetectExtensions("audio/mpeg")
	require.NoError(t, err)

	require.Equal(t, ".mp3", exts[0])
}
