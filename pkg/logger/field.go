package logger

type (
	Field struct {
		Key   string      `yaml:"key"`
		Value interface{} `yaml:"value"`
	}

	Fields []Field
)

func MapToFields(m map[string]any) Fields {
	if len(m) == 0 {
		return nil
	}
	fields := make(Fields, 0, len(m))
	for k, v := range m {
		fields = append(fields, Field{Key: k, Value: v})
	}
	return fields
}
