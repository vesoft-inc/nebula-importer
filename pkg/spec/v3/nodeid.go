package specv3

import (
	"strings"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/picker"
)

var supportedNodeIDFunctions = map[string]struct{}{
	"HASH": {},
}

type (
	// NodeID is the VID in 3.x
	NodeID struct {
		Name        string        `yaml:"-"`
		Type        ValueType     `yaml:"type"`
		Index       int           `yaml:"index"`
		ConcatItems []interface{} `yaml:"concatItems,omitempty"` // only support string and int, string for constant, int is for Index
		Function    *string       `yaml:"function"`

		picker picker.Picker
	}
)

func IsSupportedNodeIDFunction(function string) bool {
	_, ok := supportedNodeIDFunctions[strings.ToUpper(function)]
	return ok
}

func (id *NodeID) Complete() {
	if id.Type == "" {
		id.Type = ValueTypeDefault
	}
}

func (id *NodeID) Validate() error {
	if id.Name == "" {
		return id.importError(errors.ErrNoNodeIDName)
	}
	if !IsSupportedNodeIDValueType(id.Type) {
		return id.importError(errors.ErrUnsupportedValueType, "unsupported type %s", id.Type)
	}
	if id.Function != nil && !IsSupportedNodeIDFunction(*id.Function) {
		return id.importError(errors.ErrUnsupportedFunction, "unsupported function %s", *id.Function)
	}
	if err := id.initPicker(); err != nil {
		return id.importError(err, "init picker failed")
	}

	return nil
}

func (id *NodeID) Value(record Record) (string, error) {
	val, err := id.picker.Pick(record)
	if err != nil {
		if len(id.ConcatItems) > 0 {
			return "", id.importError(err, "record concat items %v pick failed", id.ConcatItems).SetRecord(record)
		}
		return "", id.importError(err, "record index %d pick failed", id.Index).SetRecord(record)
	}
	defer val.Release()
	return val.Val, nil
}

func (id *NodeID) initPicker() error {
	pickerConfig := picker.Config{
		Type:     string(id.Type),
		Function: id.Function,
	}

	if len(id.ConcatItems) > 0 {
		pickerConfig.ConcatItems = id.ConcatItems
	} else {
		pickerConfig.Indices = []int{id.Index}
	}

	var err error
	id.picker, err = pickerConfig.Build()
	return err
}

func (id *NodeID) importError(err error, formatWithArgs ...any) *errors.ImportError {
	return errors.AsOrNewImportError(err, formatWithArgs...).SetPropName(id.Name)
}
