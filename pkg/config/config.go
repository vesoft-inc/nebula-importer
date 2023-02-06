package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vesoft-inc/nebula-importer/v3/pkg/base"
	ierrors "github.com/vesoft-inc/nebula-importer/v3/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/v3/pkg/logger"
	"github.com/vesoft-inc/nebula-importer/v3/pkg/picker"
	"github.com/vesoft-inc/nebula-importer/v3/pkg/utils"
	"gopkg.in/yaml.v2"
)

const (
	dbNULL = "NULL"
)

type NebulaClientConnection struct {
	User     *string `json:"user" yaml:"user"`
	Password *string `json:"password" yaml:"password"`
	Address  *string `json:"address" yaml:"address"`
}

type NebulaPostStart struct {
	Commands    *string `json:"commands" yaml:"commands"`
	AfterPeriod *string `json:"afterPeriod" yaml:"afterPeriod"`
}

type NebulaPreStop struct {
	Commands *string `json:"commands" yaml:"commands"`
}

type NebulaClientSettings struct {
	Retry             *int                    `json:"retry" yaml:"retry"`
	Concurrency       *int                    `json:"concurrency" yaml:"concurrency"`
	ChannelBufferSize *int                    `json:"channelBufferSize" yaml:"channelBufferSize"`
	Space             *string                 `json:"space" yaml:"space"`
	Connection        *NebulaClientConnection `json:"connection" yaml:"connection"`
	PostStart         *NebulaPostStart        `json:"postStart" yaml:"postStart"` // from v1
	PreStop           *NebulaPreStop          `json:"preStop" yaml:"preStop"`     // from v1
}

type Prop struct {
	Name               *string `json:"name" yaml:"name"`
	Type               *string `json:"type" yaml:"type"`
	Index              *int    `json:"index" yaml:"index"`
	Nullable           bool    `json:"nullable" yaml:"nullable"`
	NullValue          string  `json:"nullValue" yaml:"nullValue"`
	AlternativeIndices []int   `json:"alternativeIndices" yaml:"alternativeIndices"`
	DefaultValue       *string `json:"defaultValue" yaml:"defaultValue"`
	picker             picker.Picker
}

type VID struct {
	Index       *int          `json:"index" yaml:"index"`
	ConcatItems []interface{} `json:"concatItems" yaml:"concatItems"` // only string and int is support, int is for Index
	Function    *string       `json:"function" yaml:"function"`
	Type        *string       `json:"type" yaml:"type"`
	Prefix      *string       `json:"prefix" yaml:"prefix"`
	picker      picker.Picker
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
	LazyQuotes *bool   `json:"lazyQuotes" yaml:"lazyQuotes"`
}

type File struct {
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
	RemoveTempFiles      *bool                 `json:"removeTempFiles" yaml:"removeTempFiles"` // from v1
	NebulaClientSettings *NebulaClientSettings `json:"clientSettings" yaml:"clientSettings"`
	LogPath              *string               `json:"logPath" yaml:"logPath"`
	WorkingDirectory     *string               `json:"workingDir" yaml:"workingDir"`
	Files                []*File               `json:"files" yaml:"files"`
}

var (
	kDefaultVidType   = "string"
	kDefaultConnAddr  = "127.0.0.1:9669"
	kDefaultUser      = "root"
	kDefaultPassword  = "nebula"
	kDefaultBatchSize = 128
	supportedVersions = []string{"v1rc1", "v1rc2", "v1", "v2"}
)

func isSupportedVersion(ver string) bool {
	for _, v := range supportedVersions {
		if v == ver {
			return true
		}
	}
	return false
}

func Parse(filename string, runnerLogger logger.Logger) (*YAMLConfig, error) {
	logger.SetLogger(runnerLogger)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, ierrors.Wrap(ierrors.InvalidConfigPathOrFormat, err)
	}

	var conf YAMLConfig
	if err = yaml.Unmarshal(content, &conf); err != nil {
		return nil, ierrors.Wrap(ierrors.InvalidConfigPathOrFormat, err)
	}

	if conf.Version == nil || !isSupportedVersion(*conf.Version) {
		return nil, ierrors.Wrap(ierrors.InvalidConfigPathOrFormat,
			fmt.Errorf("The supported YAML configure versions are %v, please upgrade importer.", supportedVersions))
	}
	abs, err := filepath.Abs(filename)
	if err != nil {
		return nil, ierrors.Wrap(ierrors.InvalidConfigPathOrFormat, err)
	}
	path := filepath.Dir(abs)

	if workingDir := conf.WorkingDirectory; workingDir != nil && len(*workingDir) > 0 {
		if !filepath.IsAbs(*workingDir) {
			path = filepath.Join(path, *workingDir)
		} else {
			path = *workingDir
		}
	}

	if err = conf.ValidateAndReset(path); err != nil {
		return nil, ierrors.Wrap(ierrors.ConfigError, err)
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

	if config.RemoveTempFiles == nil {
		removeTempFiles := false
		config.RemoveTempFiles = &removeTempFiles
		logger.Log.Warnf("You have not configured whether to remove generated temporary files, reset to default value. removeTempFiles: %v",
			*config.RemoveTempFiles)
	}

	if config.LogPath == nil {
		defaultPath := filepath.Join(os.TempDir(), fmt.Sprintf("nebula-importer-%d.log", time.Now().UnixNano()))
		config.LogPath = &defaultPath
		logger.Log.Warnf("You have not configured the log file path in: logPath, reset to default path: %s", *config.LogPath)
	}
	if !filepath.IsAbs(*config.LogPath) {
		absPath := filepath.Join(dir, *config.LogPath)
		config.LogPath = &absPath
	}

	if config.Files == nil || len(config.Files) == 0 {
		return errors.New("There is no files in configuration")
	}

	//TODO(yuyu): check each item in config.Files
	// if item is a directory, iter this directory and replace this directory config section by filename config section
	if err := config.expandDirectoryToFiles(dir); err != nil {
		logger.Log.Errorf("%s", err)
		return err
	}
	for i := range config.Files {
		if err := config.Files[i].validateAndReset(dir, fmt.Sprintf("files[%d]", i)); err != nil {
			return err
		}
	}

	return nil
}

func (config *YAMLConfig) expandDirectoryToFiles(dir string) (err error) {
	var newFiles []*File

	for _, file := range config.Files {
		files, err := file.expandFiles(dir)
		if err != nil {
			logger.Log.Errorf("error when expand file: %s", err)
			return err
		}
		for _, f := range files {
			newFiles = append(newFiles, f)
		}
	}
	config.Files = newFiles

	return err
}

func (n *NebulaPostStart) validateAndReset(prefix string) error {
	if n.AfterPeriod != nil {
		_, err := time.ParseDuration(*n.AfterPeriod)
		if err != nil {
			return err
		}
	} else {
		period := "0s"
		n.AfterPeriod = &period
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
		logger.Log.Warnf("Invalid retry option in %s.retry, reset to %d ", prefix, *n.Retry)
	}

	if n.Concurrency == nil {
		d := 10
		n.Concurrency = &d
		logger.Log.Warnf("Invalid client concurrency in %s.concurrency, reset to %d", prefix, *n.Concurrency)
	}

	if n.ChannelBufferSize == nil {
		d := 128
		n.ChannelBufferSize = &d
		logger.Log.Warnf("Invalid client channel buffer size in %s.channelBufferSize, reset to %d", prefix, *n.ChannelBufferSize)
	}

	if n.Connection == nil {
		return fmt.Errorf("Please configure the connection information in: %s.connection", prefix)
	}
	if err := n.Connection.validateAndReset(fmt.Sprintf("%s.connection", prefix)); err != nil {
		return err
	}

	if n.PostStart != nil {
		return n.PostStart.validateAndReset(fmt.Sprintf("%s.postStart", prefix))
	}
	return nil
}

func (c *NebulaClientConnection) validateAndReset(prefix string) error {
	if c.Address == nil {
		c.Address = &kDefaultConnAddr
		logger.Log.Warnf("%s.address: %s", prefix, *c.Address)
	}

	if c.User == nil {
		c.User = &kDefaultUser
		logger.Log.Warnf("%s.user: %s", prefix, *c.User)
	}

	if c.Password == nil {
		c.Password = &kDefaultPassword
		logger.Log.Warnf("%s.password: %s", prefix, *c.Password)
	}
	return nil
}

func (f *File) IsInOrder() bool {
	return (f.InOrder != nil && *f.InOrder) || (f.CSV != nil && f.CSV.WithLabel != nil && *f.CSV.WithLabel)
}

func (f *File) validateAndReset(dir, prefix string) error {
	if f.Path == nil {
		return fmt.Errorf("Please configure file path in: %s.path", prefix)
	}

	if base.HasHttpPrefix(*f.Path) {
		if _, err := url.ParseRequestURI(*f.Path); err != nil {
			return err
		}

		if _, _, err := base.ExtractFilename(*f.Path); err != nil {
			return err
		}

		if f.FailDataPath == nil {
			failDataPath := filepath.Join(os.TempDir(), fmt.Sprintf("nebula-importer-err-data-%d", time.Now().UnixNano()))
			f.FailDataPath = &failDataPath
			logger.Log.Warnf("You have not configured the failed data output file path in: %s.failDataPath, reset to tmp path: %s",
				prefix, *f.FailDataPath)
		} else {
			if !filepath.IsAbs(*f.FailDataPath) {
				absPath := filepath.Join(dir, *f.FailDataPath)
				f.FailDataPath = &absPath
			}
		}
	} else {
		if !filepath.IsAbs(*f.Path) {
			absPath := filepath.Join(dir, *f.Path)
			f.Path = &absPath
		}
		if !base.FileExists(*f.Path) {
			return fmt.Errorf("File(%s) doesn't exist", *f.Path)
		}

		if f.FailDataPath == nil {
			p := filepath.Join(filepath.Dir(*f.Path), "err", filepath.Base(*f.Path))
			f.FailDataPath = &p
			logger.Log.Warnf("You have not configured the failed data output file path in: %s.failDataPath, reset to default path: %s",
				prefix, *f.FailDataPath)
		} else {
			if !filepath.IsAbs(*f.FailDataPath) {
				absPath := filepath.Join(dir, *f.FailDataPath)
				f.FailDataPath = &absPath
			}
		}
	}

	if f.BatchSize == nil {
		f.BatchSize = &kDefaultBatchSize
		logger.Log.Infof("Invalid batch size in file(%s), reset to %d", *f.Path, *f.BatchSize)
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

func (f *File) expandFiles(dir string) ([]*File, error) {
	var files []*File
	if base.HasHttpPrefix(*f.Path) {
		if f.FailDataPath != nil {
			_, filename, err := base.ExtractFilename(*f.Path)
			if err != nil {
				return nil, err
			}
			failedDataPath := filepath.Join(*f.FailDataPath, filename)
			f.FailDataPath = &failedDataPath
			logger.Log.Infof("Failed data path: %v", failedDataPath)
		}
		files = append(files, f)
	} else {
		if !filepath.IsAbs(*f.Path) {
			absPath := filepath.Join(dir, *f.Path)
			f.Path = &absPath
		}

		fileNames, err := filepath.Glob(*f.Path)
		if err != nil || len(fileNames) == 0 {
			logger.Log.Errorf("error file path: %s", *f.Path)
			return files, err
		}

		for i := range fileNames {
			var failedDataPath *string = nil
			if f.FailDataPath != nil {
				base := filepath.Base(fileNames[i])
				tmp := filepath.Join(*f.FailDataPath, base)
				failedDataPath = &tmp
				logger.Log.Infof("Failed data path: %v", *failedDataPath)
			}
			eachConf := *f
			eachConf.Path = &fileNames[i]
			eachConf.FailDataPath = failedDataPath
			files = append(files, &eachConf)
			logger.Log.Infof("find file: %v", *eachConf.Path)
		}
	}

	return files, nil
}

func (c *CSVConfig) validateAndReset(prefix string) error {
	if c.WithHeader == nil {
		h := false
		c.WithHeader = &h
		logger.Log.Infof("%s.withHeader: %v", prefix, false)
	}

	if c.WithLabel == nil {
		l := false
		c.WithLabel = &l
		logger.Log.Infof("%s.withLabel: %v", prefix, false)
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

func (s *Schema) CollectEmptyPropsTagNames() []string {
	if !s.IsVertex() || s.Vertex == nil {
		return nil
	}
	var tagNames []string
	for _, tag := range s.Vertex.Tags {
		if len(tag.Props) == 0 {
			tagNames = append(tagNames, *tag.Name)
			continue
		}
		for _, prop := range tag.Props {
			if prop != nil {
				continue
			}
		}
		tagNames = append(tagNames, *tag.Name)
	}
	return tagNames
}

func (s *Schema) validateAndReset(prefix string) error {
	var err error = nil
	switch strings.ToLower(*s.Type) {
	case "edge":
		if s.Edge == nil {
			logger.Log.Infof("%s.edge is nil", prefix)
			s.Edge = &Edge{}
		}
		err = s.Edge.validateAndReset(fmt.Sprintf("%s.edge", prefix))
	case "vertex":
		if s.Vertex == nil {
			logger.Log.Infof("%s.vertex is nil", prefix)
			s.Vertex = &Vertex{}
		}
		err = s.Vertex.validateAndReset(fmt.Sprintf("%s.vertex", prefix))
	default:
		err = fmt.Errorf("Error schema type(%s) in %s.type only edge and vertex are supported", *s.Type, prefix)
	}
	return err
}

func (v *VID) ParseFunction(str string) (err error) {
	i := strings.Index(str, "(")
	j := strings.Index(str, ")")
	if i < 0 && j < 0 {
		v.Function = nil
		v.Type = &kDefaultVidType
	} else if i > 0 && j > i {
		strs := strings.ToLower(str[i+1 : j])
		fnType := strings.Split(strs, "+")
		if len(fnType) == 2 {
			v.Function = &fnType[0]
			v.Type = &fnType[1]
			v.Prefix = nil
		} else if len(fnType) == 3 {
			v.Function = &fnType[0]
			v.Type = &fnType[1]
			v.Prefix = &fnType[2]
		} else {
			v.Function = nil
			v.Type = &fnType[0]
			v.Prefix = nil
		}
	} else {
		err = fmt.Errorf("Invalid function format: %s", str)
	}
	return
}

func (v *VID) String(vid string) string {
	if (v.Function != nil && *v.Function != "") && (v.Prefix != nil && *v.Prefix != "") {
		return fmt.Sprintf("%s(%s+%s+%s)", vid, *v.Function, *v.Type, *v.Prefix)
	} else if (v.Function == nil || *v.Function == "") && (v.Prefix != nil && *v.Prefix != "") {
		return fmt.Sprintf("%s(%s+%s+%s)", vid, "", *v.Type, *v.Prefix)
	} else if (v.Function != nil && *v.Function != "") && (v.Prefix == nil || *v.Prefix == "") {
		return fmt.Sprintf("%s(%s+%s)", vid, *v.Function, *v.Type)
	} else {
		return fmt.Sprintf("%s(%s)", vid, *v.Type)
	}
}

func (v *VID) FormatValue(record base.Record) (string, error) {
	value, err := v.picker.Pick(record)
	if err != nil {
		return "", err
	}
	return value.Val, nil
}

func (v *VID) checkFunction(prefix string) error {
	if v.Function != nil {
		switch strings.ToLower(*v.Function) {
		case "", "hash":
		default:
			return fmt.Errorf("Invalid %s.function: %s, only following values are supported: \"\", hash", prefix, *v.Function)
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
	if v.Type != nil {
		vidType := strings.TrimSpace(strings.ToLower(*v.Type))
		if vidType != "string" && vidType != "int" {
			return fmt.Errorf("vid type must be `string' or `int', now is %s", vidType)
		}
	} else {
		v.Type = &kDefaultVidType
		logger.Log.Warnf("Not set %s.Type, reset to default value `%s'", prefix, *v.Type)
	}

	return v.InitPicker()
}

func (v *VID) InitPicker() error {
	pickerConfig := picker.Config{
		Type:     *v.Type,
		Function: v.Function,
	}

	hasPrefix := v.Prefix != nil && *v.Prefix != ""

	if len(v.ConcatItems) > 0 {
		if hasPrefix {
			pickerConfig.ConcatItems.AddConstant(*v.Prefix)
		}
		for i, item := range v.ConcatItems {
			switch val := item.(type) {
			case int:
				pickerConfig.ConcatItems.AddIndex(val)
			case string:
				pickerConfig.ConcatItems.AddConstant(val)
			default:
				return fmt.Errorf("ConcatItems only support int or string, but the %d is %v", i, val)
			}
		}
	} else if hasPrefix {
		pickerConfig.ConcatItems.AddConstant(*v.Prefix)
		pickerConfig.ConcatItems.AddIndex(*v.Index)
	} else {
		pickerConfig.Indices = []int{*v.Index}
	}

	if (v.Function == nil || *v.Function == "") && strings.EqualFold(*v.Type, "int") {
		pickerConfig.CheckOnPost = func(v *picker.Value) error {
			return checkVidFormat(v.Val, true)
		}
	}

	var err error
	v.picker, err = pickerConfig.Build()
	return err
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

func checkVidFormat(vid string, isInt bool) error {
	if isInt {
		if utils.IsInteger(vid) {
			return nil
		}
		vidLen := len(vid)
		if vidLen > 8 /* hash("") */ && strings.HasSuffix(vid, "\")") && strings.HasPrefix(vid, "hash(\"") {
			return nil
		}
		return fmt.Errorf("Invalid vid format: " + vid)
	}
	return nil
}

func (e *Edge) FormatValues(record base.Record) (string, error) {
	var cells []string
	for i, prop := range e.Props {
		if c, err := prop.FormatValue(record); err != nil {
			return "", fmt.Errorf("edge: %s, column: %d, error: %v", e.String(), i, err)
		} else {
			cells = append(cells, c)
		}
	}
	rank := ""
	if e.Rank != nil && e.Rank.Index != nil {
		rank = fmt.Sprintf("@%s", record[*e.Rank.Index])
	}
	srcVID, err := e.SrcVID.FormatValue(record)
	if err != nil {
		return "", err
	}
	dstVID, err := e.DstVID.FormatValue(record)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(" %s->%s%s:(%s) ", srcVID, dstVID, rank, strings.Join(cells, ",")), nil
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

	if e.SrcVID == nil {
		index := 0
		e.SrcVID = &VID{Index: &index, Type: &kDefaultVidType}
	}
	if err := e.SrcVID.validateAndReset(fmt.Sprintf("%s.srcVID", prefix), 0); err != nil {
		return err
	}

	if e.DstVID == nil {
		index := 1
		e.DstVID = &VID{Index: &index, Type: &kDefaultVidType}
	}
	if err := e.DstVID.validateAndReset(fmt.Sprintf("%s.dstVID", prefix), 1); err != nil {
		return err
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
			logger.Log.Errorf("prop %d of edge %s is nil", i, *e.Name)
		}
	}
	return nil
}

func (v *Vertex) FormatValues(record base.Record) (string, error) {
	var cells []string
	for _, tag := range v.Tags {
		str, noProps, err := tag.FormatValues(record)
		if err != nil {
			return "", err
		}
		if !noProps {
			cells = append(cells, str)
		}
	}
	vid, err := v.VID.FormatValue(record)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(" %s: (%s)", vid, strings.Join(cells, ",")), nil
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
	if v.VID == nil {
		index := 0
		v.VID = &VID{Index: &index, Type: &kDefaultVidType}
	}
	if err := v.VID.validateAndReset(fmt.Sprintf("%s.vid", prefix), 0); err != nil {
		return err
	}
	j := 1
	for i := range v.Tags {
		if v.Tags[i] != nil {
			if err := v.Tags[i].validateAndReset(fmt.Sprintf("%s.tags[%d]", prefix, i), j); err != nil {
				return err
			}
			j = j + len(v.Tags[i].Props)
		} else {
			logger.Log.Errorf("tag %d is nil", i)
		}
	}
	return nil
}

func (p *Prop) IsStringType() bool {
	return strings.ToLower(*p.Type) == "string"
}

func (p *Prop) IsDateOrTimeType() bool {
	t := strings.ToLower(*p.Type)
	return t == "date" || t == "time" || t == "datetime" || t == "timestamp"
}

func (p *Prop) IsTimestampType() bool {
	t := strings.ToLower(*p.Type)
	return t == "timestamp"
}

func (p *Prop) IsGeographyType() bool {
	t := strings.ToLower(*p.Type)
	return t == "geography" || t == "geography(point)" || t == "geography(linestring)" || t == "geography(polygon)"
}

func (p *Prop) FormatValue(record base.Record) (string, error) {
	value, err := p.picker.Pick(record)
	if err != nil {
		return "", err
	}
	return value.Val, nil
}

func (p *Prop) getValue(record base.Record) (string, bool, error) {
	if p.Index != nil && *p.Index >= len(record) {
		return "", false, fmt.Errorf("Prop index %d out range %d of record(%v)", *p.Index, len(record), record)
	}

	r := record[*p.Index]
	if !p.Nullable {
		return r, false, nil
	}

	if r != p.NullValue {
		return r, false, nil
	}

	for _, idx := range p.AlternativeIndices {
		if idx >= len(record) {
			return "", false, fmt.Errorf("Prop index %d out range %d of record(%v)", idx, len(record), record)
		}
		r = record[idx]
		if r != p.NullValue {
			return r, false, nil
		}
	}
	if p.DefaultValue != nil {
		return *p.DefaultValue, false, nil
	}
	return dbNULL, true, nil
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
			return fmt.Errorf("Invalid prop index: %d, name: %s, type: %s", *p.Index, *p.Name, *p.Type)
		}
	}
	return p.InitPicker()
}

func (p *Prop) InitPicker() error {
	pickerConfig := picker.Config{
		Indices: []int{*p.Index},
		Type:    *p.Type,
	}

	if p.Nullable {
		pickerConfig.Nullable = func(s string) bool {
			return s == p.NullValue
		}
		pickerConfig.NullValue = dbNULL
		if len(p.AlternativeIndices) > 0 {
			pickerConfig.Indices = append(pickerConfig.Indices, p.AlternativeIndices...)
		}
		pickerConfig.DefaultValue = p.DefaultValue
	}

	var err error
	p.picker, err = pickerConfig.Build()
	return err
}

func (t *Tag) FormatValues(record base.Record) (string, bool, error) {
	var cells []string
	noProps := true
	for _, p := range t.Props {
		if p == nil {
			continue
		}
		noProps = false
		if c, err := p.FormatValue(record); err != nil {
			return "", noProps, fmt.Errorf("tag: %v, error: %v", *t, err)
		} else {
			cells = append(cells, c)
		}
	}
	return strings.Join(cells, ","), noProps, nil
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
			logger.Log.Errorf("prop %d of tag %s is nil", i, *t.Name)
		}
	}
	return nil
}
