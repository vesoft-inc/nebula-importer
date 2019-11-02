package csv

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/yixinglu/nebula-importer/pkg/clientmgr"
	"github.com/yixinglu/nebula-importer/pkg/config"
)

type CSVReader struct {
	File    config.File
	DataChs []chan clientmgr.Record
}

func (r *CSVReader) Read() {
	log.Printf("Start to read CSV data file: %s", r.File.Path)

	file, err := os.Open(r.File.Path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))

	lineNum, numFailedLines, len := 0, 0, len(r.DataChs)

	for {
		line, err := reader.Read()
		if err == io.EOF {
			for i := range r.DataChs {
				r.DataChs[i] <- clientmgr.DoneRecord()
			}
			log.Printf("Total lines of file(%s) is: %d, failed: %d", r.File.Path, lineNum, numFailedLines)
			lineNum, numFailedLines = 0, 0
			break
		}

		lineNum++

		if err != nil {
			log.Printf("Fail to read line %d, error: %s", lineNum, err.Error())
			numFailedLines++
			continue
		}

		if len(line) == 0 || len(line[0]) == 0 {
			log.Printf("Line %d or its vid is empty", lineNum)
			numFailedLines++
			continue
		}

		chanId, err := getChanId(line[0], len)
		if err != nil {
			log.Printf("Error vid: %s", line[0])
			numFailedLines++
			continue
		}
		r.DataChs[chanId] <- line
	}
}

func getChanId(idStr string, numChans int) (int, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return id % numChans, nil
}
