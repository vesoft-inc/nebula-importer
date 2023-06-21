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

var _ = Describe("gcsSource", func() {
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
			GCS: &GCSConfig{
				Endpoint: httpServer.URL,
				Bucket:   "bucket",
				Key:      "key",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&gcsSource{}))

		Expect(s.Name()).To(Equal(fmt.Sprintf("gcs %s bucket/key", httpServer.URL)))

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

	It("GetObject failed", func() {
		httpMux.HandleFunc("/bucket/key", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusMethodNotAllowed)
		})
		c := Config{
			GCS: &GCSConfig{
				Endpoint:        httpServer.URL,
				CredentialsJSON: `{"foo":"bar"}`,
				Bucket:          "bucket",
				Key:             "key",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&gcsSource{}))

		err = s.Open()
		Expect(err).To(HaveOccurred())
	})
})
