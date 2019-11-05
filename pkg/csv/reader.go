package csv

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/yixinglu/nebula-importer/pkg/base"
	"github.com/yixinglu/nebula-importer/pkg/config"
)

type CSVReader struct {
	File    config.File
	DataChs []chan base.Data
}

func (r *CSVReader) Read() {
	log.Printf("\nStart to read CSV data file: %s", r.File.Path)

	file, err := os.Open(r.File.Path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.ReuseRecord = true

	lineNum, numErrorLines, length := 0, 0, len(r.DataChs)

	re := regexp.MustCompile(`^[+-]?\d+$`)

	for {
		line, err := reader.Read()
		if err == io.EOF {
			// Notify all data channels to finish
			for i := range r.DataChs {
				r.DataChs[i] <- base.FinishData()
			}
			log.Printf("\nTotal lines of file(%s) is: %d, error lines: %d", r.File.Path, lineNum, numErrorLines)
			break
		}

		lineNum++

		if err != nil {
			log.Printf("Fail to read line %d, error: %s", lineNum, err.Error())
			numErrorLines++
			continue
		}

		if len(line) == 0 {
			log.Printf("Line %d is empty", lineNum)
			numErrorLines++
			continue
		}

		// TODO: handle header line

		var vidIdx int = 0
		if r.File.CSV.WithLabel {
			vidIdx = 1
		}

		if len(line) <= vidIdx || !re.MatchString(line[vidIdx]) {
			log.Printf("Invalid record(%d): %v", lineNum, line)
			numErrorLines++
			continue
		}

		chanId, err := getChanId(line[vidIdx], length)
		if err != nil {
			log.Printf("Error vid: %s", line[vidIdx])
			numErrorLines++
			continue
		}

		var data base.Data
		if r.File.CSV.WithLabel {
			switch line[0] {
			case "+":
				data = base.InsertData(line[1:])
			case "-":
				data = base.DeleteData(line[1:])
			default:
				log.Printf("Invalid label: %s", line[0])
				numErrorLines++
				continue
			}
		} else {
			data = base.InsertData(line)
		}

		r.DataChs[chanId] <- data
	}
}

func getChanId(idStr string, numChans int) (int, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(id % int64(numChans)), nil
}
