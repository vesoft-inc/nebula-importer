//go:build linux

package client

import (
	stderrors "errors"

	nebula "github.com/vesoft-inc/nebula-go/v3"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/logger"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SessionV3", func() {
	It("success", func() {
		session := newSessionV3(HostAddress{}, "user", "password", "", nil, nil)
		pool := &nebula.ConnectionPool{}
		nSession := &nebula.Session{}

		patches := gomonkey.NewPatches()
		defer patches.Reset()

		patches.ApplyFuncReturn(nebula.NewSslConnectionPool, pool, nil)
		patches.ApplyMethodReturn(pool, "GetSession", nSession, nil)

		patches.ApplyMethodReturn(nSession, "Execute", &nebula.ResultSet{}, nil)
		patches.ApplyMethodReturn(nSession, "Release")

		err := session.Open()
		Expect(err).NotTo(HaveOccurred())
		resp, err := session.Execute("")
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).NotTo(BeNil())

		err = session.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("failed", func() {
		session := newSessionV3(HostAddress{}, "user", "password", "", nil, logger.NopLogger)
		pool := &nebula.ConnectionPool{}
		nSession := &nebula.Session{}

		patches := gomonkey.NewPatches()
		defer patches.Reset()

		patches.ApplyFuncReturn(nebula.NewSslConnectionPool, nil, stderrors.New("new connection pool failed"))

		err := session.Open()
		Expect(err).To(HaveOccurred())

		patches.Reset()

		patches.ApplyFuncReturn(nebula.NewSslConnectionPool, pool, nil)
		patches.ApplyMethodReturn(pool, "GetSession", nil, stderrors.New("get session failed"))

		err = session.Open()
		Expect(err).To(HaveOccurred())

		patches.Reset()

		patches.ApplyFuncReturn(nebula.NewSslConnectionPool, pool, nil)
		patches.ApplyMethodReturn(pool, "GetSession", nSession, nil)

		patches.ApplyMethodReturn(nSession, "Execute", nil, stderrors.New("execute failed"))
		patches.ApplyMethodReturn(nSession, "Release")

		err = session.Open()
		Expect(err).NotTo(HaveOccurred())
		resp, err := session.Execute("")
		Expect(err).To(HaveOccurred())
		Expect(resp).To(BeNil())

		err = session.Close()
		Expect(err).NotTo(HaveOccurred())
	})
})
