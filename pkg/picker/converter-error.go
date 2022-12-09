package picker

var _ Converter = ErrorConverter{}

type ErrorConverter struct {
	Err error
}

func (ec ErrorConverter) Convert(v *Value) (*Value, error) {
	return nil, ec.Err
}
