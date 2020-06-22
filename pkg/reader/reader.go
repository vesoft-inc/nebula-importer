package reader

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/csv"
	"github.com/vesoft-inc/nebula-importer/pkg/errors"
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
	localFile   bool
	cleanup     bool
	WithHeader  bool
	DataReader  DataFileReader
	Concurrency int
	BatchMgr    *BatchMgr
	StopFlag    bool
}

func New(fileIdx int, file *config.File, cleanup bool, clientRequestChs []chan base.ClientRequest, errCh chan<- base.ErrData) (*FileReader, error) {
	switch strings.ToLower(*file.Type) {
	case "csv":
		r := csv.CSVReader{CSVConfig: file.CSV}
		reader := FileReader{
			FileIdx:    fileIdx,
			DataReader: &r,
			File:       file,
			WithHeader: *file.CSV.WithHeader,
			StopFlag:   false,
			cleanup:    cleanup,
		}
		reader.BatchMgr = NewBatchMgr(file.Schema, *file.BatchSize, clientRequestChs, errCh)
		if !reader.WithHeader {
			if err := reader.BatchMgr.InitSchema(strings.Split(file.Schema.String(), ",")); err != nil {
				return nil, err
			}
		}
		return &reader, nil
	default:
		return nil, fmt.Errorf("Wrong file type: %s", file.Type)
	}
}

func (r *FileReader) startLog() {
	fpath, _ := base.FormatFilePath(*r.File.Path)
	logger.Infof("Start to read file(%d): %s, schema: < %s >", r.FileIdx, fpath, r.BatchMgr.Schema.String())
}

func (r *FileReader) Stop() {
	r.StopFlag = true
}

func (r *FileReader) prepareDataFile() (*string, error) {
	local, filename, err := base.ExtractFilename(*r.File.Path)
	r.localFile = local
	if r.localFile {
		// Do nothing for local file, so it wouldn't throw any errors
		return &filename, nil
	}
	if err != nil {
		return nil, errors.Wrap(errors.DownloadError, err)
	}

	if _, err := url.ParseRequestURI(*r.File.Path); err != nil {
		return nil, errors.Wrap(errors.DownloadError, err)
	}

	// Download data file from internet to `/tmp` directory and return the path
	file, err := ioutil.TempFile("", fmt.Sprintf("*_%s", filename))
	if err != nil {
		return nil, errors.Wrap(errors.UnknownError, err)
	}
	defer file.Close()

	client := http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(*r.File.Path)
	if err != nil {
		return nil, errors.Wrap(errors.DownloadError, err)
	}
	defer resp.Body.Close()

	n, err := io.Copy(file, resp.Body)
	if err != nil {
		return nil, errors.Wrap(errors.DownloadError, err)
	}

	filepath := file.Name()

	fpath, _ := base.FormatFilePath(*r.File.Path)
	logger.Infof("File(%s) has been downloaded to \"%s\", size: %d", fpath, filepath, n)

	return &filepath, nil
}

func (r *FileReader) Read() error {
	filePath, err := r.prepareDataFile()
	if err != nil {
		return err
	}
	file, err := os.Open(*filePath)
	if err != nil {
		return errors.Wrap(errors.ConfigError, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Errorf("Fail to close opened data file: %s", *filePath)
			return
		}
		if !r.localFile && r.cleanup {
			if err := os.Remove(*filePath); err != nil {
				logger.Errorf("Fail to remove temp data file: %s", *filePath)
			} else {
				logger.Infof("Temp downloaded data file has been removed: %s", *filePath)
			}
		}
	}()

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
				err = r.BatchMgr.InitSchema(data.Record)
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
			fpath, _ := base.FormatFilePath(*r.File.Path)
			logger.Errorf("Fail to read file(%s) line %d, error: %s", fpath, lineNum, err.Error())
			numErrorLines++
		}

		if r.StopFlag || (r.File.Limit != nil && *r.File.Limit > 0 && *r.File.Limit <= lineNum) {
			break
		}
	}

	r.BatchMgr.Done()
	fpath, _ := base.FormatFilePath(*r.File.Path)
	logger.Infof("Total lines of file(%s) is: %d, error lines: %d", fpath, lineNum, numErrorLines)

	return nil
}
