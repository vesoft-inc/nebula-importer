package specv3

import (
	stderrors "errors"
	"fmt"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Edge", func() {
	Describe(".Complete", func() {
		It("should complete", func() {
			edge := NewEdge(
				"name",
				WithEdgeSrc(&EdgeNodeRef{
					Name: "srcNodeName",
					ID: &NodeID{
						Name: "id",
						Type: ValueTypeInt,
					},
				}),
				WithEdgeDst(&EdgeNodeRef{
					Name: "dstNodeName",
					ID: &NodeID{
						Name: "id",
						Type: ValueTypeString,
					},
				}),
				WithEdgeProps(&Prop{Name: "prop1", Type: ValueTypeString}),
				WithEdgeProps(&Prop{Name: "prop2", Type: ValueTypeInt}),
			)
			edge.Complete()

			Expect(edge.Name).To(Equal("name"))

			Expect(edge.Src.Name).To(Equal(strSrc))
			Expect(edge.Src.ID.Name).To(Equal(strVID))
			Expect(edge.Src.ID.Type).To(Equal(ValueTypeInt))

			Expect(edge.Dst.Name).To(Equal(strDst))
			Expect(edge.Dst.ID.Name).To(Equal(strVID))
			Expect(edge.Dst.ID.Type).To(Equal(ValueTypeString))

			Expect(edge.Props).To(HaveLen(2))
			Expect(edge.Props[0].Name).To(Equal("prop1"))
			Expect(edge.Props[0].Type).To(Equal(ValueTypeString))
			Expect(edge.Props[1].Name).To(Equal("prop2"))
			Expect(edge.Props[1].Type).To(Equal(ValueTypeInt))
		})
	})

	Describe(".Validate", func() {
		It("no name", func() {
			edge := NewEdge("")
			err := edge.Validate()
			Expect(stderrors.Is(err, errors.ErrNoEdgeName)).To(BeTrue())
		})

		It("no src", func() {
			edge := NewEdge("name")
			err := edge.Validate()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrNoEdgeSrc)).To(BeTrue())
		})

		It("src validate failed", func() {
			edge := NewEdge("name", WithEdgeSrc(&EdgeNodeRef{
				Name: "node",
			}))
			err := edge.Validate()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrNoNodeID)).To(BeTrue())
		})

		It("no dst", func() {
			edge := NewEdge("name", WithEdgeSrc(&EdgeNodeRef{
				Name: "srcNodeName",
				ID: &NodeID{
					Name: "id",
					Type: ValueTypeInt,
				},
			}))
			err := edge.Validate()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrNoEdgeDst)).To(BeTrue())
		})

		It("dst validate failed", func() {
			edge := NewEdge("name", WithEdgeSrc(&EdgeNodeRef{
				Name: "srcNodeName",
				ID: &NodeID{
					Name: "id",
					Type: ValueTypeInt,
				},
			}), WithEdgeDst(&EdgeNodeRef{
				Name: "dstNodeName",
			}))
			err := edge.Validate()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrNoNodeID)).To(BeTrue())
		})

		It("dst validate failed 2", func() {
			edge := NewEdge("name", WithEdgeSrc(&EdgeNodeRef{
				Name: "srcNodeName",
				ID: &NodeID{
					Name: "id",
					Type: ValueTypeInt,
				},
			}), WithEdgeDst(&EdgeNodeRef{
				Name: "dstNodeName",
				ID:   &NodeID{},
			}))
			err := edge.Validate()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrNoNodeIDName)).To(BeTrue())
		})

		It("props validate failed", func() {
			edge := NewEdge(
				"name",
				WithEdgeSrc(&EdgeNodeRef{
					Name: "srcNodeName",
					ID: &NodeID{
						Name: "id",
						Type: ValueTypeInt,
					},
				}),
				WithEdgeDst(&EdgeNodeRef{
					Name: "dstNodeName",
					ID: &NodeID{
						Name: "id",
						Type: ValueTypeString,
					},
				}),
				WithEdgeProps(&Prop{Name: "prop"}),
			)
			err := edge.Validate()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrUnsupportedValueType)).To(BeTrue())
		})

		It("success without props", func() {
			edge := NewEdge(
				"name",
				WithEdgeSrc(&EdgeNodeRef{
					Name: "srcNodeName",
					ID: &NodeID{
						Name: "id",
						Type: ValueTypeInt,
					},
				}),
				WithEdgeDst(&EdgeNodeRef{
					Name: "dstNodeName",
					ID: &NodeID{
						Name: "id",
						Type: ValueTypeString,
					},
				}),
			)
			err := edge.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		It("success with props", func() {
			edge := NewEdge(
				"name",
				WithEdgeSrc(&EdgeNodeRef{
					Name: "srcNodeName",
					ID: &NodeID{
						Name: "id",
						Type: ValueTypeInt,
					},
				}),
				WithEdgeDst(&EdgeNodeRef{
					Name: "dstNodeName",
					ID: &NodeID{
						Name: "id",
						Type: ValueTypeString,
					},
				}),
				WithEdgeProps(&Prop{Name: "prop", Type: ValueTypeString}),
			)
			err := edge.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		It("WithRank failed", func() {
			edge := NewEdge(
				"name",
				WithEdgeSrc(&EdgeNodeRef{
					Name: "srcNodeName",
					ID: &NodeID{
						Name: "id",
						Type: ValueTypeInt,
					},
				}),
				WithEdgeDst(&EdgeNodeRef{
					Name: "dstNodeName",
					ID: &NodeID{
						Name: "id",
						Type: ValueTypeString,
					},
				}),
				WithRank(&Rank{Index: -1}),
			)
			err := edge.Validate()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrInvalidIndex)).To(BeTrue())
		})

		It("WithRank successfully", func() {
			edge := NewEdge(
				"name",
				WithEdgeSrc(&EdgeNodeRef{
					Name: "srcNodeName",
					ID: &NodeID{
						Name: "id",
						Type: ValueTypeInt,
					},
				}),
				WithEdgeDst(&EdgeNodeRef{
					Name: "dstNodeName",
					ID: &NodeID{
						Name: "id",
						Type: ValueTypeString,
					},
				}),
				WithRank(&Rank{Index: 0}),
			)
			err := edge.Validate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe(".InsertStatement", func() {
		When("no props", func() {
			var edge *Edge
			BeforeEach(func() {
				edge = NewEdge(
					"name",
					WithEdgeSrc(&EdgeNodeRef{
						Name: "srcNodeName",
						ID: &NodeID{
							Name:  "id",
							Type:  ValueTypeInt,
							Index: 0,
						},
					}),
					WithEdgeDst(&EdgeNodeRef{
						Name: "dstNodeName",
						ID: &NodeID{
							Name:  "id",
							Type:  ValueTypeString,
							Index: 1,
						},
					}),
				)
				edge.Complete()
				err := edge.Validate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("one record", func() {
				statement, err := edge.InsertStatement([]string{"1", "id1", "1.1", "str1"})
				Expect(err).NotTo(HaveOccurred())
				Expect(statement).To(Equal("INSERT EDGE IGNORE_EXISTED_INDEX `name`() VALUES 1->\"id1\":()"))
			})

			It("two record", func() {
				statement, err := edge.InsertStatement([]string{"1", "id1", "1.1", "str1"}, []string{"2", "id2", "2.2", "str2"})
				Expect(err).NotTo(HaveOccurred())
				Expect(statement).To(Equal("INSERT EDGE IGNORE_EXISTED_INDEX `name`() VALUES 1->\"id1\":(), 2->\"id2\":()"))
			})

			It("src failed", func() {
				statement, err := edge.InsertStatement([]string{})
				Expect(err).To(HaveOccurred())
				Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
				Expect(statement).To(BeEmpty())
			})

			It("dst failed", func() {
				statement, err := edge.InsertStatement([]string{"1"})
				Expect(err).To(HaveOccurred())
				Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
				Expect(statement).To(BeEmpty())
			})
		})

		When("one prop", func() {
			var edge *Edge
			BeforeEach(func() {
				edge = NewEdge(
					"name",
					WithEdgeSrc(&EdgeNodeRef{
						Name: "srcNodeName",
						ID: &NodeID{
							Name:  "id",
							Type:  ValueTypeInt,
							Index: 0,
						},
					}),
					WithEdgeDst(&EdgeNodeRef{
						Name: "dstNodeName",
						ID: &NodeID{
							Name:  "id",
							Type:  ValueTypeString,
							Index: 1,
						},
					}),
					WithEdgeProps(
						&Prop{Name: "prop1", Type: ValueTypeString, Index: 3},
					),
				)
				edge.Complete()
				err := edge.Validate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("one record", func() {
				statement, err := edge.InsertStatement([]string{"1", "id1", "1.1", "str1"})
				Expect(err).NotTo(HaveOccurred())
				Expect(statement).To(Equal("INSERT EDGE IGNORE_EXISTED_INDEX `name`(`prop1`) VALUES 1->\"id1\":(\"str1\")"))
			})

			It("two record", func() {
				statement, err := edge.InsertStatement([]string{"1", "id1", "1.1", "str1"}, []string{"2", "id2", "2.2", "str2"})
				Expect(err).NotTo(HaveOccurred())
				Expect(statement).To(Equal("INSERT EDGE IGNORE_EXISTED_INDEX `name`(`prop1`) VALUES 1->\"id1\":(\"str1\"), 2->\"id2\":(\"str2\")"))
			})

			It("src failed", func() {
				statement, err := edge.InsertStatement([]string{})
				Expect(err).To(HaveOccurred())
				Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
				Expect(statement).To(BeEmpty())
			})

			It("dst failed", func() {
				statement, err := edge.InsertStatement([]string{"1"})
				Expect(err).To(HaveOccurred())
				Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
				Expect(statement).To(BeEmpty())
			})

			It("props failed", func() {
				statement, err := edge.InsertStatement([]string{"1", "id1"})
				Expect(err).To(HaveOccurred())
				Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
				Expect(statement).To(BeEmpty())
			})
		})

		When("many props", func() {
			var edge *Edge
			BeforeEach(func() {
				edge = NewEdge(
					"name",
					WithEdgeSrc(&EdgeNodeRef{
						Name: "srcNodeName",
						ID: &NodeID{
							Name:  "id",
							Type:  ValueTypeInt,
							Index: 0,
						},
					}),
					WithEdgeDst(&EdgeNodeRef{
						Name: "dstNodeName",
						ID: &NodeID{
							Name:  "id",
							Type:  ValueTypeString,
							Index: 1,
						},
					}),
					WithEdgeProps(
						&Prop{Name: "prop1", Type: ValueTypeString, Index: 3},
						&Prop{Name: "prop2", Type: ValueTypeDouble, Index: 2},
					),
				)
				edge.Complete()
				err := edge.Validate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("one record", func() {
				statement, err := edge.InsertStatement([]string{"1", "id1", "1.1", "str1"})
				Expect(err).NotTo(HaveOccurred())
				Expect(statement).To(Equal("INSERT EDGE IGNORE_EXISTED_INDEX `name`(`prop1`, `prop2`) VALUES 1->\"id1\":(\"str1\", 1.1)"))
			})

			It("two record", func() {
				statement, err := edge.InsertStatement([]string{"1", "id1", "1.1", "str1"}, []string{"2", "id2", "2.2", "str2"})
				Expect(err).NotTo(HaveOccurred())
				Expect(statement).To(Equal("INSERT EDGE IGNORE_EXISTED_INDEX `name`(`prop1`, `prop2`) VALUES 1->\"id1\":(\"str1\", 1.1), 2->\"id2\":(\"str2\", 2.2)"))
			})

			It("src failed", func() {
				statement, err := edge.InsertStatement([]string{})
				Expect(err).To(HaveOccurred())
				Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
				Expect(statement).To(BeEmpty())
			})

			It("dst failed", func() {
				statement, err := edge.InsertStatement([]string{"1"})
				Expect(err).To(HaveOccurred())
				Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
				Expect(statement).To(BeEmpty())
			})

			It("props failed", func() {
				statement, err := edge.InsertStatement([]string{"1", "id1"})
				Expect(err).To(HaveOccurred())
				Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
				Expect(statement).To(BeEmpty())
			})
		})

		When("WithRank", func() {
			var edge *Edge
			BeforeEach(func() {
				edge = NewEdge(
					"name",
					WithEdgeSrc(&EdgeNodeRef{
						Name: "srcNodeName",
						ID: &NodeID{
							Name:  "id",
							Type:  ValueTypeInt,
							Index: 0,
						},
					}),
					WithEdgeDst(&EdgeNodeRef{
						Name: "dstNodeName",
						ID: &NodeID{
							Name:  "id",
							Type:  ValueTypeString,
							Index: 1,
						},
					}),
					WithRank(&Rank{Index: 2}),
					WithEdgeProps(
						&Prop{Name: "prop1", Type: ValueTypeString, Index: 4},
						&Prop{Name: "prop2", Type: ValueTypeDouble, Index: 3},
					),
				)
				edge.Complete()
				err := edge.Validate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("one record", func() {
				statement, err := edge.InsertStatement([]string{"1", "id1", "1", "1.1", "str1"})
				Expect(err).NotTo(HaveOccurred())
				Expect(statement).To(Equal("INSERT EDGE IGNORE_EXISTED_INDEX `name`(`prop1`, `prop2`) VALUES 1->\"id1\"@1:(\"str1\", 1.1)"))
			})

			It("two record", func() {
				statement, err := edge.InsertStatement([]string{"1", "id1", "1", "1.1", "str1"}, []string{"2", "id2", "2", "2.2", "str2"})
				Expect(err).NotTo(HaveOccurred())
				Expect(statement).To(Equal("INSERT EDGE IGNORE_EXISTED_INDEX `name`(`prop1`, `prop2`) VALUES 1->\"id1\"@1:(\"str1\", 1.1), 2->\"id2\"@2:(\"str2\", 2.2)"))
			})

			It("src failed", func() {
				statement, err := edge.InsertStatement([]string{})
				Expect(err).To(HaveOccurred())
				Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
				Expect(statement).To(BeEmpty())
			})

			It("dst failed", func() {
				statement, err := edge.InsertStatement([]string{"1"})
				Expect(err).To(HaveOccurred())
				Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
				Expect(statement).To(BeEmpty())
			})

			It("rank failed", func() {
				statement, err := edge.InsertStatement([]string{"1", "id1"})
				Expect(err).To(HaveOccurred())
				Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
				Expect(statement).To(BeEmpty())
			})

			It("props failed", func() {
				statement, err := edge.InsertStatement([]string{"1", "id1", "1"})
				Expect(err).To(HaveOccurred())
				Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
				Expect(statement).To(BeEmpty())
			})
		})

		When("WithEdgeIgnoreExistedIndex", func() {
			It("WithEdgeIgnoreExistedIndex false", func() {
				edge := NewEdge(
					"name",
					WithEdgeSrc(&EdgeNodeRef{
						Name: "srcNodeName",
						ID: &NodeID{
							Name:  "id",
							Type:  ValueTypeInt,
							Index: 0,
						},
					}),
					WithEdgeDst(&EdgeNodeRef{
						Name: "dstNodeName",
						ID: &NodeID{
							Name:  "id",
							Type:  ValueTypeString,
							Index: 1,
						},
					}),
					WithEdgeIgnoreExistedIndex(false),
				)
				edge.Complete()
				err := edge.Validate()
				Expect(err).NotTo(HaveOccurred())

				statement, err := edge.InsertStatement([]string{"1", "id1"})
				Expect(err).NotTo(HaveOccurred())
				Expect(statement).To(Equal("INSERT EDGE `name`() VALUES 1->\"id1\":()"))
			})
			It("WithEdgeIgnoreExistedIndex true", func() {
				edge := NewEdge(
					"name",
					WithEdgeSrc(&EdgeNodeRef{
						Name: "srcNodeName",
						ID: &NodeID{
							Name:  "id",
							Type:  ValueTypeInt,
							Index: 0,
						},
					}),
					WithEdgeDst(&EdgeNodeRef{
						Name: "dstNodeName",
						ID: &NodeID{
							Name:  "id",
							Type:  ValueTypeString,
							Index: 1,
						},
					}),
					WithEdgeIgnoreExistedIndex(true),
				)
				edge.Complete()
				err := edge.Validate()
				Expect(err).NotTo(HaveOccurred())

				statement, err := edge.InsertStatement([]string{"1", "id1"})
				Expect(err).NotTo(HaveOccurred())
				Expect(statement).To(Equal("INSERT EDGE IGNORE_EXISTED_INDEX `name`() VALUES 1->\"id1\":()"))
			})
		})
	})
})

var _ = Describe("Edges", func() {
	Describe(".Complete", func() {
		It("should complete", func() {
			edges := Edges{
				NewEdge(
					"name1",
					WithEdgeSrc(&EdgeNodeRef{
						Name: "srcNodeName",
						ID: &NodeID{
							Name: "id",
							Type: ValueTypeInt,
						},
					}),
					WithEdgeDst(&EdgeNodeRef{
						Name: "dstNodeName",
						ID: &NodeID{
							Name: "id",
							Type: ValueTypeString,
						},
					}),
					WithEdgeProps(&Prop{}),
				),
				NewEdge(
					"name2",
					WithEdgeSrc(&EdgeNodeRef{
						Name: "srcNodeName",
						ID: &NodeID{
							Name: "id",
							Type: ValueTypeInt,
						},
					}),
					WithEdgeDst(&EdgeNodeRef{
						Name: "dstNodeName",
						ID: &NodeID{
							Name: "id",
							Type: ValueTypeString,
						},
					}),
					WithEdgeProps(&Prop{}),
				),
			}
			edges.Complete()
			Expect(edges).To(HaveLen(2))
			Expect(edges[0].Name).To(Equal("name1"))
			Expect(edges[1].Name).To(Equal("name2"))
		})
	})

	Describe(".Validate", func() {
		var edges Edges
		BeforeEach(func() {
			for i := 1; i <= 4; i++ {
				edges = append(edges, NewEdge(
					fmt.Sprintf("name%d", i),
					WithEdgeSrc(&EdgeNodeRef{
						Name: "srcNodeName",
						ID: &NodeID{
							Name: "id",
							Type: ValueTypeInt,
						},
					}),
					WithEdgeDst(&EdgeNodeRef{
						Name: "dstNodeName",
						ID: &NodeID{
							Name: "id",
							Type: ValueTypeString,
						},
					}),
				))
			}
		})
		DescribeTable("table cases",
			func(getEdges func() Edges, failedIndex int) {
				es := getEdges()
				err := es.Validate()
				if failedIndex >= 0 {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(es[failedIndex].Validate()))
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},
			Entry("empty nodes",
				func() Edges { return Edges{} },
				-1,
			),
			Entry("success",
				func() Edges { return edges },
				-1,
			),
			Entry("failed at 0",
				func() Edges {
					return Edges{
						NewEdge(""),
						edges[0],
						edges[1],
						edges[2],
						edges[3],
					}
				},
				0,
			),
			Entry("failed at 1",
				func() Edges {
					return Edges{
						edges[0],
						NewEdge("failed"),
						edges[1],
						edges[2],
						edges[3],
					}
				},
				1,
			),
			Entry("failed at end",
				func() Edges {
					return Edges{
						edges[0],
						edges[1],
						edges[2],
						edges[3],
						NewEdge("failed", WithEdgeSrc(&EdgeNodeRef{})),
					}
				},
				4,
			),
			Entry("failed at id validate",
				func() Edges {
					return Edges{
						edges[0],
						edges[1],
						edges[2],
						edges[3],
						NewEdge("failed", WithEdgeSrc(&EdgeNodeRef{ID: &NodeID{
							Type: "unsupported",
						}})),
					}
				},
				4,
			),
		)
	})
})
