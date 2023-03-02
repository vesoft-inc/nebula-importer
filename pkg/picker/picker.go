package picker

var (
	_ Picker = PickerFunc(nil)
	_ Picker = ConverterPicker{}
	_ Picker = NullablePickers{}
)

type (
	Picker interface {
		Pick([]string) (*Value, error)
	}

	PickerFunc func(record []string) (*Value, error)

	ConverterPicker struct {
		picker    Picker
		converter Converter
	}

	NullablePickers []Picker
)

func (f PickerFunc) Pick(record []string) (*Value, error) {
	return f(record)
}

func (cp ConverterPicker) Pick(record []string) (*Value, error) {
	v, err := cp.picker.Pick(record)
	if err != nil {
		return nil, err
	}
	if cp.converter == nil {
		return v, nil
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
