package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	yaml "gopkg.in/yaml.v2"
)

type NebulaClientConnection struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Address  string `yaml:"address"`
}

type NebulaClientSettings struct {
	Concurrency       int                    `yaml:"concurrency"`
	ChannelBufferSize int                    `yaml:"channelBufferSize"`
	Space             string                 `yaml:"space"`
	Connection        NebulaClientConnection `yaml:"connection"`
}

type Prop struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Ignore bool   `yaml:"ignore"`
	Index  int
}

type Edge struct {
	Name        string `yaml:"name"`
	WithRanking bool   `yaml:"withRanking"`
	Props       []Prop `yaml:"props"`
}

type Tag struct {
	Name  string `yaml:"name"`
	Props []Prop `yaml:"props"`
}

type Vertex struct {
	Tags []Tag `yaml:"tags"`
}

type Schema struct {
	Type   string `yaml:"type"`
	Edge   Edge   `yaml:"edge"`
	Vertex Vertex `yaml:"vertex"`
}

type CSVConfig struct {
	WithHeader bool `yaml:"withHeader"`
	WithLabel  bool `yaml:"withLabel"`
}

type File struct {
	Path         string    `yaml:"path"`
	FailDataPath string    `yaml:"failDataPath"`
	BatchSize    int       `yaml:"batchSize"`
	Type         string    `yaml:"type"`
	CSV          CSVConfig `yaml:"csv"`
	Schema       Schema    `yaml:"schema"`
}

type YAMLConfig struct {
	Version              string               `yaml:"version"`
	Description          string               `yaml:"description"`
	NebulaClientSettings NebulaClientSettings `yaml:"clientSettings"`
	LogPath              string               `yaml:"logPath"`
	Files                []File               `yaml:"files"`
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
	if err := config.NebulaClientSettings.validateAndReset(); err != nil {
		return err
	}

	if config.LogPath == "" {
		config.LogPath = "/tmp/nebula-importer.log"
		log.Printf("You have not configured the log file path in: logPath, reset to default path: %s", config.LogPath)
	}

	if len(config.Files) == 0 {
		return errors.New("There is no files in configuration")
	}

	for i := range config.Files {
		if err := config.Files[i].validateAndReset(dir, fmt.Sprintf("files[%d]", i)); err != nil {
			return err
		}
	}
	return nil
}

func (n *NebulaClientSettings) validateAndReset() error {
	if n.Space == "" {
		return errors.New("Please configure the space name in: clientSettings.space")
	}
	if n.Concurrency <= 0 {
		log.Printf("Invalide client concurrency: %d in clientSettings.concurrency, reset to default 40", n.Concurrency)
		n.Concurrency = 40
	}

	if n.ChannelBufferSize <= 0 {
		log.Printf("Invalide client channel buffer size: %d in clientSettings.channelBufferSize, reset to default 128", n.ChannelBufferSize)
		n.ChannelBufferSize = 128
	}

	if n.Connection.Address == "" {
		n.Connection.Address = "127.0.0.1:3699"
		log.Printf("Client connection address: %s", n.Connection.Address)
	}

	if n.Connection.User == "" {
		n.Connection.User = "user"
		log.Printf("Client connection user: %s", n.Connection.User)
	}

	if n.Connection.Password == "" {
		n.Connection.Password = "password"
		log.Printf("Client connection password: %s", n.Connection.Password)
	}
	return nil
}

func (f *File) validateAndReset(dir, prefix string) error {
	if f.Path == "" {
		return fmt.Errorf("Please configure file path in: %s.path", prefix)
	}
	if !base.FileExists(f.Path) {
		path := filepath.Join(dir, f.Path)
		if !base.FileExists(path) {
			return fmt.Errorf("File(%s) doesn't exist", f.Path)
		} else {
			f.Path = path
		}
	}
	if f.FailDataPath == "" {
		if d, err := filepath.Abs(filepath.Dir(f.Path)); err != nil {
			return err
		} else {
			f.FailDataPath = filepath.Join(d, "err", filepath.Base(f.Path))
			log.Printf("You have not configured the failed data output file path in: %s.failDataPath, reset to default path: %s", prefix, f.FailDataPath)
		}
	}
	if f.BatchSize <= 0 {
		log.Printf("Invalide batch size: %d in file(%s), reset to default 128", f.BatchSize, f.Path)
		f.BatchSize = 128
	}
	if strings.ToLower(f.Type) != "csv" {
		// TODO: Now only support csv import
		return fmt.Errorf("Invalid file data type: %s, reset to csv", f.Type)
	}

	return f.Schema.validateAndReset(fmt.Sprintf("%s.schema", prefix))
}

func (s *Schema) String() string {
	if strings.ToUpper(s.Type) == "VERTEX" {
		return s.Vertex.String()
	} else {
		return s.Edge.String()
	}
}

func (s *Schema) validateAndReset(prefix string) error {
	var err error = nil
	switch strings.ToLower(s.Type) {
	case "edge":
		err = s.Edge.validateAndReset(fmt.Sprintf("%s.edge", prefix))
	case "vertex":
		err = s.Vertex.validateAndReset(fmt.Sprintf("%s.vertex", prefix))
	default:
		err = fmt.Errorf("Error schema type(%s) in %s.type only edge and vertex are supported", s.Type, prefix)
	}
	return err
}

func (e *Edge) String() string {
	var cells []string
	cells = append(cells, ":SRC_VID", ":DST_VID")
	if e.WithRanking {
		cells = append(cells, ":RANK")
	}
	for _, prop := range e.Props {
		cells = append(cells, prop.String(e.Name))
	}
	return strings.Join(cells, ",")
}

func (e *Edge) validateAndReset(prefix string) error {
	if e.Name == "" {
		fmt.Errorf("Please configure edge name in: %s.name", prefix)
	}
	for i := range e.Props {
		if err := e.Props[i].validateAndReset(fmt.Sprintf("%s.prop[%d]", prefix, i)); err != nil {
			return err
		}
	}
	return nil
}

func (v *Vertex) String() string {
	var cells []string
	cells = append(cells, ":VID")
	for _, tag := range v.Tags {
		for _, prop := range tag.Props {
			cells = append(cells, prop.String(tag.Name))
		}
	}
	return strings.Join(cells, ",")
}

func (v *Vertex) validateAndReset(prefix string) error {
	for i := range v.Tags {
		if err := v.Tags[i].validateAndReset(fmt.Sprintf("%s.tags[%d]", prefix, i)); err != nil {
			return err
		}
	}
	return nil
}

func (p *Prop) String(prefix string) string {
	if p.Ignore {
		return ":IGNORE"
	} else {
		return fmt.Sprintf("%s.%s:%s", prefix, p.Name, p.Type)
	}
}

func (p *Prop) validateAndReset(prefix string) error {
	p.Type = strings.ToLower(p.Type)
	if base.IsValidType(p.Type) {
		return nil
	} else {
		return fmt.Errorf("Error property type of %s.type: %s", prefix, p.Type)
	}
}

func (t *Tag) validateAndReset(prefix string) error {
	if t.Name == "" {
		return fmt.Errorf("Please configure the vertex tag name in: %s.name", prefix)
	}

	for i := range t.Props {
		if err := t.Props[i].validateAndReset(fmt.Sprintf("%s.props[%d]", prefix, i)); err != nil {
			return err
		}
	}
	return nil
}
