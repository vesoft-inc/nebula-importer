package picker

import (
	stderrors "errors"
	"fmt"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	It("build failed", func() {
		var c Config
		p, err := c.Build()
		Expect(err).To(HaveOccurred())
		Expect(p).To(BeNil())
	})

	Describe("Config cases", func() {
		var (
			strEmpty   = ""
			strStr1    = "str1"
			strInt1    = "1"
			strFunHash = "hash"
		)
		type recordCase struct {
			record        []string
			wantValue     *Value
			wantErr       error
			wantErrString string
		}
		testcases := []struct {
			name     string
			c        Config
			buildErr error
			cases    []recordCase
		}{
			{
				name: "index BOOL",
				c: Config{
					Indices: []int{1},
					Type:    "BOOL",
				},
				cases: []recordCase{
					{
						record:  nil,
						wantErr: errors.ErrNoRecord,
					},
					{
						record:  []string{},
						wantErr: errors.ErrNoRecord,
					},
					{
						record:  []string{"0"},
						wantErr: errors.ErrNoRecord,
					},
					{
						record:    []string{"0", "1"},
						wantValue: &Value{Val: "1", IsNull: false},
					},
					{
						record:    []string{"0", "1", "2"},
						wantValue: &Value{Val: "1", IsNull: false},
					},
				},
			},
			{
				name: "index iNt",
				c: Config{
					Indices: []int{1},
					Type:    "iNt",
				},
				cases: []recordCase{
					{
						record:  nil,
						wantErr: errors.ErrNoRecord,
					},
					{
						record:  []string{},
						wantErr: errors.ErrNoRecord,
					},
					{
						record:  []string{"0"},
						wantErr: errors.ErrNoRecord,
					},
					{
						record:    []string{"0", "1"},
						wantValue: &Value{Val: "1", IsNull: false},
					},
					{
						record:    []string{"0", "1", "2"},
						wantValue: &Value{Val: "1", IsNull: false},
					},
				},
			},
			{
				name: "index Float",
				c: Config{
					Indices: []int{2},
					Type:    "Float",
				},
				cases: []recordCase{
					{
						record:    []string{"0", "1.1", "2.2"},
						wantValue: &Value{Val: "2.2", IsNull: false},
					},
				},
			},
			{
				name: "index double",
				c: Config{
					Indices: []int{3},
					Type:    "double",
				},
				cases: []recordCase{
					{
						record:    []string{"0", "1.1", "2.2", "3.3"},
						wantValue: &Value{Val: "3.3", IsNull: false},
					},
				},
			},
			{
				name: "index string",
				c: Config{
					Indices: []int{1},
					Type:    "string",
				},
				cases: []recordCase{
					{
						record:    []string{"0", "str1", "str2"},
						wantValue: &Value{Val: "\"str1\"", IsNull: false},
					},
				},
			},
			{
				name: "index date",
				c: Config{
					Indices: []int{0},
					Type:    "date",
				},
				cases: []recordCase{
					{
						record:    []string{"2020-01-02"},
						wantValue: &Value{Val: "DATE(\"2020-01-02\")", IsNull: false},
					},
				},
			},
			{
				name: "index time",
				c: Config{
					Indices: []int{0},
					Type:    "time",
				},
				cases: []recordCase{
					{
						record:    []string{"18:38:23.284"},
						wantValue: &Value{Val: "TIME(\"18:38:23.284\")", IsNull: false},
					},
				},
			},
			{
				name: "index datetime",
				c: Config{
					Indices: []int{0},
					Type:    "datetime",
				},
				cases: []recordCase{
					{
						record:    []string{"2020-01-11T19:28:23.284"},
						wantValue: &Value{Val: "DATETIME(\"2020-01-11T19:28:23.284\")", IsNull: false},
					},
				},
			},
			{
				name: "index timestamp",
				c: Config{
					Indices: []int{0},
					Type:    "timestamp",
				},
				cases: []recordCase{
					{
						record:    []string{"2020-01-11T19:28:23"},
						wantValue: &Value{Val: "TIMESTAMP(\"2020-01-11T19:28:23\")", IsNull: false},
					},
					{
						record:    []string{"1578770903"},
						wantValue: &Value{Val: "TIMESTAMP(1578770903)", IsNull: false},
					},
					{
						record:    []string{""},
						wantValue: &Value{Val: "TIMESTAMP(\"\")", IsNull: false},
					},
					{
						record:    []string{"0"},
						wantValue: &Value{Val: "TIMESTAMP(0)", IsNull: false},
					},
					{
						record:    []string{"12"},
						wantValue: &Value{Val: "TIMESTAMP(12)", IsNull: false},
					},
					{
						record:    []string{"0x"},
						wantValue: &Value{Val: "TIMESTAMP(\"0x\")", IsNull: false},
					},
					{
						record:    []string{"0X"},
						wantValue: &Value{Val: "TIMESTAMP(\"0X\")", IsNull: false},
					},
					{
						record:    []string{"0123456789"},
						wantValue: &Value{Val: "TIMESTAMP(0123456789)", IsNull: false},
					},
					{
						record:    []string{"9876543210"},
						wantValue: &Value{Val: "TIMESTAMP(9876543210)", IsNull: false},
					},
					{
						record:    []string{"0x0123456789abcdef"},
						wantValue: &Value{Val: "TIMESTAMP(0x0123456789abcdef)", IsNull: false},
					},
					{
						record:    []string{"0X0123456789ABCDEF"},
						wantValue: &Value{Val: "TIMESTAMP(0X0123456789ABCDEF)", IsNull: false},
					},
				},
			},
			{
				name: "index geography",
				c: Config{
					Indices: []int{0},
					Type:    "geography",
				},
				cases: []recordCase{
					{
						record:    []string{"Polygon((-85.1 34.8,-80.7 28.4,-76.9 34.9,-85.1 34.8))"},
						wantValue: &Value{Val: "ST_GeogFromText(\"Polygon((-85.1 34.8,-80.7 28.4,-76.9 34.9,-85.1 34.8))\")", IsNull: false},
					},
				},
			},
			{
				name: "index geography(point)",
				c: Config{
					Indices: []int{0},
					Type:    "geography(point)",
				},
				cases: []recordCase{
					{
						record:    []string{"Point(0.0 0.0)"},
						wantValue: &Value{Val: "ST_GeogFromText(\"Point(0.0 0.0)\")", IsNull: false},
					},
				},
			},
			{
				name: "index geography(linestring)",
				c: Config{
					Indices: []int{0},
					Type:    "geography(linestring)",
				},
				cases: []recordCase{
					{
						record:    []string{"linestring(0 1, 179.99 89.99)"},
						wantValue: &Value{Val: "ST_GeogFromText(\"linestring(0 1, 179.99 89.99)\")", IsNull: false},
					},
				},
			},
			{
				name: "index geography(polygon)",
				c: Config{
					Indices: []int{0},
					Type:    "geography(polygon)",
				},
				cases: []recordCase{
					{
						record:    []string{"polygon((0 1, 2 4, 3 5, 4 9, 0 1))"},
						wantValue: &Value{Val: "ST_GeogFromText(\"polygon((0 1, 2 4, 3 5, 4 9, 0 1))\")", IsNull: false},
					},
				},
			},
			{
				name: "index unsupported type",
				c: Config{
					Indices: []int{0},
					Type:    "unsupported",
				},
				buildErr: errors.ErrUnsupportedValueType,
			},
			{
				name: "index invalid",
				c: Config{
					Indices: []int{-1},
				},
				buildErr: errors.ErrInvalidIndex,
			},
			{
				name: "concat items index invalid",
				c: Config{
					ConcatItems: []any{"str", -1},
				},
				buildErr: errors.ErrInvalidIndex,
			},
			{
				name: "index Nullable",
				c: Config{
					Indices: []int{1},
					Type:    "string",
					Nullable: func(s string) bool {
						return s == ""
					},
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "", "str2", "str3"},
						wantValue: &Value{Val: "", IsNull: true},
					},
				},
			},
			{
				name: "index Nullable value",
				c: Config{
					Indices: []int{1},
					Type:    "string",
					Nullable: func(s string) bool {
						return s == ""
					},
					NullValue: "",
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "", "str2", "str3"},
						wantValue: &Value{Val: "", IsNull: true},
					},
				},
			},
			{
				name: "index Nullable value changed",
				c: Config{
					Indices: []int{1},
					Type:    "string",
					Nullable: func(s string) bool {
						return s == "__NULL__"
					},
					NullValue: "NULL",
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "__NULL__", "str2", "str3"},
						wantValue: &Value{Val: "NULL", IsNull: true},
					},
				},
			},
			{
				name: "index not Nullable",
				c: Config{
					Indices:   []int{1},
					Type:      "string",
					Nullable:  nil,
					NullValue: "NULL",
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "", "str2", "str3"},
						wantValue: &Value{Val: "\"\"", IsNull: false},
					},
				},
			},
			{
				name: "index not Nullable defaultValue",
				c: Config{
					Indices:      []int{1},
					Type:         "string",
					Nullable:     nil,
					NullValue:    "NULL",
					DefaultValue: &strStr1,
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "", "str2", "str3"},
						wantValue: &Value{Val: "\"\"", IsNull: false},
					},
				},
			},
			{
				name: "index defaultValue string",
				c: Config{
					Indices: []int{1},
					Type:    "string",
					Nullable: func(s string) bool {
						return s == ""
					},
					NullValue:    "NULL",
					DefaultValue: &strStr1,
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "", "str2", "str3"},
						wantValue: &Value{Val: "\"str1\"", IsNull: false},
					},
				},
			},
			{
				name: "index defaultValue string empty",
				c: Config{
					Indices: []int{1},
					Type:    "string",
					Nullable: func(s string) bool {
						return s == "_NULL_"
					},
					NullValue:    "NULL",
					DefaultValue: &strEmpty,
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "_NULL_", "str2", "str3"},
						wantValue: &Value{Val: "\"\"", IsNull: false},
					},
				},
			},
			{
				name: "index defaultValue int",
				c: Config{
					Indices: []int{1},
					Type:    "int",
					Nullable: func(s string) bool {
						return s == ""
					},
					NullValue:    "NULL",
					DefaultValue: &strInt1,
				},
				cases: []recordCase{
					{
						record:    []string{"0", "", "2", "3"},
						wantValue: &Value{Val: "1", IsNull: false},
					},
				},
			},
			{
				name: "index Function string",
				c: Config{
					Indices:  []int{1},
					Type:     "string",
					Function: &strFunHash,
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "str1"},
						wantValue: &Value{Val: "hash(\"str1\")", IsNull: false},
					},
				},
			},
			{
				name: "index Function int",
				c: Config{
					Indices:  []int{1, 2, 3},
					Type:     "int",
					Function: &strFunHash,
				},
				cases: []recordCase{
					{
						record:    []string{"0", "1"},
						wantValue: &Value{Val: "hash(\"1\")", IsNull: false},
					},
				},
			},
			{
				name: "index Function Nullable",
				c: Config{
					Indices: []int{1},
					Type:    "string",
					Nullable: func(s string) bool {
						return s == ""
					},
					NullValue: "NULL",
					Function:  &strFunHash,
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "", "str2", "str3"},
						wantValue: &Value{Val: "NULL", IsNull: true},
					},
				},
			},
			{
				name: "index Function defaultValue",
				c: Config{
					Indices: []int{1},
					Type:    "string",
					Nullable: func(s string) bool {
						return s == ""
					},
					NullValue:    "NULL",
					DefaultValue: &strStr1,
					Function:     &strFunHash,
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "", "str2", "str3"},
						wantValue: &Value{Val: "hash(\"str1\")", IsNull: false},
					},
				},
			},
			{
				name: "indices",
				c: Config{
					Indices: []int{1, 2, 3},
					Type:    "string",
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "", "str2", "str3"},
						wantValue: &Value{Val: "\"\"", IsNull: false},
					},
				},
			},
			{
				name: "indices unsupported type",
				c: Config{
					Indices: []int{1, 2, 3},
					Type:    "unsupported",
				},
				buildErr: errors.ErrUnsupportedValueType,
			},
			{
				name: "indices Nullable unsupported type",
				c: Config{
					Indices: []int{1, 2, 3},
					Type:    "unsupported",
					Nullable: func(s string) bool {
						return s == ""
					},
					DefaultValue: &strEmpty,
				},
				buildErr: errors.ErrUnsupportedValueType,
			},
			{
				name: "indices Nullable",
				c: Config{
					Indices: []int{1, 2, 3},
					Type:    "string",
					Nullable: func(s string) bool {
						return s == ""
					},
				},
				cases: []recordCase{
					{
						record:  []string{"str0", "", ""},
						wantErr: errors.ErrNoRecord,
					},
					{
						record:    []string{"str0", "", "", "str3"},
						wantValue: &Value{Val: "\"str3\"", IsNull: false},
					},
					{
						record:    []string{"str0", "", "", ""},
						wantValue: &Value{Val: "", IsNull: true},
					},
				},
			},
			{
				name: "indices Nullable value",
				c: Config{
					Indices: []int{1, 2, 3},
					Type:    "string",
					Nullable: func(s string) bool {
						return s == ""
					},
					NullValue: "",
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "", "", ""},
						wantValue: &Value{Val: "", IsNull: true},
					},
				},
			},
			{
				name: "indices Nullable value changed",
				c: Config{
					Indices: []int{1, 2, 3},
					Type:    "string",
					Nullable: func(s string) bool {
						return s == "__NULL__"
					},
					NullValue: "NULL",
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "__NULL__", "__NULL__", "__NULL__"},
						wantValue: &Value{Val: "NULL", IsNull: true},
					},
				},
			},
			{
				name: "indices not Nullable",
				c: Config{
					Indices:   []int{1, 2, 3},
					Type:      "string",
					Nullable:  nil,
					NullValue: "NULL",
				},
				cases: []recordCase{
					{
						record:  []string{""},
						wantErr: errors.ErrNoRecord,
					},
					{
						record:    []string{"str0", "", "", ""},
						wantValue: &Value{Val: "\"\"", IsNull: false},
					},
				},
			},
			{
				name: "indices not Nullable defaultValue",
				c: Config{
					Indices:      []int{1, 2, 3},
					Type:         "string",
					Nullable:     nil,
					NullValue:    "NULL",
					DefaultValue: &strStr1,
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "", "", ""},
						wantValue: &Value{Val: "\"\"", IsNull: false},
					},
				},
			},
			{
				name: "indices defaultValue string",
				c: Config{
					Indices: []int{1, 2, 3},
					Type:    "string",
					Nullable: func(s string) bool {
						return s == ""
					},
					NullValue:    "NULL",
					DefaultValue: &strStr1,
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "", "", ""},
						wantValue: &Value{Val: "\"str1\"", IsNull: false},
					},
				},
			},
			{
				name: "indices defaultValue string empty",
				c: Config{
					Indices: []int{1, 2, 3},
					Type:    "string",
					Nullable: func(s string) bool {
						return s == "_NULL_"
					},
					NullValue:    "NULL",
					DefaultValue: &strEmpty,
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "_NULL_", "_NULL_", "_NULL_"},
						wantValue: &Value{Val: "\"\"", IsNull: false},
					},
				},
			},
			{
				name: "indices defaultValue int",
				c: Config{
					Indices: []int{1, 2, 3},
					Type:    "int",
					Nullable: func(s string) bool {
						return s == ""
					},
					NullValue:    "NULL",
					DefaultValue: &strInt1,
				},
				cases: []recordCase{
					{
						record:    []string{"0", "", "", ""},
						wantValue: &Value{Val: "1", IsNull: false},
					},
				},
			},
			{
				name: "indices Function string",
				c: Config{
					Indices:  []int{1, 2, 3},
					Type:     "string",
					Function: &strFunHash,
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "str1"},
						wantValue: &Value{Val: "hash(\"str1\")", IsNull: false},
					},
				},
			},
			{
				name: "indices Function int",
				c: Config{
					Indices:  []int{1, 2, 3},
					Type:     "int",
					Function: &strFunHash,
				},
				cases: []recordCase{
					{
						record:    []string{"0", "1"},
						wantValue: &Value{Val: "hash(\"1\")", IsNull: false},
					},
				},
			},
			{
				name: "indices Function Nullable",
				c: Config{
					Indices: []int{1, 2, 3},
					Type:    "string",
					Nullable: func(s string) bool {
						return s == ""
					},
					NullValue: "NULL",
					Function:  &strFunHash,
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "", "", ""},
						wantValue: &Value{Val: "NULL", IsNull: true},
					},
				},
			},
			{
				name: "indices Function defaultValue",
				c: Config{
					Indices: []int{1, 2, 3},
					Type:    "string",
					Nullable: func(s string) bool {
						return s == ""
					},
					NullValue:    "NULL",
					DefaultValue: &strStr1,
					Function:     &strFunHash,
				},
				cases: []recordCase{
					{
						record:    []string{"str0", "", "", ""},
						wantValue: &Value{Val: "hash(\"str1\")", IsNull: false},
					},
				},
			},
			{
				name: "concat items",
				c: Config{
					ConcatItems: []any{"c1", 4, 5, "c2", 6, "c3"},
					Indices:     []int{1, 2, 3},
					Type:        "string",
					Nullable: func(s string) bool {
						return s == ""
					},
					NullValue:    "NULL",
					DefaultValue: &strStr1,
				},
				cases: []recordCase{
					{
						record:  []string{"str0", "str1", "str2", "str3", "str4", "str5"},
						wantErr: errors.ErrNoRecord,
					},
					{
						record:    []string{"str0", "str1", "str2", "str3", "str4", "str5", "str6"},
						wantValue: &Value{Val: "\"c1str4str5c2str6c3\"", IsNull: false},
					},
					{
						record:    []string{"", "", "", "", "", "", ""},
						wantValue: &Value{Val: "\"c1c2c3\"", IsNull: false},
					},
					{
						record:    []string{"", "", "", "", "str4", "", ""},
						wantValue: &Value{Val: "\"c1str4c2c3\"", IsNull: false},
					},
				},
			},
			{
				name: "concat items Function",
				c: Config{
					ConcatItems: []any{"c1", 4, 5, "c2", 6, "c3"},
					Indices:     []int{1, 2, 3},
					Type:        "string",
					Nullable: func(s string) bool {
						return s == ""
					},
					NullValue:    "NULL",
					DefaultValue: &strStr1,
					Function:     &strFunHash,
				},
				cases: []recordCase{
					{
						record:  []string{"str0", "str1", "str2", "str3", "str4", "str5"},
						wantErr: errors.ErrNoRecord,
					},
					{
						record:    []string{"str0", "str1", "str2", "str3", "str4", "str5", "str6"},
						wantValue: &Value{Val: "hash(\"c1str4str5c2str6c3\")", IsNull: false},
					},
					{
						record:    []string{"", "", "", "", "", "", ""},
						wantValue: &Value{Val: "hash(\"c1c2c3\")", IsNull: false},
					},
					{
						record:    []string{"", "", "", "", "str4", "", ""},
						wantValue: &Value{Val: "hash(\"c1str4c2c3\")", IsNull: false},
					},
				},
			},
			{
				name: "check",
				c: Config{
					Indices: []int{1},
					Type:    "string",
					CheckOnPost: func(value *Value) error {
						return nil
					},
				},
				cases: []recordCase{
					{
						record:    []string{"0", "str1", "str2"},
						wantValue: &Value{Val: "\"str1\"", IsNull: false},
					},
				},
			},
			{
				name: "check failed",
				c: Config{
					Indices: []int{1},
					Type:    "string",
					CheckOnPost: func(value *Value) error {
						return fmt.Errorf("check failed")
					},
				},
				cases: []recordCase{
					{
						record:        []string{"0", "str1", "str2"},
						wantErrString: "check failed",
					},
				},
			},
		}

		for _, tc := range testcases {
			It(tc.name, func() {
				p, err := tc.c.Build()
				if tc.buildErr != nil {
					Expect(err).To(HaveOccurred())
					Expect(stderrors.Is(err, tc.buildErr)).To(BeTrue())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}

				for i, c := range tc.cases {
					v, err := p.Pick(c.record)
					if c.wantErr == nil && c.wantErrString == "" {
						Expect(err).NotTo(HaveOccurred(), "%d %v", i, c.record)
						// isSetNull must equal to IsNull
						c.wantValue.isSetNull = c.wantValue.IsNull
						Expect(c.wantValue).To(Equal(v), "%d %v", i, c.record)
					} else {
						Expect(err).To(HaveOccurred(), "%d %v", i, c.record)
						if c.wantErr != nil {
							Expect(stderrors.Is(err, c.wantErr)).To(BeTrue(), "%d %v", i, c.record)
						}
						if c.wantErrString != "" {
							Expect(err.Error()).To(ContainSubstring(c.wantErrString), "%d %v", i, c.record)
						}
						Expect(v).To(BeNil())
					}
				}
			})
		}
	})
})
