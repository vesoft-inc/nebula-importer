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
	// Node is VERTEX in 3.x
	Node struct {
		Name  string  `yaml:"name"`
		ID    *NodeID `yaml:"id"`
		Props Props   `yaml:"props,omitempty"`

		IgnoreExistedIndex *bool `yaml:"ignoreExistedIndex,omitempty"`

		Filter *specbase.Filter `yaml:"filter,omitempty"`

		insertPrefix string // // "INSERT EDGE name(prop_name, ..., prop_name) VALUES "
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

func (n *Node) Options(opts ...NodeOption) *Node {
	for _, opt := range opts {
		opt(n)
	}
	return n
}

func (n *Node) Complete() {
	if n.ID != nil {
		n.ID.Complete()
		n.ID.Name = strVID
	}
	n.Props.Complete()

	// default enable IGNORE_EXISTED_INDEX
	insertPrefixFmt := "INSERT VERTEX IGNORE_EXISTED_INDEX %s(%s) VALUES "
	if n.IgnoreExistedIndex != nil && !*n.IgnoreExistedIndex {
		insertPrefixFmt = "INSERT VERTEX %s(%s) VALUES "
	}
	n.insertPrefix = fmt.Sprintf(
		insertPrefixFmt,
		utils.ConvertIdentifier(n.Name),
		strings.Join(n.Props.NameList(), ", "),
	)
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

	return nil
}

func (n *Node) InsertStatement(records ...Record) (statement string, nRecord int, err error) {
	buff := bytebufferpool.Get()
	defer bytebufferpool.Put(buff)

	buff.SetString(n.insertPrefix)

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

		// "%s:(%s)"
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
