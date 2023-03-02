//go:generate mockgen -source=session.go -destination session_mock.go -package client Session
package client

type Session interface {
	Open() error
	Execute(statement string) (Response, error)
	Close() error
}
