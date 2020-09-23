package reader

import (
	"fmt"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
)

type Batch struct {
	errCh           chan<- base.ErrData
	clientRequestCh chan base.ClientRequest
	bufferSize      int
	currentIndex    int
	buffer          []base.Data
	batchMgr        *BatchMgr
}

func NewBatch(mgr *BatchMgr, bufferSize int, clientReq chan base.ClientRequest, errCh chan<- base.ErrData) *Batch {
	b := Batch{
		errCh:           errCh,
		clientRequestCh: clientReq,
		bufferSize:      bufferSize,
		currentIndex:    0,
		buffer:          make([]base.Data, bufferSize),
		batchMgr:        mgr,
	}
	return &b
}

func (b *Batch) IsFull() bool {
	return b.currentIndex == b.bufferSize
}

func (b *Batch) Add(data base.Data) {
	if b.IsFull() {
		b.requestClient()
	}
	b.buffer[b.currentIndex] = data
	b.currentIndex++
}

func (b *Batch) Done() {
	if b.currentIndex > 0 {
		b.requestClient()
	}

	b.clientRequestCh <- base.ClientRequest{
		ErrCh: b.errCh,
		Stmt:  base.STAT_FILEDONE,
	}
}

func (b *Batch) requestClient() {
	var stmt string
	var err error
	if b.batchMgr.Schema.IsVertex() {
		stmt, err = b.batchMgr.MakeVertexStmt(b.buffer[:b.currentIndex])
	} else {
		stmt, err = b.batchMgr.MakeEdgeStmt(b.buffer[:b.currentIndex])
	}

	if err != nil {
		stmt = fmt.Sprintf("%s(%s)", "THERE_ARE_SOME_ERRORS", err.Error())
	}

	b.clientRequestCh <- base.ClientRequest{
		Stmt:  stmt,
		ErrCh: b.errCh,
		Data:  b.buffer[:b.currentIndex],
	}

	b.currentIndex = 0
}

func (b *Batch) SendErrorData(d base.Data, err error) {
	b.errCh <- base.ErrData{
		Error: err,
		Data:  []base.Data{d},
	}
}
