package picker

import (
	"strconv"
	"strings"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/utils"
)

var (
	_ Converter = TypeBoolConverter{}
	_ Converter = TypeIntConverter{}
	_ Converter = TypeFloatConverter{}
	_ Converter = TypeDoubleConverter{}
	_ Converter = TypeStringConverter{}
	_ Converter = TypeDateConverter{}
	_ Converter = TypeTimeConverter{}
	_ Converter = TypeDatetimeConverter{}
	_ Converter = TypeTimestampConverter{}
	_ Converter = TypeGeoConverter{}
	_ Converter = TypeGeoPointConverter{}
	_ Converter = TypeGeoLineStringConverter{}
	_ Converter = TypeGeoPolygonConverter{}
)

type (
	TypeBoolConverter = NonConverter

	TypeIntConverter = NonConverter

	TypeFloatConverter = NonConverter

	TypeDoubleConverter = NonConverter

	TypeStringConverter struct{}

	TypeDateConverter = FunctionStringConverter

	TypeTimeConverter = FunctionStringConverter

	TypeDatetimeConverter = FunctionStringConverter

	TypeTimestampConverter struct {
		fc  FunctionConverter
		fsc FunctionStringConverter
	}

	TypeGeoConverter = FunctionStringConverter

	TypeGeoPointConverter = FunctionStringConverter

	TypeGeoLineStringConverter = FunctionStringConverter

	TypeGeoPolygonConverter = FunctionStringConverter
)

func NewTypeConverter(t string) (Converter, error) {
	switch strings.ToUpper(t) {
	case "BOOL":
		return TypeBoolConverter{}, nil
	case "INT":
		return TypeIntConverter{}, nil
	case "FLOAT":
		return TypeFloatConverter{}, nil
	case "DOUBLE":
		return TypeDoubleConverter{}, nil
	case "STRING":
		return TypeStringConverter{}, nil
	case "DATE":
		return TypeDateConverter{
			Name: "DATE",
		}, nil
	case "TIME":
		return TypeTimeConverter{
			Name: "TIME",
		}, nil
	case "DATETIME":
		return TypeDatetimeConverter{
			Name: "DATETIME",
		}, nil
	case "TIMESTAMP":
		return TypeTimestampConverter{
			fc: FunctionConverter{
				Name: "TIMESTAMP",
			},
			fsc: FunctionStringConverter{
				Name: "TIMESTAMP",
			},
		}, nil
	case "GEOGRAPHY":
		return TypeGeoConverter{
			Name: "ST_GeogFromText",
		}, nil
	case "GEOGRAPHY(POINT)":
		return TypeGeoPointConverter{
			Name: "ST_GeogFromText",
		}, nil
	case "GEOGRAPHY(LINESTRING)":
		return TypeGeoLineStringConverter{
			Name: "ST_GeogFromText",
		}, nil
	case "GEOGRAPHY(POLYGON)":
		return TypeGeoPolygonConverter{
			Name: "ST_GeogFromText",
		}, nil
	}
	return nil, errors.ErrUnsupportedValueType
}

func (TypeStringConverter) Convert(v *Value) (*Value, error) {
	v.Val = strconv.Quote(v.Val)
	return v, nil
}

func (tc TypeTimestampConverter) Convert(v *Value) (*Value, error) {
	if utils.IsUnsignedInteger(v.Val) {
		return tc.fc.Convert(v)
	}
	return tc.fsc.Convert(v)
}
