package specv3

import (
	stderrors "errors"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Graph", func() {
	Describe(".Complete", func() {
		It("should complete", func() {
			graph := NewGraph(
				"graphName",
				WithGraphNodes(
					NewNode(
						"node1",
						WithNodeID(&NodeID{
							Name: "id1",
							Type: ValueTypeString,
						}),
					),
				),
				WithGraphNodes(
					NewNode(
						"node2",
						WithNodeID(&NodeID{
							Name: "id2",
							Type: ValueTypeInt,
						}),
					),
				),
				WithGraphEdges(
					NewEdge(
						"edge1",
						WithEdgeSrc(&EdgeNodeRef{
							Name: "node1",
							ID: &NodeID{
								Name: "id1",
								Type: ValueTypeInt,
							},
						}),
						WithEdgeDst(&EdgeNodeRef{
							Name: "node1",
							ID: &NodeID{
								Name: "id1",
								Type: ValueTypeInt,
							},
						}),
					),
				),
				WithGraphEdges(
					NewEdge(
						"edge2",
						WithEdgeSrc(&EdgeNodeRef{
							Name: "node2",
							ID: &NodeID{
								Name: "id2",
								Type: ValueTypeString,
							},
						}),
						WithEdgeDst(&EdgeNodeRef{
							Name: "node2",
							ID: &NodeID{
								Name: "id2",
								Type: ValueTypeString,
							},
						}),
					),
				),
			)
			graph.Complete()
			Expect(graph.Name).To(Equal("graphName"))

			Expect(graph.Nodes).To(HaveLen(2))
			Expect(graph.Nodes[0].Name).To(Equal("node1"))
			Expect(graph.Nodes[0].ID.Name).To(Equal(strVID))
			Expect(graph.Nodes[1].Name).To(Equal("node2"))
			Expect(graph.Nodes[1].ID.Name).To(Equal(strVID))

			Expect(graph.Edges).To(HaveLen(2))
			Expect(graph.Edges[0].Name).To(Equal("edge1"))
			Expect(graph.Edges[0].Src.Name).To(Equal(strSrc))
			Expect(graph.Edges[0].Src.ID.Name).To(Equal(strVID))
			Expect(graph.Edges[0].Dst.Name).To(Equal(strDst))
			Expect(graph.Edges[0].Dst.ID.Name).To(Equal(strVID))
			Expect(graph.Edges[1].Name).To(Equal("edge2"))
			Expect(graph.Edges[1].Src.Name).To(Equal(strSrc))
			Expect(graph.Edges[1].Src.ID.Name).To(Equal(strVID))
			Expect(graph.Edges[1].Dst.Name).To(Equal(strDst))
			Expect(graph.Edges[1].Dst.ID.Name).To(Equal(strVID))
		})
	})

	Describe(".Validate", func() {
		It("no name", func() {
			graph := NewGraph("")
			err := graph.Validate()
			Expect(stderrors.Is(err, errors.ErrNoSpaceName)).To(BeTrue())
		})

		It("nodes validate failed", func() {
			graph := NewGraph("graphName", WithGraphNodes(NewNode("")))
			err := graph.Validate()
			Expect(stderrors.Is(err, errors.ErrNoNodeName)).To(BeTrue())
		})

		It("nodes validate failed", func() {
			graph := NewGraph("graphName", WithGraphEdges(NewEdge("")))
			err := graph.Validate()
			Expect(stderrors.Is(err, errors.ErrNoEdgeName)).To(BeTrue())
		})

		It("success", func() {
			graph := NewGraph(
				"graphName",
				WithGraphNodes(
					NewNode(
						"node1",
						WithNodeID(&NodeID{
							Name: "id",
							Type: ValueTypeInt,
						}),
					),
				),
				WithGraphEdges(
					NewEdge(
						"edge1",
						WithEdgeSrc(&EdgeNodeRef{
							Name: "node1",
							ID: &NodeID{
								Name: "id",
								Type: ValueTypeInt,
							},
						}),
						WithEdgeDst(&EdgeNodeRef{
							Name: "node1",
							ID: &NodeID{
								Name: "id",
								Type: ValueTypeInt,
							},
						}),
					),
				),
			)
			graph.Complete()
			err := graph.Validate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe(".NodeStatement", func() {
		var graph *Graph
		BeforeEach(func() {
			graph = NewGraph(
				"graphName",
				WithGraphNodes(
					NewNode(
						"node1",
						WithNodeID(&NodeID{
							Name:  "id",
							Type:  ValueTypeInt,
							Index: 0,
						}),
					),
				),
			)
			graph.Complete()
			err := graph.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		It("success", func() {
			node := graph.Nodes[0]
			statement, nRecord, err := graph.NodeStatement(node, []string{"1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(nRecord).To(Equal(1))
			Expect(statement).To(Equal("INSERT VERTEX IGNORE_EXISTED_INDEX `node1`() VALUES 1:()"))

			b := graph.NodeStatementBuilder(node)
			statement, nRecord, err = b.Build([]string{"1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(nRecord).To(Equal(1))
			Expect(statement).To(Equal("INSERT VERTEX IGNORE_EXISTED_INDEX `node1`() VALUES 1:()"))
		})

		It("failed", func() {
			node := graph.Nodes[0]
			statement, nRecord, err := graph.NodeStatement(node, []string{})
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
			Expect(nRecord).To(Equal(0))
			Expect(statement).To(Equal(""))

			b := graph.NodeStatementBuilder(node)
			statement, nRecord, err = b.Build([]string{})
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
			Expect(nRecord).To(Equal(0))
			Expect(statement).To(Equal(""))
		})
	})

	Describe(".EdgeStatement", func() {
		var graph *Graph
		BeforeEach(func() {
			graph = NewGraph(
				"graphName",
				WithGraphEdges(
					NewEdge(
						"edge1",
						WithEdgeSrc(&EdgeNodeRef{
							Name: "node1",
							ID: &NodeID{
								Name:  "id",
								Type:  ValueTypeInt,
								Index: 0,
							},
						}),
						WithEdgeDst(&EdgeNodeRef{
							Name: "node1",
							ID: &NodeID{
								Name:  "id",
								Type:  ValueTypeInt,
								Index: 1,
							},
						}),
					),
				),
			)
			graph.Complete()
			err := graph.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		It("success", func() {
			edge := graph.Edges[0]
			statement, nRecord, err := graph.EdgeStatement(edge, []string{"1", "2"})
			Expect(err).NotTo(HaveOccurred())
			Expect(nRecord).To(Equal(1))
			Expect(statement).To(Equal("INSERT EDGE IGNORE_EXISTED_INDEX `edge1`() VALUES 1->2:()"))

			b := graph.EdgeStatementBuilder(edge)
			statement, nRecord, err = b.Build([]string{"1", "2"})
			Expect(err).NotTo(HaveOccurred())
			Expect(nRecord).To(Equal(1))
			Expect(statement).To(Equal("INSERT EDGE IGNORE_EXISTED_INDEX `edge1`() VALUES 1->2:()"))
		})

		It("failed", func() {
			edge := graph.Edges[0]
			statement, nRecord, err := graph.EdgeStatement(edge, []string{})
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
			Expect(nRecord).To(Equal(0))
			Expect(statement).To(Equal(""))

			b := graph.EdgeStatementBuilder(edge)
			statement, nRecord, err = b.Build([]string{})
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
			Expect(nRecord).To(Equal(0))
			Expect(statement).To(Equal(""))
		})
	})

	Describe("", func() {
		var graph *Graph
		BeforeEach(func() {
			graph = NewGraph("graphName", WithGraphNodes(NewNode("node1")))
		})

		It("exists", func() {
			node, ok := graph.GetNodeByName("node1")
			Expect(ok).To(BeTrue())
			Expect(node).NotTo(BeNil())
			Expect(node.Name).To(Equal("node1"))
		})

		It("not exists", func() {
			node, ok := graph.GetNodeByName("not-exists")
			Expect(ok).To(BeFalse())
			Expect(node).To(BeNil())
		})
	})

	Describe("", func() {
		var graph *Graph
		BeforeEach(func() {
			graph = NewGraph("graphName", WithGraphEdges(NewEdge("edge1")))
		})

		It("exists", func() {
			edge, ok := graph.GetEdgeByName("edge1")
			Expect(ok).To(BeTrue())
			Expect(edge).NotTo(BeNil())
			Expect(edge.Name).To(Equal("edge1"))
		})

		It("not exists", func() {
			edge, ok := graph.GetEdgeByName("not-exists")
			Expect(ok).To(BeFalse())
			Expect(edge).To(BeNil())
		})
	})
})
