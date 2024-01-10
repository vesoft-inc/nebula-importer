package client

import (
	"crypto/tls"
	"strings"
	"time"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/logger"
)

const (
	DefaultUser                     = "root"
	DefaultPassword                 = "nebula"
	DefaultReconnectInitialInterval = time.Second
	DefaultReconnectMaxInterval     = 2 * time.Minute
	DefaultRetry                    = 3
	DefaultRetryInitialInterval     = time.Second
	DefaultRetryMaxInterval         = 2 * time.Minute
	DefaultRetryRandomizationFactor = 0.1
	DefaultRetryMultiplier          = 1.5
	DefaultRetryMaxElapsedTime      = time.Hour
	DefaultConcurrencyPerAddress    = 10
	DefaultQueueSize                = 1000
)

type (
	Option func(*options)

	options struct {
		// for client
		addresses            []string
		user                 string
		password             string
		handshakeKey         string
		tlsConfig            *tls.Config
		retry                int
		retryInitialInterval time.Duration
		logger               logger.Logger
		fnNewSession         NewSessionFunc
		clientInitFunc       func(Client) error
		// for pool
		reconnectInitialInterval time.Duration
		concurrencyPerAddress    int
		queueSize                int
		fnNewClientWithOptions   func(o *options) Client // for convenience of testing in Pool
	}
)

func WithV3() Option {
	return func(o *options) {
		WithNewSessionFunc(func(hostAddress HostAddress) Session {
			return newSessionV3(hostAddress, o.user, o.password, o.handshakeKey, o.tlsConfig.Clone(), o.logger)
		})(o)
	}
}

func WithAddress(addresses ...string) Option {
	return func(o *options) {
		for _, addr := range addresses {
			if strings.IndexByte(addr, ',') != -1 {
				o.addresses = append(o.addresses, strings.Split(addr, ",")...)
			} else {
				o.addresses = append(o.addresses, addr)
			}
		}
	}
}

func WithUser(user string) Option {
	return func(o *options) {
		o.user = user
	}
}

func WithPassword(password string) Option {
	return func(o *options) {
		o.password = password
	}
}

func WithUserPassword(user, password string) Option {
	return func(o *options) {
		WithUser(user)(o)
		WithPassword(password)(o)
	}
}

func WithHandshakeKey(handshakeKey string) Option {
	return func(o *options) {
		o.handshakeKey = handshakeKey
	}
}

func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(o *options) {
		o.tlsConfig = tlsConfig
	}
}

func WithRetry(retry int) Option {
	return func(o *options) {
		if retry > 0 {
			o.retry = retry
		}
	}
}

func WithRetryInitialInterval(interval time.Duration) Option {
	return func(o *options) {
		if interval > 0 {
			o.retryInitialInterval = interval
		}
	}
}

func WithLogger(l logger.Logger) Option {
	return func(o *options) {
		o.logger = l
	}
}

func WithNewSessionFunc(fn NewSessionFunc) Option {
	return func(o *options) {
		o.fnNewSession = fn
	}
}

func WithClientInitFunc(fn func(Client) error) Option {
	return func(o *options) {
		o.clientInitFunc = fn
	}
}

func WithReconnectInitialInterval(interval time.Duration) Option {
	return func(o *options) {
		if interval > 0 {
			o.reconnectInitialInterval = interval
		}
	}
}

func WithConcurrencyPerAddress(concurrencyPerAddress int) Option {
	return func(o *options) {
		if concurrencyPerAddress > 0 {
			o.concurrencyPerAddress = concurrencyPerAddress
		}
	}
}

func WithQueueSize(queueSize int) Option {
	return func(o *options) {
		if queueSize > 0 {
			o.queueSize = queueSize
		}
	}
}

func newOptions(opts ...Option) *options {
	var defaultOptions = &options{
		user:                     DefaultUser,
		password:                 DefaultPassword,
		reconnectInitialInterval: DefaultReconnectInitialInterval,
		retry:                    DefaultRetry,
		retryInitialInterval:     DefaultRetryInitialInterval,
		concurrencyPerAddress:    DefaultConcurrencyPerAddress,
		queueSize:                DefaultQueueSize,
	}

	defaultOptions.withOptions(opts...)

	return defaultOptions
}

func (o *options) withOptions(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}

	if o.logger == nil {
		o.logger = logger.NopLogger
	}

	if o.fnNewSession == nil {
		WithV3()(o)
	}

	if o.fnNewClientWithOptions == nil {
		o.fnNewClientWithOptions = newClientWithOptions
	}
}

func (o *options) clone() *options {
	cpy := *o
	return &cpy
}
