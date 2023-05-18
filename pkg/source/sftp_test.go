package source

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	stderrors "errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var _ = Describe("sftpSource", func() {
	var (
		tmpdir            string
		host              = "127.0.0.1"
		port              = 0
		user              = "user"
		password          = "password"
		keyFile           = ""
		keyData           = ""
		keyFilePassphrase = ""
		keyDataPassphrase = ""
		passphrase        = "ssh passphrase"
		sftpServer        *TestSFTPServer
	)
	BeforeEach(func() {
		var err error
		tmpdir, err = os.MkdirTemp("", "test")
		Expect(err).NotTo(HaveOccurred())

		// Generate a new RSA private key
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		Expect(err).NotTo(HaveOccurred())
		privateKeyPEM := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}

		keyData = string(pem.EncodeToMemory(privateKeyPEM))
		keyFile = filepath.Join(tmpdir, "id_rsa")
		signer, err := getSSHSigner(keyData, "")
		Expect(err).NotTo(HaveOccurred())
		err = os.WriteFile(keyFile, []byte(keyData), 0600)
		Expect(err).NotTo(HaveOccurred())

		encryptedPEM, err := x509.EncryptPEMBlock(rand.Reader, privateKeyPEM.Type, privateKeyPEM.Bytes, []byte(passphrase), x509.PEMCipherAES256) //nolint:staticcheck
		Expect(err).NotTo(HaveOccurred())
		keyDataPassphrase = string(pem.EncodeToMemory(encryptedPEM))
		keyFilePassphrase = filepath.Join(tmpdir, "id_rsa_passphrase")
		signerPassphrase, err := getSSHSigner(keyDataPassphrase, passphrase)
		Expect(err).NotTo(HaveOccurred())
		err = os.WriteFile(keyFilePassphrase, []byte(keyDataPassphrase), 0600)
		Expect(err).NotTo(HaveOccurred())

		sftpServer = &TestSFTPServer{
			ListenAddress: fmt.Sprintf("%s:%d", host, port),
			User:          user,
			Password:      password,
			PrivateKeys:   []ssh.Signer{signer, signerPassphrase},
		}

		err = sftpServer.Start()
		Expect(err).NotTo(HaveOccurred())

		_, portStr, err := net.SplitHostPort(sftpServer.Addr())
		Expect(err).NotTo(HaveOccurred())
		port, _ = strconv.Atoi(portStr)
	})
	AfterEach(func() {
		var err error
		sftpServer.Stop()
		err = os.RemoveAll(tmpdir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("successfully password", func() {
		content := []byte("Hello")
		file := filepath.Join(tmpdir, "file")
		err := os.WriteFile(file, content, 0600)
		Expect(err).NotTo(HaveOccurred())

		for _, c := range []Config{
			{ // password
				SFTP: &SFTPConfig{
					Host:     host,
					Port:     port,
					User:     user,
					Password: password,
					Path:     file,
				},
			},
			{ // key file
				SFTP: &SFTPConfig{
					Host:    host,
					Port:    port,
					User:    user,
					KeyFile: keyFile,
					Path:    file,
				},
			},
			{ // key file with passphrase
				SFTP: &SFTPConfig{
					Host:       host,
					Port:       port,
					User:       user,
					KeyFile:    keyFilePassphrase,
					Passphrase: passphrase,
					Path:       file,
				},
			},
		} {
			c := c
			s, err := New(&c)
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(BeAssignableToTypeOf(&sftpSource{}))

			Expect(s.Name()).To(Equal(fmt.Sprintf("sftp 127.0.0.1:%d %s", port, file)))

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
		}
	})

	It("get size failed", func() {
		content := []byte("Hello")
		file := filepath.Join(tmpdir, "file")
		err := os.WriteFile(file, content, 0600)
		Expect(err).NotTo(HaveOccurred())

		c := Config{
			SFTP: &SFTPConfig{
				Host:       host,
				Port:       port,
				User:       user,
				KeyFile:    keyFilePassphrase,
				Passphrase: passphrase,
				Path:       file,
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&sftpSource{}))

		Expect(s.Name()).To(Equal(fmt.Sprintf("sftp 127.0.0.1:%d %s", port, file)))

		Expect(s.Config()).NotTo(BeNil())

		err = s.Open()
		Expect(err).NotTo(HaveOccurred())

		sftpServer.Stop()

		sz, err := s.Size()
		Expect(err).To(HaveOccurred())
		Expect(sz).To(Equal(int64(0)))
	})

	It("read key file failed", func() {
		c := Config{
			SFTP: &SFTPConfig{
				Host:       host,
				Port:       port,
				User:       user,
				KeyFile:    keyFilePassphrase + "x",
				Passphrase: passphrase,
				Path:       "",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&sftpSource{}))

		err = s.Open()
		Expect(err).To(HaveOccurred())
	})

	It("getSSHAuthMethod failed", func() {
		c := Config{
			SFTP: &SFTPConfig{
				Host:       host,
				Port:       port,
				User:       user,
				KeyFile:    keyFilePassphrase,
				Passphrase: passphrase + "x",
				Path:       "",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&sftpSource{}))

		err = s.Open()
		Expect(err).To(HaveOccurred())
	})

	It("ssh.Dial failed", func() {
		c := Config{
			SFTP: &SFTPConfig{
				Host:       host,
				Port:       0,
				User:       user,
				KeyFile:    keyFilePassphrase,
				Passphrase: passphrase,
				Path:       "",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&sftpSource{}))

		err = s.Open()
		Expect(err).To(HaveOccurred())
	})

	It("sftp.NewClient failed", func() {
		sftpServer.DisableSubsystem = true

		c := Config{
			SFTP: &SFTPConfig{
				Host:       host,
				Port:       port,
				User:       user,
				KeyFile:    keyFilePassphrase,
				Passphrase: passphrase,
				Path:       "",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&sftpSource{}))

		err = s.Open()
		Expect(err).To(HaveOccurred())
	})

	It("Open file failed", func() {
		c := Config{
			SFTP: &SFTPConfig{
				Host:       host,
				Port:       port,
				User:       user,
				KeyFile:    keyFilePassphrase,
				Passphrase: passphrase,
				Path:       "x",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&sftpSource{}))

		err = s.Open()
		Expect(err).To(HaveOccurred())
	})
})

// The following is mock sftp server

type (
	TestSFTPServer struct {
		ListenAddress    string
		User             string
		Password         string
		PrivateKeys      []ssh.Signer
		DisableSubsystem bool

		serverConfig *ssh.ServerConfig
		listener     net.Listener
		conns        []net.Conn
	}
)

func (s *TestSFTPServer) Start() error {
	serverConfig := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == s.User && string(pass) == s.Password {
				return nil, nil
			}
			return nil, stderrors.New("bad username or password")
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if conn.User() == s.User {
				for _, privateKey := range s.PrivateKeys {
					if bytes.Equal(key.Marshal(), privateKey.PublicKey().Marshal()) {
						return nil, nil
					}
				}
			}
			return nil, fmt.Errorf("pubkey for %q not acceptable", conn.User())
		},
	}

	for _, privateKey := range s.PrivateKeys {
		serverConfig.AddHostKey(privateKey)
	}

	s.serverConfig = serverConfig

	if err := s.listen(); err != nil {
		return err
	}

	go s.acceptLoop()

	return nil
}

func (s *TestSFTPServer) Addr() string {
	return s.listener.Addr().String()
}

func (s *TestSFTPServer) Stop() {
	s.listener.Close()
	for _, conn := range s.conns {
		conn.Close()
	}
	s.conns = nil
}

func (s *TestSFTPServer) listen() error {
	listener, err := net.Listen("tcp", s.ListenAddress)
	if err != nil {
		return err
	}
	s.listener = listener
	return nil
}

func (s *TestSFTPServer) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("accept failed %v", err)
			return
		}
		s.conns = append(s.conns, conn)
		go s.handlerConn(conn)
	}
}

func (s *TestSFTPServer) handlerConn(conn net.Conn) {
	defer conn.Close()
	serverConn, chans, reqs, err := ssh.NewServerConn(conn, s.serverConfig)
	if err != nil {
		log.Printf("create ssh session conn failed %v", err)
		return
	}

	defer serverConn.Close()
	go ssh.DiscardRequests(reqs)
	for newChannel := range chans {
		go s.handlerNewChannel(newChannel)
	}
}

func (s *TestSFTPServer) handlerNewChannel(newChannel ssh.NewChannel) {
	if newChannel.ChannelType() != "session" {
		newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
		log.Printf("unknown channel type: %s", newChannel.ChannelType())
		return
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("accept channel %v", err)
		return
	}
	defer channel.Close()

	go func(in <-chan *ssh.Request) {
		for req := range in {
			ok := false
			switch req.Type { //nolint:gocritic
			// Here we handle only the "subsystem" request.
			case "subsystem":
				if !s.DisableSubsystem && string(req.Payload[4:]) == "sftp" {
					ok = true
				}
			}
			req.Reply(ok, nil)
		}
	}(requests)

	server, err := sftp.NewServer(
		channel,
		sftp.ReadOnly(),
	)
	if err != nil {
		log.Printf("create sftp server failed %v", err)
		return
	}
	defer server.Close()

	if err = server.Serve(); err != io.EOF {
		log.Printf("sftp server failed %v", err)
	}
}
