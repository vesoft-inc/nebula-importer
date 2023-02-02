package picker

import "fmt"

var (
	_ Picker = IndexPicker(0)
)

type (
	IndexPicker int
)

func (ip IndexPicker) Pick(record []string) (*Value, error) {
	index := int(ip)
	if index < 0 || index >= len(record) {
		return nil, fmt.Errorf("prop index %d out range %d of record(%v)", index, len(record), record)
	}
	return &Value{
		Val:    record[index],
		IsNull: false,
	}, nil
}
