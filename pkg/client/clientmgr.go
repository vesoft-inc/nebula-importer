package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/vesoft-inc/nebula-go/graph"
	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
	"github.com/vesoft-inc/nebula-importer/pkg/stats"
)

type NebulaClientMgr struct {
	config  config.NebulaClientSettings
	file    config.File
	errCh   chan base.ErrData
	statsCh chan<- stats.Stats
	pool    *ClientPool
}

func NewNebulaClientMgr(settings config.NebulaClientSettings, statsCh chan<- stats.Stats) (*NebulaClientMgr, error) {
	mgr := NebulaClientMgr{
		config:  settings,
		errCh:   make(chan base.ErrData),
		statsCh: statsCh,
	}

	if mgr.pool, err = NewClientPool(settings); err != nil {
		return nil, err
	}

	logger.Log.Printf("Create %d Nebula Graph clients", mgr.config.Concurrency)

	return &mgr, nil
}

func (m *NebulaClientMgr) Close() {
	m.pool.Close()
	close(m.errCh)
}

func (m *NebulaClientMgr) GetDataChans() []chan base.Data {
	return m.pool.DataChs
}

func (m *NebulaClientMgr) GetErrChan() <-chan base.ErrData {
	return m.errCh
}

func (m *NebulaClientMgr) InitFile(file config.File) error {
	m.file = file
	for i := 0; i < m.config.Concurrency; i++ {
		stmt := fmt.Sprintf("USE %s;", file.Schema.Space)
		resp, err := m.pool.Conns[i].Execute(stmt)
		if err != nil {
			return fmt.Errorf("Client %d can not switch space %s, error: %v, %s",
				i, file.Schema.Space, resp.GetErrorCode(), resp.GetErrorMsg())
		}
	}
	return m.startWorkers()
}

func (m *NebulaClientMgr) isVertex() (bool, error) {
	switch strings.ToUpper(m.file.Schema.Type) {
	case "VERTEX":
		return true, nil
	case "EDGE":
		return false, nil
	default:
		return false, fmt.Errorf("Error schema type: %s", m.file.Schema.Type)
	}
}

func (m *NebulaClientMgr) startWorkers() error {
	isVertex, err := m.isVertex()
	if err != nil {
		return err
	}

	for i := 0; i < m.config.Concurrency; i++ {
		go func(i int) {
			batchSize := 0
			batch := make([]base.Data, m.file.BatchSize)
			for {
				data, ok := <-m.pool.DataChs[i]
				if !ok {
					break
				}

				if data.Type == base.DONE {
					if batchSize == 0 {
						break
					}
				} else if data.Type == base.HEADER {
					// TODO:
					logger.Log.Fatal("Unsupported HEADER data type")
				} else {
					batch[batchSize] = data
					batchSize++

					if batchSize < m.file.BatchSize {
						continue
					}
				}

				var stmt string
				if isVertex {
					stmt = m.makeVertexStmtWithoutHeaderLine(batch[:batchSize])
				} else {
					stmt = m.makeEdgeStmtWithoutHeaderLine(batch[:batchSize])
				}

				now := time.Now()
				if resp, err := m.pool.Conns[i].Execute(stmt); err != nil {
					m.errCh <- base.ErrData{
						Error: err,
						Data:  batch[:batchSize],
					}
				} else {
					if resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
						err = fmt.Errorf("Client %d fail to execute: %s, ErrMsg: %s, ErrCode: %v", i, stmt, resp.GetErrorMsg(), resp.GetErrorCode())
						m.errCh <- base.ErrData{
							Error: err,
							Data:  batch[:batchSize],
						}
					} else {
						m.statsCh <- stats.Stats{
							Latency:   uint64(resp.GetLatencyInUs()),
							ReqTime:   time.Since(now).Seconds(),
							BatchSize: batchSize,
						}
					}
				}

				batchSize = 0

				if data.Type == base.DONE {
					break
				}
			}
			// Send nil error to notify error handler of finishing to handle this file
			m.errCh <- base.ErrData{Error: nil}
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
		logger.Log.Fatalf("Invalid data type: %s", batch[length-1].Type)
		return ""
	}
}

func (m *NebulaClientMgr) makeEdgeBatchStmt(batch []base.Data) string {
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

func (m *NebulaClientMgr) makeVertexStmtWithoutHeaderLine(batch []base.Data) string {
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

func (m *NebulaClientMgr) makeVertexInsertStmtWithoutHeaderLine(data []base.Data) string {
	propTypeMap := make(map[int]string)
	var numProps int = 0
	var builder strings.Builder
	builder.WriteString("INSERT VERTEX ")
	for i, tag := range m.file.Schema.Vertex.Tags {
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
		if i < len(m.file.Schema.Vertex.Tags)-1 {
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

func (m *NebulaClientMgr) makeVertexDeleteStmtWithoutHeaderLine(data []base.Data) string {
	var builder strings.Builder
	for _, d := range data {
		// TODO: delete vertex in batch
		builder.WriteString(fmt.Sprintf("DELETE VERTEX %s;", d.Record[0]))
	}
	return builder.String()
}

func (m *NebulaClientMgr) makeEdgeStmtWithoutHeaderLine(batch []base.Data) string {
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

func (m *NebulaClientMgr) makeEdgeInsertStmtWithoutHeaderLine(batch []base.Data) string {
	numProps := 0
	propTypeMap := make(map[int]string)
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("INSERT EDGE %s(", m.file.Schema.Edge.Name))
	for i, prop := range m.file.Schema.Edge.Props {
		builder.WriteString(prop.Name)
		propTypeMap[numProps] = prop.Type
		numProps++
		if i < len(m.file.Schema.Edge.Props)-1 {
			builder.WriteString(",")
		}
	}
	builder.WriteString(") VALUES ")
	batchSize := len(batch)
	for i := 0; i < batchSize; i++ {
		builder.WriteString(fmt.Sprintf("%s->%s", batch[i].Record[0], batch[i].Record[1]))
		if m.file.Schema.Edge.WithRanking {
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
			builder.WriteString("\"")
			builder.WriteString(strings.Replace(record[i], "\"", "\\\"", -1))
			builder.WriteString("\"")
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

func (m *NebulaClientMgr) fillEdgePropsValues(builder *strings.Builder, record base.Record, isEnd bool, propTypeMap map[int]string) {
	fromIdx := 2
	if m.file.Schema.Edge.WithRanking {
		fromIdx = 3
	}
	if fromIdx > len(record) {
		logger.Log.Fatalf("Invalid record for edge: %v", record)
	}
	builder.WriteString("(")
	for i := fromIdx; i < len(record); i++ {
		if strings.ToLower(propTypeMap[i-fromIdx]) == "string" {
			builder.WriteString("\"")
			builder.WriteString(strings.Replace(record[i], "\"", "\\\"", -1))
			builder.WriteString("\"")
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
