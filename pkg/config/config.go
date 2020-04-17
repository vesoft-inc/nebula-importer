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
	User     *string `json:"user" yaml:"user"`
	Password *string `json:"password" yaml:"password"`
	Address  *string `json:"address" yaml:"address"`
}

type NebulaClientSettings struct {
	Retry             *int                    `json:"retry" yaml:"retry"`
	Concurrency       *int                    `json:"concurrency" yaml:"concurrency"`
	ChannelBufferSize *int                    `json:"channelBufferSize" yaml:"channelBufferSize"`
	Space             *string                 `json:"space" yaml:"space"`
	Connection        *NebulaClientConnection `json:"connection" yaml:"connection"`
}

type Prop struct {
	Name  *string `json:"name" yaml:"name"`
	Type  *string `json:"type" yaml:"type"`
	Index *int    `json:"index" yaml:"index"`
}

type VID struct {
	Index    *int    `json:"index" yaml:"index"`
	Function *string `json:"function" yaml:"function"`
}

type Rank struct {
	Index *int `json:"index" yaml:"index"`
}

type Edge struct {
	Name        *string `json:"name" yaml:"name"`
	WithRanking *bool   `json:"withRanking" yaml:"withRanking"`
	Props       []*Prop `json:"props" yaml:"props"`
	SrcVID      *VID    `json:"srcVID" yaml:"srcVID"`
	DstVID      *VID    `json:"dstVID" yaml:"dstVID"`
	Rank        *Rank   `json:"rank" yaml:"rank"`
}

type Tag struct {
	Name  *string `json:"name" yaml:"name"`
	Props []*Prop `json:"props" yaml:"props"`
}

type Vertex struct {
	VID  *VID   `json:"vid" yaml:"vid"`
	Tags []*Tag `json:"tags" yaml:"tags"`
}

type Schema struct {
	Type   *string `json:"type" yaml:"type"`
	Edge   *Edge   `json:"edge" yaml:"edge"`
	Vertex *Vertex `json:"vertex" yaml:"vertex"`
}

type CSVConfig struct {
	WithHeader *bool   `json:"withHeader" yaml:"withHeader"`
	WithLabel  *bool   `json:"withLabel" yaml:"withLabel"`
	Delimiter  *string `json:"delimiter" yaml:"delimiter"`
}

type File struct {
	Paths        []string
	Path         *string    `json:"path" yaml:"path"`
	FailDataPath *string    `json:"failDataPath" yaml:"failDataPath"`
	BatchSize    *int       `json:"batchSize" yaml:"batchSize"`
	Limit        *int       `json:"limit" yaml:"limit"`
	InOrder      *bool      `json:"inOrder" yaml:"inOrder"`
	Type         *string    `json:"type" yaml:"type"`
	CSV          *CSVConfig `json:"csv" yaml:"csv"`
	Schema       *Schema    `json:"schema" yaml:"schema"`
}

type YAMLConfig struct {
	Version              *string               `json:"version" yaml:"version"`
	Description          *string               `json:"description" yaml:"description"`
	NebulaClientSettings *NebulaClientSettings `json:"clientSettings" yaml:"clientSettings"`
	LogPath              *string               `json:"logPath" yaml:"logPath"`
	Files                []*File               `json:"files" yaml:"files"`
}

var version string = "v1rc2"

func Parse(filename string) (*YAMLConfig, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var conf YAMLConfig
	if err = yaml.Unmarshal(content, &conf); err != nil {
		return nil, err
	}

	if conf.Version == nil && *conf.Version != version {
		return nil, fmt.Errorf("The YAML configure version must be %s", version)
	}

	path, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return nil, err
	}
	if err = conf.ValidateAndReset(path); err != nil {
		return nil, err
	}

	return &conf, nil
}

func (config *YAMLConfig) ValidateAndReset(dir string) error {
	if config.NebulaClientSettings == nil {
		return errors.New("please configure clientSettings")
	}
	if err := config.NebulaClientSettings.validateAndReset("clientSettings"); err != nil {
		return err
	}

	if config.LogPath == nil {
		defaultPath := "/tmp/nebula-importer.log"
		config.LogPath = &defaultPath
		logger.Warnf("You have not configured the log file path in: logPath, reset to default path: %s", *config.LogPath)
	}

	if config.Files == nil || len(config.Files) == 0 {
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

	if n.Retry == nil {
		retry := 1
		n.Retry = &retry
		logger.Warnf("Invalid retry option in %s.retry, reset to %d ", prefix, *n.Retry)
	}

	if n.Concurrency == nil {
		d := 10
		n.Concurrency = &d
		logger.Warnf("Invalid client concurrency in %s.concurrency, reset to %d", prefix, *n.Concurrency)
	}

	if n.ChannelBufferSize == nil {
		d := 128
		n.ChannelBufferSize = &d
		logger.Warnf("Invalid client channel buffer size in %s.channelBufferSize, reset to %d", prefix, *n.ChannelBufferSize)
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
	if !base.PathExists(*f.Path) {
		path := filepath.Join(dir, *f.Path)
		if !base.PathExists(path) {
			return fmt.Errorf("File(%s) doesn't exist", *f.Path)
		} else {
			f.Path = &path
		}
	}
	f.Paths, _ = base.PathFileList(*f.Path)

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
		logger.Infof("Invalid batch size in path(%s), reset to %d", *f.Path, *f.BatchSize)
	}
	if f.InOrder == nil {
		inOrder := false
		f.InOrder = &inOrder
	}
	if strings.ToLower(*f.Type) != "csv" {
		// TODO: Now only support csv import
		return fmt.Errorf("Invalid file data type: %s, reset to csv", *f.Type)
	}

	if f.CSV != nil {
		err := f.CSV.validateAndReset(fmt.Sprintf("%s.csv", prefix))
		if err != nil {
			return err
		}
	}

	if f.Schema == nil {
		return fmt.Errorf("Please configure file schema: %s.schema", prefix)
	}
	return f.Schema.validateAndReset(fmt.Sprintf("%s.schema", prefix))
}

func (c *CSVConfig) validateAndReset(prefix string) error {
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

	if c.Delimiter != nil {
		if len(*c.Delimiter) == 0 {
			return fmt.Errorf("%s.delimiter is empty string", prefix)
		}
	}

	return nil
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

func (v *VID) ParseFunction(str string) {
	i := strings.Index(str, "(")
	j := strings.Index(str, ")")
	if i < 0 && j < 0 {
		v.Function = nil
	} else if i > 0 && j > i {
		function := strings.ToLower(str[i+1 : j])
		v.Function = &function
	} else {
		logger.Fatalf("Invalid function format: %s", str)
	}
}

func (v *VID) String(vid string) string {
	if v.Function == nil || *v.Function == "" {
		return vid
	} else {
		return fmt.Sprintf("%s(%s)", vid, *v.Function)
	}
}

func (v *VID) checkFunction(prefix string) error {
	if v.Function != nil {
		switch strings.ToLower(*v.Function) {
		case "", "hash", "uuid":
		default:
			return fmt.Errorf("Invalid %s.function: %s, only following values are supported: \"\", hash, uuid", prefix, *v.Function)
		}
	}
	return nil
}

func (v *VID) validateAndReset(prefix string, defaultVal int) error {
	if v.Index == nil {
		v.Index = &defaultVal
	}
	if *v.Index < 0 {
		return fmt.Errorf("Invalid %s.index: %d", prefix, *v.Index)
	}
	if err := v.checkFunction(prefix); err != nil {
		return err
	}
	return nil
}

func (r *Rank) validateAndReset(prefix string, defaultVal int) error {
	if r.Index == nil {
		r.Index = &defaultVal
	}
	if *r.Index < 0 {
		return fmt.Errorf("Invalid %s.index: %d", prefix, *r.Index)
	}
	return nil
}

func (e *Edge) FormatValues(record base.Record) string {
	var cells []string
	for i, prop := range e.Props {
		if c, err := prop.FormatValue(record); err != nil {
			logger.Fatalf("edge: %s, column: %d, error: %v", e.String(), i, err)
		} else {
			cells = append(cells, c)
		}
	}
	rank := ""
	if e.Rank != nil && e.Rank.Index != nil {
		rank = fmt.Sprintf("@%s", record[*e.Rank.Index])
	}
	var srcVID string
	if e.SrcVID.Function != nil {
		//TODO(yee): differentiate string and integer column type, find and compare src/dst vertex column with property
		srcVID = fmt.Sprintf("%s(%q)", *e.SrcVID.Function, record[*e.SrcVID.Index])
	} else {
		srcVID = base.TryConvInt64(record[*e.SrcVID.Index])
	}
	var dstVID string
	if e.DstVID.Function != nil {
		dstVID = fmt.Sprintf("%s(%q)", *e.DstVID.Function, record[*e.DstVID.Index])
	} else {
		dstVID = base.TryConvInt64(record[*e.DstVID.Index])
	}
	return fmt.Sprintf(" %s->%s%s:(%s) ", srcVID, dstVID, rank, strings.Join(cells, ","))
}

func (e *Edge) maxIndex() int {
	maxIdx := 0
	if e.SrcVID != nil && e.SrcVID.Index != nil && *e.SrcVID.Index > maxIdx {
		maxIdx = *e.SrcVID.Index
	}

	if e.DstVID != nil && e.DstVID.Index != nil && *e.DstVID.Index > maxIdx {
		maxIdx = *e.DstVID.Index
	}

	if e.Rank != nil && e.Rank.Index != nil && *e.Rank.Index > maxIdx {
		maxIdx = *e.Rank.Index
	}

	for _, p := range e.Props {
		if p != nil && p.Index != nil && *p.Index > maxIdx {
			maxIdx = *p.Index
		}
	}

	return maxIdx
}

func combine(cell, val string) string {
	if len(cell) > 0 {
		return fmt.Sprintf("%s/%s", cell, val)
	} else {
		return val
	}
}

func (e *Edge) String() string {
	cells := make([]string, e.maxIndex()+1)
	if e.SrcVID != nil && e.SrcVID.Index != nil {
		cells[*e.SrcVID.Index] = combine(cells[*e.SrcVID.Index], e.SrcVID.String(base.LABEL_SRC_VID))
	}
	if e.DstVID != nil && e.DstVID.Index != nil {
		cells[*e.DstVID.Index] = combine(cells[*e.DstVID.Index], e.DstVID.String(base.LABEL_DST_VID))
	}
	if e.Rank != nil && e.Rank.Index != nil {
		cells[*e.Rank.Index] = combine(cells[*e.Rank.Index], base.LABEL_RANK)
	}
	for _, prop := range e.Props {
		if prop.Index != nil {
			cells[*prop.Index] = combine(cells[*prop.Index], prop.String(*e.Name))
		}
	}
	for i := range cells {
		if cells[i] == "" {
			cells[i] = base.LABEL_IGNORE
		}
	}
	return strings.Join(cells, ",")
}

func (e *Edge) validateAndReset(prefix string) error {
	if e.Name == nil {
		return fmt.Errorf("Please configure edge name in: %s.name", prefix)
	}
	if e.SrcVID != nil {
		if err := e.SrcVID.validateAndReset(fmt.Sprintf("%s.srcVID", prefix), 0); err != nil {
			return err
		}
	} else {
		index := 0
		e.SrcVID = &VID{Index: &index}
	}
	if e.DstVID != nil {
		if err := e.DstVID.validateAndReset(fmt.Sprintf("%s.dstVID", prefix), 1); err != nil {
			return err
		}
	} else {
		index := 1
		e.DstVID = &VID{Index: &index}
	}
	start := 2
	if e.Rank != nil {
		if err := e.Rank.validateAndReset(fmt.Sprintf("%s.rank", prefix), 2); err != nil {
			return err
		}
		start++
	} else {
		if e.WithRanking != nil && *e.WithRanking {
			index := 2
			e.Rank = &Rank{Index: &index}
			start++
		}
	}
	for i := range e.Props {
		if e.Props[i] != nil {
			if err := e.Props[i].validateAndReset(fmt.Sprintf("%s.prop[%d]", prefix, i), i+start); err != nil {
				return err
			}
		} else {
			logger.Errorf("prop %d of edge %s is nil", i, *e.Name)
		}
	}
	return nil
}

func (v *Vertex) FormatValues(record base.Record) string {
	var cells []string
	for _, tag := range v.Tags {
		cells = append(cells, tag.FormatValues(record))
	}
	var vid string
	if v.VID.Function != nil {
		vid = fmt.Sprintf("%s(%q)", *v.VID.Function, record[*v.VID.Index])
	} else {
		vid = base.TryConvInt64(record[*v.VID.Index])
	}
	return fmt.Sprintf(" %s: (%s)", vid, strings.Join(cells, ","))
}

func (v *Vertex) maxIndex() int {
	maxIdx := 0
	if v.VID != nil && v.VID.Index != nil && *v.VID.Index > maxIdx {
		maxIdx = *v.VID.Index
	}
	for _, tag := range v.Tags {
		if tag != nil {
			for _, prop := range tag.Props {
				if prop != nil && prop.Index != nil && *prop.Index > maxIdx {
					maxIdx = *prop.Index
				}
			}
		}
	}

	return maxIdx
}

func (v *Vertex) String() string {
	cells := make([]string, v.maxIndex()+1)
	if v.VID != nil && v.VID.Index != nil {
		cells[*v.VID.Index] = v.VID.String(base.LABEL_VID)
	}
	for _, tag := range v.Tags {
		for _, prop := range tag.Props {
			if prop != nil && prop.Index != nil {
				cells[*prop.Index] = combine(cells[*prop.Index], prop.String(*tag.Name))
			}
		}
	}

	for i := range cells {
		if cells[i] == "" {
			cells[i] = base.LABEL_IGNORE
		}
	}
	return strings.Join(cells, ",")
}

func (v *Vertex) validateAndReset(prefix string) error {
	// if v.Tags == nil {
	// 	return fmt.Errorf("Please configure %.tags", prefix)
	// }
	if v.VID != nil {
		if err := v.VID.validateAndReset(fmt.Sprintf("%s.vid", prefix), 0); err != nil {
			return err
		}
	} else {
		index := 0
		v.VID = &VID{Index: &index}
	}
	j := 1
	for i := range v.Tags {
		if v.Tags[i] != nil {
			if err := v.Tags[i].validateAndReset(fmt.Sprintf("%s.tags[%d]", prefix, i), j); err != nil {
				return err
			}
			j = j + len(v.Tags[i].Props)
		} else {
			logger.Errorf("tag %d is nil", i)
		}
	}
	return nil
}

func (p *Prop) IsStringType() bool {
	return strings.ToLower(*p.Type) == "string"
}

func (p *Prop) IsIntType() bool {
	return strings.ToLower(*p.Type) == "int"
}

func (p *Prop) IsDateTimestampType() bool {
	return strings.HasPrefix(strings.ToLower(*p.Type), "date-timestamp")
}

func (p *Prop) FormatValue(record base.Record) (string, error) {
	if p.Index != nil && *p.Index >= len(record) {
		return "", fmt.Errorf("Prop index %d out range %d of record(%v)", *p.Index, len(record), record)
	}
	r := record[*p.Index]
	if p.IsStringType() {
		return fmt.Sprintf("%q", r), nil
	}
	if p.IsIntType() {
		return base.TryConvInt64(r), nil
	}
	if p.IsDateTimestampType() {
		return base.TryConvDateTimestamp(r, strings.Split(*p.Type, ":")[1]), nil
	}
	return r, nil
}

func (p *Prop) String(prefix string) string {
	return fmt.Sprintf("%s.%s:%s", prefix, *p.Name, *p.Type)
}

func (p *Prop) validateAndReset(prefix string, val int) error {
	*p.Type = strings.ToLower(*p.Type)
	if !base.IsValidType(*p.Type) {
		return fmt.Errorf("Error property type of %s.type: %s", prefix, *p.Type)
	}
	if p.Index == nil {
		p.Index = &val
	} else {
		if *p.Index < 0 {
			logger.Fatalf("Invalid prop index: %d, name: %s, type: %s", *p.Index, *p.Name, *p.Type)
		}
	}
	return nil
}

func (t *Tag) FormatValues(record base.Record) string {
	var cells []string
	for _, p := range t.Props {
		if c, err := p.FormatValue(record); err != nil {
			logger.Fatalf("tag: %v, error: %v", *t, err)
		} else {
			cells = append(cells, c)
		}
	}
	return strings.Join(cells, ",")
}

func (t *Tag) validateAndReset(prefix string, start int) error {
	if t.Name == nil {
		return fmt.Errorf("Please configure the vertex tag name in: %s.name", prefix)
	}

	for i := range t.Props {
		if t.Props[i] != nil {
			if err := t.Props[i].validateAndReset(fmt.Sprintf("%s.props[%d]", prefix, i), i+start); err != nil {
				return err
			}
		} else {
			logger.Errorf("prop %d of tag %s is nil", i, *t.Name)
		}
	}
	return nil
}
