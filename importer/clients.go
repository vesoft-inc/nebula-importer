package nebula_importer

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	nebula "github.com/vesoft-inc/nebula-go"
	graph "github.com/vesoft-inc/nebula-go/graph"
)

func NewNebulaClientPool(config *YAMLConfig, nth int, recordCh *<-chan []string, errDataCh *chan<- []string, errLogCh *chan<- string) {
	for i := 0; i < config.Settings.Concurrency; i++ {
		go func() {
			client, err := nebula.NewClient(config.Settings.Connection.Address)
			if err != nil {
				log.Println(err)
				return
			}

			if err = client.Connect(config.Settings.Connection.User, config.Settings.Connection.Password); err != nil {
				log.Println(err)
				return
			}
			defer client.Disconnect()

			for {
				record := <-*recordCh

				stmt, err := makeStmt(config, nth, record)
				if err != nil {
					*errLogCh <- fmt.Sprintf("Fail to make nGQL statement, %s", err.Error())
					*errDataCh <- record
					continue
				}

				// TODO: Add some metrics for response latency, succeededCount, failedCount
				resp, err := client.Execute(stmt)
				if err != nil {
					*errLogCh <- fmt.Sprintf("Client execute error: %s", err.Error())
					*errDataCh <- record
					continue
				}

				if resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
					*errLogCh <- fmt.Sprintf("Fail to execute: %s, error: %s", stmt, resp.GetErrorMsg())
					*errDataCh <- record
					continue
				}
			}

		}()
	}
}

func makeStmt(config *YAMLConfig, nth int, record []string) (string, error) {
	file := config.Files[nth]
	schemaType := strings.ToUpper(file.Schema.Type)

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("INSERT %s %s(", schemaType, file.Schema.Name))

	for idx, prop := range file.Schema.Props {
		builder.WriteString(prop.Name)
		if idx < len(file.Schema.Props)-1 {
			builder.WriteString(",")
		}
	}
	builder.WriteString(") VALUES ")

	isEdge := schemaType == "EDGE"
	fromIndex := 1
	if isEdge {
		fromIndex = 2
	}

	if err := writeVID(isEdge, record, &builder); err != nil {
		return "", err
	}

	builder.WriteString(":(")
	for idx, val := range record[fromIndex:] {
		builder.WriteString(val)
		if idx < len(record)-1 {
			builder.WriteString(",")
		}
	}
	builder.WriteString(");")
	return builder.String(), nil
}

func writeVID(isEdge bool, record []string, builder *strings.Builder) error {
	vid, err := strconv.ParseInt(record[0], 10, 64)
	if err != nil {
		return err
	}
	builder.WriteString(fmt.Sprintf("%d", vid))

	if isEdge {
		dstVID, err := strconv.ParseInt(record[1], 10, 64)
		if err != nil {
			return err
		}
		builder.WriteString(fmt.Sprintf(" -> %d", dstVID))
	}
	return nil
}
