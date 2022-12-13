package picker

import "fmt"

var (
	_ Converter = FunctionConverter{}
	_ Converter = FunctionStringConverter{}
)

type (
	FunctionConverter struct {
		Name string
	}
	FunctionStringConverter struct {
		Name string
	}
)

func (fc FunctionConverter) Convert(v *Value) (*Value, error) {
	v.Val = getFuncValue(fc.Name, v.Val)
	return v, nil
}

func (fc FunctionStringConverter) Convert(v *Value) (*Value, error) {
	v.Val = getFuncValue(fc.Name, fmt.Sprintf("%q", v.Val))
	return v, nil
}

func getFuncValue(name, value string) string {
	return name + "(" + value + ")"
}
