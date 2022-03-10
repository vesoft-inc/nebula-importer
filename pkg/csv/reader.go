package csv

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
	"io"
	"os"
)

type CSVReader struct {
	CSVConfig *config.CSVConfig
	reader    *csv.Reader
	lineNum   uint64
	rr        *recordReader
	br        *bufio.Reader
	totalBytes int64
	initComplete bool
}

type recordReader struct {
	io.Reader
	remainingBytes int
}

func (r *recordReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.remainingBytes += n
	return
}

func (r *CSVReader) InitReader(file *os.File) {
	r.rr = &recordReader{
		Reader: file,
	}
	r.br = bufio.NewReader(r.rr)
	r.reader = csv.NewReader(r.br)
	if r.CSVConfig.Delimiter != nil {
		d := []rune(*r.CSVConfig.Delimiter)
		if len(d) > 0 {
			r.reader.Comma = d[0]
			logger.Infof("The delimiter of %s is %#U", file.Name(), r.reader.Comma)
		}
	}
	stat, err := file.Stat()
	if err != nil {
		logger.Infof("The stat of %s is wrong, %s", file.Name(), err)
	}
	r.totalBytes = stat.Size()
	r.initComplete = true
}

func (r *CSVReader) ReadLine() (base.Data, error) {
	line, err := r.reader.Read()

	if err != nil {
		return base.Data{}, err
	}

	r.lineNum++
	n := r.rr.remainingBytes - r.br.Buffered()
	r.rr.remainingBytes -= n

	if *r.CSVConfig.WithHeader && r.lineNum == 1 {
		if *r.CSVConfig.WithLabel {
			return base.HeaderData(line[1:], n), nil
		} else {
			return base.HeaderData(line, n), nil
		}
	}

	if *r.CSVConfig.WithLabel {
		switch line[0] {
		case "+":
			return base.InsertData(line[1:], n), nil
		case "-":
			return base.DeleteData(line[1:], n), nil
		default:
			return base.Data{
				Bytes: n,
			}, fmt.Errorf("Invalid label: %s", line[0])
		}
	} else {
		return base.InsertData(line, n), nil
	}
}

func (r *CSVReader) TotalBytes() (int64) {
	for {
		if r.initComplete {
			return r.totalBytes
		}
	}
}
