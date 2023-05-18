package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("path", func() {
	DescribeTable("RelativePathBaseOn",
		func(basePath, filePath, ExpectFilePath string) {
			Expect(RelativePathBaseOn(basePath, filePath)).To(Equal(ExpectFilePath))
		},
		EntryDescription("RelativePathBaseOn(%[1]q, %[2]q) == %[3]q"),
		Entry(nil, ".", "f", "f"),
		Entry(nil, "./d1", "f", "d1/f"),
		Entry(nil, "./d1/", "f", "d1/f"),
		Entry(nil, "./d1/d2", "f", "d1/d2/f"),

		Entry(nil, "/", "f", "/f"),
		Entry(nil, "/d1", "f", "/d1/f"),
		Entry(nil, "/d1/", "f", "/d1/f"),
		Entry(nil, "/d1/d2", "f", "/d1/d2/f"),

		Entry(nil, "/", "d3/f", "/d3/f"),
		Entry(nil, "/d1", "d3/f", "/d1/d3/f"),
		Entry(nil, "/d1/", "d3/f", "/d1/d3/f"),
		Entry(nil, "/d1/d2", "d3/f", "/d1/d2/d3/f"),

		Entry(nil, "/", "./d3/f", "/d3/f"),
		Entry(nil, "/d1", "./d3/f", "/d1/d3/f"),
		Entry(nil, "/d1/", "./d3/f", "/d1/d3/f"),
		Entry(nil, "/d1/d2", "./d3/f", "/d1/d2/d3/f"),

		Entry(nil, "/", "../f", "/f"),
		Entry(nil, "/d1", "../f", "/f"),
		Entry(nil, "/d1/", "../f", "/f"),
		Entry(nil, "/d1/d2", "../f", "/d1/f"),

		Entry(nil, "/", "../d3/f", "/d3/f"),
		Entry(nil, "/d1", "../d3/f", "/d3/f"),
		Entry(nil, "/d1/", "../d3/f", "/d3/f"),
		Entry(nil, "/d1/d2", "../d3/f", "/d1/d3/f"),

		Entry(nil, "/", "/f", "/f"),
		Entry(nil, "/d1", "/f", "/f"),
		Entry(nil, "/d1/", "/f", "/f"),
		Entry(nil, "/d1/d2", "/f", "/f"),

		Entry(nil, "/", "/d3/f", "/d3/f"),
		Entry(nil, "/d1", "/d3/f", "/d3/f"),
		Entry(nil, "/d1/", "/d3/f", "/d3/f"),
		Entry(nil, "/d1/d2", "/d3/f", "/d3/f"),
	)
})
