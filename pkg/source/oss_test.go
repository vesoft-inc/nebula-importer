package source

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ossSource", func() {
	var (
		httpMux    *http.ServeMux
		httpServer *httptest.Server
	)
	BeforeEach(func() {
		httpMux = http.NewServeMux()
		httpServer = httptest.NewServer(httpMux)
	})
	AfterEach(func() {
		httpServer.Close()
	})
	It("successfully", func() {
		content := []byte("Hello")
		httpMux.HandleFunc("/bucket/key", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				_, _ = w.Write(content)
			case http.MethodHead:
				w.Header().Set("Content-Length", strconv.Itoa(len(content)))
			default:
				Panic()
			}
		})

		c := Config{
			OSS: &OSSConfig{
				Endpoint:  httpServer.URL,
				AccessKey: "accessKey",
				SecretKey: "secretKey",
				Bucket:    "bucket",
				Key:       "key",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&ossSource{}))

		Expect(s.Name()).To(Equal(fmt.Sprintf("oss %s bucket/key", httpServer.URL)))

		Expect(s.Config()).NotTo(BeNil())

		err = s.Open()
		Expect(err).NotTo(HaveOccurred())

		sz, err := s.Size()
		Expect(err).NotTo(HaveOccurred())
		Expect(sz).To(Equal(int64(len(content))))

		var p [32]byte
		n, err := s.Read(p[:])
		Expect(err).To(Equal(io.EOF))
		Expect(n).To(Equal(len(content)))
		Expect(p[:n]).To(Equal(content))

		for i := 0; i < 2; i++ {
			n, err = s.Read(p[:])
			Expect(err).To(Equal(io.EOF))
			Expect(n).To(Equal(0))
		}

		err = s.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("oss.New failed", func() {
		c := Config{
			OSS: &OSSConfig{
				Endpoint:  "\t",
				AccessKey: "accessKey",
				SecretKey: "secretKey",
				Bucket:    "bucket",
				Key:       "key",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&ossSource{}))

		err = s.Open()
		Expect(err).To(HaveOccurred())
	})

	It("Bucket failed", func() {
		c := Config{
			OSS: &OSSConfig{
				Endpoint:  httpServer.URL,
				AccessKey: "accessKey",
				SecretKey: "secretKey",
				Bucket:    "b",
				Key:       "key",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&ossSource{}))

		err = s.Open()
		Expect(err).To(HaveOccurred())
	})

	It("GetObject failed", func() {
		httpMux.HandleFunc("/bucket/key", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusMethodNotAllowed)
		})
		c := Config{
			OSS: &OSSConfig{
				Endpoint:  httpServer.URL,
				AccessKey: "accessKey",
				SecretKey: "secretKey",
				Bucket:    "bucket",
				Key:       "key",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&ossSource{}))

		err = s.Open()
		Expect(err).To(HaveOccurred())
	})

	It("Size failed", func() {
		content := []byte("Hello")
		httpMux.HandleFunc("/bucket/key", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				_, _ = w.Write(content)
			case http.MethodHead:
				w.WriteHeader(http.StatusMethodNotAllowed)
			default:
				Panic()
			}
		})
		c := Config{
			OSS: &OSSConfig{
				Endpoint:  httpServer.URL,
				AccessKey: "accessKey",
				SecretKey: "secretKey",
				Bucket:    "bucket",
				Key:       "key",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&ossSource{}))

		err = s.Open()
		Expect(err).NotTo(HaveOccurred())

		sz, err := s.Size()
		Expect(err).To(HaveOccurred())
		Expect(sz).To(Equal(int64(0)))
	})

	It("Size failed", func() {
		content := []byte("Hello")
		httpMux.HandleFunc("/bucket/key", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				_, _ = w.Write(content)
			case http.MethodHead:
				w.WriteHeader(http.StatusOK)
			default:
				Panic()
			}
		})
		c := Config{
			OSS: &OSSConfig{
				Endpoint:  httpServer.URL,
				AccessKey: "accessKey",
				SecretKey: "secretKey",
				Bucket:    "bucket",
				Key:       "key",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&ossSource{}))

		err = s.Open()
		Expect(err).NotTo(HaveOccurred())

		sz, err := s.Size()
		Expect(err).To(HaveOccurred())
		Expect(sz).To(Equal(int64(0)))
	})
})
