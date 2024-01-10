//go:generate mockgen -source=session.go -destination session_mock.go -package client Session
package client

import (
	"crypto/tls"
	"fmt"
	"time"

	nebula "github.com/vesoft-inc/nebula-go/v3"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/logger"
)

type (
	defaultSessionV3 struct {
		session      *nebula.Session
		hostAddress  nebula.HostAddress
		user         string
		password     string
		handshakeKey string
		tlsConfig    *tls.Config
		logger       logger.Logger
	}
)

//revive:disable-next-line:argument-limit
func newSessionV3(hostAddress HostAddress, user, password, handshakeKey string, tlsConfig *tls.Config, l logger.Logger) Session {
	if l == nil {
		l = logger.NopLogger
	}
	return &defaultSessionV3{
		hostAddress: nebula.HostAddress{
			Host: hostAddress.Host,
			Port: hostAddress.Port,
		},
		user:         user,
		password:     password,
		handshakeKey: handshakeKey,
		tlsConfig:    tlsConfig,
		logger:       l,
	}
}

func (s *defaultSessionV3) Open() error {
	hostAddress := s.hostAddress
	pool, err := nebula.NewSslConnectionPool(
		[]nebula.HostAddress{hostAddress},
		nebula.PoolConfig{
			MaxConnPoolSize: 1,
			HandshakeKey:    s.handshakeKey,
		},
		s.tlsConfig,
		newNebulaLogger(s.logger.With(logger.Field{
			Key:   "address",
			Value: fmt.Sprintf("%s:%d", hostAddress.Host, hostAddress.Port),
		})),
	)
	if err != nil {
		return err
	}

	session, err := pool.GetSession(s.user, s.password)
	if err != nil {
		return err
	}

	s.session = session

	return nil
}

func (s *defaultSessionV3) Execute(statement string) (Response, error) {
	startTime := time.Now()
	rs, err := s.session.Execute(statement)
	if err != nil {
		return nil, err
	}
	return newResponseV3(rs, time.Since(startTime)), nil
}

func (s *defaultSessionV3) Close() error {
	s.session.Release()
	return nil
}
