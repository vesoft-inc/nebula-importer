package picker

import (
	"strings"
)

var _ Picker = ConcatPicker{}

type (
	ConcatItems struct {
		pickers NullablePickers
	}

	ConcatPicker struct {
		items ConcatItems
	}
)

func (ci *ConcatItems) AddIndex(index int) *ConcatItems {
	ci.pickers = append(ci.pickers, IndexPicker(index))
	return ci
}

func (ci *ConcatItems) AddConstant(constant string) *ConcatItems {
	ci.pickers = append(ci.pickers, ConstantPicker(constant))
	return ci
}

func (ci ConcatItems) Len() int {
	return len(ci.pickers)
}

func (cp ConcatPicker) Pick(record []string) (*Value, error) {
	var sb strings.Builder
	for _, p := range cp.items.pickers {
		v, err := p.Pick(record)
		if err != nil {
			return nil, err
		}
		sb.WriteString(v.Val)
	}

	return &Value{
		Val:    sb.String(),
		IsNull: false,
	}, nil
}
