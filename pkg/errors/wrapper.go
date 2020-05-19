package errors

const (
	kUnknownErrorCode        = -1
	kConfigErrorCode         = 1
	kInvalidConfigPathOrFile = 2
	kDownloadErrorCode       = 100
	kNebulaErrorCode         = 200
)

type ImporterError struct {
	ErrCode int
	ErrMsg  error
}

func (e ImporterError) Error() string {
	return e.ErrMsg.Error()
}

type DownloadError struct {
	ImporterError
}

func (e DownloadError) Error() string {
	return e.ImporterError.Error()
}

func NewDownloadError(err error) DownloadError {
	return DownloadError{
		ImporterError: ImporterError{
			ErrCode: kDownloadErrorCode,
			ErrMsg:  err,
		},
	}
}

type ConfigError struct {
	ImporterError
}

func (e ConfigError) Error() string {
	return e.ImporterError.Error()
}

func NewConfigError(err error) ConfigError {
	return ConfigError{
		ImporterError: ImporterError{
			ErrCode: kConfigErrorCode,
			ErrMsg:  err,
		},
	}
}

type NebulaError struct {
	ImporterError
}

func (e NebulaError) Error() string {
	return e.ImporterError.Error()
}

func NewNebulaError(err error) NebulaError {
	return NebulaError{
		ImporterError: ImporterError{
			ErrCode: kNebulaErrorCode,
			ErrMsg:  err,
		},
	}
}

type UnknownError struct {
	ImporterError
}

func (e UnknownError) Error() string {
	return e.ImporterError.Error()
}

func NewUnknownError(err error) UnknownError {
	return UnknownError{
		ImporterError: ImporterError{
			ErrCode: kUnknownErrorCode,
			ErrMsg:  err,
		},
	}
}

type InvalidConfigPathOrFormat struct {
	ImporterError
}

func (e InvalidConfigPathOrFormat) Error() string {
	return e.ImporterError.Error()
}

func NewInvalidConfigPathOrFormat(err error) InvalidConfigPathOrFormat {
	return InvalidConfigPathOrFormat{
		ImporterError: ImporterError{
			ErrCode: kInvalidConfigPathOrFile,
			ErrMsg:  err,
		},
	}
}
