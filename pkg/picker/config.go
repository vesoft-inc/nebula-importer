package picker

import (
	"strings"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"
)

// Config is the configuration to build Picker
// The priority is as follows:
//
//	ConcatItems > Indices
//	Nullable
//	DefaultValue
//	NullValue, if set to null, subsequent conversions will be skipped.
//	Type
//	Function
//	CheckOnPost
type Config struct {
	ConcatItems  []any              // Concat index column, constant, or mixed. int for index column, string for constant.
	Indices      []int              // Set index columns, the first non-null.
	Nullable     func(string) bool  // Determine whether it is null. Optional.
	NullValue    string             // Set null value when it is null. Optional.
	DefaultValue *string            // Set default value when it is null. Optional.
	Type         string             // Set the type of value.
	Function     *string            // Set the conversion function of value.
	CheckOnPost  func(*Value) error // Set the value check function on post.
}

//revive:disable-next-line:cyclomatic
func (c *Config) Build() (Picker, error) {
	for i := range c.Indices {
		if c.Indices[i] < 0 {
			return nil, errors.ErrInvalidIndex
		}
	}
	var retPicker Picker
	var nullHandled bool
	switch {
	case len(c.ConcatItems) > 0:
		concatItems := ConcatItems{}
		if err := concatItems.Add(c.ConcatItems...); err != nil {
			return nil, err
		}
		retPicker = ConcatPicker{
			items: concatItems,
		}
	case len(c.Indices) == 1:
		retPicker = IndexPicker(c.Indices[0])
	case len(c.Indices) > 1:
		if c.Nullable == nil {
			// the first must be picked
			retPicker = IndexPicker(c.Indices[0])
		} else {
			pickers := make(NullablePickers, 0, len(c.Indices))
			for _, index := range c.Indices {
				pickers = append(pickers, ConverterPicker{
					picker: IndexPicker(index),
					converter: NullableConverters{
						NullableConverter{
							Nullable: c.Nullable,
						},
					},
				})
			}
			retPicker = pickers
		}
		nullHandled = true
	default:
		return nil, errors.ErrNoIndicesOrConcatItems
	}

	var converters []Converter

	if c.Nullable != nil {
		if !nullHandled {
			converters = append(converters, NullableConverter{
				Nullable: c.Nullable,
			})
		}

		if c.DefaultValue != nil {
			converters = append(converters, DefaultConverter{
				Value: *c.DefaultValue,
			})
		} else {
			converters = append(converters, NullConverter{
				Value: c.NullValue,
			})
		}
	}
	typeConverter, err := NewTypeConverter(c.Type)
	if err != nil {
		return nil, err
	}
	converters = append(converters, typeConverter)

	if c.Function != nil && *c.Function != "" {
		var functionConverter Converter = FunctionConverter{
			Name: *c.Function,
		}
		if strings.EqualFold(*c.Function, "hash") && !strings.EqualFold(c.Type, "string") {
			functionConverter = FunctionStringConverter{
				Name: *c.Function,
			}
		}
		converters = append(converters, functionConverter)
	}

	if c.CheckOnPost != nil {
		converters = append(converters, ConverterFunc(func(v *Value) (*Value, error) {
			if err := c.CheckOnPost(v); err != nil {
				v.Release()
				return nil, err
			}
			return v, nil
		}))
	}

	var converter Converter = Converters(converters)
	if c.Nullable != nil {
		converter = NullableConverters(converters)
	}

	return ConverterPicker{
		picker:    retPicker,
		converter: converter,
	}, nil
}
