package specv3

import (
	"fmt"
	"sort"
	"strings"

	nebula "github.com/vesoft-inc/nebula-go/v3"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/bytebufferpool"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/manager"
	specbase "github.com/vesoft-inc/nebula-importer/v4/pkg/spec/base"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/utils"
)

type (
	// Node is VERTEX in 3.x
	Node struct {
		Name  string  `yaml:"name"`
		ID    *NodeID `yaml:"id"`
		Props Props   `yaml:"props,omitempty"`

		IgnoreExistedIndex *bool `yaml:"ignoreExistedIndex,omitempty"`

		Filter *specbase.Filter `yaml:"filter,omitempty"`

		Mode         specbase.Mode          `yaml:"mode,omitempty"`
		DynamicParam *specbase.DynamicParam `yaml:"dynamicParam,omitempty"`

		fnStatement        func(records ...Record) (string, int, error)
		dynamicFnStatement func(pool *nebula.SessionPool, records ...Record) (string, int, error)
		// "INSERT VERTEX name(prop_name, ..., prop_name) VALUES "
		// "UPDATE VERTEX ON name "
		// "DELETE TAG name FROM "
		statementPrefix string
		// session for batch update
	}

	Nodes []*Node

	NodeOption func(*Node)
)

func NewNode(name string, opts ...NodeOption) *Node {
	n := &Node{
		Name: name,
	}
	n.Options(opts...)

	return n
}

func WithNodeID(id *NodeID) NodeOption {
	return func(n *Node) {
		n.ID = id
	}
}

func WithNodeProps(props ...*Prop) NodeOption {
	return func(n *Node) {
		n.Props = append(n.Props, props...)
	}
}

func WithNodeIgnoreExistedIndex(ignore bool) NodeOption {
	return func(n *Node) {
		n.IgnoreExistedIndex = &ignore
	}
}

func WithNodeFilter(f *specbase.Filter) NodeOption {
	return func(n *Node) {
		n.Filter = f
	}
}

func WithNodeMode(m specbase.Mode) NodeOption {
	return func(n *Node) {
		n.Mode = m
	}
}

func (n *Node) Options(opts ...NodeOption) *Node {
	for _, opt := range opts {
		opt(n)
	}
	return n
}

//nolint:dupl
func (n *Node) Complete() {
	if n.ID != nil {
		n.ID.Complete()
		n.ID.Name = strVID
	}
	n.Props.Complete()
	n.Mode = n.Mode.Convert()

	switch n.Mode {
	case specbase.InsertMode:
		n.fnStatement = n.insertStatement
		// default enable IGNORE_EXISTED_INDEX
		insertPrefixFmt := "INSERT VERTEX IGNORE_EXISTED_INDEX %s(%s) VALUES "
		if n.IgnoreExistedIndex != nil && !*n.IgnoreExistedIndex {
			insertPrefixFmt = "INSERT VERTEX %s(%s) VALUES "
		}
		n.statementPrefix = fmt.Sprintf(
			insertPrefixFmt,
			utils.ConvertIdentifier(n.Name),
			strings.Join(n.Props.NameList(), ", "),
		)
	case specbase.UpdateMode:
		n.fnStatement = n.updateStatement
		n.statementPrefix = fmt.Sprintf("UPDATE VERTEX ON %s ", utils.ConvertIdentifier(n.Name))
	case specbase.DeleteMode:
		n.fnStatement = n.deleteStatement
		n.statementPrefix = fmt.Sprintf("DELETE TAG %s FROM ", utils.ConvertIdentifier(n.Name))
	case specbase.BatchUpdateMode:
		//batch update, would fetch the node first.
		//and then update the node with the props
		//statementPrefix should be modified after fetch the node
		n.fnStatement = n.updateBatchStatement
	}
}

func (n *Node) Validate() error {
	if n.Name == "" {
		return n.importError(errors.ErrNoNodeName)
	}

	if n.ID == nil {
		return n.importError(errors.ErrNoNodeID)
	}

	if err := n.ID.Validate(); err != nil {
		return n.importError(err)
	}

	if err := n.Props.Validate(); err != nil {
		return n.importError(err)
	}

	if n.Filter != nil {
		if err := n.Filter.Build(); err != nil {
			return n.importError(errors.ErrFilterSyntax, "%s", err)
		}
	}

	if !n.Mode.IsSupport() {
		return n.importError(errors.ErrUnsupportedMode)
	}

	if n.Mode == specbase.UpdateMode && len(n.Props) == 0 {
		return n.importError(errors.ErrNoProps)
	}

	return nil
}

func (n *Node) Statement(records ...Record) (statement string, nRecord int, err error) {
	return n.fnStatement(records...)
}

func (n *Node) insertStatement(records ...Record) (statement string, nRecord int, err error) {
	buff := bytebufferpool.Get()
	defer bytebufferpool.Put(buff)

	buff.SetString(n.statementPrefix)

	for _, record := range records {
		if n.Filter != nil {
			ok, err := n.Filter.Filter(record)
			if err != nil {
				return "", 0, n.importError(err)
			}
			if !ok { // skipping those return false by Filter
				continue
			}
		}
		idValue, err := n.ID.Value(record)
		if err != nil {
			return "", 0, n.importError(err)
		}
		propsValueList, err := n.Props.ValueList(record)
		if err != nil {
			return "", 0, n.importError(err)
		}

		if nRecord > 0 {
			_, _ = buff.WriteString(", ")
		}

		// id:(prop_value1, prop_value2, ...)
		_, _ = buff.WriteString(idValue)
		_, _ = buff.WriteString(":(")
		_, _ = buff.WriteStringSlice(propsValueList, ", ")
		_, _ = buff.WriteString(")")

		nRecord++
	}

	if nRecord == 0 {
		return "", 0, nil
	}

	return buff.String(), nRecord, nil
}

func (n *Node) updateStatement(records ...Record) (statement string, nRecord int, err error) {
	buff := bytebufferpool.Get()
	defer bytebufferpool.Put(buff)

	for _, record := range records {
		if n.Filter != nil {
			ok, err := n.Filter.Filter(record)
			if err != nil {
				return "", 0, n.importError(err)
			}
			if !ok { // skipping those return false by Filter
				continue
			}
		}
		idValue, err := n.ID.Value(record)
		if err != nil {
			return "", 0, n.importError(err)
		}
		propsSetValueList, err := n.Props.SetValueList(record)
		if err != nil {
			return "", 0, n.importError(err)
		}

		// "UPDATE VERTEX ON name id SET prop_name1 = prop_value1, prop_name1 = prop_value1, ...;"
		_, _ = buff.WriteString(n.statementPrefix)
		_, _ = buff.WriteString(idValue)
		_, _ = buff.WriteString(" SET ")
		_, _ = buff.WriteStringSlice(propsSetValueList, ", ")
		_, _ = buff.WriteString(";")

		nRecord++
	}

	return buff.String(), nRecord, nil
}

func (n *Node) deleteStatement(records ...Record) (statement string, nRecord int, err error) {
	buff := bytebufferpool.Get()
	defer bytebufferpool.Put(buff)

	for _, record := range records {
		if n.Filter != nil {
			ok, err := n.Filter.Filter(record)
			if err != nil {
				return "", 0, n.importError(err)
			}
			if !ok { // skipping those return false by Filter
				continue
			}
		}
		idValue, err := n.ID.Value(record)
		if err != nil {
			return "", 0, n.importError(err)
		}

		// "DELETE TAG name FROM id;"
		_, _ = buff.WriteString(n.statementPrefix)
		_, _ = buff.WriteString(idValue)
		_, _ = buff.WriteString(";")

		nRecord++
	}

	return buff.String(), nRecord, nil
}

func (n *Node) importError(err error, formatWithArgs ...any) *errors.ImportError {
	return errors.AsOrNewImportError(err, formatWithArgs...).SetNodeName(n.Name)
}

func (ns Nodes) Complete() {
	for i := range ns {
		ns[i].Complete()
	}
}

func (ns Nodes) Validate() error {
	for i := range ns {
		if err := ns[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) updateBatchStatement(records ...Record) (statement string, nRecord int, err error) {
	if n.DynamicParam == nil {
		return "", 0, errors.ErrNoDynamicParam
	}
	buff := bytebufferpool.Get()
	defer bytebufferpool.Put(buff)
	var (
		idValues          []string
		cols              []string
		needUpdateRecords []Record
	)

	for _, record := range records {
		idValue, err := n.ID.Value(record)
		if err != nil {
			return "", 0, n.importError(err)
		}
		idValues = append(idValues, idValue)
		propsSetValueList, err := n.Props.ValueList(record)
		if err != nil {
			return "", 0, err
		}
		needUpdateRecords = append(needUpdateRecords, propsSetValueList)
	}
	for _, prop := range n.Props {
		cols = append(cols, prop.Name)
	}

	updatedCols, updatedRecords, err := n.genDynamicUpdateRecord(manager.DefaultSessionPool, idValues, cols, needUpdateRecords)
	if err != nil {
		return "", 0, err
	}

	// batch insert
	// INSERT VERTEX %s(%s) VALUES
	prefix := fmt.Sprintf("INSERT VERTEX %s(%s) VALUES ", utils.ConvertIdentifier(n.Name), strings.Join(updatedCols, ", "))
	buff.SetString(prefix)

	for index, record := range updatedRecords {
		idValue := idValues[index]

		if nRecord > 0 {
			_, _ = buff.WriteString(", ")
		}

		// id:(prop_value1, prop_value2, ...)
		_, _ = buff.WriteString(idValue)
		_, _ = buff.WriteString(":(")
		_, _ = buff.WriteStringSlice(record, ", ")
		_, _ = buff.WriteString(")")

		nRecord++
	}
	return buff.String(), nRecord, nil
}

// genDynamicUpdateRecord generate the update record for batch update
// return column values and records
func (n *Node) genDynamicUpdateRecord(pool *nebula.SessionPool, idValues []string, cols []string, records []Record) ([]string, []Record, error) {
	stat := fmt.Sprintf("FETCH PROP ON %s %s YIELD VERTEX as v;", utils.ConvertIdentifier(n.Name), strings.Join(idValues, ","))
	var (
		rs             *nebula.ResultSet
		err            error
		updatedCols    []string
		updatedRecords []Record
	)
	for i := 0; i < 3; i++ {
		rs, err = pool.Execute(stat)
		if err != nil {
			continue
		}
		if !rs.IsSucceed() {
			continue
		}
	}
	if err != nil {
		return nil, nil, err
	}
	if !rs.IsSucceed() {
		return nil, nil, fmt.Errorf(rs.GetErrorMsg())
	}
	fetchData, err := n.getNebulaFetchData(rs)
	for _, property := range fetchData {
		updatedCols = n.getDynamicUpdateCols(cols, property)
		break
	}
	for index, id := range idValues {
		originalData, ok := fetchData[id]
		if !ok {
			return nil, nil, fmt.Errorf("cannot find id, id: %s", id)
		}
		r := n.getUpdateRocord(originalData, updatedCols, records[index])
		updatedRecords = append(updatedRecords, r)
	}
	return updatedCols, updatedRecords, nil
}

// append the need update column to the end of the cols
func (n *Node) getDynamicUpdateCols(updateCols []string, properties map[string]*nebula.ValueWrapper) []string {
	needUpdate := make(map[string]struct{})
	for _, c := range updateCols {
		needUpdate[c] = struct{}{}
	}
	var cols []string
	for k, _ := range properties {
		if _, ok := needUpdate[k]; !ok {
			cols = append(cols, k)
		}
	}
	sort.Slice(cols, func(i, j int) bool {
		return cols[i] < cols[j]
	})
	cols = append(cols, updateCols...)
	return cols
}

func (n *Node) getNebulaFetchData(rs *nebula.ResultSet) (map[string]map[string]*nebula.ValueWrapper, error) {
	m := make(map[string]map[string]*nebula.ValueWrapper)
	for i := 0; i < rs.GetRowSize(); i++ {
		row, err := rs.GetRowValuesByIndex(i)
		if err != nil {
			return nil, err
		}
		cell, err := row.GetValueByIndex(0)
		if err != nil {
			return nil, err
		}
		node, err := cell.AsNode()
		
		if err != nil {
			return nil, err
		}
		property, err := node.Properties(n.Name)
		if err != nil {
			return nil, err
		}
		m[node.GetID().String()] = property
	}
	return m, nil
}

func (n *Node) getUpdateRocord(original map[string]*nebula.ValueWrapper, Columns []string, update Record) Record {
	r := make(Record, 0, len(Columns))
	var vStr string
	for _, c := range Columns {
		value := original[c]

		switch value.GetType() {
		// TODO should handle other type
		case "datetime":
			vStr = fmt.Sprintf("datetime(\"%s\")", value.String())
		default:
			vStr = value.String()
		}
		r = append(r, vStr)
	}
	// update
	for i := 0; i < len(update); i++ {
		r[len(Columns)-len(update)+i] = update[i]
	}
	return r
}
