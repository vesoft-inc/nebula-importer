package client

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vesoft-inc/nebula-go/graph"
	"github.com/yixinglu/nebula-importer/pkg/base"
	"github.com/yixinglu/nebula-importer/pkg/config"
	"github.com/yixinglu/nebula-importer/pkg/stats"
)

type NebulaClientMgr struct {
	config  config.NebulaClientSettings
	file    config.File
	errCh   chan base.ErrData
	statsCh chan<- stats.Stats
	pool    *ClientPool
}

func NewNebulaClientMgr(settings config.NebulaClientSettings, statsCh chan<- stats.Stats) *NebulaClientMgr {
	mgr := NebulaClientMgr{
		config:  settings,
		errCh:   make(chan base.ErrData),
		statsCh: statsCh,
	}

	mgr.pool = NewClientPool(settings)
	mgr.startWorkers()

	log.Printf("Create %d Nebula Graph clients", mgr.config.Concurrency)

	return &mgr
}

func (m *NebulaClientMgr) Close() {
	m.statsCh <- stats.Stats{Type: stats.PRINT}
	m.pool.Close()
	close(m.errCh)
}

func (m *NebulaClientMgr) GetDataChans() []chan base.Data {
	return m.pool.DataChs
}

func (m *NebulaClientMgr) GetErrChan() <-chan base.ErrData {
	return m.errCh
}

func (m *NebulaClientMgr) InitFile(file config.File) {
	m.file = file
	for i := 0; i < m.config.Concurrency; i++ {
		stmt := fmt.Sprintf("Use %d;", file.Schema.Space)
		resp, err := m.pool.Conns[i].Execute(stmt)
		if err != nil {
			log.Fatalf("Client %d can not switch space %s, error: %v, %s",
				i, file.Schema.Space, resp.GetErrorCode(), resp.GetErrorMsg())
		}
	}
}

func (m *NebulaClientMgr) closeErrChan(opType base.OpType) {
	if opType == base.DONE {
		m.errCh <- base.ErrData{Error: nil}
	}
}

func (m *NebulaClientMgr) startWorkers() {
	for i := 0; i < m.config.Concurrency; i++ {
		go func(i int) {
			batchSize := 0
			batch := make([]base.Data, m.file.BatchSize)
			for {
				data, ok := <-m.pool.DataChs[i]
				if !ok {
					break
				}

				switch data.Type {
				case base.DONE:
					// Need not to notify error handler. Reset it in outside main program
					if batchSize == 0 {
						m.closeErrChan(data.Type)
						continue
					}
				case base.HEADER:
					// TODO:
				default:
					batch[batchSize] = data
					batchSize++

					if batchSize < m.file.BatchSize {
						continue
					}
				}

				var stmt string
				switch strings.ToUpper(m.file.Schema.Type) {
				case "VERTEX":
					stmt = m.makeVertexStmtWithoutHeaderLine(batch)
				case "EDGE":
					stmt = m.makeEdgeStmtWithoutHeaderLine(batch)
				default:
					log.Fatalf("Error schema type: %s", m.file.Schema.Type)
				}

				now := time.Now()
				resp, err := m.pool.Conns[i].Execute(stmt)
				reqTime := time.Since(now).Seconds()

				if err != nil {
					m.errCh <- base.ErrData{
						Error: err,
						Data:  data,
					}
					m.closeErrChan(data.Type)
					continue
				}

				if resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
					errMsg := fmt.Sprintf("Fail to execute: %s, ErrMsg: %s, ErrCode: %v", stmt, resp.GetErrorMsg(), resp.GetErrorCode())
					m.errCh <- base.ErrData{
						Error: errors.New(errMsg),
						Data:  data,
					}
					m.closeErrChan(data.Type)
					continue
				}

				m.statsCh <- stats.Stats{
					Latency: uint64(resp.GetLatencyInUs()),
					ReqTime: reqTime,
				}

				m.closeErrChan(data.Type)
				batchSize = 0
			}
		}(i)
	}
}

func (m *NebulaClientMgr) makeVertexBatchStmt(batch []base.Data) string {
	length := len(batch)
	switch batch[length-1].Type {
	case base.INSERT:
		return m.makeVertexInsertStmtWithoutHeaderLine(batch)
	case base.DELETE:
		return m.makeVertexDeleteStmtWithoutHeaderLine(batch)
	default:
		log.Fatalf("Invalid data type: %s", batch[length-1].Type)
		return ""
	}
}

func (m *NebulaClientMgr) makeEdgeBatchStmt(batch []base.Data) string {
	length := len(batch)
	switch batch[length-1].Type {
	case base.INSERT:
		return m.makeEdgeInsertStmtWithoutHeaderLine(batch)
	case base.DELETE:
		log.Fatal("Unsupported delete edge")
	default:
		log.Fatalf("Invalid data type: %s", batch[length-1].Type)
	}
	return ""
}

func (m *NebulaClientMgr) makeVertexStmtWithoutHeaderLine(batch []base.Data) string {
	if len(batch) == 0 {
		log.Fatal("Make vertex stmt for empty batch")
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

func (m *NebulaClientMgr) makeVertexInsertStmtWithoutHeaderLine(data []base.Data) string {
	var builder strings.Builder
	builder.WriteString("INSERT VERTEX ")
	for i, tag := range m.file.Schema.Vertex.Tags {
		builder.WriteString(fmt.Sprintf("%s(", tag.Name))
		for j, prop := range tag.Props {
			builder.WriteString(prop.Name)
			if j < len(tag.Props)-1 {
				builder.WriteString(",")
			} else {
				builder.WriteString(")")
			}
		}
		if i < len(m.file.Schema.Vertex.Tags)-1 {
			builder.WriteString(",")
		} else {
			builder.WriteString(" VALUES ")
		}
	}
	batchSize := len(data)
	for i := 0; i < batchSize; i++ {
		builder.WriteString(fmt.Sprintf(" %s: ", data[i].Record[0]))
		fillVertexPropsValues(&builder, data[i].Record, i == batchSize-1)
	}

	return builder.String()
}

func (m *NebulaClientMgr) makeVertexDeleteStmtWithoutHeaderLine(data []base.Data) string {
	var builder strings.Builder
	builder.WriteString("DELETE VERTEX ")
	for i, d := range data {
		builder.WriteString(d.Record[0])
		if i == len(data)-1 {
			builder.WriteString(";")
		} else {
			builder.WriteString(",")
		}
	}
	return builder.String()
}

func (m *NebulaClientMgr) makeEdgeStmtWithoutHeaderLine(batch []base.Data) string {
	if len(batch) == 0 {
		log.Fatal("Fail to make edge stmt for empty batch")
	}

	if len(batch) == 1 {
		return m.makeEdgeBatchStmt(batch)
	}

	var builder strings.Builder
	lastIdx := 0
	for i := range batch {
		if batch[i-1].Type != batch[i].Type {
			builder.WriteString(m.makeEdgeBatchStmt(batch[lastIdx:i]))
			lastIdx = i
		}
	}
	builder.WriteString(m.makeEdgeBatchStmt(batch[lastIdx:]))
	return builder.String()
}

func (m *NebulaClientMgr) makeEdgeInsertStmtWithoutHeaderLine(batch []base.Data) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("INSERT EDGE %s(", m.file.Schema.Edge.Name))
	for i, prop := range m.file.Schema.Edge.Props {
		builder.WriteString(prop.Name)
		if i < len(m.file.Schema.Edge.Props)-1 {
			builder.WriteString(",")
		} else {
			builder.WriteString(")")
		}
	}
	builder.WriteString(" VALUES ")
	batchSize := len(batch)
	for i := 0; i < batchSize; i++ {
		builder.WriteString(fmt.Sprintf("%s->%s", batch[i].Record[0], batch[i].Record[1]))
		if m.file.Schema.Edge.WithRanking {
			builder.WriteString(fmt.Sprintf("@%s", batch[i].Record[2]))
		}
		builder.WriteString(":")
		m.fillEdgePropsValues(&builder, batch[i].Record, i == batchSize-1)
	}
	return builder.String()
}

func fillVertexPropsValues(builder *strings.Builder, record base.Record, isEnd bool) {
	builder.WriteString("(")
	for i := 1; i < len(record); i++ {
		builder.WriteString(record[i])
		if i < len(record)-1 {
			builder.WriteString(",")
		} else {
			builder.WriteString(")")
		}
	}
	if isEnd {
		builder.WriteString(";")
	} else {
		builder.WriteString(",")
	}
}

func (m *NebulaClientMgr) fillEdgePropsValues(builder *strings.Builder, record base.Record, isEnd bool) {
	fromIdx := 2
	if m.file.Schema.Edge.WithRanking {
		fromIdx = 3
	}
	if fromIdx >= len(record) {
		log.Fatalf("Invalid record for edge: %v", record)
	}
	builder.WriteString("(")
	for i := fromIdx; i < len(record); i++ {
		builder.WriteString(record[i])
		if i < len(record)-1 {
			builder.WriteString(",")
		} else {
			builder.WriteString(")")
		}
	}
	if isEnd {
		builder.WriteString(";")
	} else {
		builder.WriteString(",")
	}
}
