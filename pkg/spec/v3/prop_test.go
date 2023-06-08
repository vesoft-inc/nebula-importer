package specv3

import (
	stderrors "errors"
	"strings"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Prop", func() {
	Describe(".Complete", func() {
		It("no value type", func() {
			prop := &Prop{}
			prop.Complete()
			Expect(prop.Type).To(Equal(ValueTypeDefault))

			prop = &Prop{
				Type: ValueTypeInt,
			}
			prop.Complete()
			Expect(prop.Type).To(Equal(ValueTypeInt))
		})
		It("have value type", func() {
			prop := &Prop{
				Type: ValueTypeInt,
			}
			prop.Complete()
			Expect(prop.Type).To(Equal(ValueTypeInt))
		})
	})

	DescribeTable(".Validate",
		func(prop *Prop, expectErr error) {
			err := prop.Validate()
			if expectErr != nil {
				if Expect(err).To(HaveOccurred()) {
					Expect(stderrors.Is(err, expectErr)).To(BeTrue())
					e, ok := errors.AsImportError(err)
					Expect(ok).To(BeTrue())
					Expect(e.Cause()).To(Equal(expectErr))
					Expect(e.PropName()).To(Equal(prop.Name))
				}
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		},
		Entry("no prop name", &Prop{}, errors.ErrNoPropName),
		Entry("unsupported value type", &Prop{Name: "a", Type: "x"}, errors.ErrUnsupportedValueType),
		Entry("supported value type", &Prop{Name: "a", Type: ValueTypeDefault}, nil),
	)

	It(".Validate init picker failed", func() {
		prop := &Prop{Name: "a", Type: "unsupported"}
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		patches.ApplyGlobalVar(&supportedPropValueTypes, map[ValueType]struct{}{
			ValueType(strings.ToUpper(string(prop.Type))): {},
		})

		err := prop.Validate()
		Expect(err).To(HaveOccurred())
		Expect(stderrors.Is(err, errors.ErrUnsupportedValueType)).To(BeTrue())
	})

	DescribeTable(".Value",
		func(p *Prop, record Record, expectValue string, expectErr error) {
			val, err := func() (string, error) {
				p.Complete()
				err := p.Validate()
				if err != nil {
					return "", err
				}
				return p.Value(record)
			}()
			if expectErr != nil {
				if Expect(err).To(HaveOccurred()) {
					Expect(stderrors.Is(err, expectErr)).To(BeTrue())
					e, ok := errors.AsImportError(err)
					Expect(ok).To(BeTrue())
					Expect(e.Cause()).To(Equal(expectErr))
					Expect(e.PropName()).To(Equal(p.Name))
				}
				Expect(val).To(Equal(expectValue))
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(expectValue))
			}
		},
		Entry("no record empty",
			&Prop{Name: "p1"},
			Record([]string{}),
			"",
			errors.ErrNoRecord,
		),
		Entry("no record",
			&Prop{
				Name:  "p1",
				Type:  ValueTypeInt,
				Index: 1,
			},
			Record([]string{"0"}),
			"",
			errors.ErrNoRecord,
		),
		Entry("record int",
			&Prop{
				Name:  "p1",
				Type:  ValueTypeInt,
				Index: 0,
			},
			Record([]string{"1"}),
			"1",
			nil,
		),
		Entry("record string",
			&Prop{
				Name:  "p1",
				Type:  ValueTypeString,
				Index: 0,
			},
			Record([]string{"str"}),
			"\"str\"",
			nil,
		),
		Entry("record double",
			&Prop{
				Name:  "p1",
				Type:  ValueTypeDouble,
				Index: 0,
			},
			Record([]string{"1.1"}),
			"1.1",
			nil,
		),
		Entry("Nullable",
			&Prop{
				Name:     "p1",
				Type:     ValueTypeInt,
				Index:    0,
				Nullable: true,
			},
			Record([]string{""}),
			"NULL",
			nil,
		),
		Entry("Nullable not null",
			&Prop{
				Name:     "p1",
				Type:     ValueTypeInt,
				Index:    0,
				Nullable: true,
			},
			Record([]string{"1"}),
			"1",
			nil,
		),
		Entry("NullValue",
			&Prop{
				Name:      "p1",
				Type:      ValueTypeInt,
				Index:     0,
				Nullable:  true,
				NullValue: "N/A",
			},
			Record([]string{"N/A"}),
			"NULL",
			nil,
		),
		Entry("NullValue not null",
			&Prop{
				Name:      "p1",
				Type:      ValueTypeInt,
				Index:     0,
				Nullable:  true,
				NullValue: "N/A",
			},
			Record([]string{"1"}),
			"1",
			nil,
		),
		Entry("AlternativeIndices 0",
			&Prop{
				Name:               "p1",
				Type:               ValueTypeInt,
				Index:              0,
				Nullable:           true,
				NullValue:          "N/A",
				AlternativeIndices: []int{},
			},
			Record([]string{"1"}),
			"1",
			nil,
		),
		Entry("AlternativeIndices 1 pick index",
			&Prop{
				Name:               "p1",
				Type:               ValueTypeInt,
				Index:              0,
				Nullable:           true,
				NullValue:          "N/A",
				AlternativeIndices: []int{1},
			},
			Record([]string{"1"}),
			"1",
			nil,
		),
		Entry("AlternativeIndices 1 pick failed",
			&Prop{
				Name:               "p1",
				Type:               ValueTypeInt,
				Index:              0,
				Nullable:           true,
				NullValue:          "N/A",
				AlternativeIndices: []int{1},
			},
			Record([]string{"N/A"}),
			"",
			errors.ErrNoRecord,
		),
		Entry("AlternativeIndices 1 pick 1",
			&Prop{
				Name:               "p1",
				Type:               ValueTypeInt,
				Index:              0,
				Nullable:           true,
				NullValue:          "N/A",
				AlternativeIndices: []int{1},
			},
			Record([]string{"N/A", "1"}),
			"1",
			nil,
		),
		Entry("AlternativeIndices 1 pick null",
			&Prop{
				Name:               "p1",
				Type:               ValueTypeInt,
				Index:              0,
				Nullable:           true,
				NullValue:          "N/A",
				AlternativeIndices: []int{1},
			},
			Record([]string{"N/A", "N/A"}),
			"NULL",
			nil,
		),
		Entry("AlternativeIndices 1 pick default",
			&Prop{
				Name:               "p1",
				Type:               ValueTypeInt,
				Index:              0,
				Nullable:           true,
				NullValue:          "N/A",
				AlternativeIndices: []int{1},
				DefaultValue:       func() *string { s := "0"; return &s }(),
			},
			Record([]string{"N/A", "N/A"}),
			"0",
			nil,
		),
		Entry("unsupported value type",
			&Prop{
				Name: "p1",
				Type: "unsupported",
			},
			Record([]string{"1"}),
			nil,
			errors.ErrUnsupportedValueType,
		),
	)
	DescribeTable(".SetValue",
		func(p *Prop, record Record, expectValue string, expectErr error) {
			setVal, err := func() (string, error) {
				p.Complete()
				err := p.Validate()
				if err != nil {
					return "", err
				}
				return p.SetValue(record)
			}()
			if expectErr != nil {
				if Expect(err).To(HaveOccurred()) {
					Expect(stderrors.Is(err, expectErr)).To(BeTrue())
					e, ok := errors.AsImportError(err)
					Expect(ok).To(BeTrue())
					Expect(e.Cause()).To(Equal(expectErr))
					Expect(e.PropName()).To(Equal(p.Name))
				}
				Expect(setVal).To(Equal(expectValue))
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(setVal).To(Equal(expectValue))
			}
		},
		Entry("no record empty",
			&Prop{Name: "p1"},
			Record([]string{}),
			"",
			errors.ErrNoRecord,
		),
		Entry("no record",
			&Prop{
				Name:  "p1",
				Type:  ValueTypeInt,
				Index: 1,
			},
			Record([]string{"0"}),
			"",
			errors.ErrNoRecord,
		),
		Entry("record int",
			&Prop{
				Name:  "p1",
				Type:  ValueTypeInt,
				Index: 0,
			},
			Record([]string{"1"}),
			"`p1` = 1",
			nil,
		),
		Entry("record string",
			&Prop{
				Name:  "p1",
				Type:  ValueTypeString,
				Index: 0,
			},
			Record([]string{"str"}),
			"`p1` = \"str\"",
			nil,
		),
		Entry("record double",
			&Prop{
				Name:  "p1",
				Type:  ValueTypeDouble,
				Index: 0,
			},
			Record([]string{"1.1"}),
			"`p1` = 1.1",
			nil,
		),
	)
})

var _ = Describe("Props", func() {
	Describe(".Complete", func() {
		It("default value type", func() {
			prop1 := Prop{}
			prop2 := Prop{
				Type: ValueTypeInt,
			}
			prop3 := Prop{
				Type: ValueTypeDouble,
			}
			p1, p2, p3 := prop1, prop2, prop3
			props := Props{&p1, &p2, &p3}
			props.Complete()
			Expect(props).To(HaveLen(3))

			p1.Complete()
			Expect(props[0]).To(Equal(&p1))
			p2.Complete()
			Expect(props[1]).To(Equal(&p2))
			p3.Complete()
			Expect(props[2]).To(Equal(&p3))
		})
	})

	DescribeTable(".Validate",
		func(props Props, failedIndex int) {
			err := props.Validate()
			if failedIndex >= 0 {
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(props[failedIndex].Validate()))
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		},
		Entry("empty props",
			Props{},
			-1,
		),
		Entry("success",
			Props{
				&Prop{Name: "a", Type: ValueTypeInt},
				&Prop{Name: "b", Type: ValueTypeString},
				&Prop{Name: "c", Type: ValueTypeDouble},
			},
			-1,
		),
		Entry("failed at 0",
			Props{
				&Prop{Name: ""},
				&Prop{Name: "a", Type: ValueTypeInt},
				&Prop{Name: "b", Type: ValueTypeString},
				&Prop{Name: "c", Type: ValueTypeDouble},
			},
			0,
		),
		Entry("failed at 1",
			Props{
				&Prop{Name: "a", Type: ValueTypeInt},
				&Prop{Name: "failed"},
				&Prop{Name: "b", Type: ValueTypeString},
				&Prop{Name: "c", Type: ValueTypeDouble},
			},
			1,
		),
		Entry("failed at end",
			Props{
				&Prop{Name: "a", Type: ValueTypeInt},
				&Prop{Name: "b", Type: ValueTypeString},
				&Prop{Name: "c", Type: ValueTypeDouble},
				&Prop{Name: "failed"},
			},
			3,
		),
	)

	DescribeTable(".ValueList",
		func(props Props, record Record, expectValueList []string, failedIndex int) {
			valueList, err := func() ([]string, error) {
				props.Complete()
				if err := props.Validate(); err != nil {
					return nil, err
				}
				return props.ValueList(record)
			}()
			if failedIndex >= 0 {
				if Expect(err).To(HaveOccurred()) {
					_, expectErr := props[failedIndex].Value(record)
					Expect(err).To(Equal(expectErr))
				}
				Expect(valueList).To(Equal(expectValueList))
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(valueList).To(Equal(expectValueList))
			}
		},
		Entry("empty props",
			Props{},
			[]string{"1", "1.1", "str"},
			[]string{},
			-1,
		),
		Entry("success",
			Props{
				&Prop{Name: "a", Type: ValueTypeInt, Index: 0},
				&Prop{Name: "b", Type: ValueTypeString, Index: 2},
				&Prop{Name: "c", Type: ValueTypeDouble, Index: 1},
			},
			[]string{"1", "1.1", "str"},
			[]string{"1", "\"str\"", "1.1"},
			-1,
		),
		Entry("failed",
			Props{
				&Prop{Name: "a", Type: ValueTypeInt, Index: 0},
			},
			nil,
			nil,
			0,
		),
	)

	DescribeTable(".SetValueList",
		func(props Props, record Record, expectValueList []string, failedIndex int) {
			setValueList, err := func() ([]string, error) {
				props.Complete()
				if err := props.Validate(); err != nil {
					return nil, err
				}
				return props.SetValueList(record)
			}()
			if failedIndex >= 0 {
				if Expect(err).To(HaveOccurred()) {
					_, expectErr := props[failedIndex].Value(record)
					Expect(err).To(Equal(expectErr))
				}
				Expect(setValueList).To(Equal(expectValueList))
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(setValueList).To(Equal(expectValueList))
			}
		},
		Entry("empty props",
			Props{},
			[]string{"1", "1.1", "str"},
			[]string{},
			-1,
		),
		Entry("success",
			Props{
				&Prop{Name: "a", Type: ValueTypeInt, Index: 0},
				&Prop{Name: "b", Type: ValueTypeString, Index: 2},
				&Prop{Name: "c", Type: ValueTypeDouble, Index: 1},
			},
			[]string{"1", "1.1", "str"},
			[]string{"`a` = 1", "`b` = \"str\"", "`c` = 1.1"},
			-1,
		),
		Entry("failed",
			Props{
				&Prop{Name: "a", Type: ValueTypeInt, Index: 0},
			},
			nil,
			nil,
			0,
		),
	)

	DescribeTable(".NameList",
		func(props Props, expectValueList []string) {
			props.Complete()
			valueList := props.NameList()
			Expect(valueList).To(Equal(expectValueList))
		},
		Entry("empty props",
			Props{},
			[]string{},
		),
		Entry("one",
			Props{
				&Prop{Name: "a", Type: ValueTypeInt, Index: 0},
			},
			[]string{"`a`"},
		),
		Entry("many",
			Props{
				&Prop{Name: "a", Type: ValueTypeInt, Index: 0},
				&Prop{Name: "b", Type: ValueTypeString, Index: 2},
				&Prop{Name: "c", Type: ValueTypeDouble, Index: 1},
			},
			[]string{"`a`", "`b`", "`c`"},
		),
	)

	DescribeTable(".Append",
		func(l, r Props) {
			lLen, rLen := len(l), len(r)
			props := l.Append(r...)
			Expect(props).To(HaveLen(lLen + rLen))
		},
		Entry("nil + nil",
			nil,
			nil,
		),
		Entry("nil + non-nil",
			nil,
			Props{&Prop{}},
		),
		Entry("non-nil + nil",
			Props{&Prop{}, &Prop{}},
			nil,
		),
		Entry("non-nil + non-nil",
			Props{&Prop{}, &Prop{}, &Prop{}},
			Props{&Prop{}, &Prop{}, &Prop{}, &Prop{}},
		),
	)
})
