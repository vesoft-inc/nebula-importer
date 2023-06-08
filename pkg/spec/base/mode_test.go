package specbase

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mode", func() {
	DescribeTable(".Convert",
		func(m, expect Mode) {
			Expect(m.Convert()).To(Equal(expect))
		},
		EntryDescription("%[1]s => %[2]v"),
		Entry(nil, Mode(""), DefaultMode),
		Entry(nil, DefaultMode, InsertMode),
		Entry(nil, InsertMode, InsertMode),
		Entry(nil, UpdateMode, UpdateMode),
		Entry(nil, DeleteMode, DeleteMode),
		Entry(nil, Mode("insert"), InsertMode),
		Entry(nil, Mode("Update"), UpdateMode),
		Entry(nil, Mode("DELETE"), DeleteMode),
	)
	DescribeTable(".Convert",
		func(m Mode, expect bool) {
			Expect(m.IsSupport()).To(Equal(expect))
		},
		EntryDescription("%[1]s => %[2]v"),
		Entry(nil, Mode(""), false),
		Entry(nil, DefaultMode, true),
		Entry(nil, InsertMode, true),
		Entry(nil, UpdateMode, true),
		Entry(nil, DeleteMode, true),
		Entry(nil, Mode("x"), false),
	)
})
