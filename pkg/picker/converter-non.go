package picker

var _ Converter = NonConverter{}

type NonConverter struct{}

func (NonConverter) Convert(v *Value) (*Value, error) {
	return v, nil
}
