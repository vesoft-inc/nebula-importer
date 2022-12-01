package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"gopkg.in/yaml.v2"

	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

func TestYAMLParser(t *testing.T) {
	runnerLogger := logger.NewRunnerLogger("")
	yamlConfig, err := Parse("../../examples/v2/example.yaml", runnerLogger)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range yamlConfig.Files {
		if strings.ToLower(*file.Type) != "csv" {
			t.Fatal("Error file type")
		}
		switch strings.ToLower(*file.Schema.Type) {
		case "edge":
		case "vertex":
			if file.Schema.Vertex == nil {
				continue
			}
			if len(file.Schema.Vertex.Tags) == 0 && !*file.CSV.WithHeader {
				t.Fatal("Empty tags in vertex")
			}
		default:
			t.Fatal("Error schema type")
		}
	}
}

type testYAML struct {
	Version *string `yaml:"version"`
	Files   *[]struct {
		Path *string `yaml:"path"`
	} `yaml:"files"`
}

var yamlStr string = `
version: beta
files:
  - path: ./file.csv
`

func TestTypePointer(t *testing.T) {
	ty := testYAML{}
	if err := yaml.Unmarshal([]byte(yamlStr), &ty); err != nil {
		t.Fatal(err)
	}
	t.Logf("yaml: %v, %v", *ty.Version, *ty.Files)
}

var jsonStr = `
{
  "version": "beta",
  "files": [
    { "path": "/tmp" },
    { "path": "/etc" }
	]
}
`

func TestJsonInYAML(t *testing.T) {
	conf := YAMLConfig{}
	if err := yaml.Unmarshal([]byte(jsonStr), &conf); err != nil {
		t.Fatal(err)
	}

	if conf.Version == nil || *conf.Version != "beta" {
		t.Fatal("Error version")
	}

	if conf.Files == nil || len(conf.Files) != 2 {
		t.Fatal("Error files")
	}

	paths := []string{"/tmp", "/etc"}
	for i, p := range paths {
		f := conf.Files[i]
		if f == nil || f.Path == nil || *f.Path != p {
			t.Fatalf("Error file %d path", i)
		}
	}
}

type Person struct {
	Name string `json:"name"`
}

type Man struct {
	Person
	Age int `json:"age"`
}

func TestJsonTypeEmbeding(t *testing.T) {
	man := Man{
		Person: Person{Name: "zhangsan"},
		Age:    18,
	}
	t.Logf("%v", man)
	b, _ := json.Marshal(man)
	t.Logf("%s", string(b))
}

func TestParseVersion(t *testing.T) {
	testcases := []struct {
		version string
		isError bool
	}{
		{
			version: "version: v1rc1",
			isError: false,
		},
		{
			version: "version: v1rc2",
			isError: false,
		},
		{
			version: "version: v1",
			isError: false,
		},
		{
			version: "version: v2",
			isError: false,
		},
		{
			version: "",
			isError: true,
		},
		{
			version: "version: vx",
			isError: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.version, func(t *testing.T) {
			ast := assert.New(t)

			tmpl, err := template.ParseFiles("testdata/test-parse-version.yaml")
			ast.NoError(err)

			f, err := os.CreateTemp("testdata", ".test-parse-version.yaml")
			ast.NoError(err)
			filename := f.Name()
			defer func() {
				_ = f.Close()
				_ = os.Remove(filename)
			}()

			err = tmpl.ExecuteTemplate(f, "test-parse-version.yaml", map[string]string{
				"Version": tc.version,
			})
			ast.NoError(err)

			_, err = Parse(filename, logger.NewRunnerLogger(""))
			if tc.isError {
				ast.Error(err)
			} else {
				ast.NoError(err)
			}
		})
	}
}

func TestParseAfterPeriod(t *testing.T) {
	testcases := []struct {
		afterPeriod string
		isError     bool
	}{
		{
			afterPeriod: "",
			isError:     false,
		},
		{
			afterPeriod: "afterPeriod: 1s",
			isError:     false,
		},
		{
			afterPeriod: "afterPeriod: 1m",
			isError:     false,
		},
		{
			afterPeriod: "afterPeriod: 3m4s",
			isError:     false,
		},
		{
			afterPeriod: "afterPeriod: 1ss",
			isError:     true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.afterPeriod, func(t *testing.T) {
			ast := assert.New(t)

			tmpl, err := template.ParseFiles("testdata/test-parse-after-period.yaml")
			ast.NoError(err)

			f, err := os.CreateTemp("testdata", ".test-parse-after-period.yaml")
			ast.NoError(err)
			filename := f.Name()
			defer func() {
				_ = f.Close()
				_ = os.Remove(filename)
			}()

			err = tmpl.ExecuteTemplate(f, "test-parse-after-period.yaml", map[string]string{
				"AfterPeriod": tc.afterPeriod,
			})
			ast.NoError(err)

			_, err = Parse(filename, logger.NewRunnerLogger(""))
			if tc.isError {
				ast.Error(err)
			} else {
				ast.NoError(err)
			}
		})
	}
}

func TestParseLogPath(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	testcases := []struct {
		logPath    string
		isRelative bool
		clean      func()
	}{
		{
			logPath: "",
		},
		{
			logPath:    "logPath: ./nebula-importer.log",
			isRelative: true,
		},
		{
			logPath:    "logPath: ./not-exists/nebula-importer.log",
			isRelative: true,
		},
		{
			logPath: fmt.Sprintf("logPath: %s/nebula-importer.log", tmpdir),
		},
		{
			logPath: fmt.Sprintf("logPath: %s/not-exists/nebula-importer.log", tmpdir),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.logPath, func(t *testing.T) {
			ast := assert.New(t)

			tmpl, err := template.ParseFiles("testdata/test-parse-log-path.yaml")
			ast.NoError(err)

			f, err := os.CreateTemp("testdata", ".test-parse-log-path.yaml")
			ast.NoError(err)
			filename := f.Name()
			defer func() {
				_ = f.Close()
				_ = os.Remove(filename)
			}()

			err = tmpl.ExecuteTemplate(f, "test-parse-log-path.yaml", map[string]string{
				"LogPath": tc.logPath,
			})
			ast.NoError(err)

			c, err := Parse(filename, logger.NewRunnerLogger(""))
			ast.NoError(err)
			ast.NotNil(c.LogPath)
			ast.Truef(filepath.IsAbs(*c.LogPath), "%s is abs path", *c.LogPath)

			logContent := []string{"first log", "second log"}
			for i, s := range logContent {
				runnerLogger := logger.NewRunnerLogger(*c.LogPath)
				ast.FileExists(*c.LogPath)
				runnerLogger.Error(s)

				// test first create and append
				for j := 0; j <= i; j++ {
					content, err := os.ReadFile(*c.LogPath)
					ast.NoError(err)
					ast.Contains(string(content), logContent[i])
				}
			}

			if tc.isRelative {
				removePath := *c.LogPath
				if strings.Contains(*c.LogPath, "/not-exists/") {
					removePath = filepath.Dir(removePath)
				}
				_ = os.RemoveAll(removePath)
			}
		})
	}
}

func TestParseNoFiles(t *testing.T) {
	_, err := Parse("./testdata/test-parse-no-files.yaml", logger.NewRunnerLogger(""))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no files")
}

func TestVidType(t *testing.T) {
	testcases := []struct {
		typ       string
		isSupport bool
	}{
		{
			typ:       "int",
			isSupport: true,
		},
		{
			typ:       "INT",
			isSupport: true,
		},
		{
			typ:       "iNt",
			isSupport: true,
		},
		{
			typ:       " iNt ",
			isSupport: true,
		},
		{
			typ:       "string",
			isSupport: true,
		},
		{
			typ:       "aaa",
			isSupport: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.typ, func(t *testing.T) {
			ast := assert.New(t)
			vid := VID{
				Type: &tc.typ,
			}
			err := vid.validateAndReset("", 0)
			if tc.isSupport {
				ast.NoError(err)
			} else {
				ast.Error(err)
				ast.Contains(err.Error(), "vid type must be")
			}
		})
	}
}

func TestVidFormatValue(t *testing.T) {
	var (
		idx0    = 0
		idx1    = 1
		fHash   = "hash"
		tInt    = "int"
		tString = "string"
		prefix  = "p_"
	)
	testcases := []struct {
		name          string
		vid           VID
		record        base.Record
		want          string
		wantErrString string
	}{
		{
			name: "index out of range",
			vid: VID{
				Index: &idx1,
			},
			want:          "",
			record:        base.Record{""},
			wantErrString: "out of range record length",
		},
		{
			name: "type string",
			vid: VID{
				Index: &idx0,
				Type:  &tString,
			},
			record: base.Record{"str"},
			want:   "\"str\"",
		},
		{
			name: "type int",
			vid: VID{
				Index: &idx0,
				Type:  &tInt,
			},
			record: base.Record{"1"},
			want:   "1",
		},
		{
			name: "type int d",
			vid: VID{
				Index: &idx0,
				Type:  &tInt,
			},
			record: base.Record{"1"},
			want:   "1",
		},
		{
			name: "type int 0d",
			vid: VID{
				Index: &idx1,
				Type:  &tInt,
			},
			record: base.Record{"", "070"},
			want:   "070",
		},
		{
			name: "type int 0x",
			vid: VID{
				Index: &idx0,
				Type:  &tInt,
			},
			record: base.Record{"0x0F"},
			want:   "0x0F",
		},
		{
			name: "type int 0X",
			vid: VID{
				Index: &idx0,
				Type:  &tInt,
			},
			record: base.Record{"0XF0"},
			want:   "0XF0",
		},
		{
			name: "type int format err",
			vid: VID{
				Index: &idx0,
				Type:  &tInt,
			},
			record:        base.Record{"F0"},
			want:          "",
			wantErrString: "Invalid vid format",
		},
		{
			name: "function hash",
			vid: VID{
				Index:    &idx0,
				Function: &fHash,
			},
			record: base.Record{"str"},
			want:   "hash(\"str\")",
		},
		{
			name: "prefix",
			vid: VID{
				Index:  &idx0,
				Type:   &tString,
				Prefix: &prefix,
			},
			record: base.Record{"str"},
			want:   prefix + "str",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ast := assert.New(t)
			str, err := tc.vid.FormatValue(tc.record)
			if tc.wantErrString != "" {
				ast.Error(err)
				ast.Contains(err.Error(), tc.wantErrString)
			} else {
				ast.NoError(err)
				ast.Contains(str, tc.want)
			}
		})
	}
}

func TestPropType(t *testing.T) {
	testcases := []struct {
		typ       string
		isSupport bool
	}{
		{
			typ:       "int",
			isSupport: true,
		},
		{
			typ:       "INT",
			isSupport: true,
		},
		{
			typ:       "iNt",
			isSupport: true,
		},
		{
			typ:       "string",
			isSupport: true,
		},
		{
			typ:       "float",
			isSupport: true,
		},
		{
			typ:       "double",
			isSupport: true,
		},
		{
			typ:       "bool",
			isSupport: true,
		},
		{
			typ:       "date",
			isSupport: true,
		},
		{
			typ:       "time",
			isSupport: true,
		},
		{
			typ:       "datetime",
			isSupport: true,
		},
		{
			typ:       "timestamp",
			isSupport: true,
		},
		{
			typ:       "geography",
			isSupport: true,
		},
		{
			typ:       "geography(point)",
			isSupport: true,
		},
		{
			typ:       "geography(linestring)",
			isSupport: true,
		},
		{
			typ:       "geography(polygon)",
			isSupport: true,
		},
		{
			typ:       "aaa",
			isSupport: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.typ, func(t *testing.T) {
			ast := assert.New(t)
			prop := Prop{
				Type: &tc.typ,
			}
			err := prop.validateAndReset("", 0)
			if tc.isSupport {
				ast.NoError(err)
			} else {
				ast.Error(err)
				ast.Contains(err.Error(), "Error property type")
			}
		})
	}
}

func TestPropFormatValue(t *testing.T) {
	var (
		idx0                 = 0
		idx1                 = 1
		tBool                = "bool"
		tInt                 = "int"
		tFloat               = "float"
		tDouble              = "double"
		tString              = "string"
		tTime                = "time"
		tTimestamp           = "timestamp"
		tDate                = "date"
		tDatetime            = "datetime"
		tGeography           = "geography"
		tGeographyPoint      = "geography(point)"
		tGeographyLineString = "geography(linestring)"
		tGeographyPolygon    = "geography(polygon)"
	)

	testcases := []struct {
		name          string
		prop          Prop
		record        base.Record
		want          string
		wantErrString string
	}{
		{
			name: "index out of range",
			prop: Prop{
				Index: &idx1,
			},
			want:          "",
			record:        base.Record{""},
			wantErrString: "out range",
		},
		{
			name: "type bool",
			prop: Prop{
				Index: &idx0,
				Type:  &tBool,
			},
			record: base.Record{"false"},
			want:   "false",
		},
		{
			name: "type int",
			prop: Prop{
				Index: &idx0,
				Type:  &tInt,
			},
			record: base.Record{"1"},
			want:   "1",
		},
		{
			name: "type float",
			prop: Prop{
				Index: &idx0,
				Type:  &tFloat,
			},
			record: base.Record{"1.1"},
			want:   "1.1",
		},
		{
			name: "type double",
			prop: Prop{
				Index: &idx0,
				Type:  &tDouble,
			},
			record: base.Record{"2.2"},
			want:   "2.2",
		},
		{
			name: "type string",
			prop: Prop{
				Index: &idx0,
				Type:  &tString,
			},
			record: base.Record{"str"},
			want:   "\"str\"",
		},
		{
			name: "type time",
			prop: Prop{
				Index: &idx0,
				Type:  &tTime,
			},
			record: base.Record{"18:38:23.284"},
			want:   "time(\"18:38:23.284\")",
		},
		{
			name: "type timestamp",
			prop: Prop{
				Index: &idx0,
				Type:  &tTimestamp,
			},
			record: base.Record{"2020-01-11T19:28:23"},
			want:   "timestamp(\"2020-01-11T19:28:23\")",
		},
		{
			name: "type date",
			prop: Prop{
				Index: &idx0,
				Type:  &tDate,
			},
			record: base.Record{"2020-01-02"},
			want:   "date(\"2020-01-02\")",
		},
		{
			name: "type datetime",
			prop: Prop{
				Index: &idx0,
				Type:  &tDatetime,
			},
			record: base.Record{"2020-01-11T19:28:23.284"},
			want:   "datetime(\"2020-01-11T19:28:23.284\")",
		},
		{
			name: "type geography",
			prop: Prop{
				Index: &idx0,
				Type:  &tGeography,
			},
			record: base.Record{"Polygon((-85.1 34.8,-80.7 28.4,-76.9 34.9,-85.1 34.8))"},
			want:   "ST_GeogFromText(\"Polygon((-85.1 34.8,-80.7 28.4,-76.9 34.9,-85.1 34.8))\")",
		},
		{
			name: "type geography(point)",
			prop: Prop{
				Index: &idx0,
				Type:  &tGeographyPoint,
			},
			record: base.Record{"Point(0.0 0.0)"},
			want:   "ST_GeogFromText(\"Point(0.0 0.0)\")",
		},
		{
			name: "type geography(linestring)",
			prop: Prop{
				Index: &idx0,
				Type:  &tGeographyLineString,
			},
			record: base.Record{"linestring(0 1, 179.99 89.99)"},
			want:   "ST_GeogFromText(\"linestring(0 1, 179.99 89.99)\")",
		},
		{
			name: "type geography(polygon)",
			prop: Prop{
				Index: &idx0,
				Type:  &tGeographyPolygon,
			},
			record: base.Record{"polygon((0 1, 2 4, 3 5, 4 9, 0 1))"},
			want:   "ST_GeogFromText(\"polygon((0 1, 2 4, 3 5, 4 9, 0 1))\")",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ast := assert.New(t)
			str, err := tc.prop.FormatValue(tc.record)
			if tc.wantErrString != "" {
				ast.Error(err)
				ast.Contains(err.Error(), tc.wantErrString)
			} else {
				ast.NoError(err)
				ast.Contains(str, tc.want)
			}
		})
	}
}

func TestParseFunction(t *testing.T) {
	var (
		tString = "string"
		tInt    = "int"
		fHash   = "hash"
		prefix  = "prefix"
	)
	testcases := []struct {
		str       string
		vid       VID
		isSupport bool
	}{
		{
			str: ":VID",
			vid: VID{
				Type: &tString,
			},
			isSupport: true,
		},
		{
			str: ":VID(string)",
			vid: VID{
				Type: &tString,
			},
			isSupport: true,
		},
		{
			str: ":VID(int)",
			vid: VID{
				Type: &tInt,
			},
			isSupport: true,
		},
		{
			str: ":VID(hash+int)",
			vid: VID{
				Function: &fHash,
				Type:     &tInt,
			},
			isSupport: true,
		},
		{
			str: ":VID(hash+int+prefix)",
			vid: VID{
				Function: &fHash,
				Type:     &tInt,
				Prefix:   &prefix,
			},
			isSupport: true,
		},
		{
			str:       ":VID(",
			isSupport: false,
		},
		{
			str:       ":VID)int(",
			isSupport: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.str, func(t *testing.T) {
			ast := assert.New(t)
			vid := VID{}
			err := vid.ParseFunction(tc.str)
			if tc.isSupport {
				ast.NoError(err)
				ast.Equal(vid, tc.vid)
			} else {
				ast.Error(err)
				ast.Contains(err.Error(), "Invalid function format")
			}
		})
	}
}
