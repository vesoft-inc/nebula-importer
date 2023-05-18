package picker

var _ Converter = DefaultConverter{}

type DefaultConverter struct {
	Value string
}

func (dc DefaultConverter) Convert(v *Value) (*Value, error) {
	if v.IsNull {
		v.Val = dc.Value
		v.IsNull = false
	}
	return v, nil
}
