package logger

import (
	"encoding/json"
	stderrors "errors"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ = Describe("zapLogger", func() {
	Describe("logs", func() {
		var (
			tmpdir string
			file1  string
			file2  string
			l      Logger
		)
		BeforeEach(func() {
			var err error
			tmpdir, err = os.MkdirTemp("", "test")
			Expect(err).NotTo(HaveOccurred())
			file1 = filepath.Join(tmpdir, "1.log")
			file2 = filepath.Join(tmpdir, "2.log")
			var zl *zapLogger
			zl, err = newZapLogger(&options{
				level:   InfoLevel,
				fields:  Fields{{Key: "key1", Value: "value1"}, {Key: "key2", Value: "value2"}},
				console: false,
				files:   []string{file1, file2},
			})
			l = zl
			// Set fatal hook to prevent exit
			zl.l = zl.l.WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))
			Expect(err).NotTo(HaveOccurred())
			Expect(l).NotTo(BeNil())
		})
		AfterEach(func() {
			var err error
			Expect(l).NotTo(BeNil())
			err = l.Close()
			Expect(err).NotTo(HaveOccurred())
			err = os.RemoveAll(tmpdir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("debug", func() {
			var (
				err                error
				content1, content2 []byte
			)
			l.Debug("debug message")
			err = l.Sync()
			Expect(err).NotTo(HaveOccurred())
			content1, err = os.ReadFile(file1)
			Expect(err).NotTo(HaveOccurred())
			Expect(content1).To(BeEmpty())
			content2, err = os.ReadFile(file2)
			Expect(err).NotTo(HaveOccurred())
			Expect(content2).To(BeEmpty())
		})

		It("info", func() {
			var (
				err                error
				content1, content2 []byte
			)
			l.Info("info message")
			err = l.Sync()
			Expect(err).NotTo(HaveOccurred())
			content1, err = os.ReadFile(file1)
			Expect(err).NotTo(HaveOccurred())
			content2, err = os.ReadFile(file2)
			Expect(err).NotTo(HaveOccurred())
			Expect(content1).To(Equal(content2))
			m := map[string]any{}
			err = json.Unmarshal(content1, &m)
			Expect(err).NotTo(HaveOccurred())
			Expect(m).To(HaveKeyWithValue("caller", ContainSubstring("logger/zap_test.go")))
			Expect(m).To(HaveKeyWithValue("msg", "info message"))
			Expect(m).To(HaveKeyWithValue("level", "info"))
			Expect(m).To(HaveKeyWithValue("key1", "value1"))
			Expect(m).To(HaveKeyWithValue("key2", "value2"))
		})

		It("warn", func() {
			var (
				err                error
				content1, content2 []byte
			)
			l.Warn("warn message")
			err = l.Sync()
			Expect(err).NotTo(HaveOccurred())
			content1, err = os.ReadFile(file1)
			Expect(err).NotTo(HaveOccurred())
			content2, err = os.ReadFile(file2)
			Expect(err).NotTo(HaveOccurred())
			Expect(content1).To(Equal(content2))
			m := map[string]any{}
			err = json.Unmarshal(content1, &m)
			Expect(err).NotTo(HaveOccurred())
			Expect(m).To(HaveKeyWithValue("caller", ContainSubstring("logger/zap_test.go")))
			Expect(m).To(HaveKeyWithValue("msg", "warn message"))
			Expect(m).To(HaveKeyWithValue("level", "warn"))
			Expect(m).To(HaveKeyWithValue("key1", "value1"))
			Expect(m).To(HaveKeyWithValue("key2", "value2"))
		})

		It("error", func() {
			var (
				err                error
				content1, content2 []byte
			)
			l.Error("error message")
			err = l.Sync()
			Expect(err).NotTo(HaveOccurred())
			content1, err = os.ReadFile(file1)
			Expect(err).NotTo(HaveOccurred())
			content2, err = os.ReadFile(file2)
			Expect(err).NotTo(HaveOccurred())
			Expect(content1).To(Equal(content2))
			m := map[string]any{}
			err = json.Unmarshal(content1, &m)
			Expect(err).NotTo(HaveOccurred())
			Expect(m).To(HaveKeyWithValue("caller", ContainSubstring("logger/zap_test.go")))
			Expect(m).To(HaveKeyWithValue("msg", "error message"))
			Expect(m).To(HaveKeyWithValue("level", "error"))
			Expect(m).To(HaveKeyWithValue("key1", "value1"))
			Expect(m).To(HaveKeyWithValue("key2", "value2"))
		})

		It("panic", func() {
			var (
				err                error
				content1, content2 []byte
			)
			done := make(chan struct{})
			go func() {
				defer func() {
					r := recover()
					Expect(r).NotTo(BeNil())
					done <- struct{}{}
				}()
				l.Panic("panic message")
			}()
			<-done
			err = l.Sync()
			Expect(err).NotTo(HaveOccurred())
			content1, err = os.ReadFile(file1)
			Expect(err).NotTo(HaveOccurred())
			content2, err = os.ReadFile(file2)
			Expect(err).NotTo(HaveOccurred())
			Expect(content1).To(Equal(content2))
			m := map[string]any{}
			err = json.Unmarshal(content1, &m)
			Expect(err).NotTo(HaveOccurred())
			Expect(m).To(HaveKeyWithValue("caller", ContainSubstring("logger/zap_test.go")))
			Expect(m).To(HaveKeyWithValue("msg", "panic message"))
			Expect(m).To(HaveKeyWithValue("level", "panic"))
			Expect(m).To(HaveKeyWithValue("key1", "value1"))
			Expect(m).To(HaveKeyWithValue("key2", "value2"))
		})

		It("fatal", func() {
			var (
				err                error
				content1, content2 []byte
			)
			done := make(chan struct{})
			go func() {
				defer func() {
					r := recover()
					Expect(r).NotTo(BeNil())
					done <- struct{}{}
				}()
				l.Fatal("fatal message")
			}()
			<-done
			err = l.Sync()
			Expect(err).NotTo(HaveOccurred())
			content1, err = os.ReadFile(file1)
			Expect(err).NotTo(HaveOccurred())
			content2, err = os.ReadFile(file2)
			Expect(err).NotTo(HaveOccurred())
			Expect(content1).To(Equal(content2))
			m := map[string]any{}
			err = json.Unmarshal(content1, &m)
			Expect(err).NotTo(HaveOccurred())
			Expect(m).To(HaveKeyWithValue("caller", ContainSubstring("logger/zap_test.go")))
			Expect(m).To(HaveKeyWithValue("msg", "fatal message"))
			Expect(m).To(HaveKeyWithValue("level", "fatal"))
			Expect(m).To(HaveKeyWithValue("key1", "value1"))
			Expect(m).To(HaveKeyWithValue("key2", "value2"))
		})

		It("SkipCaller", func() {
			var (
				err                error
				content1, content2 []byte
			)
			l.SkipCaller(-1).Info("info message")
			err = l.Sync()
			Expect(err).NotTo(HaveOccurred())
			content1, err = os.ReadFile(file1)
			Expect(err).NotTo(HaveOccurred())
			content2, err = os.ReadFile(file2)
			Expect(err).NotTo(HaveOccurred())
			Expect(content1).To(Equal(content2))
			m := map[string]any{}
			err = json.Unmarshal(content1, &m)
			Expect(err).NotTo(HaveOccurred())
			Expect(m).To(HaveKeyWithValue("caller", Not(ContainSubstring("logger/zap_test.go"))))
			Expect(m).To(HaveKeyWithValue("msg", "info message"))
			Expect(m).To(HaveKeyWithValue("level", "info"))
			Expect(m).To(HaveKeyWithValue("key1", "value1"))
			Expect(m).To(HaveKeyWithValue("key2", "value2"))
		})

		It("With", func() {
			var (
				err                error
				content1, content2 []byte
			)
			l = l.With(Field{Key: "key3", Value: "value3"}, Field{Key: "key4", Value: "value4"})
			l.Info("info message")
			err = l.Sync()
			Expect(err).NotTo(HaveOccurred())
			content1, err = os.ReadFile(file1)
			Expect(err).NotTo(HaveOccurred())
			content2, err = os.ReadFile(file2)
			Expect(err).NotTo(HaveOccurred())
			Expect(content1).To(Equal(content2))
			m := map[string]any{}
			err = json.Unmarshal(content1, &m)
			Expect(err).NotTo(HaveOccurred())
			Expect(m).To(HaveKeyWithValue("caller", ContainSubstring("logger/zap_test.go")))
			Expect(m).To(HaveKeyWithValue("msg", "info message"))
			Expect(m).To(HaveKeyWithValue("level", "info"))
			Expect(m).To(HaveKeyWithValue("key1", "value1"))
			Expect(m).To(HaveKeyWithValue("key2", "value2"))
			Expect(m).To(HaveKeyWithValue("key3", "value3"))
			Expect(m).To(HaveKeyWithValue("key4", "value4"))
		})

		It("WithError", func() {
			var (
				err                error
				content1, content2 []byte
			)
			l.WithError(stderrors.New("test error")).Info("info message")
			err = l.Sync()
			Expect(err).NotTo(HaveOccurred())
			content1, err = os.ReadFile(file1)
			Expect(err).NotTo(HaveOccurred())
			content2, err = os.ReadFile(file2)
			Expect(err).NotTo(HaveOccurred())
			Expect(content1).To(Equal(content2))
			m := map[string]any{}
			err = json.Unmarshal(content1, &m)
			Expect(err).NotTo(HaveOccurred())
			Expect(m).To(HaveKeyWithValue("caller", ContainSubstring("logger/zap_test.go")))
			Expect(m).To(HaveKeyWithValue("msg", "info message"))
			Expect(m).To(HaveKeyWithValue("level", "info"))
			Expect(m).To(HaveKeyWithValue("key1", "value1"))
			Expect(m).To(HaveKeyWithValue("key2", "value2"))
			Expect(m).To(HaveKeyWithValue("error", "test error"))
		})
	})

	DescribeTable("toZapLevel",
		func(lvl Level, zapLvl zapcore.Level) {
			Expect(toZapLevel(lvl)).To(Equal(zapLvl))
		},
		EntryDescription("%[1]s"),
		Entry(nil, DebugLevel-1, zap.InfoLevel),
		Entry(nil, DebugLevel, zap.DebugLevel),
		Entry(nil, InfoLevel, zap.InfoLevel),
		Entry(nil, WarnLevel, zap.WarnLevel),
		Entry(nil, ErrorLevel, zap.ErrorLevel),
		Entry(nil, PanicLevel, zap.PanicLevel),
		Entry(nil, FatalLevel, zap.FatalLevel),
	)

	It("enable console", func() {
		var (
			l   Logger
			err error
		)
		l, err = newZapLogger(&options{
			console: true,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(l).NotTo(BeNil())
		_ = l.Close()
	})

	It("enable console", func() {
		var (
			l   Logger
			err error
		)
		l, err = newZapLogger(&options{
			timeLayout: time.RFC3339,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(l).NotTo(BeNil())
		_ = l.Close()
	})

	It("test sync failed console", func() {
		var (
			l      Logger
			tmpdir string
			err    error
		)
		tmpdir, err = os.MkdirTemp("", "test")
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			err = os.RemoveAll(tmpdir)
			Expect(err).NotTo(HaveOccurred())
		}()

		l, err = newZapLogger(&options{
			files: []string{filepath.Join(tmpdir, "1.log")},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(l).NotTo(BeNil())

		err = l.Close()
		Expect(err).NotTo(HaveOccurred())

		err = l.Sync()
		Expect(err).To(HaveOccurred())
	})

	It("open file error", func() {
		l, err := newZapLogger(&options{
			files: []string{"not-exists/1.log"},
		})
		Expect(err).To(HaveOccurred())
		Expect(l).To(BeNil())
	})
})
