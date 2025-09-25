package specv3

import (
	stderrors "errors"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NodeID", func() {
	Describe(".Complete", func() {
		It("empty prop", func() {
			nodeID := &NodeID{}
			nodeID.Complete()
			Expect(nodeID.Type).To(Equal(ValueTypeDefault))
		})
	})

	Describe(".Validate", func() {
		It("failed", func() {
			nodeID := &NodeID{Name: "id", Type: "unsupported"}
			err := nodeID.Validate()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrUnsupportedValueType)).To(BeTrue())
		})

		It("success", func() {
			nodeID := &NodeID{Name: "id", Type: ValueTypeDefault}
			err := nodeID.Validate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	DescribeTable(".Value",
		func(nodeID *NodeID, record Record, expectValue string, expectErr error) {
			if nodeID.Function != nil {
				patches := gomonkey.NewPatches()
				defer patches.Reset()
				patches.ApplyGlobalVar(&supportedNodeIDFunctions, map[string]struct{}{
					"HASH": {},
				})
			}

			val, err := func() (string, error) {
				nodeID.Complete()
				err := nodeID.Validate()
				if err != nil {
					return "", err
				}
				return nodeID.Value(record)
			}()

			if expectErr != nil {
				if Expect(err).To(HaveOccurred()) {
					Expect(stderrors.Is(err, expectErr)).To(BeTrue())
					e, ok := errors.AsImportError(err)
					Expect(ok).To(BeTrue())
					Expect(e.Cause()).To(Equal(expectErr))
					Expect(e.PropName()).To(Equal(nodeID.Name))
				}
				Expect(val).To(Equal(expectValue))
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(expectValue))
			}
		},
		Entry("no record empty",
			&NodeID{Name: "id"},
			Record([]string{}),
			"",
			errors.ErrNoRecord,
		),
		Entry("no record",
			&NodeID{
				Name:  "id",
				Type:  ValueTypeInt,
				Index: 1,
			},
			Record([]string{"0"}),
			"",
			errors.ErrNoRecord,
		),
		Entry("record int",
			&NodeID{
				Name:  "id",
				Type:  ValueTypeInt,
				Index: 0,
			},
			Record([]string{"1"}),
			"1",
			nil,
		),
		Entry("record string",
			&NodeID{
				Name:  "id",
				Type:  ValueTypeString,
				Index: 0,
			},
			Record([]string{"id"}),
			"\"id\"",
			nil,
		),
		Entry("ConcatItems",
			&NodeID{
				Name:        "id",
				Type:        ValueTypeString,
				ConcatItems: []any{"c1", 3, "c2", 1, 2, "c3", 0},
			},
			Record([]string{"s0", "s1", "s2", "s3"}),
			"\"c1s3c2s1s2c3s0\"",
			nil,
		),
		Entry("ConcatItems failed type",
			&NodeID{
				Name:        "id",
				Type:        ValueTypeString,
				ConcatItems: []any{true},
			},
			Record([]string{"1"}),
			"",
			errors.ErrUnsupportedConcatItemType,
		),
		Entry("ConcatItems failed no record",
			&NodeID{
				Name:        "id",
				Type:        ValueTypeString,
				ConcatItems: []any{"c1", 3, "c2", 1, 2, "c3", 0, 10},
			},
			Record([]string{"s0", "s1", "s2", "s3"}),
			"",
			errors.ErrNoRecord,
		),
		Entry("Function",
			&NodeID{
				Name:        "id",
				Type:        ValueTypeInt,
				ConcatItems: []any{"c1", 3, "c2", 1, 2, "c3", 0},
				Function:    func() *string { s := "hash"; return &s }(),
			},
			Record([]string{"s0", "s1", "s2", "s3"}),
			"hash(\"c1s3c2s1s2c3s0\")",
			nil,
		),
		Entry("unsupported value type",
			&NodeID{
				Name:  "id",
				Type:  ValueTypeDouble,
				Index: 0,
			},
			Record([]string{"1.1"}),
			nil,
			errors.ErrUnsupportedValueType,
		),
		Entry("unsupported function",
			&NodeID{
				Name:     "id",
				Type:     ValueTypeInt,
				Index:    0,
				Function: func() *string { s := "unsupported"; return &s }(),
			},
			Record([]string{"1"}),
			nil,
			errors.ErrUnsupportedFunction,
		),
	)
})
