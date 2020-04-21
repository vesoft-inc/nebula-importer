package client

import (
	"time"

	nebula "github.com/vesoft-inc/nebula-go"
)

func NewNebulaConnection(addr, user, password string) (*nebula.GraphClient, error) {
	opts := nebula.WithTimeout(10 * time.Second)
	client, err := nebula.NewClient(addr, opts)
	if err != nil {
		return nil, err
	}

	if err = client.Connect(user, password); err != nil {
		return nil, err
	}
	return client, nil
}
