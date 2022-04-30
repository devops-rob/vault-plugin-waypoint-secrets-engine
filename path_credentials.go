package waypointsecrets

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// pathCredentials extends the Vault API with a `/creds`
// endpoint for a role. You can choose whether
// or not certain attributes should be displayed,
// required, and named.
func pathCredentials(b *waypointBackend) *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeLowerCaseString,
				Description: "Name of the role",
				Required:    true,
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathCredentialsRead,
			logical.UpdateOperation: b.pathCredentialsRead,
		},
		HelpSynopsis:    pathCredentialsHelpSyn,
		HelpDescription: pathCredentialsHelpDesc,
	}
}

// pathCredentialsRead creates a new Waypoint user token each time it is called if a
// role exists.
func (b *waypointBackend) pathCredentialsRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	roleName := d.Get("name").(string)

	roleEntry, err := b.getRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving role: %w", err)
	}

	if roleEntry == nil {
		return nil, errors.New("error retrieving role: role is nil")
	}

	return b.createUserCreds(ctx, req, roleEntry)
}

// createUserCreds creates a new Waypoint user token to store into the Vault backend, generates
// a response with the secrets information, and checks the TTL and MaxTTL attributes.
func (b *waypointBackend) createUserCreds(ctx context.Context, req *logical.Request, role *waypointRoleEntry) (*logical.Response, error) {
	account, err := b.createUser(ctx, req.Storage, role)
	if err != nil {
		return nil, err
	}

	// The response is divided into two objects (1) internal data and (2) data.
	// If you want to reference any information in your code, you need to
	// store it in internal data!
	resp := b.Secret(User).Response(map[string]interface{}{
		"user_id": account.UserId,
		"token":   account.Token,
	}, map[string]interface{}{
		"user_id": account.UserId,
	})

	if role.TTL > 0 {
		resp.Secret.TTL = role.TTL
	}

	if role.MaxTTL > 0 {
		resp.Secret.MaxTTL = role.MaxTTL
	}

	return resp, nil
}

// createUser uses the Waypoint client to create a new user
func (b *waypointBackend) createUser(ctx context.Context, s logical.Storage, roleEntry *waypointRoleEntry) (*waypointUser, error) {
	client, err := b.getClient(ctx, s)
	if err != nil {
		return nil, err
	}

	var token *waypointUser

	token, err = createUser(ctx, client, roleEntry.Name)
	if err != nil {
		return nil, fmt.Errorf("error creating Waypoint user: %w", err)
	}

	if token == nil {
		return nil, errors.New("error creating Waypoint user")
	}

	return token, nil

}

const pathCredentialsHelpSyn = `
Generate a Waypoint user from a specific Vault role.
`

const pathCredentialsHelpDesc = `
This path generates a Waypoint token
based on a particular role.
`
