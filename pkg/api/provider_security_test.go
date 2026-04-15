// SPDX-Licence-Identifier: EUPL-1.2

package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormaliseMarketplaceCode_Good(t *testing.T) {
	code, err := normaliseMarketplaceCode("analytics")
	require.NoError(t, err)
	assert.Equal(t, "analytics", code)
}

func TestNormaliseMarketplaceCode_Bad(t *testing.T) {
	_, err := normaliseMarketplaceCode("analytics;rm")
	assert.Error(t, err)
}

func TestNormaliseMarketplaceCode_Bad_EncodedTraversal(t *testing.T) {
	_, err := normaliseMarketplaceCode("analytics%2f..%2Fescape")
	assert.Error(t, err)
}
