package reader

import (
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type BatchMgr struct {
	Schema            config.Schema
	Batches           []*Batch
	InsertStmtPrefix  string
	initializedSchema bool
}

func NewBatchMgr(schema config.Schema, batchSize int, clientRequestChs []chan base.ClientRequest, errCh chan<- base.ErrData) *BatchMgr {
	bm := BatchMgr{
		Schema:            schema,
		Batches:           make([]*Batch, len(clientRequestChs)),
		initializedSchema: false,
	}

	if schema.IsVertex() {
		bm.Schema.Vertex.Tags = nil
	} else {
		bm.Schema.Edge.Props = nil
	}

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
	if bm.initializedSchema {
		logger.Log.Println("Batch manager schema has been initialized!")
		return
	}
	bm.initializedSchema = true
	for i, h := range header {
		switch strings.ToUpper(h) {
		case base.LABEL_LABEL:
		case base.LABEL_VID:
			bm.Schema.Vertex.VID.Index = i
		case base.LABEL_SRC_VID:
			bm.Schema.Edge.SrcVID.Index = i
		case base.LABEL_DST_VID:
			bm.Schema.Edge.DstVID.Index = i
		case base.LABEL_RANK:
			bm.Schema.Edge.Rank.Index = i
		case base.LABEL_IGNORE:
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

var re = regexp.MustCompile(`^([+-]?\d+|hash\("(.+)"\)|uuid\("(.+)"\))$`)

func (bm *BatchMgr) Add(data base.Data) error {
	var vid string
	if bm.Schema.IsVertex() {
		vid = data.Record[bm.Schema.Vertex.VID.Index]
	} else {
		vid = data.Record[bm.Schema.Edge.SrcVID.Index]
	}
	if !re.MatchString(vid) {
		return fmt.Errorf("Invalid vid format: %s", vid)
	}
	batchIdx := getBatchId(vid, len(bm.Batches))
	bm.Batches[batchIdx].Add(data)
	return nil
}

var h = fnv.New32a()

func getBatchId(idStr string, numChans int) uint32 {
	h.Write([]byte(idStr))
	return h.Sum32() % uint32(numChans)
}

func makeStmt(batch []base.Data, f func([]base.Data) string) string {
	if len(batch) == 0 {
		logger.Log.Fatal("Make stmt for empty batch")
	}

	if len(batch) == 1 {
		return f(batch)
	}

	var builder strings.Builder
	lastIdx, length := 0, len(batch)
	for i := 1; i < length; i++ {
		if batch[i-1].Type != batch[i].Type {
			builder.WriteString(f(batch[lastIdx:i]))
			lastIdx = i
		}
	}
	builder.WriteString(f(batch[lastIdx:]))
	return builder.String()
}

func (m *BatchMgr) MakeVertexStmt(batch []base.Data) string {
	return makeStmt(batch, m.makeVertexBatchStmt)
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
		builder.WriteString(m.Schema.Vertex.FormatValues(data[i].Record))
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
		builder.WriteString(fmt.Sprintf("DELETE VERTEX %s;", d.Record[m.Schema.Vertex.VID.Index]))
	}
	return builder.String()
}

func (m *BatchMgr) MakeEdgeStmt(batch []base.Data) string {
	return makeStmt(batch, m.makeEdgeBatchStmt)
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
		builder.WriteString(m.Schema.Edge.FormatValues(batch[i].Record))
		if i < batchSize-1 {
			builder.WriteString(",")
		} else {
			builder.WriteString(";")
		}
	}
	return builder.String()
}
