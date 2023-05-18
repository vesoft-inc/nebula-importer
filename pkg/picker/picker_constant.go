package picker

var (
	_ Picker = ConstantPicker("")
)

type ConstantPicker string

func (cp ConstantPicker) Pick(_ []string) (*Value, error) {
	return NewValue(string(cp)), nil
}
