package waypointsecrets

import (
	"errors"
	waypoint "github.com/hashicorp-dev-advocates/waypoint-client/pkg/client"
	"log"
)

type waypointClient struct {
	waypoint.Waypoint
}

func newClient(config *waypointConfig) (*waypointClient, error) {

	if config == nil {
		return nil, errors.New("client configuration was nil")
	}

	if config.Token == "" {
		return nil, errors.New("token was not defined")
	}

	if config.Addr == "" {
		return nil, errors.New("waypoint server address was not defined")
	}

	// create a client
	conf := waypoint.DefaultConfig()
	conf.Token = config.Token
	conf.Address = config.Addr

	c, err := waypoint.New(conf)
	if err != nil {
		log.Fatal(err)
	}

	return &waypointClient{c}, nil
}
