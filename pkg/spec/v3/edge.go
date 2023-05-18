package specv3

import (
	"fmt"
	"strings"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/bytebufferpool"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"
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

		fnInsertStatement func(records ...Record) (string, error)
		insertPrefix      string // "INSERT EDGE name(prop_name, ..., prop_name) VALUES "
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

func (e *Edge) Options(opts ...EdgeOption) *Edge {
	for _, opt := range opts {
		opt(e)
	}
	return e
}

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

	e.fnInsertStatement = e.insertStatementWithoutRank
	if e.Rank != nil {
		e.Rank.Complete()
		e.fnInsertStatement = e.insertStatementWithRank
	}

	e.Props.Complete()

	// default enable IGNORE_EXISTED_INDEX
	insertPrefixFmt := "INSERT EDGE IGNORE_EXISTED_INDEX %s(%s) VALUES "
	if e.IgnoreExistedIndex != nil && !*e.IgnoreExistedIndex {
		insertPrefixFmt = "INSERT EDGE %s(%s) VALUES "
	}

	e.insertPrefix = fmt.Sprintf(
		insertPrefixFmt,
		utils.ConvertIdentifier(e.Name),
		strings.Join(e.Props.NameList(), ", "),
	)
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

	return nil
}

func (e *Edge) InsertStatement(records ...Record) (string, error) {
	return e.fnInsertStatement(records...)
}

func (e *Edge) insertStatementWithoutRank(records ...Record) (string, error) {
	buff := bytebufferpool.Get()
	defer bytebufferpool.Put(buff)

	buff.SetString(e.insertPrefix)

	for i, record := range records {
		srcIDValue, err := e.Src.IDValue(record)
		if err != nil {
			return "", e.importError(err)
		}
		dstIDValue, err := e.Dst.IDValue(record)
		if err != nil {
			return "", e.importError(err)
		}
		propsValueList, err := e.Props.ValueList(record)
		if err != nil {
			return "", e.importError(err)
		}

		if i > 0 {
			_, _ = buff.WriteString(", ")
		}

		// "%s->%s:(%s)"
		_, _ = buff.WriteString(srcIDValue)
		_, _ = buff.WriteString("->")
		_, _ = buff.WriteString(dstIDValue)
		_, _ = buff.WriteString(":(")
		_, _ = buff.WriteStringSlice(propsValueList, ", ")
		_, _ = buff.WriteString(")")
	}
	return buff.String(), nil
}

func (e *Edge) insertStatementWithRank(records ...Record) (string, error) {
	buff := bytebufferpool.Get()
	defer bytebufferpool.Put(buff)

	buff.SetString(e.insertPrefix)

	for i, record := range records {
		srcIDValue, err := e.Src.IDValue(record)
		if err != nil {
			return "", e.importError(err)
		}
		dstIDValue, err := e.Dst.IDValue(record)
		if err != nil {
			return "", e.importError(err)
		}
		rankValue, err := e.Rank.Value(record)
		if err != nil {
			return "", e.importError(err)
		}
		propsValueList, err := e.Props.ValueList(record)
		if err != nil {
			return "", e.importError(err)
		}

		if i > 0 {
			_, _ = buff.WriteString(", ")
		}

		// "%s->%s@%s:(%s)"
		_, _ = buff.WriteString(srcIDValue)
		_, _ = buff.WriteString("->")
		_, _ = buff.WriteString(dstIDValue)
		_, _ = buff.WriteString("@")
		_, _ = buff.WriteString(rankValue)
		_, _ = buff.WriteString(":(")
		_, _ = buff.WriteStringSlice(propsValueList, ", ")
		_, _ = buff.WriteString(")")
	}
	return buff.String(), nil
}

func (e *Edge) importError(err error, formatWithArgs ...any) *errors.ImportError { //nolint:unparam
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
