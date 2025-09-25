package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/client"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/config"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/logger"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/manager"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/version"

	"github.com/spf13/cobra"
)

type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

type (
	ImporterOptions struct {
		IOStreams
		Arguments    []string
		ConfigFile   string
		cfg          config.Configurator
		logger       logger.Logger
		useNopLogger bool // for test
		pool         client.Pool
		mgr          manager.Manager
	}
)

func NewImporterOptions(streams IOStreams) *ImporterOptions {
	return &ImporterOptions{
		IOStreams: streams,
	}
}

func NewDefaultImporterCommand() *cobra.Command {
	o := NewImporterOptions(IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	})
	return NewImporterCommand(o)
}

func NewImporterCommand(o *ImporterOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nebula-importer",
		Short: `The NebulaGraph Importer Tool.`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				if err != nil {
					l := o.logger

					if l == nil || o.useNopLogger {
						l = logger.NopLogger
					}

					e := errors.NewImportError(err)
					fields := logger.MapToFields(e.Fields())
					l.SkipCaller(1).WithError(e.Cause()).Error("failed to execute", fields...)
				}
				if o.pool != nil {
					_ = o.pool.Close()
				}
				if o.logger != nil {
					_ = o.logger.Sync()
					_ = o.logger.Close()
				}
			}()
			err = o.Complete(cmd, args)
			if err != nil {
				return err
			}
			err = o.Validate()
			if err != nil {
				return err
			}
			return o.Run(cmd, args)
		},
		Version:       version.GetVersion().String(),
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.SetVersionTemplate("{{.Version}}")

	o.AddFlags(cmd)
	return cmd
}

func (*ImporterOptions) Complete(_ *cobra.Command, _ []string) error {
	return nil
}

func (o *ImporterOptions) Validate() error {
	cfg, err := config.FromFile(o.ConfigFile)
	if err != nil {
		return err
	}

	if err = cfg.Optimize(o.ConfigFile); err != nil {
		return err
	}

	if err = cfg.Build(); err != nil {
		return err
	}

	o.cfg = cfg
	o.logger = cfg.GetLogger()
	o.pool = cfg.GetClientPool()
	o.mgr = cfg.GetManager()

	return nil
}

func (o *ImporterOptions) Run(_ *cobra.Command, _ []string) error {
	if err := o.mgr.Start(); err != nil {
		return err
	}
	//revive:disable-next-line:if-return
	if err := o.mgr.Wait(); err != nil {
		return err
	}
	if o.mgr.Stats().IsFailed() {
		return fmt.Errorf("failed to import")
	}
	return nil
}

func (o *ImporterOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ConfigFile, "config", "c", o.ConfigFile,
		"specify nebula-importer configure file")
}
