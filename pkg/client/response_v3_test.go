//go:build linux

package client

import (
	"time"

	nebula "github.com/vesoft-inc/nebula-go/v3"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("defaultResponseV3", func() {
	It("newResponseV3", func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		rs := nebula.ResultSet{}
		resp := newResponseV3(&rs, time.Second)

		patches.ApplyMethodReturn(rs, "GetErrorCode", nebula.ErrorCode_SUCCEEDED)
		patches.ApplyMethodReturn(rs, "GetErrorMsg", "")

		err := resp.GetError()
		Expect(err).NotTo(HaveOccurred())

		patches.Reset()

		patches.ApplyMethodReturn(rs, "GetLatency", int64(1))
		patches.ApplyMethodReturn(rs, "GetErrorCode", nebula.ErrorCode_E_DISCONNECTED)
		patches.ApplyMethodReturn(rs, "GetErrorMsg", "test msg")

		err = resp.GetError()
		Expect(err).To(HaveOccurred())

		Expect(resp.GetLatency()).To(Equal(time.Microsecond))
		Expect(resp.GetRespTime()).To(Equal(time.Second))
		Expect(resp.IsPermanentError()).To(BeFalse())
		Expect(resp.IsRetryMoreError()).To(BeFalse())
	})

	DescribeTable("IsPermanentError",
		func(errorCode nebula.ErrorCode, isPermanentError bool) {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			rs := nebula.ResultSet{}
			resp := newResponseV3(&rs, time.Second)

			patches.ApplyMethodReturn(rs, "GetErrorCode", errorCode)

			Expect(resp.IsPermanentError()).To(Equal(isPermanentError))
		},
		EntryDescription("%[1]s -> %[2]t"),
		Entry(nil, nebula.ErrorCode_E_SYNTAX_ERROR, true),
		Entry(nil, nebula.ErrorCode_E_SEMANTIC_ERROR, true),
		Entry(nil, nebula.ErrorCode_E_DISCONNECTED, false),
	)

	DescribeTable("IsPermanentError",
		func(errorMsg string, isPermanentError bool) {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			rs := nebula.ResultSet{}
			resp := newResponseV3(&rs, time.Second)

			patches.ApplyMethodReturn(rs, "GetErrorMsg", errorMsg)

			Expect(resp.IsRetryMoreError()).To(Equal(isPermanentError))
		},
		EntryDescription("%[1]s -> %[2]t"),
		Entry(nil, "x raft buffer is full x", true),
		Entry(nil, "x x", false),
	)
})
