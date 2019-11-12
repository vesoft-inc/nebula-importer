package reader

import (
	"strconv"
	"strings"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
)

type BatchMgr struct {
	Schema  config.Schema
	Batches []*Batch
}

func NewBatchMgr(schema config.Schema, batchSize int, clientRequestChs []chan base.ClientRequest, statsCh chan<- base.Stats, errCh chan<- base.ErrData) *BatchMgr {
	bm := BatchMgr{
		Schema:  schema,
		Batches: make([]*Batch, len(clientRequestChs)),
	}

	isVertex := true
	if strings.ToUpper(bm.Schema.Type) == "EDGE" {
		isVertex = false
	}

	for i := range bm.Batches {
		bm.Batches[i] = NewBatch(&bm, batchSize, isVertex, clientRequestChs[i], statsCh, errCh)
	}
	return &bm
}

func (bm *BatchMgr) Done() {
	for i := range bm.Batches {
		bm.Batches[i].Done()
	}
}

func (bm *BatchMgr) InitSchema(record base.Record) {

}

func (bm *BatchMgr) Add(data base.Data) error {
	if batchIdx, err := getBatchId(data.Record[0], len(bm.Batches)); err != nil {
		return err
	} else {
		bm.Batches[batchIdx].Add(data)
		return nil
	}
}

func getBatchId(idStr string, numChans int) (uint, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	if id < 0 {
		id = -id
	}
	return uint(id % int64(numChans)), nil
}
