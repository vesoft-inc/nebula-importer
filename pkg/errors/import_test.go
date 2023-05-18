package errors

import (
	stderrors "errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ImportError", func() {
	Describe("NewImportError", func() {
		It("nil", func() {
			importError := NewImportError(nil)
			Expect(importError).NotTo(BeNil())
			Expect(importError.Cause()).To(BeNil())
			Expect(importError.Unwrap()).To(BeNil())
			Expect(importError.Messages).To(BeNil())
			Expect(importError.fields).To(BeEmpty())
		})

		It("err", func() {
			err := stderrors.New("test error")
			importError := NewImportError(err)
			Expect(importError).NotTo(BeNil())
			Expect(importError.Cause()).To(Equal(err))
			Expect(importError.Unwrap()).To(Equal(err))
			Expect(importError.Messages).To(BeNil())
			Expect(importError.fields).To(BeEmpty())
		})

		It("err with message empty", func() {
			err := stderrors.New("test error")
			importError := NewImportError(err, "")
			Expect(importError).NotTo(BeNil())
			Expect(importError.Cause()).To(Equal(err))
			Expect(importError.Unwrap()).To(Equal(err))
			Expect(importError.Messages).To(BeNil())
			Expect(importError.fields).To(BeEmpty())
		})

		It("err with message", func() {
			err := stderrors.New("test error")
			importError := NewImportError(err, "test message")
			Expect(importError).NotTo(BeNil())
			Expect(importError.Cause()).To(Equal(err))
			Expect(importError.Unwrap()).To(Equal(err))
			Expect(importError.Messages).To(Equal([]string{"test message"}))
			Expect(importError.fields).To(BeEmpty())
		})

		It("err with message and format", func() {
			err := stderrors.New("test error")
			importError := NewImportError(err, "test message %d", 1)
			Expect(importError).NotTo(BeNil())
			Expect(importError.Cause()).To(Equal(err))
			Expect(importError.Unwrap()).To(Equal(err))
			Expect(importError.Messages).To(Equal([]string{"test message 1"}))
			Expect(importError.fields).To(BeEmpty())
		})

		It("err with message not string", func() {
			err := stderrors.New("test error")
			importError := NewImportError(err, 1)
			Expect(importError).NotTo(BeNil())
			Expect(importError.Cause()).To(Equal(err))
			Expect(importError.Unwrap()).To(Equal(err))
			Expect(importError.Messages).To(Equal([]string{"1"}))
			Expect(importError.fields).To(BeEmpty())
		})
	})

	Describe("AsImportError", func() {
		It("nil", func() {
			err, ok := AsImportError(nil)
			Expect(ok).To(BeFalse())
			Expect(err).To(BeNil())
		})

		It("no import error", func() {
			err, ok := AsImportError(stderrors.New("test error"))
			Expect(ok).To(BeFalse())
			Expect(err).To(BeNil())
		})

		It("import error", func() {
			importError, ok := AsImportError(NewImportError(nil, "test message"))
			Expect(ok).To(BeTrue())
			Expect(importError).NotTo(BeNil())
			Expect(importError.Messages).To(Equal([]string{"test message"}))
		})
	})

	Describe("AsOrNewImportError", func() {
		It("nil", func() {
			importError := AsOrNewImportError(nil)
			Expect(importError).NotTo(BeNil())
		})

		It("no import error", func() {
			importError := AsOrNewImportError(stderrors.New("test error"))
			Expect(importError).NotTo(BeNil())
		})

		It("import error", func() {
			importError := AsOrNewImportError(
				NewImportError(nil, "test message"),
				"test message %d",
				1,
			)
			Expect(importError).NotTo(BeNil())
			Expect(importError.Messages).To(Equal([]string{"test message", "test message 1"}))
		})
	})

	It("Fields", func() {
		importError := AsOrNewImportError(stderrors.New("test error"))
		Expect(importError.Fields()).To(BeEmpty())

		importError.AppendMessage("")
		Expect(importError.Messages).To(BeEmpty())

		importError.SetGraphName("")
		Expect(importError.GraphName()).To(BeEmpty())

		importError.SetNodeName("")
		Expect(importError.NodeName()).To(BeEmpty())

		importError.SetEdgeName("")
		Expect(importError.EdgeName()).To(BeEmpty())

		importError.SetNodeIDName("")
		Expect(importError.NodeIDName()).To(BeEmpty())

		importError.SetPropName("")
		Expect(importError.PropName()).To(BeEmpty())

		importError.SetRecord(nil)
		Expect(importError.Record()).To(BeEmpty())

		importError.SetStatement("")
		Expect(importError.Statement()).To(BeEmpty())

		Expect(importError.Fields()).To(BeEmpty())

		importError.AppendMessage("test message")
		importError.AppendMessage("test message %d", 1)
		Expect(importError.Messages).To(Equal([]string{"test message", "test message 1"}))

		importError.SetGraphName("graphName")
		Expect(importError.GraphName()).To(Equal("graphName"))

		importError.SetNodeName("nodeName")
		Expect(importError.NodeName()).To(Equal("nodeName"))

		importError.SetEdgeName("edgeName")
		Expect(importError.EdgeName()).To(Equal("edgeName"))

		importError.SetNodeIDName("nodeIDName")
		Expect(importError.NodeIDName()).To(Equal("nodeIDName"))

		importError.SetPropName("propName")
		Expect(importError.PropName()).To(Equal("propName"))

		importError.SetRecord([]string{"record1", "record2"})
		Expect(importError.Record()).To(Equal([]string{"record1", "record2"}))

		importError.SetStatement("test statement")
		Expect(importError.Statement()).To(Equal("test statement"))

		Expect(importError.Fields()).To(Equal(map[string]any{
			"messages":  []string{"test message", "test message 1"},
			"graph":     "graphName",
			"node":      "nodeName",
			"edge":      "edgeName",
			"nodeID":    "nodeIDName",
			"prop":      "propName",
			"record":    []string{"record1", "record2"},
			"statement": "test statement",
		}))
		Expect(importError.Error()).To(Equal("graph(graphName): node(nodeName): edge(edgeName): nodeID(nodeIDName): prop(propName): record([record1 record2]): statement(test statement): messagestest message, test message 1: test error"))
	})

	It("withField", func() {
		importError := AsOrNewImportError(stderrors.New("test error"))
		importError.withField("f1", "str")
		Expect(importError.getFieldStringSlice("f1")).To(BeEmpty())
		importError.withField("f2", []string{"str"})
		Expect(importError.getFieldString("f2")).To(BeEmpty())
	})
})
