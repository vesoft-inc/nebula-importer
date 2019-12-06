package reader

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/csv"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type DataFileReader interface {
	InitReader(*os.File)
	ReadLine() (base.Data, error)
}

// FIXME: private fields
type FileReader struct {
	File        *config.File
	DataReader  DataFileReader
	Concurrency int
	BatchMgr    *BatchMgr
}

func New(file *config.File, clientRequestChs []chan base.ClientRequest, errCh chan<- base.ErrData) (*FileReader, error) {
	switch strings.ToLower(*file.Type) {
	case "csv":
		r := csv.CSVReader{CSVConfig: file.CSV}
		reader := FileReader{
			DataReader: &r,
			File:       file,
		}
		reader.BatchMgr = NewBatchMgr(file.Schema, *file.BatchSize, clientRequestChs, errCh)
		if !*file.CSV.WithHeader {
			reader.BatchMgr.InitSchema(strings.Split(file.Schema.String(), ","))
		}
		return &reader, nil
	default:
		return nil, fmt.Errorf("Wrong file type: %s", file.Type)
	}
}

func (r *FileReader) Read() error {
	file, err := os.Open(*r.File.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	r.DataReader.InitReader(file)

	lineNum, numErrorLines := 0, 0

	logger.Infof("Start to read file: %s", *r.File.Path)

	for {
		data, err := r.DataReader.ReadLine()
		if err == io.EOF {
			r.BatchMgr.Done()
			logger.Infof("Total lines of file(%s) is: %d, error lines: %d, schema: <%s>", *r.File.Path, lineNum, numErrorLines, r.BatchMgr.Schema.String())
			break
		}

		lineNum++

		if err == nil {
			if data.Type == base.HEADER {
				r.BatchMgr.InitSchema(data.Record)
			} else {
				err = r.BatchMgr.Add(data)
			}
		}

		if err != nil {
			logger.Errorf("Fail to read line %d, error: %s", lineNum, err.Error())
			numErrorLines++
		}
	}

	return nil
}
