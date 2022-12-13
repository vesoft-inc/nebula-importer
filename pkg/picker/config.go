package picker

import (
	"fmt"
	"strings"
)

// Config is the configuration to build Picker
// The priority is as follows:
//		ConcatItems > Indices
// 		Nullable
// 		DefaultValue
//    NullValue, if set to null, subsequent conversions will be skipped.
// 		Type
// 		Function
// 		CheckOnPost
type Config struct {
	ConcatItems  ConcatItems        // Concat index column, constant, or mixed.
	Indices      []int              // Set index columns, the first non-null.
	Nullable     func(string) bool  // Determine whether it is null. Optional.
	NullValue    string             // Set null value when it is null. Optional.
	DefaultValue *string            // Set default value when it is null. Optional.
	Type         string             // Set the type of value.
	Function     *string            // Set the conversion function of value.
	CheckOnPost  func(*Value) error // Set the value check function on post.
}

func (c *Config) Build() (Picker, error) {
	var retPicker Picker
	var nullHandled bool
	switch {
	case c.ConcatItems.Len() > 0:
		retPicker = ConcatPicker{
			items: c.ConcatItems,
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
		return nil, fmt.Errorf("no indices or concat items")
	}

	var converters []Converter

	if !nullHandled && c.Nullable != nil {
		converters = append(converters, NullableConverter{
			Nullable: c.Nullable,
		})
	}

	if c.Nullable != nil {
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

	converters = append(converters, NewTypeConverter(c.Type))

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
