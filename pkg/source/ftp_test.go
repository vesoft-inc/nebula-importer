package source

import (
	"crypto/tls"
	stderrors "errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"

	ftpserverlib "github.com/fclairamb/ftpserverlib"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("ftpSource", func() {
	var (
		host            = "127.0.0.1"
		port            = 0
		user            = "user"
		password        = "Password"
		fs              = afero.NewMemMapFs()
		ftpServerDriver *TestFTPServerDriver
		ftpServer       *ftpserverlib.FtpServer
		wgFTPServer     sync.WaitGroup
	)
	BeforeEach(func() {
		ftpServerDriver = &TestFTPServerDriver{
			User:     user,
			Password: password,
			Settings: &ftpserverlib.Settings{
				ListenAddr: fmt.Sprintf("%s:%d", host, port),
			},
			Fs: fs,
		}
		ftpServer = ftpserverlib.NewFtpServer(ftpServerDriver)
		err := ftpServer.Listen()
		Expect(err).NotTo(HaveOccurred())

		_, portStr, err := net.SplitHostPort(ftpServer.Addr())
		Expect(err).NotTo(HaveOccurred())
		port, _ = strconv.Atoi(portStr)

		wgFTPServer.Add(1)
		go func() {
			defer wgFTPServer.Done()
			err = ftpServer.Serve()
			Expect(err).NotTo(HaveOccurred())
		}()
	})
	AfterEach(func() {
		_ = ftpServer.Stop()
		wgFTPServer.Wait()
	})
	It("successfully", func() {
		content := []byte("Hello")
		f, err := fs.Create("/file")
		Expect(err).NotTo(HaveOccurred())
		_, _ = f.Write(content)
		_ = f.Close()

		c := Config{
			FTP: &FTPConfig{
				Host:     host,
				Port:     port,
				User:     user,
				Password: password,
				Path:     "/file",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&ftpSource{}))

		Expect(s.Name()).To(Equal(fmt.Sprintf("ftp 127.0.0.1:%d /file", port)))

		Expect(s.Config()).NotTo(BeNil())

		err = s.Open()
		Expect(err).NotTo(HaveOccurred())

		sz, err := s.Size()
		Expect(err).NotTo(HaveOccurred())
		Expect(sz).To(Equal(int64(len(content))))

		var p [32]byte
		n, err := s.Read(p[:])
		Expect(err).NotTo(HaveOccurred())
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

	It("ftp.Dial failed", func() {
		c := Config{
			FTP: &FTPConfig{
				Host:     host,
				Port:     0,
				User:     user,
				Password: password,
				Path:     "/file",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&ftpSource{}))

		err = s.Open()
		Expect(err).To(HaveOccurred())
	})

	It("Login failed", func() {
		c := Config{
			FTP: &FTPConfig{
				Host:     host,
				Port:     port,
				User:     user,
				Password: password + "p",
				Path:     "/file",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&ftpSource{}))

		err = s.Open()
		Expect(err).To(HaveOccurred())
	})

	It("FileSize failed", func() {
		c := Config{
			FTP: &FTPConfig{
				Host:     host,
				Port:     port,
				User:     user,
				Password: password,
				Path:     "/file-not-exists",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&ftpSource{}))

		err = s.Open()
		Expect(err).To(HaveOccurred())
	})

	It("Retr failed", func() {
		c := Config{
			FTP: &FTPConfig{
				Host:     host,
				Port:     port,
				User:     user,
				Password: password,
				Path:     "/file",
			},
		}

		ftpServerDriver.OpenFileFunc = func(_ string, _ int, _ os.FileMode) (afero.File, error) {
			return nil, stderrors.New("test error")
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&ftpSource{}))

		err = s.Open()
		Expect(err).To(HaveOccurred())
	})
})

// The following is mock ftp server

var (
	_ ftpserverlib.MainDriver   = (*TestFTPServerDriver)(nil)
	_ ftpserverlib.ClientDriver = (*TestFTPClientDriver)(nil)
)

type (
	TestFTPServerDriver struct {
		User     string
		Password string
		Settings *ftpserverlib.Settings
		Fs       afero.Fs

		OpenFileFunc func(name string, flag int, perm os.FileMode) (afero.File, error)
	}
	TestFTPClientDriver struct {
		afero.Fs
		serverDriver *TestFTPServerDriver
	}
)

func (d *TestFTPServerDriver) GetSettings() (*ftpserverlib.Settings, error) {
	return d.Settings, nil
}

func (d *TestFTPServerDriver) ClientConnected(_ ftpserverlib.ClientContext) (string, error) {
	return "TEST Server", nil
}

func (d *TestFTPServerDriver) ClientDisconnected(cc ftpserverlib.ClientContext) {}

func (d *TestFTPServerDriver) AuthUser(_ ftpserverlib.ClientContext, user, pass string) (ftpserverlib.ClientDriver, error) {
	if user == d.User && pass == d.Password {
		return &TestFTPClientDriver{
			Fs:           d.Fs,
			serverDriver: d,
		}, nil
	}
	return nil, stderrors.New("bad username or password")
}

func (d *TestFTPServerDriver) GetTLSConfig() (*tls.Config, error) {
	return nil, stderrors.New("TLS is not configured")
}

func (d *TestFTPClientDriver) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if d.serverDriver.OpenFileFunc != nil {
		return d.serverDriver.OpenFileFunc(name, flag, perm)
	}
	return d.Fs.OpenFile(name, flag, perm)
}
