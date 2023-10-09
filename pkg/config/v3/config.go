package configv3

import (
	"fmt"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/client"
	configbase "github.com/vesoft-inc/nebula-importer/v4/pkg/config/base"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/logger"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/manager"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/utils"
)

var _ configbase.Configurator = (*Config)(nil)

type (
	Client = configbase.Client
	Log    = configbase.Log

	Config struct {
		Client  `yaml:"client"`
		Manager `yaml:"manager"`
		Sources `yaml:"sources"`
		*Log    `yaml:"log,omitempty"`

		logger logger.Logger
		pool   client.Pool
		mgr    manager.Manager
	}
)

func (c *Config) Optimize(configPath string) error {
	if err := c.Client.OptimizePath(configPath); err != nil {
		return err
	}

	if err := c.Log.OptimizePath(configPath); err != nil {
		return err
	}

	if err := c.Sources.OptimizePath(configPath); err != nil {
		return err
	}

	//revive:disable-next-line:if-return
	if err := c.Sources.OptimizePathWildCard(); err != nil {
		return err
	}

	return nil
}

func (c *Config) Build() error {
	var (
		err  error
		l    logger.Logger
		pool client.Pool
		mgr  manager.Manager
	)
	defer func() {
		if err != nil {
			if pool != nil {
				_ = pool.Close()
			}
			if l != nil {
				_ = l.Close()
			}
		}
	}()

	l, err = c.BuildLogger()
	if err != nil {
		return err
	}
	pool, err = c.BuildClientPool(
		client.WithLogger(l),
		client.WithClientInitFunc(c.clientInitFunc),
	)
	if err != nil {
		return err
	}
	mgr, err = c.Manager.BuildManager(l, pool, c.Sources,
		manager.WithGetClientOptions(client.WithClientInitFunc(nil)), // clean the USE SPACE in 3.x
	)
	if err != nil {
		return err
	}

	c.logger = l
	c.pool = pool
	c.mgr = mgr

	return nil
}

func (c *Config) GetLogger() logger.Logger {
	return c.logger
}

func (c *Config) GetClientPool() client.Pool {
	return c.pool
}

func (c *Config) GetManager() manager.Manager {
	return c.mgr
}

func (c *Config) clientInitFunc(cli client.Client) error {
	resp, err := cli.Execute(fmt.Sprintf("USE %s", utils.ConvertIdentifier(c.Manager.GraphName)))
	if err != nil {
		return err
	}
	if !resp.IsSucceed() {
		return resp.GetError()
	}
	return nil
}
