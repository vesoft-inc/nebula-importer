package source

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/colinmarc/hdfs/v2"
	"github.com/colinmarc/hdfs/v2/hadoopconf"
	krb "github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/credentials"
	"github.com/jcmturner/gokrb5/v8/keytab"
)

const defaultKrb5ConfigFile = "/etc/krb5.conf"

var (
	_ Source  = (*hdfsSource)(nil)
	_ Globber = (*hdfsSource)(nil)
)

type (
	HDFSConfig struct {
		Address                string `yaml:"address,omitempty"`
		User                   string `yaml:"user,omitempty"`
		ServicePrincipalName   string `yaml:"servicePrincipalName,omitempty"`
		Krb5ConfigFile         string `yaml:"krb5ConfigFile,omitempty"`
		CCacheFile             string `yaml:"ccacheFile,omitempty"`
		KeyTabFile             string `yaml:"keyTabFile,omitempty"`
		Password               string `yaml:"password,omitempty"`
		DataTransferProtection string `yaml:"dataTransferProtection,omitempty"`
		DisablePAFXFAST        bool   `yaml:"disablePAFXFAST,omitempty"`
		Path                   string `yaml:"path,omitempty"`
	}

	hdfsSource struct {
		c   *Config
		cli *hdfs.Client
		r   *hdfs.FileReader
	}
)

func newHDFSSource(c *Config) Source {
	return &hdfsSource{
		c: c,
	}
}

func (s *hdfsSource) Name() string {
	return s.c.HDFS.String()
}

func (s *hdfsSource) Connect() error {
	conf, err := hadoopconf.LoadFromEnvironment()
	if err != nil {
		return err
	}

	options := hdfs.ClientOptionsFromConf(conf)
	if s.c.HDFS.Address != "" {
		options.Addresses = strings.Split(s.c.HDFS.Address, ",")
	}

	if s.c.HDFS.ServicePrincipalName != "" {
		options.KerberosClient, err = s.c.HDFS.getKerberosClient()
		if err != nil {
			return err
		}

		options.KerberosServicePrincipleName = s.c.HDFS.ServicePrincipalName
		if s.c.HDFS.DataTransferProtection != "" {
			options.DataTransferProtection = s.c.HDFS.DataTransferProtection
		}
	} else {
		options.User = s.c.HDFS.User
	}

	cli, err := hdfs.NewClient(options)
	if err != nil {
		return err
	}

	s.cli = cli

	return nil
}

func (s *hdfsSource) Open() error {
	if err := s.Connect(); err != nil {
		return err
	}

	r, err := s.cli.Open(s.c.HDFS.Path)
	if err != nil {
		return err
	}

	s.r = r

	return nil
}

func (s *hdfsSource) IsDir(dir string) (isDir bool, err error) {
	if err = s.Connect(); err != nil {
		return false, err
	}

	fi, err := s.cli.Stat(dir)
	if err != nil {
		return false, err
	}

	return fi.IsDir(), nil
}

func (s *hdfsSource) Readdirnames(dir string) (names []string, err error) {
	if err = s.Connect(); err != nil {
		return nil, err
	}

	fis, err := s.cli.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	names = make([]string, 0, len(fis))
	for _, fi := range fis {
		names = append(names, fi.Name())
	}
	return names, nil
}

func (s *hdfsSource) Glob() ([]*Config, error) {
	matches, err := sourceGlob(s, s.c.HDFS.Path)
	if err != nil {
		return nil, err
	}

	cs := make([]*Config, 0, len(matches))
	for _, match := range matches {
		cpy := s.c.Clone()
		cpy.HDFS.Path = match
		cs = append(cs, cpy)
	}
	return cs, nil
}

func (s *hdfsSource) Config() *Config {
	return s.c
}

func (s *hdfsSource) Size() (int64, error) {
	return s.r.Stat().Size(), nil
}

func (s *hdfsSource) Read(p []byte) (int, error) {
	return s.r.Read(p)
}

func (s *hdfsSource) Close() (err error) {
	if s.r != nil {
		err = s.r.Close()
	}
	if s.cli != nil {
		_ = s.cli.Close()
	}
	return err
}

func (c *HDFSConfig) String() string {
	return fmt.Sprintf("hdfs %s %s", c.Address, c.Path)
}

func (c *HDFSConfig) getKerberosClient() (*krb.Client, error) {
	krb5ConfigFile := c.Krb5ConfigFile
	if krb5ConfigFile == "" {
		krb5ConfigFile = os.Getenv("KRB5_CONFIG")
	}
	if krb5ConfigFile == "" {
		krb5ConfigFile = defaultKrb5ConfigFile
	}
	krb5conf, err := config.Load(krb5ConfigFile)
	if err != nil {
		return nil, err
	}

	settings := []func(*krb.Settings){
		krb.DisablePAFXFAST(c.DisablePAFXFAST),
	}

	var krb5client *krb.Client
	var needLogin = true
	if c.Password != "" {
		krb5client = krb.NewWithPassword(c.User, krb5conf.LibDefaults.DefaultRealm, c.Password, krb5conf, settings...)
	} else if c.KeyTabFile != "" {
		var kt *keytab.Keytab
		if kt, err = keytab.Load(c.KeyTabFile); err != nil {
			return nil, err
		}
		krb5client = krb.NewWithKeytab(c.User, krb5conf.LibDefaults.DefaultRealm, kt, krb5conf, settings...)
	} else {
		ccacheFile := c.CCacheFile
		if ccacheFile == "" {
			ccacheFile = os.Getenv("KRB5CCNAME")
			if strings.Contains(ccacheFile, ":") {
				if strings.HasPrefix(ccacheFile, "FILE:") {
					ccacheFile = strings.SplitN(ccacheFile, ":", 2)[1]
				}
			}
		}

		if ccacheFile == "" {
			var u *user.User
			if u, err = user.Current(); err != nil {
				return nil, err
			}
			ccacheFile = fmt.Sprintf("/tmp/krb5cc_%s", u.Uid)
		}
		var ccache *credentials.CCache
		if ccache, err = credentials.LoadCCache(ccacheFile); err != nil {
			return nil, err
		}
		krb5client, err = krb.NewFromCCache(ccache, krb5conf, settings...)
		if err != nil {
			return nil, err
		}
		needLogin = false
	}

	if needLogin {
		if err = krb5client.Login(); err != nil {
			return nil, err
		}
	}
	return krb5client, nil
}
