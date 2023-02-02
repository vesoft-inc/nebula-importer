package picker

var (
	_ Picker = ConstantPicker("")
)

type ConstantPicker string

func (cp ConstantPicker) Pick(_ []string) (v *Value, err error) {
	return &Value{
		Val:    string(cp),
		IsNull: false,
	}, nil
}
