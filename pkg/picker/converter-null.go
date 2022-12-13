package picker

var (
	_ Converter = NullableConverter{}
	_ Converter = NullConverter{}
)

type (
	NullableConverter struct {
		Nullable func(string) bool
	}

	NullConverter struct {
		Value string
	}
)

func (nc NullableConverter) Convert(v *Value) (*Value, error) {
	if !v.IsNull && nc.Nullable(v.Val) {
		v.IsNull = true
	}
	return v, nil
}

func (nc NullConverter) Convert(v *Value) (*Value, error) {
	if v.IsNull {
		v.Val = nc.Value
		v.isSetNull = true
	}
	return v, nil
}
