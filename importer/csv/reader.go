package csv

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
	file importer.File
}

func NewCSVReader(file importer.File) importer.DataFileReader {
	return &CSVReader{file: file}
}

func (r *CSVReader) InitFileReader(path string, stmtChs []chan importer.Stmt, doneCh chan<- bool) {
	for _, ch := range stmtChs {
		ch <- importer.Stmt{
			Stmt: "USE ?;",
			Data: []interface{}{r.file.Schema.Space},
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

func (r *CSVReader) convertRecords(records [][]string) [][]interface{} {
	if r.file.BatchSize != len(records) {
		log.Fatalf("records length is not equal to batch size: %d != %d", len(records), r.file.BatchSize)
	}
	data := make([][]interface{}, len(records))
	for i := range records {
		data[i] = make([]interface{}, len(records[i]))
		for j := range records[i] {
			data[i][j] = records[i][j]
		}
	}
	return data
}

func (r *CSVReader) makeVertexInsertStmtWithoutHeaderLine() string {
	var builder strings.Builder
	builder.WriteString("INSERT VERTEX ")
	numProps := 0
	for i, tag := range r.file.Schema.Vertex.Tags {
		builder.WriteString(fmt.Sprintf("%s(", tag.Name))
		for j, prop := range tag.Props {
			builder.WriteString(prop.Name)
			if j < len(tag.Props)-1 {
				builder.WriteString(",")
			} else {
				builder.WriteString(")")
			}
			numProps++
		}
		if i < len(r.file.Schema.Vertex.Tags)-1 {
			builder.WriteString(",")
		} else {
			builder.WriteString(" VALUES ")
		}
	}
	for i := 0; i < r.file.BatchSize; i++ {
		builder.WriteString(" ?: ")
		fillPropsPlaceholder(&builder, numProps, i == r.file.BatchSize-1)
	}

	return builder.String()
}

func (r *CSVReader) makeEdgeInsertStmtWithoutHeaderLine() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("INSERT EDGE %s(", r.file.Schema.Edge.Name))
	numProps := 0
	for i, prop := range r.file.Schema.Edge.Props {
		builder.WriteString(prop.Name)
		if i < len(r.file.Schema.Edge.Props)-1 {
			builder.WriteString(",")
		} else {
			builder.WriteString(")")
		}
		numProps++
	}
	builder.WriteString(" VALUES ")
	for i := 0; i < r.file.BatchSize; i++ {
		builder.WriteString("?->?: ")
		if r.file.Schema.Edge.WithRanking {
			builder.WriteString("(?)")
		}
		fillPropsPlaceholder(&builder, numProps, i == r.file.BatchSize-1)
	}
	return builder.String()
}

func fillPropsPlaceholder(builder *strings.Builder, numProps int, isEnd bool) {
	builder.WriteString("(")
	builder.WriteString(strings.TrimSuffix(strings.Repeat("?,", numProps), ","))
	builder.WriteString(")")
	if isEnd {
		builder.WriteString(";")
	} else {
		builder.WriteString(",")
	}
}

func (r *CSVReader) MakeStmt(records [][]string) importer.Stmt {
	switch strings.ToUpper(r.file.Schema.Type) {
	case "EDGE":
		return importer.Stmt{
			Stmt: r.makeEdgeInsertStmtWithoutHeaderLine(),
			Data: r.convertRecords(records),
		}
	case "VERTEX":
		return importer.Stmt{
			Stmt: r.makeVertexInsertStmtWithoutHeaderLine(),
			Data: r.convertRecords(records),
		}
	default:
		log.Fatalf("Wrong schema type: %s", r.file.Schema.Type)
		return importer.Stmt{}
	}
}
