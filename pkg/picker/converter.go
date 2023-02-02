package picker

var _ Converter = Converters(nil)

type (
	Converter interface {
		Convert(*Value) (*Value, error)
	}

	ConverterFunc func(v *Value) (*Value, error)

	Converters         []Converter
	NullableConverters []Converter
)

func (f ConverterFunc) Convert(v *Value) (*Value, error) {
	return f(v)
}

func (cs Converters) Convert(v *Value) (*Value, error) {
	switch len(cs) {
	case 0:
		return v, nil
	case 1:
		return cs[0].Convert(v)
	}
	return cs.convertSlow(v)
}

func (cs Converters) convertSlow(v *Value) (*Value, error) {
	var err error
	for _, c := range cs {
		v, err = c.Convert(v)
		if err != nil {
			return nil, err
		}
	}
	return v, nil
}

func (ncs NullableConverters) Convert(v *Value) (*Value, error) {
	if v.isSetNull {
		return v, nil
	}
	switch len(ncs) {
	case 0:
		return v, nil
	case 1:
		return ncs[0].Convert(v)
	}
	return ncs.convertSlow(v)
}

func (ncs NullableConverters) convertSlow(v *Value) (*Value, error) {
	var err error
	for _, c := range ncs {
		v, err = c.Convert(v)
		if err != nil {
			return nil, err
		}
		if v.isSetNull {
			return v, nil
		}
	}
	return v, nil
}
