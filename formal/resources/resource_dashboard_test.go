package resource

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuppressEquivalentDashboardDefinition(t *testing.T) {
	definition := `{"layout":{"columns":2},"widgets":[{"key":"a","layout":{"span":1},"logs":{"visualization":"DASHBOARD_VISUALIZATION_BARS","group_by":["llm.provider"]}},{"key":"b","layout":{"span":1},"note":{"markdown":"x"}}]}`
	reformatted := `{"widgets":[{"key":"a","layout":{"span":1},"logs":{"visualization":3,"groupBy":["llm.provider"]}},{"key":"b","layout":{"span":1},"note":{"markdown":"x"}}],"layout":{"columns":2}}`
	reordered := `{"layout":{"columns":2},"widgets":[{"key":"b","layout":{"span":1},"note":{"markdown":"x"}},{"key":"a","layout":{"span":1},"logs":{"visualization":"DASHBOARD_VISUALIZATION_BARS","group_by":["llm.provider"]}}]}`

	require.True(t, suppressEquivalentDashboardDefinition("definition", definition, reformatted, nil))
	require.False(t, suppressEquivalentDashboardDefinition("definition", definition, reordered, nil))
}
