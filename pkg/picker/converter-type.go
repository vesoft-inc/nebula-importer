package picker

import (
	"fmt"
	"strings"

	"github.com/vesoft-inc/nebula-importer/pkg/utils"
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

func NewTypeConverter(t string) Converter {
	switch strings.ToUpper(t) {
	case "BOOL":
		return TypeBoolConverter{}
	case "INT":
		return TypeIntConverter{}
	case "FLOAT":
		return TypeFloatConverter{}
	case "DOUBLE":
		return TypeDoubleConverter{}
	case "STRING":
		return TypeStringConverter{}
	case "DATE":
		return TypeDateConverter{
			Name: "DATE",
		}
	case "TIME":
		return TypeTimeConverter{
			Name: "TIME",
		}
	case "DATETIME":
		return TypeDatetimeConverter{
			Name: "DATETIME",
		}
	case "TIMESTAMP":
		return TypeTimestampConverter{
			fc: FunctionConverter{
				Name: "TIMESTAMP",
			},
			fsc: FunctionStringConverter{
				Name: "TIMESTAMP",
			},
		}
	case "GEOGRAPHY":
		return TypeGeoConverter{
			Name: "ST_GeogFromText",
		}
	case "GEOGRAPHY(POINT)":
		return TypeGeoPointConverter{
			Name: "ST_GeogFromText",
		}
	case "GEOGRAPHY(LINESTRING)":
		return TypeGeoLineStringConverter{
			Name: "ST_GeogFromText",
		}
	case "GEOGRAPHY(POLYGON)":
		return TypeGeoPolygonConverter{
			Name: "ST_GeogFromText",
		}
	}
	return ErrorConverter{
		Err: fmt.Errorf("unsupported type %s", t),
	}
}

func (tc TypeStringConverter) Convert(v *Value) (*Value, error) {
	v.Val = fmt.Sprintf("%q", v.Val)
	return v, nil
}

func (tc TypeTimestampConverter) Convert(v *Value) (*Value, error) {
	if utils.IsUnsignedInteger(v.Val) {
		return tc.fc.Convert(v)
	}
	return tc.fsc.Convert(v)
}
