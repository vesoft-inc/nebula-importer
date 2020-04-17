package base

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type Stmt struct {
	Stmt string
	Data [][]interface{}
}

type Record []string

type OpType int

const (
	DONE   OpType = 0
	INSERT OpType = 1
	DELETE OpType = 2
	HEADER OpType = 100
)

func (op OpType) String() string {
	switch op {
	case 0:
		return "DONE"
	case 1:
		return "INSERT"
	case 2:
		return "DELETE"
	case 100:
		return "HEADER"
	default:
		return "UNKNOWN"
	}
}

type Data struct {
	Type   OpType
	Record Record
}

func InsertData(record Record) Data {
	return Data{
		Type:   INSERT,
		Record: record,
	}
}

func DeleteData(record Record) Data {
	return Data{
		Type:   DELETE,
		Record: record,
	}
}

func HeaderData(record Record) Data {
	return Data{
		Type:   HEADER,
		Record: record,
	}
}

var done = Data{
	Type:   DONE,
	Record: nil,
}

func FinishData() Data {
	return done
}

type ErrData struct {
	Error error
	Data  []Data
}

type ResponseData struct {
	Error error
	Stats Stats
}

type ClientRequest struct {
	Stmt  string
	ErrCh chan<- ErrData
	Data  []Data
}

const (
	LABEL_LABEL   = ":LABEL"
	LABEL_VID     = ":VID"
	LABEL_SRC_VID = ":SRC_VID"
	LABEL_DST_VID = ":DST_VID"
	LABEL_RANK    = ":RANK"
	LABEL_IGNORE  = ":IGNORE"
)

func TryConvInt64(cell string) string {
	n, _ := strconv.ParseUint(cell, 10, 64)
	if n > math.MaxInt64 {
		return fmt.Sprint(int64(n))
	} else {
		return cell
	}
}

func TryConvDateTimestamp(cell string, format string) string {
	f := strings.TrimSpace(format)
	c := strings.TrimSpace(cell)
	if len(c) < len(f) {
		f = f[0:len(c)]
	} else {
		c = c[0:len(f)]
	}
	tm, _ := time.Parse(f, c)
	return fmt.Sprint(tm.Unix())
}
