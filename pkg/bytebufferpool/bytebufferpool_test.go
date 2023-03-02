package bytebufferpool

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ByteBuffer", func() {
	It("", func() {
		buff := Get()
		defer Put(buff)

		Expect(buff).NotTo(BeNil())
		Expect(buff.Len()).To(Equal(0))
		Expect(buff.Bytes()).To(Equal([]byte(nil)))
		Expect(buff.String()).To(Equal(""))

		buff.Write([]byte("a")) //nolint:gocritic
		Expect(buff.Len()).To(Equal(1))
		Expect(buff.Bytes()).To(Equal([]byte("a")))
		Expect(buff.String()).To(Equal("a"))

		buff.WriteString("b")
		Expect(buff.Len()).To(Equal(2))
		Expect(buff.Bytes()).To(Equal([]byte("ab")))
		Expect(buff.String()).To(Equal("ab"))

		buff.Set([]byte("c"))
		Expect(buff.Len()).To(Equal(1))
		Expect(buff.Bytes()).To(Equal([]byte("c")))
		Expect(buff.String()).To(Equal("c"))

		buff.SetString("d")
		Expect(buff.Len()).To(Equal(1))
		Expect(buff.Bytes()).To(Equal([]byte("d")))
		Expect(buff.String()).To(Equal("d"))

		buff.Reset()
		Expect(buff.Len()).To(Equal(0))
		Expect(buff.Bytes()).To(Equal([]byte(nil)))
		Expect(buff.String()).To(Equal(""))

		buff.WriteStringSlice(nil, ",")
		Expect(buff.Len()).To(Equal(0))
		Expect(buff.Bytes()).To(Equal([]byte(nil)))
		Expect(buff.String()).To(Equal(""))

		buff.WriteStringSlice([]string{}, ",")
		Expect(buff.Len()).To(Equal(0))
		Expect(buff.Bytes()).To(Equal([]byte(nil)))
		Expect(buff.String()).To(Equal(""))

		buff.WriteStringSlice([]string{"a"}, ",")
		Expect(buff.Len()).To(Equal(1))
		Expect(buff.Bytes()).To(Equal([]byte("a")))
		Expect(buff.String()).To(Equal("a"))

		buff.WriteStringSlice([]string{"b", "c", "d"}, ",")
		Expect(buff.Len()).To(Equal(6))
		Expect(buff.Bytes()).To(Equal([]byte("ab,c,d")))
		Expect(buff.String()).To(Equal("ab,c,d"))
	})
})
