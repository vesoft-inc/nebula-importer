package reader

import (
	"errors"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type BatchMgr struct {
	Schema             *config.Schema
	Batches            []*Batch
	InsertStmtPrefix   string
	initializedSchema  bool
	emptyPropsTagNames []string
	runnerLogger       *logger.RunnerLogger
}

func NewBatchMgr(schema *config.Schema, batchSize int, clientRequestChs []chan base.ClientRequest, errCh chan<- base.ErrData) *BatchMgr {
	bm := BatchMgr{
		Schema:             &config.Schema{},
		Batches:            make([]*Batch, len(clientRequestChs)),
		initializedSchema:  false,
		emptyPropsTagNames: schema.CollectEmptyPropsTagNames(),
	}

	bm.Schema.Type = schema.Type

	if bm.Schema.IsVertex() {
		index := 0
		bm.Schema.Vertex = &config.Vertex{
			VID:  &config.VID{Index: &index},
			Tags: []*config.Tag{},
		}
	} else {
		srcIdx, dstIdx := 0, 1
		bm.Schema.Edge = &config.Edge{
			Name:   schema.Edge.Name,
			SrcVID: &config.VID{Index: &srcIdx},
			DstVID: &config.VID{Index: &dstIdx},
			Rank:   nil,
			Props:  []*config.Prop{},
		}
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

func (bm *BatchMgr) InitSchema(header base.Record, runnerLogger *logger.RunnerLogger) (err error) {
	err = nil
	if bm.initializedSchema {
		logger.Log.Info("Batch manager schema has been initialized!")
		return
	}
	bm.initializedSchema = true
	for i, hh := range header {
		for _, h := range strings.Split(hh, "/") {
			switch c := strings.ToUpper(h); {
			case c == base.LABEL_LABEL:
				err = fmt.Errorf("Invalid schema: %v", header)
			case strings.HasPrefix(c, base.LABEL_VID):
				*bm.Schema.Vertex.VID.Index = i
				err = bm.Schema.Vertex.VID.ParseFunction(c)
			case strings.HasPrefix(c, base.LABEL_SRC_VID):
				*bm.Schema.Edge.SrcVID.Index = i
				err = bm.Schema.Edge.SrcVID.ParseFunction(c)
			case strings.HasPrefix(c, base.LABEL_DST_VID):
				*bm.Schema.Edge.DstVID.Index = i
				err = bm.Schema.Edge.DstVID.ParseFunction(c)
			case c == base.LABEL_RANK:
				if bm.Schema.Edge.Rank == nil {
					rank := i
					bm.Schema.Edge.Rank = &config.Rank{Index: &rank}
				} else {
					*bm.Schema.Edge.Rank.Index = i
				}
			case c == base.LABEL_IGNORE:
			default:
				if bm.Schema.IsVertex() {
					bm.addVertexTags(h, i)
				} else {
					bm.addEdgeProps(h, i)
				}
			}
		}
	}

	for _, tagName := range bm.emptyPropsTagNames {
		bm.getOrCreateVertexTagByName(tagName)
	}

	bm.generateInsertStmtPrefix()
	return
}

func (bm *BatchMgr) addVertexTags(r string, i int) {
	columnName, columnType := bm.parseProperty(r)
	tagName, prop := bm.parseTag(columnName)
	if tagName == "" {
		return
	}
	tag := bm.getOrCreateVertexTagByName(tagName)
	p := config.Prop{
		Name:  &prop,
		Type:  &columnType,
		Index: &i,
	}
	tag.Props = append(tag.Props, &p)
}

func (bm *BatchMgr) addEdgeProps(r string, i int) {
	columnName, columnType := bm.parseProperty(r)
	res := strings.SplitN(columnName, ".", 2)
	prop := res[0]
	if len(res) > 1 {
		prop = res[1]
	}
	p := config.Prop{
		Name:  &prop,
		Type:  &columnType,
		Index: &i,
	}
	bm.Schema.Edge.Props = append(bm.Schema.Edge.Props, &p)
}

func (bm *BatchMgr) generateInsertStmtPrefix() {
	var builder strings.Builder
	if bm.Schema.IsVertex() {
		builder.WriteString("INSERT VERTEX ")
		for i, tag := range bm.Schema.Vertex.Tags {
			builder.WriteString(fmt.Sprintf("`%s`(%s)", *tag.Name, bm.GeneratePropsString(tag.Props)))
			if i < len(bm.Schema.Vertex.Tags)-1 {
				builder.WriteString(",")
			}
		}
		builder.WriteString(" VALUES ")
	} else {
		edge := bm.Schema.Edge
		builder.WriteString(fmt.Sprintf("INSERT EDGE `%s`(%s) VALUES ", *edge.Name, bm.GeneratePropsString(edge.Props)))
	}
	bm.InsertStmtPrefix = builder.String()
}

func (bm *BatchMgr) GeneratePropsString(props []*config.Prop) string {
	var builder strings.Builder
	for i, prop := range props {
		builder.WriteString("`")
		builder.WriteString(*prop.Name)
		builder.WriteString("`")
		if i < len(props)-1 {
			builder.WriteString(",")
		}
	}
	return builder.String()
}

func (bm *BatchMgr) getOrCreateVertexTagByName(name string) *config.Tag {
	for i := range bm.Schema.Vertex.Tags {
		if strings.EqualFold(*bm.Schema.Vertex.Tags[i].Name, name) {
			return bm.Schema.Vertex.Tags[i]
		}
	}
	newTag := config.Tag{
		Name: &name,
	}
	idx := len(bm.Schema.Vertex.Tags)
	bm.Schema.Vertex.Tags = append(bm.Schema.Vertex.Tags, &newTag)
	return bm.Schema.Vertex.Tags[idx]
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

func (bm *BatchMgr) Add(data base.Data, runnerLogger *logger.RunnerLogger) error {
	var vid string
	if bm.Schema.IsVertex() {
		vid = data.Record[*bm.Schema.Vertex.VID.Index]
	} else {
		vid = data.Record[*bm.Schema.Edge.SrcVID.Index]
	}
	batchIdx := getBatchId(vid, len(bm.Batches), runnerLogger)
	bm.Batches[batchIdx].Add(data)
	return nil
}

var h = fnv.New32a()

func getBatchId(idStr string, numChans int, runnerLogger *logger.RunnerLogger) uint32 {
	_, err := h.Write([]byte(idStr))
	if err != nil {
		logger.Log.Error(err)
	}
	return h.Sum32() % uint32(numChans)
}

func makeStmt(batch []base.Data, f func([]base.Data) (string, error)) (string, error) {
	if len(batch) == 0 {
		return "", errors.New("Make stmt for empty batch")
	}

	if len(batch) == 1 {
		return f(batch)
	}

	var builder strings.Builder
	lastIdx, length := 0, len(batch)
	for i := 1; i < length; i++ {
		if batch[i-1].Type != batch[i].Type {
			str, err := f(batch[lastIdx:i])
			if err != nil {
				return "", err
			}
			builder.WriteString(str)
			lastIdx = i
		}
	}
	str, err := f(batch[lastIdx:])
	if err != nil {
		return "", err
	}
	builder.WriteString(str)
	return builder.String(), nil
}

func (m *BatchMgr) MakeVertexStmt(batch []base.Data) (string, error) {
	return makeStmt(batch, m.makeVertexBatchStmt)
}

func (m *BatchMgr) makeVertexBatchStmt(batch []base.Data) (string, error) {
	length := len(batch)
	switch batch[length-1].Type {
	case base.INSERT:
		return m.makeVertexInsertStmt(batch)
	case base.DELETE:
		return m.makeVertexDeleteStmt(batch)
	default:
		return "", fmt.Errorf("Invalid data type: %s", batch[length-1].Type)
	}
}

func (m *BatchMgr) makeVertexInsertStmt(data []base.Data) (string, error) {
	var builder strings.Builder
	builder.WriteString(m.InsertStmtPrefix)
	batchSize := len(data)
	for i := 0; i < batchSize; i++ {
		str, err := m.Schema.Vertex.FormatValues(data[i].Record)
		if err != nil {
			return "", err
		}
		builder.WriteString(str)
		if i < batchSize-1 {
			builder.WriteString(",")
		} else {
			builder.WriteString(";")
		}
	}

	return builder.String(), nil
}

func (m *BatchMgr) makeVertexDeleteStmt(data []base.Data) (string, error) {
	var idList []string
	for _, d := range data {
		vid, err := m.Schema.Vertex.VID.FormatValue(d.Record)
		if err != nil {
			return "", err
		}
		idList = append(idList, vid)
	}
	return fmt.Sprintf("DELETE VERTEX %s;", strings.Join(idList, ",")), nil
}

func (m *BatchMgr) MakeEdgeStmt(batch []base.Data) (string, error) {
	return makeStmt(batch, m.makeEdgeBatchStmt)
}

func (m *BatchMgr) makeEdgeBatchStmt(batch []base.Data) (string, error) {
	length := len(batch)
	switch batch[length-1].Type {
	case base.INSERT:
		return m.makeEdgeInsertStmt(batch)
	case base.DELETE:
		return m.makeEdgeDeleteStmt(batch)
	default:
		return "", fmt.Errorf("Invalid data type: %s", batch[length-1].Type)
	}
}

func (m *BatchMgr) makeEdgeInsertStmt(batch []base.Data) (string, error) {
	var builder strings.Builder
	builder.WriteString(m.InsertStmtPrefix)
	batchSize := len(batch)
	for i := 0; i < batchSize; i++ {
		str, err := m.Schema.Edge.FormatValues(batch[i].Record)
		if err != nil {
			return "", err
		}
		builder.WriteString(str)
		if i < batchSize-1 {
			builder.WriteString(",")
		} else {
			builder.WriteString(";")
		}
	}
	return builder.String(), nil
}

func (m *BatchMgr) makeEdgeDeleteStmt(batch []base.Data) (string, error) {
	var idList []string
	for _, d := range batch {
		var id string
		srcVid, err := m.Schema.Edge.SrcVID.FormatValue(d.Record)
		if err != nil {
			return "", err
		}
		dstVid, err := m.Schema.Edge.DstVID.FormatValue(d.Record)
		if err != nil {
			return "", err
		}
		if m.Schema.Edge.Rank != nil {
			rank := d.Record[*m.Schema.Edge.Rank.Index]
			id = fmt.Sprintf("%s->%s@%s", srcVid, dstVid, rank)
		} else {
			id = fmt.Sprintf("%s->%s", srcVid, dstVid)
		}
		idList = append(idList, id)
	}
	return fmt.Sprintf("DELETE EDGE %s %s;", *m.Schema.Edge.Name, strings.Join(idList, ",")), nil
}
