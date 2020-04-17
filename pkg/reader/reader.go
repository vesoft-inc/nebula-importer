package reader

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
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
		return nil, fmt.Errorf("Wrong file type: %s", file.Type)
	}
}

func (r *FileReader) startLog() {
	logger.Infof("Start to read file(%d): %s, schema: < %s >", r.FileIdx, *r.File.Path, r.BatchMgr.Schema.String())
}

func (r *FileReader) Stop() {
	r.StopFlag = true
}

func extractFilenameFromURL(uri string) (string, error) {
	base := path.Base(uri)
	index := strings.Index(base, "?")
	return url.QueryUnescape(uri[:index])
}

func (r *FileReader) handleDataFile() (*string, error) {
	if _, err := url.ParseRequestURI(*r.File.Path); err != nil {
		// This is a local path
		return r.File.Path, nil
	}

	// Download data file from internet to `/tmp` directory and return the path
	filename, err := extractFilenameFromURL(*r.File.Path)
	if err != nil {
		return nil, err
	}

	file, err := ioutil.TempFile("", filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	resp, err := http.Get(*r.File.Path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, err
	}

	filepath := file.Name()
	return &filepath, nil
}

func (r *FileReader) Read() error {
	filePath, err := r.handleDataFile()
	if err != nil {
		return err
	}
	file, err := os.Open(*filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	r.DataReader.InitReader(file)

	lineNum, numErrorLines := 0, 0

	if !r.WithHeader {
		r.startLog()
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
				r.startLog()
			} else {
				if *r.File.InOrder {
					err = r.BatchMgr.Add(data)
				} else {
					idx := lineNum % len(r.BatchMgr.Batches)
					r.BatchMgr.Batches[idx].Add(data)
				}
			}
		}

		if err != nil {
			logger.Errorf("Fail to read line %d, error: %s", lineNum, err.Error())
			numErrorLines++
		}

		if r.StopFlag || (r.File.Limit != nil && *r.File.Limit > 0 && *r.File.Limit <= lineNum) {
			break
		}
	}

	r.BatchMgr.Done()
	logger.Infof("Total lines of file(%s) is: %d, error lines: %d", *r.File.Path, lineNum, numErrorLines)

	return nil
}
