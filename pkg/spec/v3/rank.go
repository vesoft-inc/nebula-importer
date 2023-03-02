package specv3

import (
	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/picker"
)

type (
	Rank struct {
		Index int `yaml:"index"`

		picker picker.Picker
	}
)

func (*Rank) Complete() {}

func (r *Rank) Validate() error {
	//revive:disable-next-line:if-return
	if err := r.initPicker(); err != nil {
		return r.importError(err, "init picker failed")
	}
	return nil
}

func (r *Rank) Value(record Record) (string, error) {
	val, err := r.picker.Pick(record)
	if err != nil {
		return "", r.importError(err, "record index %d pick failed", r.Index).SetRecord(record)
	}
	defer val.Release()
	return val.Val, nil
}

func (r *Rank) initPicker() error {
	pickerConfig := picker.Config{
		Indices: []int{r.Index},
		Type:    string(ValueTypeInt),
	}

	var err error
	r.picker, err = pickerConfig.Build()
	return err
}

func (*Rank) importError(err error, formatWithArgs ...any) *errors.ImportError {
	return errors.AsOrNewImportError(err, formatWithArgs...)
}
