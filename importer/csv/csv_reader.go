package nebula_csv_importer

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	importer "github.com/yixinglu/nebula-importer/importer"
)

type CSVReader struct {
	Schema importer.Schema
}

func (r *CSVReader) InitFileReader(path string, stmtChs []chan importer.Stmt, doneCh chan<- bool) {
	for _, ch := range stmtChs {
		ch <- importer.Stmt{
			Stmt: "USE ?;",
			Data: []interface{}{r.Schema.Space},
		}
	}

	log.Printf("Start to read CSV data file: %s", path)

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))

	idx, len := 0, len(stmtChs)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			line, err := reader.Read()
			if err == io.EOF {
				doneCh <- true
				wg.Done()
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			stmtChs[idx%len] <- r.MakeStmt(line)
			idx++
		}
	}()
	wg.Wait()
}

func (r *CSVReader) MakeStmt(record []string) importer.Stmt {
	schemaType := strings.ToUpper(r.Schema.Type)

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("INSERT %s %s(", schemaType, r.Schema.Name))

	for idx, prop := range r.Schema.Props {
		builder.WriteString(prop.Name)
		if idx < len(r.Schema.Props)-1 {
			builder.WriteString(",")
		}
	}
	builder.WriteString(") VALUES ")

	fromIndex := writeVID(schemaType, record, &builder)

	builder.WriteString(":(")
	for idx := range record[fromIndex:] {
		builder.WriteString("?")
		if idx < len(record[fromIndex:])-1 {
			builder.WriteString(",")
		}
	}
	builder.WriteString(");")

	data := make([]interface{}, len(record))
	for i := range record {
		data[i] = record[i]
	}

	return importer.Stmt{
		Stmt: builder.String(),
		Data: data,
	}
}

func writeVID(schemaType string, record []string, builder *strings.Builder) int {
	builder.WriteString("?")
	if schemaType == "EDGE" {
		builder.WriteString(" -> ?")
		return 2
	}
	return 1
}
