package csv

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type CSVReader struct {
	CSVConfig *config.CSVConfig
	reader    *csv.Reader
	lineNum   uint64
}

func (r *CSVReader) InitReader(file *os.File) {
	r.reader = csv.NewReader(bufio.NewReader(file))
	if r.CSVConfig.Delimiter != nil {
		d := []rune(*r.CSVConfig.Delimiter)
		if len(d) > 0 {
			r.reader.Comma = d[0]
			logger.Infof("The delimiter of %s is %#U", file.Name(), r.reader.Comma)
		}
	}
}

func (r *CSVReader) ReadLine() (base.Data, error) {
	line, err := r.reader.Read()

	if err != nil {
		return base.Data{}, err
	}

	r.lineNum++

	if *r.CSVConfig.WithHeader && r.lineNum == 1 {
		if *r.CSVConfig.WithLabel {
			return base.HeaderData(line[1:]), nil
		} else {
			return base.HeaderData(line), nil
		}
	}

	if *r.CSVConfig.WithLabel {
		switch line[0] {
		case "+":
			return base.InsertData(line[1:]), nil
		case "-":
			return base.DeleteData(line[1:]), nil
		default:
			return base.Data{}, fmt.Errorf("Invalid label: %s", line[0])
		}
	} else {
		return base.InsertData(line), nil
	}
}
