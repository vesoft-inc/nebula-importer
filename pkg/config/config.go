package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
	yaml "gopkg.in/yaml.v2"
)

type NebulaClientConnection struct {
	User     *string `yaml:"user"`
	Password *string `yaml:"password"`
	Address  *string `yaml:"address"`
}

type NebulaClientSettings struct {
	Concurrency       *int                    `yaml:"concurrency"`
	ChannelBufferSize *int                    `yaml:"channelBufferSize"`
	Space             *string                 `yaml:"space"`
	Connection        *NebulaClientConnection `yaml:"connection"`
}

type Prop struct {
	Name   *string `yaml:"name"`
	Type   *string `yaml:"type"`
	Ignore *bool   `yaml:"ignore"`
	Index  *int    `yaml:"index"`
}

type VID struct {
	Index    *int    `yaml:"index"`
	Function *string `yaml:"function"`
}

type Rank struct {
	Index *int `yaml:"index"`
}

type Edge struct {
	Name        *string `yaml:"name"`
	WithRanking *bool   `yaml:"withRanking"`
	Props       []*Prop `yaml:"props"`
	SrcVID      *VID    `yaml:"srcVID"`
	DstVID      *VID    `yaml:"dstVID"`
	Rank        *Rank   `yaml:"rank"`
}

type Tag struct {
	Name  *string `yaml:"name"`
	Props []*Prop `yaml:"props"`
}

type Vertex struct {
	VID  *VID   `yaml:"vid"`
	Tags []*Tag `yaml:"tags"`
}

type Schema struct {
	Type   *string `yaml:"type"`
	Edge   *Edge   `yaml:"edge"`
	Vertex *Vertex `yaml:"vertex"`
}

type CSVConfig struct {
	WithHeader *bool `yaml:"withHeader"`
	WithLabel  *bool `yaml:"withLabel"`
}

type File struct {
	Path         *string    `yaml:"path"`
	FailDataPath *string    `yaml:"failDataPath"`
	BatchSize    *int       `yaml:"batchSize"`
	Type         *string    `yaml:"type"`
	CSV          *CSVConfig `yaml:"csv"`
	Schema       *Schema    `yaml:"schema"`
}

type YAMLConfig struct {
	Version              *string               `yaml:"version"`
	Description          *string               `yaml:"description"`
	NebulaClientSettings *NebulaClientSettings `yaml:"clientSettings"`
	LogPath              *string               `yaml:"logPath"`
	Files                []*File               `yaml:"files"`
}

func Parse(filename string) (*YAMLConfig, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var conf YAMLConfig
	if err = yaml.Unmarshal(content, &conf); err != nil {
		return nil, err
	}
	path, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return nil, err
	}
	if err = conf.validateAndReset(path); err != nil {
		return nil, err
	}

	return &conf, nil
}

func (config *YAMLConfig) validateAndReset(dir string) error {
	if err := config.NebulaClientSettings.validateAndReset("clientSettings"); err != nil {
		return err
	}

	if config.LogPath == nil {
		defaultPath := "/tmp/nebula-importer.log"
		config.LogPath = &defaultPath
		logger.Warnf("You have not configured the log file path in: logPath, reset to default path: %s", *config.LogPath)
	}

	if config.Files == nil {
		return errors.New("There is no files in configuration")
	}

	for i := range config.Files {
		if err := config.Files[i].validateAndReset(dir, fmt.Sprintf("files[%d]", i)); err != nil {
			return err
		}
	}
	return nil
}

func (n *NebulaClientSettings) validateAndReset(prefix string) error {
	if n.Space == nil {
		return fmt.Errorf("Please configure the space name in: %s.space", prefix)
	}

	if n.Concurrency == nil {
		d := 10
		n.Concurrency = &d
		logger.Warnf("Invalide client concurrency in %s.concurrency, reset to %d", prefix, *n.Concurrency)
	}

	if n.ChannelBufferSize == nil {
		d := 128
		n.ChannelBufferSize = &d
		logger.Warnf("Invalide client channel buffer size in %s.channelBufferSize, reset to %d", prefix, *n.ChannelBufferSize)
	}

	if n.Connection == nil {
		return fmt.Errorf("Please configure the connection information in: %s.connection", prefix)
	} else {
		return n.Connection.validateAndReset(fmt.Sprintf("%s.connection", prefix))
	}
}

func (c *NebulaClientConnection) validateAndReset(prefix string) error {
	if c.Address == nil {
		a := "127.0.0.1:3699"
		c.Address = &a
		logger.Warnf("%s.address: %s", prefix, *c.Address)
	}

	if c.User == nil {
		u := "user"
		c.User = &u
		logger.Warnf("%s.user: %s", prefix, *c.User)
	}

	if c.Password == nil {
		p := "password"
		c.Password = &p
		logger.Warnf("%s.password: %s", prefix, *c.Password)
	}
	return nil
}

func (f *File) validateAndReset(dir, prefix string) error {
	if f.Path == nil {
		return fmt.Errorf("Please configure file path in: %s.path", prefix)
	}
	if !base.FileExists(*f.Path) {
		path := filepath.Join(dir, *f.Path)
		if !base.FileExists(path) {
			return fmt.Errorf("File(%s) doesn't exist", *f.Path)
		} else {
			f.Path = &path
		}
	}
	if f.FailDataPath == nil {
		if d, err := filepath.Abs(filepath.Dir(*f.Path)); err != nil {
			return err
		} else {
			p := filepath.Join(d, "err", filepath.Base(*f.Path))
			f.FailDataPath = &p
			logger.Warnf("You have not configured the failed data output file path in: %s.failDataPath, reset to default path: %s", prefix, *f.FailDataPath)
		}
	}
	if f.BatchSize == nil {
		b := 128
		f.BatchSize = &b
		logger.Infof("Invalide batch size in file(%s), reset to %d", *f.Path, *f.BatchSize)
	}
	if strings.ToLower(*f.Type) != "csv" {
		// TODO: Now only support csv import
		return fmt.Errorf("Invalid file data type: %s, reset to csv", *f.Type)
	}

	if f.CSV != nil {
		f.CSV.validateAndReset(fmt.Sprintf("%s.csv", prefix))
	}

	if f.Schema == nil {
		return fmt.Errorf("Please configure file schema: %s.schema", prefix)
	}
	return f.Schema.validateAndReset(fmt.Sprintf("%s.schema", prefix))
}

func (c *CSVConfig) validateAndReset(prefix string) {
	if c.WithHeader == nil {
		h := false
		c.WithHeader = &h
		logger.Infof("%s.withHeader: %v", prefix, false)
	}

	if c.WithLabel == nil {
		l := false
		c.WithLabel = &l
		logger.Infof("%s.withLabel: %v", prefix, false)
	}
}

func (s *Schema) IsVertex() bool {
	return strings.ToUpper(*s.Type) == "VERTEX"
}

func (s *Schema) String() string {
	if s.IsVertex() {
		return s.Vertex.String()
	} else {
		return s.Edge.String()
	}
}

func (s *Schema) validateAndReset(prefix string) error {
	var err error = nil
	switch strings.ToLower(*s.Type) {
	case "edge":
		if s.Edge != nil {
			err = s.Edge.validateAndReset(fmt.Sprintf("%s.edge", prefix))
		} else {
			logger.Infof("%s.edge is nil", prefix)
		}
	case "vertex":
		if s.Vertex != nil {
			err = s.Vertex.validateAndReset(fmt.Sprintf("%s.vertex", prefix))
		} else {
			logger.Infof("%s.vertex is nil", prefix)
		}
	default:
		err = fmt.Errorf("Error schema type(%s) in %s.type only edge and vertex are supported", *s.Type, prefix)
	}
	return err
}

func (v *VID) validateAndReset(prefix string, defaultVal int) error {
	if v.Index == nil {
		v.Index = &defaultVal
		f := ""
		v.Function = &f
		logger.Infof("%s.index: %d, %s.function: %s", prefix, *v.Index, prefix, *v.Function)
	}
	return nil
}

func (r *Rank) validateAndReset(prefix string, defaultRank int) error {
	if r.Index == nil {
		r.Index = &defaultRank
		logger.Infof("%s.index: %d", prefix, *r.Index)
	}
	return nil
}

func (e *Edge) FormatValues(record base.Record) string {
	var cells []string
	for _, prop := range e.Props {
		if prop.Ignore != nil && !*prop.Ignore {
			cells = append(cells, prop.FormatValue(record))
		}
	}
	rank := ""
	if e.WithRanking != nil && *e.WithRanking {
		rank = fmt.Sprintf("@%s", record[*e.Rank.Index])
	}
	return fmt.Sprintf(" %s->%s%s:(%s) ", record[*e.SrcVID.Index], record[*e.DstVID.Index], rank, strings.Join(cells, ","))
}

func (e *Edge) String() string {
	var cells []string
	cells = append(cells, base.LABEL_SRC_VID, base.LABEL_DST_VID)
	if e.WithRanking != nil && *e.WithRanking {
		cells = append(cells, base.LABEL_RANK)
	}
	for _, prop := range e.Props {
		cells = append(cells, prop.String(*e.Name))
	}
	return strings.Join(cells, ",")
}

func (e *Edge) validateAndReset(prefix string) error {
	if e.Name == nil {
		return fmt.Errorf("Please configure edge name in: %s.name", prefix)
	}
	if e.SrcVID != nil {
		e.SrcVID.validateAndReset(fmt.Sprintf("%s.srcVID", prefix), 0)
	}
	if e.DstVID != nil {
		e.DstVID.validateAndReset(fmt.Sprintf("%s.dstVID", prefix), 1)
	}
	if e.Rank != nil {
		e.Rank.validateAndReset(fmt.Sprintf("%s.rank", prefix), 2)
	}
	for i := range e.Props {
		if err := e.Props[i].validateAndReset(fmt.Sprintf("%s.prop[%d]", prefix, i)); err != nil {
			return err
		}
	}
	return nil
}

func (v *Vertex) FormatValues(record base.Record) string {
	var cells []string
	for _, tag := range v.Tags {
		cells = append(cells, tag.FormatValues(record))
	}
	return fmt.Sprintf(" %s: (%s)", record[*v.VID.Index], strings.Join(cells, ","))
}

func (v *Vertex) String() string {
	var cells []string
	cells = append(cells, base.LABEL_VID)
	for _, tag := range v.Tags {
		for _, prop := range tag.Props {
			cells = append(cells, prop.String(*tag.Name))
		}
	}
	return strings.Join(cells, ",")
}

func (v *Vertex) validateAndReset(prefix string) error {
	if v.Tags == nil {
		return fmt.Errorf("Please configure %.tags", prefix)
	}
	if v.VID != nil {
		v.VID.validateAndReset(fmt.Sprintf("%s.vid", prefix), 0)
	}
	for i := range v.Tags {
		if err := v.Tags[i].validateAndReset(fmt.Sprintf("%s.tags[%d]", prefix, i)); err != nil {
			return err
		}
	}
	return nil
}

func (p *Prop) IsStringType() bool {
	return strings.ToLower(*p.Type) == "string"
}

func (p *Prop) FormatValue(record base.Record) string {
	if p.Index != nil && *p.Index >= len(record) {
		logger.Fatalf("Prop index %d out range %d of record(%v)", p.Index, len(record), record)
	}
	r := record[*p.Index]
	if p.IsStringType() {
		return fmt.Sprintf("%q", r)
	}
	return r
}

func (p *Prop) String(prefix string) string {
	if p.Ignore != nil && *p.Ignore {
		return base.LABEL_IGNORE
	} else {
		return fmt.Sprintf("%s.%s:%s", prefix, *p.Name, *p.Type)
	}
}

func (p *Prop) validateAndReset(prefix string) error {
	*p.Type = strings.ToLower(*p.Type)
	if base.IsValidType(*p.Type) {
		return nil
	} else {
		return fmt.Errorf("Error property type of %s.type: %s", prefix, *p.Type)
	}
}

func (t *Tag) FormatValues(record base.Record) string {
	var cells []string
	for _, p := range t.Props {
		if p.Ignore != nil && !*p.Ignore {
			cells = append(cells, p.FormatValue(record))
		}
	}
	return strings.Join(cells, ",")
}

func (t *Tag) validateAndReset(prefix string) error {
	if t.Name == nil {
		return fmt.Errorf("Please configure the vertex tag name in: %s.name", prefix)
	}

	for i := range t.Props {
		if err := t.Props[i].validateAndReset(fmt.Sprintf("%s.props[%d]", prefix, i)); err != nil {
			return err
		}
	}
	return nil
}
