package components

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLandingPageRendersRegistryComponentCount(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, LandingPage().Render(context.Background(), &buf))

	expected := strconv.Itoa(componentCount()) + " components"
	assert.Equal(t, 2, strings.Count(buf.String(), expected))
}
