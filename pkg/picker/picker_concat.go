package picker

import (
	"strings"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"
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

func (ci *ConcatItems) Add(items ...any) (err error) {
	for i := range items {
		switch v := items[i].(type) {
		case uint8:
			err = ci.AddIndex(int(v))
		case int8:
			err = ci.AddIndex(int(v))
		case uint16:
			err = ci.AddIndex(int(v))
		case int16:
			err = ci.AddIndex(int(v))
		case uint32:
			err = ci.AddIndex(int(v))
		case int32:
			err = ci.AddIndex(int(v))
		case uint64:
			err = ci.AddIndex(int(v)) //nolint:all
		case int64:
			err = ci.AddIndex(int(v))
		case int:
			err = ci.AddIndex(v)
		case uint:
			err = ci.AddIndex(int(v)) //nolint:all
		case string:
			err = ci.AddConstant(v)
		case []byte:
			err = ci.AddConstant(string(v))
		default:
			err = errors.ErrUnsupportedConcatItemType
		}
		if err != nil {
			break
		}
	}
	return err
}

func (ci *ConcatItems) AddIndex(index int) error {
	if index < 0 {
		return errors.ErrInvalidIndex
	}
	ci.pickers = append(ci.pickers, IndexPicker(index))
	return nil
}

func (ci *ConcatItems) AddConstant(constant string) error {
	ci.pickers = append(ci.pickers, ConstantPicker(constant))
	return nil
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
		_, _ = sb.WriteString(v.Val)
		v.Release()
	}
	return NewValue(sb.String()), nil
}
