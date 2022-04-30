package waypointsecrets

import (
	"context"
	waypoint "github.com/hashicorp-dev-advocates/waypoint-client/pkg/client"
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
)

const (
	envVarRunAccTests   = "VAULT_ACC"
	envVarWaypointToken = "TEST_WAYPOINT_PASSWORD"
	envVarWaypointAddr  = "TEST_WAYPOINT_ADDR"
)

// getTestBackend will help you construct a test backend object.
// Update this function with your target backend.
func getTestBackend(tb testing.TB) (*waypointBackend, logical.Storage) {
	tb.Helper()

	config := logical.TestBackendConfig()
	config.StorageView = new(logical.InmemStorage)
	config.Logger = hclog.NewNullLogger()
	config.System = logical.TestSystemView()

	b, err := Factory(context.Background(), config)
	if err != nil {
		tb.Fatal(err)
	}

	return b.(*waypointBackend), config.StorageView
}

// runAcceptanceTests will separate unit tests from
// acceptance tests, which will make active requests
// to your target API.
var runAcceptanceTests = os.Getenv(envVarRunAccTests) == "1"

// testEnv creates an object to store and track testing environment
// resources
type testEnv struct {
	Token string
	Addr  string

	Backend logical.Backend
	Context context.Context
	Storage logical.Storage

	// SecretToken tracks the API token, for checking rotations
	UserId string

	//// Tokens tracks the generated tokens, to make sure we clean up
	//AccountId string
}

// AddConfig adds the configuration to the test backend.
// Make sure data includes all of the configuration
// attributes you need and the `config` path!
func (e *testEnv) AddConfig(t *testing.T) {
	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "config",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"token": e.Token,
			"addr":  e.Addr,
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, resp)
	require.Nil(t, err)
}

// AddUserTokenRole adds a role for the Waypoint
// user token.
func (e *testEnv) AddUserTokenRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "role/test-user-token",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"user_id": e.UserId,
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, resp)
	require.Nil(t, err)
}

// ReadUserToken retrieves the user token
// based on a Vault role.
func (e *testEnv) ReadUserToken(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "creds/test-user-token",
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, err)
	require.NotNil(t, resp)

	if t, ok := resp.Data["token"]; ok {
		e.Token = t.(string)
	}
	require.NotEmpty(t, resp.Data["token"])

	if t, ok := resp.Data["user_id"]; ok {
		e.UserId = t.(string)
	}
	require.NotEmpty(t, resp.Data["user_id"])

}

// CleanupUserTokens removes the tokens
// when the test completes.
func (e *testEnv) CleanupUserTokens(t *testing.T) { // TODO - Update these fields to reflect Waypoint domain

	b := e.Backend.(*waypointBackend)
	client, err := b.getClient(e.Context, e.Storage)
	if err != nil {
		t.Fatal("fatal getting client")
	}

	_, err = client.DeleteUser(context.TODO(), waypoint.UserId(e.UserId))
	if err != nil {
		t.Fatalf("unexpected error deleting user: %s", err)
	}
}
