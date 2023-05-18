package specv3

import (
	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/picker"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/utils"
)

type (
	Prop struct {
		Name               string    `yaml:"name"`
		Type               ValueType `yaml:"type"`
		Index              int       `yaml:"index"`
		Nullable           bool      `yaml:"nullable"`
		NullValue          string    `yaml:"nullValue"`
		AlternativeIndices []int     `yaml:"alternativeIndices,omitempty"`
		DefaultValue       *string   `yaml:"defaultValue"`

		convertedName string
		picker        picker.Picker
	}

	Props []*Prop
)

func (p *Prop) Complete() {
	p.convertedName = utils.ConvertIdentifier(p.Name)
	if p.Type == "" {
		p.Type = ValueTypeDefault
	}
}

func (p *Prop) Validate() error {
	if p.Name == "" {
		return p.importError(errors.ErrNoPropName)
	}
	if !IsSupportedPropValueType(p.Type) {
		return p.importError(errors.ErrUnsupportedValueType, "unsupported type %s", p.Type)
	}
	if err := p.initPicker(); err != nil {
		return p.importError(err, "init picker failed")
	}
	return nil
}

func (p *Prop) Value(record Record) (string, error) {
	val, err := p.picker.Pick(record)
	if err != nil {
		return "", p.importError(err, "record index %d pick failed", p.Index).SetRecord(record)
	}
	defer val.Release()
	return val.Val, nil
}

func (p *Prop) initPicker() error {
	pickerConfig := picker.Config{
		Indices: []int{p.Index},
		Type:    string(p.Type),
	}

	if p.Nullable {
		pickerConfig.Nullable = func(s string) bool {
			return s == p.NullValue
		}
		pickerConfig.NullValue = dbNULL
		if len(p.AlternativeIndices) > 0 {
			pickerConfig.Indices = append(pickerConfig.Indices, p.AlternativeIndices...)
		}
		pickerConfig.DefaultValue = p.DefaultValue
	}

	var err error
	p.picker, err = pickerConfig.Build()
	return err
}

func (p *Prop) importError(err error, formatWithArgs ...any) *errors.ImportError {
	return errors.AsOrNewImportError(err, formatWithArgs...).SetPropName(p.Name)
}

func (ps Props) Complete() {
	for i := range ps {
		ps[i].Complete()
	}
}

func (ps Props) Validate() error {
	for i := range ps {
		if err := ps[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (ps Props) ValueList(record Record) ([]string, error) {
	valueList := make([]string, 0, len(ps))
	for _, prop := range ps {
		value, err := prop.Value(record)
		if err != nil {
			return nil, err
		}
		valueList = append(valueList, value)
	}
	return valueList, nil
}

func (ps Props) NameList() []string {
	nameList := make([]string, len(ps))
	for i := range ps {
		nameList[i] = ps[i].convertedName
	}
	return nameList
}

func (ps Props) Append(props ...*Prop) Props {
	cpy := make(Props, len(ps)+len(props))
	copy(cpy, ps)
	copy(cpy[len(ps):], props)
	return cpy
}
