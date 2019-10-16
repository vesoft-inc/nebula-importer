package nebula_importer

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type DataFileReader interface {
	NewFileReader(path string, stmtCh chan<- Query)
	MakeQuery([]string) Query
}

type CSVReader struct {
	Schema Schema
}

func (r *CSVReader) NewFileReader(path string, stmtCh chan<- Query) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		stmtCh <- r.MakeQuery(line)
	}
}

func (r *CSVReader) MakeQuery(record []string) Query {
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
	for idx, val := range record[fromIndex:] {
		builder.WriteString(val)
		if idx < len(record)-1 {
			builder.WriteString(",")
		}
	}
	builder.WriteString(");")

	data := make([]interface{}, len(record))
	for i := range record {
		data[i] = record[i]
	}

	return Query{
		Stmt: builder.String(),
		Data: data,
	}
}

func writeVID(schemaType string, record []string, builder *strings.Builder) int {
	builder.WriteString(fmt.Sprintf("%d", record[0]))
	if schemaType == "EDGE" {
		builder.WriteString(fmt.Sprintf(" -> %d", record[1]))
		return 2
	}
	return 1
}
