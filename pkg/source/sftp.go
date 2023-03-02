package source

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var _ Source = (*sftpSource)(nil)

type (
	SFTPConfig struct {
		Host       string `yaml:"host,omitempty"`
		Port       int    `yaml:"port,omitempty"`
		User       string `yaml:"user,omitempty"`
		Password   string `yaml:"password,omitempty"`
		KeyFile    string `yaml:"keyFile,omitempty"`
		KeyData    string `yaml:"keyData,omitempty"`
		Passphrase string `yaml:"passphrase,omitempty"`
		Path       string `yaml:"path,omitempty"`
	}

	sftpSource struct {
		c       *Config
		sshCli  *ssh.Client
		sftpCli *sftp.Client
		f       *sftp.File
	}
)

func newSFTPSource(c *Config) Source {
	return &sftpSource{
		c: c,
	}
}

func (s *sftpSource) Name() string {
	return s.c.SFTP.String()
}

func (s *sftpSource) Open() error {
	keyData := s.c.SFTP.KeyData
	if keyData == "" && s.c.SFTP.KeyFile != "" {
		keyDataBytes, err := os.ReadFile(s.c.SFTP.KeyFile)
		if err != nil {
			return err
		}
		keyData = string(keyDataBytes)
	}

	authMethod, err := getSSHAuthMethod(s.c.SFTP.Password, keyData, s.c.SFTP.Passphrase)
	if err != nil {
		return err
	}

	sshCli, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", s.c.SFTP.Host, s.c.SFTP.Port), &ssh.ClientConfig{
		User:            s.c.SFTP.User,
		Auth:            []ssh.AuthMethod{authMethod},
		Timeout:         time.Second * 5,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint: gosec
	})
	if err != nil {
		return err
	}

	sftpCli, err := sftp.NewClient(sshCli)
	if err != nil {
		_ = sshCli.Close()
		return err
	}

	f, err := sftpCli.Open(s.c.SFTP.Path)
	if err != nil {
		_ = sftpCli.Close()
		_ = sshCli.Close()
		return err
	}

	s.sshCli = sshCli
	s.sftpCli = sftpCli
	s.f = f

	return nil
}

func (s *sftpSource) Config() *Config {
	return s.c
}

func (s *sftpSource) Size() (int64, error) {
	fi, err := s.f.Stat()
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

func (s *sftpSource) Read(p []byte) (int, error) {
	return s.f.Read(p)
}

func (s *sftpSource) Close() error {
	defer func() {
		_ = s.sftpCli.Close()
		_ = s.sshCli.Close()
	}()
	return s.f.Close()
}

func getSSHAuthMethod(password, keyData, passphrase string) (ssh.AuthMethod, error) {
	if keyData != "" {
		key, err := getSSHSigner(keyData, passphrase)
		if err != nil {
			return nil, err
		}
		return ssh.PublicKeys(key), nil
	}
	return ssh.Password(password), nil
}

func getSSHSigner(keyData, passphrase string) (ssh.Signer, error) {
	if passphrase != "" {
		return ssh.ParsePrivateKeyWithPassphrase([]byte(keyData), []byte(passphrase))
	}
	return ssh.ParsePrivateKey([]byte(keyData))
}

func (c *SFTPConfig) String() string {
	return fmt.Sprintf("sftp %s:%d %s", c.Host, c.Port, c.Path)
}
