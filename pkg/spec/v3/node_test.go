package specv3

import (
	stderrors "errors"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"
	specbase "github.com/vesoft-inc/nebula-importer/v4/pkg/spec/base"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Node", func() {
	Describe(".Complete", func() {
		It("should complete", func() {
			node := NewNode(
				"name",
				WithNodeID(&NodeID{
					Name: "id",
					Type: ValueTypeInt,
				}),
				WithNodeProps(&Prop{Name: "prop1", Type: ValueTypeString}),
				WithNodeProps(&Prop{Name: "prop2", Type: ValueTypeInt}),
			)
			node.Complete()
			Expect(node.Validate()).NotTo(HaveOccurred())

			Expect(node.Name).To(Equal("name"))

			Expect(node.ID.Name).To(Equal("vid"))
			Expect(node.ID.Type).To(Equal(ValueTypeInt))

			Expect(node.Props).To(HaveLen(2))
			Expect(node.Props[0].Name).To(Equal("prop1"))
			Expect(node.Props[0].Type).To(Equal(ValueTypeString))
			Expect(node.Props[1].Name).To(Equal("prop2"))
			Expect(node.Props[1].Type).To(Equal(ValueTypeInt))
		})
	})

	Describe(".Validate", func() {
		It("no name", func() {
			node := NewNode("")
			err := node.Validate()
			Expect(stderrors.Is(err, errors.ErrNoNodeName)).To(BeTrue())
		})

		It("no id", func() {
			node := NewNode("name")
			err := node.Validate()
			Expect(stderrors.Is(err, errors.ErrNoNodeID)).To(BeTrue())
		})

		It("id validate failed", func() {
			node := NewNode(
				"name",
				WithNodeID(&NodeID{Name: "id", Type: "unsupported"}),
			)
			err := node.Validate()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrUnsupportedValueType)).To(BeTrue())
		})

		It("props validate failed", func() {
			node := NewNode(
				"name",
				WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt}),
				WithNodeProps(&Prop{Name: "prop", Type: "unsupported"}),
			)
			err := node.Validate()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrUnsupportedValueType)).To(BeTrue())
		})

		It("filter validate failed", func() {
			node := NewNode(
				"name",
				WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt}),
				WithNodeFilter(&specbase.Filter{
					Expr: "",
				}),
			)
			err := node.Validate()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrFilterSyntax)).To(BeTrue())
		})

		It("mode validate failed", func() {
			node := NewNode(
				"name",
				WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt}),
				WithNodeMode(specbase.Mode("x")),
			)
			err := node.Validate()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrUnsupportedMode)).To(BeTrue())
		})

		It("mode validate update no props failed", func() {
			node := NewNode(
				"name",
				WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt}),
				WithNodeMode(specbase.UpdateMode),
			)
			err := node.Validate()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrNoProps)).To(BeTrue())
		})

		It("success without props", func() {
			node := NewNode(
				"name",
				WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt}),
			)
			node.Complete()
			err := node.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		It("success with props", func() {
			node := NewNode(
				"name",
				WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt}),
				WithNodeProps(&Prop{Name: "prop", Type: ValueTypeString}),
			)
			node.Complete()
			err := node.Validate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe(".Statement", func() {
		When("INSERT", func() {
			When("no props", func() {
				var node *Node
				BeforeEach(func() {
					node = NewNode(
						"name",
						WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt, Index: 0}),
					)
					node.Complete()
					err := node.Validate()
					Expect(err).NotTo(HaveOccurred())
				})

				It("one record", func() {
					statement, nRecord, err := node.Statement([]string{"1", "1.1", "str1"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(1))
					Expect(statement).To(Equal("INSERT VERTEX IGNORE_EXISTED_INDEX `name`() VALUES 1:()"))
				})

				It("two record", func() {
					statement, nRecord, err := node.Statement([]string{"1", "1.1", "str1"}, []string{"2", "2.2", "str2"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(2))
					Expect(statement).To(Equal("INSERT VERTEX IGNORE_EXISTED_INDEX `name`() VALUES 1:(), 2:()"))
				})

				It("failed id no record", func() {
					statement, nRecord, err := node.Statement([]string{})
					Expect(err).To(HaveOccurred())
					Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(BeEmpty())
				})
			})

			When("one prop", func() {
				var node *Node
				BeforeEach(func() {
					node = NewNode(
						"name",
						WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt, Index: 0}),
						WithNodeProps(
							&Prop{Name: "prop1", Type: ValueTypeString, Index: 2},
						),
					)
					node.Complete()
					err := node.Validate()
					Expect(err).NotTo(HaveOccurred())
				})

				It("one record", func() {
					statement, nRecord, err := node.Statement([]string{"1", "1.1", "str1"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(1))
					Expect(statement).To(Equal("INSERT VERTEX IGNORE_EXISTED_INDEX `name`(`prop1`) VALUES 1:(\"str1\")"))
				})

				It("two record", func() {
					statement, nRecord, err := node.Statement([]string{"1", "1.1", "str1"}, []string{"2", "2.2", "str2"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(2))
					Expect(statement).To(Equal("INSERT VERTEX IGNORE_EXISTED_INDEX `name`(`prop1`) VALUES 1:(\"str1\"), 2:(\"str2\")"))
				})

				It("failed id no record", func() {
					statement, nRecord, err := node.Statement([]string{})
					Expect(err).To(HaveOccurred())
					Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(BeEmpty())
				})

				It("failed prop no record", func() {
					statement, nRecord, err := node.Statement([]string{"1"})
					Expect(err).To(HaveOccurred())
					Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(BeEmpty())
				})
			})

			When("many props", func() {
				var node *Node
				BeforeEach(func() {
					node = NewNode(
						"name",
						WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt, Index: 0}),
						WithNodeProps(
							&Prop{Name: "prop1", Type: ValueTypeString, Index: 2},
							&Prop{Name: "prop2", Type: ValueTypeDouble, Index: 1},
						),
						WithNodeMode(specbase.InsertMode),
					)
					node.Complete()
					err := node.Validate()
					Expect(err).NotTo(HaveOccurred())
				})

				It("one record", func() {
					statement, nRecord, err := node.Statement([]string{"1", "1.1", "str1"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(1))
					Expect(statement).To(Equal("INSERT VERTEX IGNORE_EXISTED_INDEX `name`(`prop1`, `prop2`) VALUES 1:(\"str1\", 1.1)"))
				})

				It("two record", func() {
					statement, nRecord, err := node.Statement([]string{"1", "1.1", "str1"}, []string{"2", "2.2", "str2"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(2))
					Expect(statement).To(Equal("INSERT VERTEX IGNORE_EXISTED_INDEX `name`(`prop1`, `prop2`) VALUES 1:(\"str1\", 1.1), 2:(\"str2\", 2.2)"))
				})

				It("failed id no record", func() {
					statement, nRecord, err := node.Statement([]string{})
					Expect(err).To(HaveOccurred())
					Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(BeEmpty())
				})

				It("failed prop no record", func() {
					statement, nRecord, err := node.Statement([]string{"1"})
					Expect(err).To(HaveOccurred())
					Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(BeEmpty())
				})
			})

			When("WithNodeIgnoreExistedIndex", func() {
				It("WithNodeIgnoreExistedIndex false", func() {
					node := NewNode(
						"name",
						WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt, Index: 0}),
						WithNodeIgnoreExistedIndex(false),
					)
					node.Complete()
					err := node.Validate()
					Expect(err).NotTo(HaveOccurred())

					statement, nRecord, err := node.Statement([]string{"1"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(1))
					Expect(statement).To(Equal("INSERT VERTEX `name`() VALUES 1:()"))
				})
				It("WithNodeIgnoreExistedIndex true", func() {
					node := NewNode(
						"name",
						WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt, Index: 0}),
						WithNodeIgnoreExistedIndex(true),
					)
					node.Complete()
					err := node.Validate()
					Expect(err).NotTo(HaveOccurred())

					statement, nRecord, err := node.Statement([]string{"1"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(1))
					Expect(statement).To(Equal("INSERT VERTEX IGNORE_EXISTED_INDEX `name`() VALUES 1:()"))
				})
			})

			When("WithNodeFilter", func() {
				It("WithNodeFilter error", func() {
					node := NewNode(
						"name",
						WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt, Index: 0}),
						WithNodeProps(
							&Prop{Name: "prop1", Type: ValueTypeString, Index: 1},
						),
						WithNodeFilter(&specbase.Filter{
							Expr: "",
						}),
						WithNodeMode(specbase.InsertMode),
					)
					node.Complete()
					err := node.Validate()
					Expect(err).To(HaveOccurred())
					Expect(stderrors.Is(err, errors.ErrFilterSyntax)).To(BeTrue())
				})
				It("WithNodeFilter successfully", func() {
					node := NewNode(
						"name",
						WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt, Index: 0}),
						WithNodeProps(
							&Prop{Name: "prop1", Type: ValueTypeString, Index: 1},
						),
						WithNodeFilter(&specbase.Filter{
							Expr: `(Record[0] == "1" or Record[0] == "2" or Record[0] == "3") and Record[1] != "A"`,
						}),
					)
					node.Complete()
					err := node.Validate()
					Expect(err).NotTo(HaveOccurred())

					// all true
					statement, nRecord, err := node.Statement([]string{"1", "B"}, []string{"2", "C"}, []string{"3", "D"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(3))
					Expect(statement).To(Equal("INSERT VERTEX IGNORE_EXISTED_INDEX `name`(`prop1`) VALUES 1:(\"B\"), 2:(\"C\"), 3:(\"D\")"))

					// partially true
					statement, nRecord, err = node.Statement([]string{"2", "A"}, []string{"3", "D"}, []string{"4", "E"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(1))
					Expect(statement).To(Equal("INSERT VERTEX IGNORE_EXISTED_INDEX `name`(`prop1`) VALUES 3:(\"D\")"))

					// all false
					statement, nRecord, err = node.Statement([]string{"1", "A"}, []string{"2", "A"}, []string{"4", "E"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(Equal(""))

					// filter failed
					statement, nRecord, err = node.Statement([]string{"1"})
					Expect(err).To(HaveOccurred())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(Equal(""))
				})
			})
		})

		When("UPDATE", func() {
			When("one prop", func() {
				var node *Node
				BeforeEach(func() {
					node = NewNode(
						"name",
						WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt, Index: 0}),
						WithNodeProps(
							&Prop{Name: "prop1", Type: ValueTypeString, Index: 2},
						),
						WithNodeMode(specbase.UpdateMode),
					)
					node.Complete()
					err := node.Validate()
					Expect(err).NotTo(HaveOccurred())
				})

				It("one record", func() {
					statement, nRecord, err := node.Statement([]string{"1", "1.1", "str1"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(1))
					Expect(statement).To(Equal("UPDATE VERTEX ON `name` 1 SET `prop1` = \"str1\";"))
				})

				It("two record", func() {
					statement, nRecord, err := node.Statement([]string{"1", "1.1", "str1"}, []string{"2", "2.2", "str2"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(2))
					Expect(statement).To(Equal("UPDATE VERTEX ON `name` 1 SET `prop1` = \"str1\";UPDATE VERTEX ON `name` 2 SET `prop1` = \"str2\";"))
				})

				It("failed id no record", func() {
					statement, nRecord, err := node.Statement([]string{})
					Expect(err).To(HaveOccurred())
					Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(BeEmpty())
				})

				It("failed prop no record", func() {
					statement, nRecord, err := node.Statement([]string{"1"})
					Expect(err).To(HaveOccurred())
					Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(BeEmpty())
				})
			})

			When("many props", func() {
				var node *Node
				BeforeEach(func() {
					node = NewNode(
						"name",
						WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt, Index: 0}),
						WithNodeProps(
							&Prop{Name: "prop1", Type: ValueTypeString, Index: 2},
							&Prop{Name: "prop2", Type: ValueTypeDouble, Index: 1},
						),
						WithNodeMode(specbase.UpdateMode),
					)
					node.Complete()
					err := node.Validate()
					Expect(err).NotTo(HaveOccurred())
				})

				It("one record", func() {
					statement, nRecord, err := node.Statement([]string{"1", "1.1", "str1"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(1))
					Expect(statement).To(Equal("UPDATE VERTEX ON `name` 1 SET `prop1` = \"str1\", `prop2` = 1.1;"))
				})

				It("two record", func() {
					statement, nRecord, err := node.Statement([]string{"1", "1.1", "str1"}, []string{"2", "2.2", "str2"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(2))
					Expect(statement).To(Equal("UPDATE VERTEX ON `name` 1 SET `prop1` = \"str1\", `prop2` = 1.1;UPDATE VERTEX ON `name` 2 SET `prop1` = \"str2\", `prop2` = 2.2;"))
				})

				It("failed id no record", func() {
					statement, nRecord, err := node.Statement([]string{})
					Expect(err).To(HaveOccurred())
					Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(BeEmpty())
				})

				It("failed prop no record", func() {
					statement, nRecord, err := node.Statement([]string{"1"})
					Expect(err).To(HaveOccurred())
					Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(BeEmpty())
				})
			})

			When("WithNodeFilter", func() {
				It("WithNodeFilter error", func() {
					node := NewNode(
						"name",
						WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt, Index: 0}),
						WithNodeProps(
							&Prop{Name: "prop1", Type: ValueTypeString, Index: 1},
						),
						WithNodeFilter(&specbase.Filter{
							Expr: "",
						}),
						WithNodeMode(specbase.UpdateMode),
					)
					node.Complete()
					err := node.Validate()
					Expect(err).To(HaveOccurred())
					Expect(stderrors.Is(err, errors.ErrFilterSyntax)).To(BeTrue())
				})
				It("WithNodeFilter successfully", func() {
					node := NewNode(
						"name",
						WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt, Index: 0}),
						WithNodeProps(
							&Prop{Name: "prop1", Type: ValueTypeString, Index: 1},
						),
						WithNodeFilter(&specbase.Filter{
							Expr: `(Record[0] == "1" or Record[0] == "2" or Record[0] == "3") and Record[1] != "A"`,
						}),
						WithNodeMode(specbase.UpdateMode),
					)
					node.Complete()
					err := node.Validate()
					Expect(err).NotTo(HaveOccurred())

					// all true
					statement, nRecord, err := node.Statement([]string{"1", "B"}, []string{"2", "C"}, []string{"3", "D"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(3))
					Expect(statement).To(Equal("UPDATE VERTEX ON `name` 1 SET `prop1` = \"B\";UPDATE VERTEX ON `name` 2 SET `prop1` = \"C\";UPDATE VERTEX ON `name` 3 SET `prop1` = \"D\";"))

					// partially true
					statement, nRecord, err = node.Statement([]string{"2", "A"}, []string{"3", "D"}, []string{"4", "E"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(1))
					Expect(statement).To(Equal("UPDATE VERTEX ON `name` 3 SET `prop1` = \"D\";"))

					// all false
					statement, nRecord, err = node.Statement([]string{"1", "A"}, []string{"2", "A"}, []string{"4", "E"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(Equal(""))

					// filter failed
					statement, nRecord, err = node.Statement([]string{"1"})
					Expect(err).To(HaveOccurred())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(Equal(""))
				})
			})
		})

		When("DELETE", func() {
			When("no props", func() {
				var node *Node
				BeforeEach(func() {
					node = NewNode(
						"name",
						WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt, Index: 0}),
						WithNodeMode(specbase.DeleteMode),
					)
					node.Complete()
					err := node.Validate()
					Expect(err).NotTo(HaveOccurred())
				})

				It("one record", func() {
					statement, nRecord, err := node.Statement([]string{"1", "1.1", "str1"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(1))
					Expect(statement).To(Equal("DELETE TAG `name` FROM 1;"))
				})

				It("two record", func() {
					statement, nRecord, err := node.Statement([]string{"1", "1.1", "str1"}, []string{"2", "2.2", "str2"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(2))
					Expect(statement).To(Equal("DELETE TAG `name` FROM 1;DELETE TAG `name` FROM 2;"))
				})

				It("failed id no record", func() {
					statement, nRecord, err := node.Statement([]string{})
					Expect(err).To(HaveOccurred())
					Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(BeEmpty())
				})
			})

			When("one prop", func() {
				var node *Node
				BeforeEach(func() {
					node = NewNode(
						"name",
						WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt, Index: 0}),
						WithNodeProps(
							&Prop{Name: "prop1", Type: ValueTypeString, Index: 2},
						),
						WithNodeMode(specbase.DeleteMode),
					)
					node.Complete()
					err := node.Validate()
					Expect(err).NotTo(HaveOccurred())
				})

				It("one record", func() {
					statement, nRecord, err := node.Statement([]string{"1", "1.1", "str1"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(1))
					Expect(statement).To(Equal("DELETE TAG `name` FROM 1;"))
				})

				It("two record", func() {
					statement, nRecord, err := node.Statement([]string{"1", "1.1", "str1"}, []string{"2", "2.2", "str2"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(2))
					Expect(statement).To(Equal("DELETE TAG `name` FROM 1;DELETE TAG `name` FROM 2;"))
				})

				It("failed id no record", func() {
					statement, nRecord, err := node.Statement([]string{})
					Expect(err).To(HaveOccurred())
					Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(BeEmpty())
				})
			})

			When("WithNodeFilter", func() {
				It("WithNodeFilter error", func() {
					node := NewNode(
						"name",
						WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt, Index: 0}),
						WithNodeProps(
							&Prop{Name: "prop1", Type: ValueTypeString, Index: 1},
						),
						WithNodeFilter(&specbase.Filter{
							Expr: "",
						}),
						WithNodeMode(specbase.DeleteMode),
					)
					node.Complete()
					err := node.Validate()
					Expect(err).To(HaveOccurred())
					Expect(stderrors.Is(err, errors.ErrFilterSyntax)).To(BeTrue())
				})
				It("WithNodeFilter successfully", func() {
					node := NewNode(
						"name",
						WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt, Index: 0}),
						WithNodeProps(
							&Prop{Name: "prop1", Type: ValueTypeString, Index: 1},
						),
						WithNodeFilter(&specbase.Filter{
							Expr: `(Record[0] == "1" or Record[0] == "2" or Record[0] == "3") and Record[1] != "A"`,
						}),
						WithNodeMode(specbase.DeleteMode),
					)
					node.Complete()
					err := node.Validate()
					Expect(err).NotTo(HaveOccurred())

					// all true
					statement, nRecord, err := node.Statement([]string{"1", "B"}, []string{"2", "C"}, []string{"3", "D"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(3))
					Expect(statement).To(Equal("DELETE TAG `name` FROM 1;DELETE TAG `name` FROM 2;DELETE TAG `name` FROM 3;"))

					// partially true
					statement, nRecord, err = node.Statement([]string{"2", "A"}, []string{"3", "D"}, []string{"4", "E"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(1))
					Expect(statement).To(Equal("DELETE TAG `name` FROM 3;"))

					// all false
					statement, nRecord, err = node.Statement([]string{"1", "A"}, []string{"2", "A"}, []string{"4", "E"})
					Expect(err).NotTo(HaveOccurred())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(Equal(""))

					// filter failed
					statement, nRecord, err = node.Statement([]string{"1"})
					Expect(err).To(HaveOccurred())
					Expect(nRecord).To(Equal(0))
					Expect(statement).To(Equal(""))
				})
			})
		})
	})
})

var _ = Describe("Nodes", func() {
	Describe(".Complete", func() {
		It("default value type", func() {
			nodes := Nodes{
				NewNode("name1", WithNodeID(&NodeID{}), WithNodeProps(&Prop{})),
				NewNode("name2", WithNodeID(&NodeID{}), WithNodeProps(&Prop{})),
			}
			nodes.Complete()
			Expect(nodes).To(HaveLen(2))
			Expect(nodes[0].Name).To(Equal("name1"))
			Expect(nodes[1].Name).To(Equal("name2"))
		})
	})

	DescribeTable(".Validate",
		func(nodes Nodes, failedIndex int) {
			nodes.Complete()
			err := nodes.Validate()
			if failedIndex >= 0 {
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(nodes[failedIndex].Validate()))
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		},
		Entry("empty nodes",
			Nodes{},
			-1,
		),
		Entry("success",
			Nodes{
				NewNode("name1", WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt})),
				NewNode("name2", WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt})),
				NewNode("name3", WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt})),
				NewNode("name4", WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt})),
			},
			-1,
		),
		Entry("failed at 0",
			Nodes{
				NewNode(""),
				NewNode("name1", WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt})),
				NewNode("name2", WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt})),
				NewNode("name3", WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt})),
				NewNode("name4", WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt})),
			},
			0,
		),
		Entry("failed at 1",
			Nodes{
				NewNode("name1", WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt})),
				NewNode("failed"),
				NewNode("name2", WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt})),
				NewNode("name3", WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt})),
				NewNode("name4", WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt})),
			},
			1,
		),
		Entry("failed at end",
			Nodes{
				NewNode("name1", WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt})),
				NewNode("name2", WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt})),
				NewNode("name3", WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt})),
				NewNode("name4", WithNodeID(&NodeID{Name: "id", Type: ValueTypeInt})),
				NewNode("failed", WithNodeID(&NodeID{Name: "id", Type: "NO"})),
			},
			4,
		),
	)
})
