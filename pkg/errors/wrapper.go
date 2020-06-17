package errors

import "fmt"

const (
	UnknownError              = -1
	ConfigError               = 1
	InvalidConfigPathOrFormat = 2
	DownloadError             = 100
	NebulaError               = 200
	NotCompleteError          = 201
)

type ImporterError struct {
	ErrCode int
	ErrMsg  error
}

func (e ImporterError) Error() string {
	return fmt.Sprintf("error code: %d, message: %s", e.ErrCode, e.ErrMsg.Error())
}

func Wrap(code int, err error) ImporterError {
	return ImporterError{
		ErrCode: code,
		ErrMsg:  err,
	}
}
