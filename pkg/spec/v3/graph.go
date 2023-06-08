package specv3

import (
	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"
	specbase "github.com/vesoft-inc/nebula-importer/v4/pkg/spec/base"
)

type (
	Graph struct {
		Name  string `yaml:"name"`
		Nodes Nodes  `yaml:"tags,omitempty"`
		Edges Edges  `yaml:"edges,omitempty"`
	}

	GraphOption func(*Graph)
)

func NewGraph(name string, opts ...GraphOption) *Graph {
	g := &Graph{
		Name: name,
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

func WithGraphNodes(nodes ...*Node) GraphOption {
	return func(g *Graph) {
		g.AddNodes(nodes...)
	}
}

func WithGraphEdges(edges ...*Edge) GraphOption {
	return func(g *Graph) {
		g.AddEdges(edges...)
	}
}

func (g *Graph) AddNodes(nodes ...*Node) {
	g.Nodes = append(g.Nodes, nodes...)
}

func (g *Graph) AddEdges(edges ...*Edge) {
	g.Edges = append(g.Edges, edges...)
}

func (g *Graph) Complete() {
	if g.Nodes != nil {
		g.Nodes.Complete()
	}
	if g.Edges != nil {
		g.Edges.Complete()
	}
}

func (g *Graph) Validate() error {
	if g.Name == "" {
		return errors.ErrNoSpaceName
	}
	if err := g.Nodes.Validate(); err != nil {
		return err
	}
	//revive:disable-next-line:if-return
	if err := g.Edges.Validate(); err != nil {
		return err
	}

	return nil
}

func (g *Graph) NodeStatement(n *Node, records ...Record) (statement string, nRecord int, err error) {
	statement, nRecord, err = n.Statement(records...)
	if err != nil {
		return "", 0, g.importError(err).SetGraphName(g.Name).SetNodeName(n.Name)
	}
	return statement, nRecord, nil
}

func (g *Graph) NodeStatementBuilder(n *Node) specbase.StatementBuilder {
	return specbase.StatementBuilderFunc(func(records ...specbase.Record) (string, int, error) {
		return g.NodeStatement(n, records...)
	})
}

func (g *Graph) EdgeStatement(e *Edge, records ...Record) (statement string, nRecord int, err error) {
	statement, nRecord, err = e.Statement(records...)
	if err != nil {
		return "", 0, g.importError(err).SetGraphName(g.Name).SetEdgeName(e.Name)
	}
	return statement, nRecord, nil
}

func (g *Graph) EdgeStatementBuilder(e *Edge) specbase.StatementBuilder {
	return specbase.StatementBuilderFunc(func(records ...specbase.Record) (string, int, error) {
		return g.EdgeStatement(e, records...)
	})
}

func (g *Graph) GetNodeByName(name string) (*Node, bool) {
	for _, n := range g.Nodes {
		if n.Name == name {
			return n, true
		}
	}
	return nil, false
}

func (g *Graph) GetEdgeByName(name string) (*Edge, bool) {
	for _, e := range g.Edges {
		if e.Name == name {
			return e, true
		}
	}
	return nil, false
}

func (g *Graph) importError(err error, formatWithArgs ...any) *errors.ImportError {
	return errors.AsOrNewImportError(err, formatWithArgs...).SetGraphName(g.Name)
}
