package reader

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type BatchMgr struct {
	Schema           config.Schema
	Batches          []*Batch
	InsertStmtPrefix string
}

func NewBatchMgr(schema config.Schema, batchSize int, clientRequestChs []chan base.ClientRequest, errCh chan<- base.ErrData) *BatchMgr {
	bm := BatchMgr{
		Schema:  schema,
		Batches: make([]*Batch, len(clientRequestChs)),
	}

	if schema.IsVertex() {
		bm.Schema.Vertex.Tags = nil
	} else {
		bm.Schema.Edge.Props = nil
	}

	bm.InitSchema(strings.Split(schema.String(), ","))

	for i := range bm.Batches {
		bm.Batches[i] = NewBatch(&bm, batchSize, clientRequestChs[i], errCh)
	}
	return &bm
}

func (bm *BatchMgr) Done() {
	for i := range bm.Batches {
		bm.Batches[i].Done()
	}
}

func (bm *BatchMgr) InitSchema(header base.Record) {
	for i, h := range header {
		switch strings.ToUpper(h) {
		case ":VID":
		case ":SRC_VID":
		case ":DST_VID":
		case ":RANK":
		case ":IGNORE":
		default:
			if bm.Schema.IsVertex() {
				bm.addVertexTags(h, i)
			} else {
				bm.addEdgeProps(h, i)
			}
		}
	}

	bm.generateInsertStmtPrefix()
}

func (bm *BatchMgr) addVertexTags(r string, i int) {
	columnName, columnType := bm.parseProperty(r)
	tagName, prop := bm.parseTag(columnName)
	if tagName == "" {
		return
	}
	tag := bm.getOrCreateVertexTagByName(tagName)
	tag.Props = append(tag.Props, config.Prop{
		Name:   prop,
		Type:   columnType,
		Index:  i,
		Ignore: prop == "",
	})
}

func (bm *BatchMgr) addEdgeProps(r string, i int) {
	columnName, columnType := bm.parseProperty(r)
	res := strings.SplitN(columnName, ".", 2)
	prop := res[0]
	if len(res) > 1 {
		prop = res[1]
	}
	bm.Schema.Edge.Props = append(bm.Schema.Edge.Props, config.Prop{
		Name:   prop,
		Type:   columnType,
		Index:  i,
		Ignore: prop == "",
	})
}

func (bm *BatchMgr) generateInsertStmtPrefix() {
	var builder strings.Builder
	if bm.Schema.IsVertex() {
		builder.WriteString("INSERT VERTEX ")
		for i, tag := range bm.Schema.Vertex.Tags {
			builder.WriteString(fmt.Sprintf("%s(%s)", tag.Name, bm.GeneratePropsString(tag.Props)))
			if i < len(bm.Schema.Vertex.Tags)-1 {
				builder.WriteString(",")
			}
		}
		builder.WriteString(" VALUES ")
	} else {
		edge := &bm.Schema.Edge
		builder.WriteString(fmt.Sprintf("INSERT EDGE %s(%s) VALUES ", edge.Name, bm.GeneratePropsString(edge.Props)))
	}
	bm.InsertStmtPrefix = builder.String()
}

func (bm *BatchMgr) GeneratePropsString(props []config.Prop) string {
	var builder strings.Builder
	for i, prop := range props {
		if !prop.Ignore {
			builder.WriteString(prop.Name)
			if i < len(props)-1 {
				builder.WriteString(",")
			}
		}
	}
	return builder.String()
}

func (bm *BatchMgr) getOrCreateVertexTagByName(name string) *config.Tag {
	for i := range bm.Schema.Vertex.Tags {
		if strings.ToLower(bm.Schema.Vertex.Tags[i].Name) == strings.ToLower(name) {
			return &bm.Schema.Vertex.Tags[i]
		}
	}
	newTag := config.Tag{Name: name}
	idx := len(bm.Schema.Vertex.Tags)
	bm.Schema.Vertex.Tags = append(bm.Schema.Vertex.Tags, newTag)
	return &bm.Schema.Vertex.Tags[idx]
}

func (bm *BatchMgr) parseTag(s string) (tag, field string) {
	res := strings.SplitN(s, ".", 2)

	if len(res) < 2 {
		return "", ""
	}

	return res[0], res[1]
}

func (bm *BatchMgr) parseProperty(r string) (columnName, columnType string) {
	res := strings.SplitN(r, ":", 2)

	if len(res) == 1 || res[1] == "" || !base.IsValidType(res[1]) {
		return res[0], "string"
	} else {
		return res[0], res[1]
	}
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

func (m *BatchMgr) MakeVertexStmt(batch []base.Data) string {
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

func (m *BatchMgr) makeVertexBatchStmt(batch []base.Data) string {
	length := len(batch)
	switch batch[length-1].Type {
	case base.INSERT:
		return m.makeVertexInsertStmt(batch)
	case base.DELETE:
		return m.makeVertexDeleteStmt(batch)
	default:
		logger.Log.Fatalf("Invalid data type: %s", batch[length-1].Type)
		return ""
	}
}

func (m *BatchMgr) makeVertexInsertStmt(data []base.Data) string {
	var builder strings.Builder
	builder.WriteString(m.InsertStmtPrefix)
	batchSize := len(data)
	for i := 0; i < batchSize; i++ {
		builder.WriteString(fmt.Sprintf(" %s: (", data[i].Record[0]))
		for j, tag := range m.Schema.Vertex.Tags {
			for k, prop := range tag.Props {
				if prop.Ignore {
					continue
				}
				r := data[i].Record[prop.Index]
				if strings.ToLower(prop.Type) == "string" {
					builder.WriteString(fmt.Sprintf("%q", r))
				} else {
					builder.WriteString(r)
				}
				if !(j == len(m.Schema.Vertex.Tags)-1 && k == len(tag.Props)-1) {
					builder.WriteString(",")
				}
			}
		}
		builder.WriteString(")")
		if i < batchSize-1 {
			builder.WriteString(",")
		} else {
			builder.WriteString(";")
		}
	}

	return builder.String()
}

func (m *BatchMgr) makeVertexDeleteStmt(data []base.Data) string {
	var builder strings.Builder
	for _, d := range data {
		// TODO: delete vertex in batch
		builder.WriteString(fmt.Sprintf("DELETE VERTEX %s;", d.Record[0]))
	}
	return builder.String()
}

func (m *BatchMgr) MakeEdgeStmt(batch []base.Data) string {
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

func (m *BatchMgr) makeEdgeBatchStmt(batch []base.Data) string {
	length := len(batch)
	switch batch[length-1].Type {
	case base.INSERT:
		return m.makeEdgeInsertStmt(batch)
	case base.DELETE:
		logger.Log.Fatal("Unsupported delete edge")
	default:
		logger.Log.Fatalf("Invalid data type: %s", batch[length-1].Type)
	}
	return ""
}

func (m *BatchMgr) makeEdgeInsertStmt(batch []base.Data) string {
	var builder strings.Builder
	builder.WriteString(m.InsertStmtPrefix)
	batchSize := len(batch)
	for i := 0; i < batchSize; i++ {
		rank := ""
		if m.Schema.Edge.WithRanking {
			// TODO: Validate rank is integer
			rank = fmt.Sprintf("@%s", batch[i].Record[2])
		}
		builder.WriteString(fmt.Sprintf("%s->%s%s: (", batch[i].Record[0], batch[i].Record[1], rank))
		for j, prop := range m.Schema.Edge.Props {
			if prop.Ignore {
				continue
			}
			r := batch[i].Record[prop.Index]
			if strings.ToLower(prop.Type) == "string" {
				builder.WriteString(fmt.Sprintf("%q", r))
			} else {
				builder.WriteString(r)
			}
			if j < len(m.Schema.Edge.Props)-1 {
				builder.WriteString(",")
			}
		}
		builder.WriteString(")")
		if i < batchSize-1 {
			builder.WriteString(",")
		} else {
			builder.WriteString(";")
		}
	}
	return builder.String()
}
