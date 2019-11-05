package client

import (
	nebula "github.com/vesoft-inc/nebula-go"

	"github.com/yixinglu/nebula-importer/pkg/config"
)

func NewNebulaConnection(conn config.NebulaClientConnection) (*nebula.GraphClient, error) {
	client, err := nebula.NewClient(conn.Address)
	if err != nil {
		return nil, err
	}

	if err = client.Connect(conn.User, conn.Password); err != nil {
		return nil, err
	}
	return client, nil
}
