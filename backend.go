package waypointsecrets

import (
	"context"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"strings"
	"sync"
)

// Factory Implements a storage backend and sets this up
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := backend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

type waypointBackend struct {
	*framework.Backend
	lock   sync.RWMutex
	client *waypointClient
}

func backend() *waypointBackend {
	var b = waypointBackend{}

	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),
		PathsSpecial: &logical.Paths{
			LocalStorage: []string{
				framework.WALPrefix,
			},
			SealWrapStorage: []string{
				"config",
				"role/*",
			},
		},
		Paths: framework.PathAppend(
			pathRole(&b),
			[]*framework.Path{
				pathConfig(&b),
				pathCredentials(&b),
			},
		),
		Secrets:     []*framework.Secret{b.waypointUser()},
		BackendType: logical.TypeLogical,
		Invalidate:  b.invalidate,
	}
	return &b
}

func (b *waypointBackend) reset() {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.client = nil
}

func (b *waypointBackend) invalidate(ctx context.Context, key string) {
	if key == "config" {
		b.reset()
	}
}

func (b *waypointBackend) getClient(ctx context.Context, s logical.Storage) (*waypointClient, error) {
	b.lock.RLock()
	unlockFunc := b.lock.RUnlock
	defer func() { unlockFunc() }()

	if b.client != nil {
		return b.client, nil
	}

	b.lock.RUnlock()
	b.lock.Lock()
	unlockFunc = b.lock.Unlock

	//
	config, err := getConfig(ctx, s)
	if err != nil {
		return nil, err
	}

	if config == nil {
		config = new(waypointConfig)
	}

	b.client, err = newClient(config)
	if err != nil {
		return nil, err
	}

	// adding to debug client connection issue
	b.Logger()

	return b.client, nil
}

const backendHelp = `
The Waypoint secrets backend dynamically generates user tokens.
After mounting this backend, credentials to manage Waypoint user tokens
must be configured with the "config/" endpoints.
`
