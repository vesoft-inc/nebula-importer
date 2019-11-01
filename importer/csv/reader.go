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
	if strings.ToUpper(file.Type) != "CSV" {
		log.Fatalf("Error file type: %s", file.Type)
	}
	return &CSVReader{file: file}
}

func (r *CSVReader) InitFileReader(stmtChs []chan importer.Stmt, doneCh chan<- bool) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for _, ch := range stmtChs {
			data := make([][]interface{}, 1)
			data[0] = make([]interface{}, 1)
			data[0][0] = r.file.Schema.Space
			ch <- importer.Stmt{
				Stmt: "USE ?;",
				Data: data,
			}
		}

		log.Printf("Start to read CSV data file: %s", r.file.Path)

		file, err := os.Open(r.file.Path)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		reader := csv.NewReader(bufio.NewReader(file))

		numBatches, batchSize, len := 0, 0, len(stmtChs)
		lines := make([][]string, r.file.BatchSize)

		for {
			line, err := reader.Read()
			if err == io.EOF {
				if batchSize > 0 {
					stmtChs[numBatches%len] <- r.MakeStmt(lines, batchSize)
				}
				doneCh <- true
				wg.Done()
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			if batchSize < r.file.BatchSize {
				lines[batchSize] = line
				batchSize++
			} else {
				batchSize = 0
				numBatches++
				stmtChs[numBatches%len] <- r.MakeStmt(lines, r.file.BatchSize)
			}
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

func (r *CSVReader) makeVertexInsertStmtWithoutHeaderLine(batchSize int) string {
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
	for i := 0; i < batchSize; i++ {
		builder.WriteString(" ?: ")
		fillPropsPlaceholder(&builder, numProps, i == batchSize-1)
	}

	return builder.String()
}

func (r *CSVReader) makeEdgeInsertStmtWithoutHeaderLine(batchSize int) string {
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
	for i := 0; i < batchSize; i++ {
		builder.WriteString("?->?: ")
		if r.file.Schema.Edge.WithRanking {
			builder.WriteString("(?)")
		}
		fillPropsPlaceholder(&builder, numProps, i == batchSize-1)
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

func (r *CSVReader) MakeStmt(records [][]string, batchSize int) importer.Stmt {
	switch strings.ToUpper(r.file.Schema.Type) {
	case "EDGE":
		return importer.Stmt{
			Stmt: r.makeEdgeInsertStmtWithoutHeaderLine(batchSize),
			Data: r.convertRecords(records),
		}
	case "VERTEX":
		return importer.Stmt{
			Stmt: r.makeVertexInsertStmtWithoutHeaderLine(batchSize),
			Data: r.convertRecords(records),
		}
	default:
		log.Fatalf("Wrong schema type: %s", r.file.Schema.Type)
		return importer.Stmt{}
	}
}
