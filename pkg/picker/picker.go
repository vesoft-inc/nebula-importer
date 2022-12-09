package picker

var _ Picker = ConverterPicker{}

type (
	Picker interface {
		Pick([]string) (*Value, error)
	}

	ConverterPicker struct {
		picker    Picker
		converter Converter
	}

	NullablePickers []Picker
)

func (cp ConverterPicker) Pick(record []string) (*Value, error) {
	v, err := cp.picker.Pick(record)
	if err != nil {
		return nil, err
	}
	return cp.converter.Convert(v)
}

func (nps NullablePickers) Pick(record []string) (v *Value, err error) {
	for _, p := range nps {
		v, err = p.Pick(record)
		if err != nil {
			return nil, err
		}
		if !v.IsNull {
			return v, nil
		}
	}
	return v, nil
}
