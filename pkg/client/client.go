package client

import (
	nebula "github.com/vesoft-inc/nebula-go"
)

func NewNebulaConnection(addr, user, password string) (*nebula.GraphClient, error) {
	client, err := nebula.NewClient(addr)
	if err != nil {
		return nil, err
	}

	if err = client.Connect(user, password); err != nil {
		return nil, err
	}
	return client, nil
}
