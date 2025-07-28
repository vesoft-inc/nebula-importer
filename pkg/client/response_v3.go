package client

import (
	"fmt"
	"strings"
	"time"

	nebula "github.com/vesoft-inc/nebula-go/v3"
)

type defaultResponseV3 struct {
	*nebula.ResultSet
	respTime time.Duration
}

func newResponseV3(rs *nebula.ResultSet, respTime time.Duration) Response {
	return defaultResponseV3{
		ResultSet: rs,
		respTime:  respTime,
	}
}

func (resp defaultResponseV3) GetLatency() time.Duration {
	return time.Duration(resp.ResultSet.GetLatency()) * time.Microsecond
}

func (resp defaultResponseV3) GetRespTime() time.Duration {
	return resp.respTime
}

func (resp defaultResponseV3) GetError() error {
	if resp.IsSucceed() {
		return nil
	}
	errorCode := resp.GetErrorCode()
	errorMsg := resp.GetErrorMsg()
	return fmt.Errorf("%d:%s", errorCode, errorMsg)
}

func (resp defaultResponseV3) IsPermanentError() bool {
	switch resp.GetErrorCode() { //nolint:exhaustive
	default:
		return false
	case nebula.ErrorCode_E_SYNTAX_ERROR:
	case nebula.ErrorCode_E_SEMANTIC_ERROR:
	}
	return true
}

func (resp defaultResponseV3) IsRetryMoreError() bool {
	errorMsg := resp.GetErrorMsg()
	// TODO: compare with E_RAFT_BUFFER_OVERFLOW
	// Can not get the E_RAFT_BUFFER_OVERFLOW inside storage now.
	return strings.Contains(errorMsg, "raft buffer is full")
}
