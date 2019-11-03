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
	errCh   chan<- base.ErrData
	statsCh chan<- stats.Stats
	doneCh  <-chan bool
	pool    *ClientPool
}

func NewNebulaClientMgr(settings config.NebulaClientSettings, errCh chan<- base.ErrData, statsCh chan<- stats.Stats, doneCh <-chan bool) *NebulaClientMgr {
	mgr := NebulaClientMgr{
		config:  settings,
		errCh:   errCh,
		statsCh: statsCh,
		doneCh:  doneCh,
	}

	mgr.pool = NewClientPool(settings)
	mgr.startWorkers()

	log.Printf("Create %d Nebula Graph clients", mgr.config.Concurrency)

	return &mgr
}

func (m *NebulaClientMgr) Close() {
	m.pool.Close()
}

func (m *NebulaClientMgr) GetDataChans() []chan base.Record {
	return m.pool.DataChs
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

func (m *NebulaClientMgr) startWorkers() {
	for i := 0; i < m.config.Concurrency; i++ {
		go func(i int) {
			batchSize, numBatches := 0, 0
			batch := make([]base.Record, m.file.BatchSize)
			for {
				data, ok := <-m.pool.DataChs[i]
				if !ok {
					break
				}

				switch strings.ToUpper(data[0]) {
				case "DONE":
					batch[batchSize] = data
					batchSize++

					if batchSize < m.file.BatchSize {
						continue
					}
				case "_SRC":
				case "_LABEL":
				default:
					// Need not to notify error handler. Reset it in outside main program
					if batchSize == 0 {
						continue
					}
				}

				var stmt string
				switch strings.ToUpper(m.file.Schema.Type) {
				case "VERTEX":
					stmt = m.makeVertexInsertStmtWithoutHeaderLine(batch)
				case "EDGE":
					stmt = m.makeEdgeInsertStmtWithoutHeaderLine(batch)
				default:
					log.Fatalf("Error schema type: %s", m.file.Schema.Type)
				}

				// TODO: Add some metrics for response latency, succeededCount, failedCount
				now := time.Now()
				resp, err := m.pool.Conns[i].Execute(stmt)
				reqTime := time.Since(now).Seconds()

				if err != nil {
					m.errCh <- base.ErrData{
						Error: err,
						Data:  data,
						Done:  false,
					}
					continue
				}

				if resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
					errMsg := fmt.Sprintf("Fail to execute: %s, ErrMsg: %s, ErrCode: %v", stmt, resp.GetErrorMsg(), resp.GetErrorCode())
					m.errCh <- base.ErrData{
						Error: errors.New(errMsg),
						Data:  data,
						Done:  false,
					}
					continue
				}

				m.statsCh <- stats.Stats{
					Latency: uint64(resp.GetLatencyInUs()),
					ReqTime: reqTime,
				}

				numBatches++
				batchSize = 0
			}
		}(i)
	}
}

func (m *NebulaClientMgr) convertRecords(records [][]string) [][]interface{} {
	if m.file.BatchSize != len(records) {
		log.Fatalf("records length is not equal to batch size: %d != %d", len(records), m.file.BatchSize)
	}
	data := make([][]interface{}, len(records))
	for i := range records {
		data[i] = make([]interface{}, len(records[i]))
		for j := range records[i] {
			data[i][j] = records[i][j]
		}
	}
	return data
}

func (m *NebulaClientMgr) makeVertexInsertStmtWithoutHeaderLine(records []base.Record) string {
	var builder strings.Builder
	builder.WriteString("INSERT VERTEX ")
	numProps := 0
	for i, tag := range m.file.Schema.Vertex.Tags {
		builder.WriteString(fmt.Sprintf("%s(", tag.Name))
		for j, prop := range tag.Props {
			builder.WriteString(prop.Name)
			if j < len(tag.Props)-1 {
				builder.WriteString(",")
			} else {
				builder.WriteString(")")
			}
			numProps++
		}
		if i < len(m.file.Schema.Vertex.Tags)-1 {
			builder.WriteString(",")
		} else {
			builder.WriteString(" VALUES ")
		}
	}
	batchSize := len(records)
	for i := 0; i < batchSize; i++ {
		builder.WriteString(fmt.Sprintf(" %s: ", records[i][0]))
		fillVertexPropsValues(&builder, records[i], i == batchSize-1)
	}

	return builder.String()
}

func (m *NebulaClientMgr) makeEdgeInsertStmtWithoutHeaderLine(batch []base.Record) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("INSERT EDGE %s(", m.file.Schema.Edge.Name))
	numProps := 0
	for i, prop := range m.file.Schema.Edge.Props {
		builder.WriteString(prop.Name)
		if i < len(m.file.Schema.Edge.Props)-1 {
			builder.WriteString(",")
		} else {
			builder.WriteString(")")
		}
		numProps++
	}
	builder.WriteString(" VALUES ")
	batchSize := len(batch)
	for i := 0; i < batchSize; i++ {
		builder.WriteString("?->?: ")
		if m.file.Schema.Edge.WithRanking {
			builder.WriteString("(?)")
		}
		fillVertexPropsValues(&builder, batch[i], i == batchSize-1)
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
