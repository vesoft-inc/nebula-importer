package reader

import (
	"fmt"
	"strings"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type Batch struct {
	statsCh         chan<- base.Stats
	errCh           chan<- base.ErrData
	clientRequestCh chan base.ClientRequest
	responseCh      chan base.ResponseData
	isVertex        bool
	bufferSize      int
	currentIndex    int
	buffer          []base.Data
	batchMgr        *BatchMgr
}

func NewBatch(mgr *BatchMgr, bufferSize int, isVertex bool, clientReq chan base.ClientRequest, statsCh chan<- base.Stats, errCh chan<- base.ErrData) *Batch {
	b := Batch{
		statsCh:         statsCh,
		errCh:           errCh,
		clientRequestCh: clientReq,
		responseCh:      make(chan base.ResponseData),
		bufferSize:      bufferSize,
		isVertex:        isVertex,
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

	b.errCh <- base.ErrData{Error: nil}
}

func (b *Batch) requestClient() {
	var stmt string
	if b.isVertex {
		stmt = b.makeVertexInsertStmtWithoutHeaderLine(b.buffer[:b.currentIndex])
	} else {
		stmt = b.makeEdgeInsertStmtWithoutHeaderLine(b.buffer[:b.currentIndex])
	}
	b.clientRequestCh <- base.ClientRequest{
		Stmt:       stmt,
		ResponseCh: b.responseCh,
	}

	if resp := <-b.responseCh; resp.Error != nil {
		b.errCh <- base.ErrData{
			Error: resp.Error,
			Data:  b.buffer[:b.currentIndex],
		}
	} else {
		stat := resp.Stats
		stat.BatchSize = b.currentIndex + 1
		b.statsCh <- stat
	}

	b.currentIndex = 0
}

func (m *Batch) makeVertexBatchStmt(batch []base.Data) string {
	length := len(batch)
	switch batch[length-1].Type {
	case base.INSERT:
		return m.makeVertexInsertStmtWithoutHeaderLine(batch)
	case base.DELETE:
		return m.makeVertexDeleteStmtWithoutHeaderLine(batch)
	default:
		logger.Log.Fatalf("Invalid data type: %s", batch[length-1].Type)
		return ""
	}
}

func (m *Batch) makeEdgeBatchStmt(batch []base.Data) string {
	length := len(batch)
	switch batch[length-1].Type {
	case base.INSERT:
		return m.makeEdgeInsertStmtWithoutHeaderLine(batch)
	case base.DELETE:
		logger.Log.Fatal("Unsupported delete edge")
	default:
		logger.Log.Fatalf("Invalid data type: %s", batch[length-1].Type)
	}
	return ""
}

func (m *Batch) makeVertexStmtWithoutHeaderLine(batch []base.Data) string {
	if len(batch) == 0 {
		logger.Log.Fatal("Make vertex stmt for empty batch")
	}

	if len(batch) == 1 {
		return m.makeVertexBatchStmt(batch)
	}

	var builder strings.Builder
	lastIdx, length := 0, len(batch)
	for i := 1; i < length; i++ {
		if batch[i-1].Type != batch[i].Type {
			builder.WriteString(m.makeVertexBatchStmt(batch[lastIdx:i]))
			lastIdx = i
		}
	}
	builder.WriteString(m.makeVertexBatchStmt(batch[lastIdx:]))
	return builder.String()
}

func (m *Batch) makeVertexInsertStmtWithoutHeaderLine(data []base.Data) string {
	propTypeMap := make(map[int]string)
	var numProps int = 0
	var builder strings.Builder
	builder.WriteString("INSERT VERTEX ")
	for i, tag := range m.batchMgr.Schema.Vertex.Tags {
		builder.WriteString(fmt.Sprintf("%s(", tag.Name))
		for j, prop := range tag.Props {
			builder.WriteString(prop.Name)
			propTypeMap[numProps] = prop.Type
			numProps++
			if j < len(tag.Props)-1 {
				builder.WriteString(",")
			}
		}
		builder.WriteString(")")
		if i < len(m.batchMgr.Schema.Vertex.Tags)-1 {
			builder.WriteString(",")
		}
	}
	builder.WriteString(" VALUES ")
	batchSize := len(data)
	for i := 0; i < batchSize; i++ {
		builder.WriteString(fmt.Sprintf(" %s: ", data[i].Record[0]))
		fillVertexPropsValues(&builder, data[i].Record, i == batchSize-1, propTypeMap)
	}

	return builder.String()
}

func (m *Batch) makeVertexDeleteStmtWithoutHeaderLine(data []base.Data) string {
	var builder strings.Builder
	for _, d := range data {
		// TODO: delete vertex in batch
		builder.WriteString(fmt.Sprintf("DELETE VERTEX %s;", d.Record[0]))
	}
	return builder.String()
}

func (m *Batch) makeEdgeStmtWithoutHeaderLine(batch []base.Data) string {
	if len(batch) == 0 {
		logger.Log.Fatal("Fail to make edge stmt for empty batch")
	}
	length := len(batch)
	if length == 1 {
		return m.makeEdgeBatchStmt(batch)
	}

	var builder strings.Builder
	lastIdx := 0

	for i := 1; i < length; i++ {
		if batch[i-1].Type != batch[i].Type {
			builder.WriteString(m.makeEdgeBatchStmt(batch[lastIdx:i]))
			lastIdx = i
		}
	}
	builder.WriteString(m.makeEdgeBatchStmt(batch[lastIdx:]))
	return builder.String()
}

func (m *Batch) makeEdgeInsertStmtWithoutHeaderLine(batch []base.Data) string {
	numProps := 0
	propTypeMap := make(map[int]string)
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("INSERT EDGE %s(", m.batchMgr.Schema.Edge.Name))
	for i, prop := range m.batchMgr.Schema.Edge.Props {
		builder.WriteString(prop.Name)
		propTypeMap[numProps] = prop.Type
		numProps++
		if i < len(m.batchMgr.Schema.Edge.Props)-1 {
			builder.WriteString(",")
		}
	}
	builder.WriteString(") VALUES ")
	batchSize := len(batch)
	for i := 0; i < batchSize; i++ {
		builder.WriteString(fmt.Sprintf("%s->%s", batch[i].Record[0], batch[i].Record[1]))
		if m.batchMgr.Schema.Edge.WithRanking {
			builder.WriteString(fmt.Sprintf("@%s", batch[i].Record[2]))
		}
		builder.WriteString(":")
		m.fillEdgePropsValues(&builder, batch[i].Record, i == batchSize-1, propTypeMap)
	}
	return builder.String()
}

func fillVertexPropsValues(builder *strings.Builder, record base.Record, isEnd bool, propTypeMap map[int]string) {
	builder.WriteString("(")
	for i := 1; i < len(record); i++ {
		if strings.ToLower(propTypeMap[i-1]) == "string" {
			builder.WriteString(fmt.Sprintf("%q", record[i]))
		} else {
			builder.WriteString(record[i])
		}

		if i < len(record)-1 {
			builder.WriteString(",")
		}
	}
	builder.WriteString(")")
	if isEnd {
		builder.WriteString(";")
	} else {
		builder.WriteString(",")
	}
}

func (m *Batch) fillEdgePropsValues(builder *strings.Builder, record base.Record, isEnd bool, propTypeMap map[int]string) {
	fromIdx := 2
	if m.batchMgr.Schema.Edge.WithRanking {
		fromIdx = 3
	}
	if fromIdx > len(record) {
		logger.Log.Fatalf("Invalid record for edge: %v", record)
	}
	builder.WriteString("(")
	for i := fromIdx; i < len(record); i++ {
		if strings.ToLower(propTypeMap[i-fromIdx]) == "string" {
			builder.WriteString(fmt.Sprintf("%q", record[i]))
		} else {
			builder.WriteString(record[i])
		}
		if i < len(record)-1 {
			builder.WriteString(",")
		}
	}
	builder.WriteString(")")
	if isEnd {
		builder.WriteString(";")
	} else {
		builder.WriteString(",")
	}
}
