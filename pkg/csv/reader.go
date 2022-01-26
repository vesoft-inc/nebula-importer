package csv

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"reflect"
	"unsafe"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type CSVReader struct {
	CSVConfig *config.CSVConfig
	reader    *csv.Reader
	lineNum   uint64
	rr        *recordReader
	br        *bufio.Reader
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
		Reader: bufio.NewReader(file),
	}
	r.reader = csv.NewReader(r.rr)
	if r.CSVConfig.Delimiter != nil {
		d := []rune(*r.CSVConfig.Delimiter)
		if len(d) > 0 {
			r.reader.Comma = d[0]
			logger.Infof("The delimiter of %s is %#U", file.Name(), r.reader.Comma)
		}
	}
	rf := reflect.ValueOf(r.reader).Elem().FieldByName("r")
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	br := rf.Interface().(*bufio.Reader)
	r.br = br
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

func CountFileBytes(path string) (int64, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		logger.Errorf("count bytes fail: %s", path)
		return 0, err
	}
	stat, err := file.Stat()
	if err != nil {
		logger.Errorf("count bytes fail: %s", path)
		return 0, err
	}
	bytesCount := stat.Size()
	return bytesCount, nil
}
