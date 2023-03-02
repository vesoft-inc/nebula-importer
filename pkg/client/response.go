//go:generate mockgen -source=response.go -destination response_mock.go -package client Response
package client

import (
	"time"
)

type Response interface {
	IsSucceed() bool
	GetLatency() time.Duration
	GetRespTime() time.Duration
	GetError() error
	IsPermanentError() bool
	IsRetryMoreError() bool
}
