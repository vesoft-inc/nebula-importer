package errors

import (
	"errors"
	"fmt"
	"strings"
)

const (
	fieldMessages   = "messages"
	fieldGraphName  = "graph"
	fieldEdgeName   = "edge"
	fieldNodeName   = "node"
	fieldNodeIDName = "nodeID"
	fieldPropName   = "prop"
	fieldRecord     = "record"
	fieldStatement  = "statement"
)

var _ error = (*ImportError)(nil)

type (
	ImportError struct {
		Err      error
		Messages []string
		fields   map[string]any
	}
)

func NewImportError(err error, formatWithArgs ...any) *ImportError {
	e := &ImportError{
		Err:    err,
		fields: map[string]any{},
	}
	return e.AppendMessage(formatWithArgs...)
}

func AsImportError(err error) (*ImportError, bool) {
	if e := new(ImportError); errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

func AsOrNewImportError(err error, formatWithArgs ...any) *ImportError {
	e, ok := AsImportError(err)
	if ok {
		return e.AppendMessage(formatWithArgs...)
	}
	return NewImportError(err, formatWithArgs...)
}

func (e *ImportError) AppendMessage(formatWithArgs ...any) *ImportError {
	if len(formatWithArgs) > 0 {
		var message string
		if format, ok := formatWithArgs[0].(string); ok {
			message = fmt.Sprintf(format, formatWithArgs[1:]...)
		} else {
			message = fmt.Sprint(formatWithArgs[0])
		}
		if message != "" {
			e.Messages = append(e.Messages, message)
		}
	}
	return e
}

func (e *ImportError) SetGraphName(graphName string) *ImportError {
	return e.withField(fieldGraphName, graphName)
}

func (e *ImportError) GraphName() string {
	return e.getFieldString(fieldGraphName)
}

func (e *ImportError) SetNodeName(nodeName string) *ImportError {
	return e.withField(fieldNodeName, nodeName)
}

func (e *ImportError) NodeName() string {
	return e.getFieldString(fieldNodeName)
}

func (e *ImportError) SetEdgeName(edgeName string) *ImportError {
	return e.withField(fieldEdgeName, edgeName)
}

func (e *ImportError) EdgeName() string {
	return e.getFieldString(fieldEdgeName)
}

func (e *ImportError) SetNodeIDName(nodeIDName string) *ImportError {
	return e.withField(fieldNodeIDName, nodeIDName)
}

func (e *ImportError) NodeIDName() string {
	return e.getFieldString(fieldNodeIDName)
}

func (e *ImportError) SetPropName(propName string) *ImportError {
	return e.withField(fieldPropName, propName)
}

func (e *ImportError) PropName() string {
	return e.getFieldString(fieldPropName)
}

func (e *ImportError) SetRecord(record []string) *ImportError {
	return e.withField(fieldRecord, record)
}

func (e *ImportError) Record() []string {
	return e.getFieldStringSlice(fieldRecord)
}

func (e *ImportError) SetStatement(statement string) *ImportError {
	return e.withField(fieldStatement, statement)
}

func (e *ImportError) Statement() string {
	return e.getFieldString(fieldStatement)
}

func (e *ImportError) Fields() map[string]any {
	m := make(map[string]any, len(e.fields)+1)
	for k, v := range e.fields {
		m[k] = v
	}
	if len(e.Messages) > 0 {
		m[fieldMessages] = e.Messages
	}
	return m
}

func (e *ImportError) withField(key string, value any) *ImportError {
	switch val := value.(type) {
	case string:
		if val == "" {
			return e
		}
	case []string:
		if len(val) == 0 {
			return e
		}
	}
	e.fields[key] = value
	return e
}

func (e *ImportError) getFieldString(key string) string {
	v, ok := e.fields[key]
	if !ok {
		return ""
	}
	vv, ok := v.(string)
	if !ok {
		return ""
	}
	return vv
}

func (e *ImportError) getFieldStringSlice(key string) []string {
	v, ok := e.fields[key]
	if !ok {
		return nil
	}
	vv, ok := v.([]string)
	if !ok {
		return nil
	}
	return vv
}

func (e *ImportError) Error() string {
	var fields []string
	if graphName := e.GraphName(); graphName != "" {
		fields = append(fields, fmt.Sprintf("%s(%s)", fieldGraphName, graphName))
	}
	if nodeName := e.NodeName(); nodeName != "" {
		fields = append(fields, fmt.Sprintf("%s(%s)", fieldNodeName, nodeName))
	}
	if edgeName := e.EdgeName(); edgeName != "" {
		fields = append(fields, fmt.Sprintf("%s(%s)", fieldEdgeName, edgeName))
	}
	if nodeIDName := e.NodeIDName(); nodeIDName != "" {
		fields = append(fields, fmt.Sprintf("%s(%s)", fieldNodeIDName, nodeIDName))
	}
	if propName := e.PropName(); propName != "" {
		fields = append(fields, fmt.Sprintf("%s(%s)", fieldPropName, propName))
	}
	if record := e.Record(); len(record) > 0 {
		fields = append(fields, fmt.Sprintf("%s(%s)", fieldRecord, record))
	}
	if statement := e.Statement(); statement != "" {
		fields = append(fields, fmt.Sprintf("%s(%s)", fieldStatement, statement))
	}
	if len(e.Messages) > 0 {
		fields = append(fields, fmt.Sprintf("%s%s", fieldMessages, strings.Join(e.Messages, ", ")))
	}
	if e.Err != nil {
		fields = append(fields, e.Err.Error())
	}

	return strings.Join(fields, ": ")
}

func (e *ImportError) Cause() error {
	return e.Err
}

func (e *ImportError) Unwrap() error {
	return e.Err
}
