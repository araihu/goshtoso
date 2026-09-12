package deployments

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateApproveAndRoundTrip(t *testing.T) {
	state := Initial()
	require.NotEmpty(t, Validate("", "", "Moon"))
	require.Empty(t, Validate("api", "v1", "Staging"))
	id := state.Create(" billing-api ", " v1.0 ", "Staging")
	state = Decode(state.Encode())
	record, ok := state.Find(id)
	require.True(t, ok)
	require.Equal(t, "billing-api", record.Service)
	require.Equal(t, "Pending review", record.Status)
	state.Approve(id)
	state.Approve(id)
	record, ok = Decode(state.Encode()).Find(id)
	require.True(t, ok)
	require.Equal(t, "Approved", record.Status)
	require.Len(t, state.Filter("BILLING"), 1)
	require.Empty(t, state.Filter("missing-service"))
}
func TestStateIsBoundedAndInvalidInputResets(t *testing.T) {
	state := Initial()
	for i := range 20 {
		state.Create(fmt.Sprintf("service-%d", i), "v1", "Production")
	}
	require.Len(t, state.Records, MaxRecords)
	require.Equal(t, state, Decode(state.Encode()))
	require.Equal(t, Initial(), Decode("not valid base64"))
	require.Equal(t, Initial(), Decode(State{Records: []Record{{1, "service", "v1", "Mars", "Approved"}}}.Encode()))
}
