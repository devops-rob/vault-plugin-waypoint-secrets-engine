package waypointsecrets

import (
	"context"
	"errors"
	"fmt"
	waypoint "github.com/hashicorp-dev-advocates/waypoint-client/pkg/client"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/sethvargo/go-password/password"
	"log"
)

const (
	User = "user"
)

type waypointUser struct {
	UserId string `json:"user_id"`
	Token  string `json:"token"`
}

func (b *waypointBackend) waypointUser() *framework.Secret {
	return &framework.Secret{
		Type:   User,
		Revoke: b.userRevoke,
		Renew:  b.userRenew,
		Fields: map[string]*framework.FieldSchema{
			"token": {
				Type:        framework.TypeString,
				Description: "Waypoint user token",
			},
			"user_id": {
				Type:        framework.TypeString,
				Description: "Waypoint User ID",
			},
		},
	}
}

// accountRevoke removes the token from the Vault storage API and calls the client to revoke the token
func (b *waypointBackend) userRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	client, err := b.getClient(ctx, req.Storage)
	if err != nil {
		return nil, fmt.Errorf("error getting client: %w", err)
	}

	userId := ""
	userIdRaw, ok := req.Secret.InternalData["user_id"]
	if ok {
		userId, ok = userIdRaw.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value for user_id in secret internal data")
		}
	}

	if err := deleteUser(ctx, client, userId); err != nil {
		return nil, fmt.Errorf("error revoking user: %w", err)
	}
	return nil, nil
}

// tokenRenew calls the client to create a new token and stores it in the Vault storage API
func (b *waypointBackend) userRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	roleRaw, ok := req.Secret.InternalData["role"]
	if !ok {
		return nil, fmt.Errorf("secret is missing role internal data")
	}

	// get the role entry
	role := roleRaw.(string)
	roleEntry, err := b.getRole(ctx, req.Storage, role)
	if err != nil {
		return nil, fmt.Errorf("error retrieving role: %w", err)
	}

	if roleEntry == nil {
		return nil, errors.New("error retrieving role: role is nil")
	}

	resp := &logical.Response{Secret: req.Secret}

	if roleEntry.TTL > 0 {
		resp.Secret.TTL = roleEntry.TTL
	}
	if roleEntry.MaxTTL > 0 {
		resp.Secret.MaxTTL = roleEntry.MaxTTL
	}

	return resp, nil
}

// createUser calls the Waypoint client and creates a new Waypoint user
func createUser(ctx context.Context, c *waypointClient, role string) (*waypointUser, error) {

	// Setting up the loginName using role_id + randomly generated string
	userNamePostfix, err := password.Generate(8, 0, 0, true, false)
	if err != nil {
		log.Fatal(err)
	}
	userName := `vault-role-` + role + `-` + userNamePostfix

	inviteUsername := userName
	inv, err := c.InviteUser(context.TODO(), inviteUsername, "30s")
	if err != nil {
		panic(err)
	}

	tok, err := c.AcceptInvitation(context.TODO(), inv)
	if err != nil {
		panic(err)
	}

	userId, err := c.GetUser(context.TODO(), waypoint.Username(userName))
	if err != nil {
		panic(err)
	}

	return &waypointUser{
		Token:  tok,
		UserId: userId.Id,
	}, nil
}

// deleteUser calls the Waypoint client to remove the user
func deleteUser(ctx context.Context, c *waypointClient, userId string) error {

	_, err := c.DeleteUser(context.TODO(), waypoint.UserId(userId))
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
