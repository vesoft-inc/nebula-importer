package specv3

import (
	"fmt"
	"strings"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/bytebufferpool"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"
	specbase "github.com/vesoft-inc/nebula-importer/v4/pkg/spec/base"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/utils"
)

type (
	Edge struct {
		Name  string       `yaml:"name"`
		Src   *EdgeNodeRef `yaml:"src"`
		Dst   *EdgeNodeRef `yaml:"dst"`
		Rank  *Rank        `yaml:"rank,omitempty"`
		Props Props        `yaml:"props,omitempty"`

		IgnoreExistedIndex *bool `yaml:"ignoreExistedIndex,omitempty"`

		Filter *specbase.Filter `yaml:"filter,omitempty"`

		Mode specbase.Mode `yaml:"mode,omitempty"`

		fnStatement func(records ...Record) (string, int, error)
		// "INSERT EDGE name(prop_name, ..., prop_name) VALUES "
		// "UPDATE EDGE ON name "
		// "DELETE EDGE name "
		statementPrefix string
	}

	EdgeNodeRef struct {
		Name string  `yaml:"-"`
		ID   *NodeID `yaml:"id"`
	}

	Edges []*Edge

	EdgeOption func(*Edge)
)

func NewEdge(name string, opts ...EdgeOption) *Edge {
	e := &Edge{
		Name: name,
	}
	e.Options(opts...)

	return e
}

func WithEdgeSrc(src *EdgeNodeRef) EdgeOption {
	return func(e *Edge) {
		e.Src = src
	}
}

func WithEdgeDst(dst *EdgeNodeRef) EdgeOption {
	return func(e *Edge) {
		e.Dst = dst
	}
}

func WithRank(rank *Rank) EdgeOption {
	return func(e *Edge) {
		e.Rank = rank
	}
}

func WithEdgeProps(props ...*Prop) EdgeOption {
	return func(e *Edge) {
		e.Props = append(e.Props, props...)
	}
}

func WithEdgeIgnoreExistedIndex(ignore bool) EdgeOption {
	return func(e *Edge) {
		e.IgnoreExistedIndex = &ignore
	}
}

func WithEdgeFilter(f *specbase.Filter) EdgeOption {
	return func(e *Edge) {
		e.Filter = f
	}
}

func WithEdgeMode(m specbase.Mode) EdgeOption {
	return func(e *Edge) {
		e.Mode = m
	}
}

func (e *Edge) Options(opts ...EdgeOption) *Edge {
	for _, opt := range opts {
		opt(e)
	}
	return e
}

//nolint:dupl
func (e *Edge) Complete() {
	if e.Src != nil {
		e.Src.Complete()
		e.Src.Name = strSrc
		if e.Src.ID != nil {
			e.Src.ID.Name = strVID
		}
	}
	if e.Dst != nil {
		e.Dst.Complete()
		e.Dst.Name = strDst
		if e.Dst.ID != nil {
			e.Dst.ID.Name = strVID
		}
	}
	if e.Rank != nil {
		e.Rank.Complete()
	}
	e.Props.Complete()
	e.Mode = e.Mode.Convert()

	switch e.Mode {
	case specbase.InsertMode:
		e.fnStatement = e.insertStatement
		// default enable IGNORE_EXISTED_INDEX
		insertPrefixFmt := "INSERT EDGE IGNORE_EXISTED_INDEX %s(%s) VALUES "
		if e.IgnoreExistedIndex != nil && !*e.IgnoreExistedIndex {
			insertPrefixFmt = "INSERT EDGE %s(%s) VALUES "
		}

		e.statementPrefix = fmt.Sprintf(
			insertPrefixFmt,
			utils.ConvertIdentifier(e.Name),
			strings.Join(e.Props.NameList(), ", "),
		)
	case specbase.UpdateMode:
		e.fnStatement = e.updateStatement
		e.statementPrefix = fmt.Sprintf("UPDATE EDGE ON %s ", utils.ConvertIdentifier(e.Name))
	case specbase.DeleteMode:
		e.fnStatement = e.deleteStatement
		e.statementPrefix = fmt.Sprintf("DELETE EDGE %s ", utils.ConvertIdentifier(e.Name))
	}
}

func (e *Edge) Validate() error {
	if e.Name == "" {
		return e.importError(errors.ErrNoEdgeName)
	}

	if e.Src == nil {
		return e.importError(errors.ErrNoEdgeSrc)
	}

	if err := e.Src.Validate(); err != nil {
		return e.importError(err)
	}

	if e.Dst == nil {
		return e.importError(errors.ErrNoEdgeDst)
	}

	if err := e.Dst.Validate(); err != nil {
		return e.importError(err)
	}

	if e.Rank != nil {
		if err := e.Rank.Validate(); err != nil {
			return err
		}
	}

	if err := e.Props.Validate(); err != nil {
		return e.importError(err)
	}

	if e.Filter != nil {
		if err := e.Filter.Build(); err != nil {
			return e.importError(errors.ErrFilterSyntax, "%s", err)
		}
	}

	if !e.Mode.IsSupport() {
		return e.importError(errors.ErrUnsupportedMode)
	}

	if e.Mode == specbase.UpdateMode && len(e.Props) == 0 {
		return e.importError(errors.ErrNoProps)
	}

	return nil
}

func (e *Edge) Statement(records ...Record) (statement string, nRecord int, err error) {
	return e.fnStatement(records...)
}

func (e *Edge) insertStatement(records ...Record) (statement string, nRecord int, err error) {
	buff := bytebufferpool.Get()
	defer bytebufferpool.Put(buff)

	buff.SetString(e.statementPrefix)

	for _, record := range records {
		if e.Filter != nil {
			ok, err := e.Filter.Filter(record)
			if err != nil {
				return "", 0, e.importError(err)
			}
			if !ok { // skipping those return false by Filter
				continue
			}
		}
		srcIDValue, err := e.Src.IDValue(record)
		if err != nil {
			return "", 0, e.importError(err)
		}
		dstIDValue, err := e.Dst.IDValue(record)
		if err != nil {
			return "", 0, e.importError(err)
		}
		var rankValueStatement string
		if e.Rank != nil {
			var rankValue string
			rankValue, err = e.Rank.Value(record)
			if err != nil {
				return "", 0, e.importError(err)
			}
			rankValueStatement = "@" + rankValue
		}
		propsValueList, err := e.Props.ValueList(record)
		if err != nil {
			return "", 0, e.importError(err)
		}

		if nRecord > 0 {
			_, _ = buff.WriteString(", ")
		}

		// src -> dst@rank:(prop_value1, prop_value2, ...)
		_, _ = buff.WriteString(srcIDValue)
		_, _ = buff.WriteString("->")
		_, _ = buff.WriteString(dstIDValue)
		_, _ = buff.WriteString(rankValueStatement)
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

func (e *Edge) updateStatement(records ...Record) (statement string, nRecord int, err error) {
	buff := bytebufferpool.Get()
	defer bytebufferpool.Put(buff)

	for _, record := range records {
		if e.Filter != nil {
			ok, err := e.Filter.Filter(record)
			if err != nil {
				return "", 0, e.importError(err)
			}
			if !ok { // skipping those return false by Filter
				continue
			}
		}
		srcIDValue, err := e.Src.IDValue(record)
		if err != nil {
			return "", 0, e.importError(err)
		}
		dstIDValue, err := e.Dst.IDValue(record)
		if err != nil {
			return "", 0, e.importError(err)
		}
		var rankValueStatement string
		if e.Rank != nil {
			var rankValue string
			rankValue, err = e.Rank.Value(record)
			if err != nil {
				return "", 0, e.importError(err)
			}
			rankValueStatement = "@" + rankValue
		}
		propsSetValueList, err := e.Props.SetValueList(record)
		if err != nil {
			return "", 0, e.importError(err)
		}

		// "UPDATE EDGE ON name "src"->"dst"@rank SET prop_name1 = prop_value1, prop_name1 = prop_value1, ...;"
		_, _ = buff.WriteString(e.statementPrefix)
		_, _ = buff.WriteString(srcIDValue)
		_, _ = buff.WriteString("->")
		_, _ = buff.WriteString(dstIDValue)
		_, _ = buff.WriteString(rankValueStatement)
		_, _ = buff.WriteString(" SET ")
		_, _ = buff.WriteStringSlice(propsSetValueList, ", ")
		_, _ = buff.WriteString(";")

		nRecord++
	}

	return buff.String(), nRecord, nil
}

func (e *Edge) deleteStatement(records ...Record) (statement string, nRecord int, err error) {
	buff := bytebufferpool.Get()
	defer bytebufferpool.Put(buff)

	buff.SetString(e.statementPrefix)

	for _, record := range records {
		if e.Filter != nil {
			ok, err := e.Filter.Filter(record)
			if err != nil {
				return "", 0, e.importError(err)
			}
			if !ok { // skipping those return false by Filter
				continue
			}
		}
		srcIDValue, err := e.Src.IDValue(record)
		if err != nil {
			return "", 0, e.importError(err)
		}
		dstIDValue, err := e.Dst.IDValue(record)
		if err != nil {
			return "", 0, e.importError(err)
		}
		var rankValueStatement string
		if e.Rank != nil {
			rankValue, err := e.Rank.Value(record)
			if err != nil {
				return "", 0, e.importError(err)
			}
			rankValueStatement = "@" + rankValue
		}

		if nRecord > 0 {
			_, _ = buff.WriteString(", ")
		}

		// src -> dst@rank
		_, _ = buff.WriteString(srcIDValue)
		_, _ = buff.WriteString("->")
		_, _ = buff.WriteString(dstIDValue)
		_, _ = buff.WriteString(rankValueStatement)

		nRecord++
	}

	if nRecord == 0 {
		return "", 0, nil
	}

	return buff.String(), nRecord, nil
}

func (e *Edge) importError(err error, formatWithArgs ...any) *errors.ImportError {
	return errors.AsOrNewImportError(err, formatWithArgs...).SetEdgeName(e.Name)
}

func (n *EdgeNodeRef) Complete() {
	if n.ID != nil {
		n.ID.Complete()
	}
}

func (n *EdgeNodeRef) Validate() error {
	if n.Name == "" {
		return n.importError(errors.ErrNoNodeName)
	}
	if n.ID == nil {
		return n.importError(errors.ErrNoNodeID)
	}
	//revive:disable-next-line:if-return
	if err := n.ID.Validate(); err != nil {
		return err
	}
	return nil
}

func (n *EdgeNodeRef) IDValue(record Record) (string, error) {
	return n.ID.Value(record)
}

func (n *EdgeNodeRef) importError(err error, formatWithArgs ...any) *errors.ImportError {
	return errors.AsOrNewImportError(err, formatWithArgs...).SetNodeName(n.Name)
}

func (es Edges) Complete() {
	for i := range es {
		es[i].Complete()
	}
}

func (es Edges) Validate() error {
	for i := range es {
		if err := es[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}
