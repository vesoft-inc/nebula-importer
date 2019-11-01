package clientmgr

import (
	nebula "github.com/vesoft-inc/nebula-go"

	"github.com/yixinglu/nebula-importer/pkg/config"
)

func NewClientPool(conn config.NebulaClientConnection) (*nebula.GraphClient, error) {
	client, err := nebula.NewClient(conn.Address)
	if err != nil {
		return client, err
	}

	if err = client.Connect(conn.User, conn.Password); err != nil {
		return client, err
	}
	return client, nil
}
