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
	FileIdx     int
	File        *config.File
	WithHeader  bool
	DataReader  DataFileReader
	Concurrency int
	BatchMgr    *BatchMgr
	StopFlag    bool
}

func New(fileIdx int, file *config.File, clientRequestChs []chan base.ClientRequest, errCh chan<- base.ErrData) (*FileReader, error) {
	switch strings.ToLower(*file.Type) {
	case "csv":
		r := csv.CSVReader{CSVConfig: file.CSV}
		reader := FileReader{
			FileIdx:    fileIdx,
			DataReader: &r,
			File:       file,
			WithHeader: *file.CSV.WithHeader,
			StopFlag:   false,
		}
		reader.BatchMgr = NewBatchMgr(file.Schema, *file.BatchSize, clientRequestChs, errCh)
		if !reader.WithHeader {
			reader.BatchMgr.InitSchema(strings.Split(file.Schema.String(), ","))
		}
		return &reader, nil
	default:
		return nil, fmt.Errorf("Wrong file type: %s", *file.Type)
	}
}

func (r *FileReader) startLog(filename string) {
	logger.Infof("Start to read file(%d): %s, schema: < %s >", r.FileIdx, filename, r.BatchMgr.Schema.String())
}

func (r *FileReader) Stop() {
	r.StopFlag = true
}

func (r *FileReader) ReadFile(filename string) (lineNum int64, numErrorLines int64, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	r.DataReader.InitReader(file)

	if !r.WithHeader {
		r.startLog(filename)
	}

	for {
		data, err := r.DataReader.ReadLine()
		if err == io.EOF {
			break
		}

		lineNum++

		if err == nil {
			if data.Type == base.HEADER {
				r.BatchMgr.InitSchema(data.Record)
				r.startLog(filename)
			} else {
				if *r.File.InOrder {
					err = r.BatchMgr.Add(data)
				} else {
					idx := lineNum % int64(len(r.BatchMgr.Batches))
					r.BatchMgr.Batches[idx].Add(data)
				}
			}
		}

		if err != nil {
			logger.Errorf("Fail to read line %d, error: %s", lineNum, err.Error())
			numErrorLines++
		}

		if r.StopFlag || (r.File.Limit != nil && *r.File.Limit > 0 && int64(*r.File.Limit) <= lineNum) {
			break
		}
	}

	return
}

func (r *FileReader) Read() error {
	var lineNumTotal int64
	var numErrorLinesTotal int64
	for _, filename := range r.File.Paths {
		lineNum, numErrorLines, err := r.ReadFile(filename)
		if err != nil {
			return err
		}
		logger.Infof("Total lines of file(%s) is: %d, error lines: %d", filename, lineNum, numErrorLines)
		lineNumTotal = lineNumTotal + lineNum
		numErrorLinesTotal = numErrorLinesTotal + numErrorLines
	}

	r.BatchMgr.Done()
	logger.Infof("Total lines of path(%s) is: %d, error lines: %d", *r.File.Path, lineNumTotal, numErrorLinesTotal)
	return nil
}
