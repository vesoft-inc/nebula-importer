package specv3

import (
	"strings"
)

const (
	dbNULL = "NULL"

	ValueTypeBool          ValueType = "BOOL"
	ValueTypeInt           ValueType = "INT"
	ValueTypeString        ValueType = "STRING"
	ValueTypeFloat         ValueType = "FLOAT"
	ValueTypeDouble        ValueType = "DOUBLE"
	ValueTypeDate          ValueType = "DATE"
	ValueTypeDateTime      ValueType = "DATETIME"
	ValueTypeTimestamp     ValueType = "TIMESTAMP"
	ValueTypeGeo           ValueType = "GEOGRAPHY"
	ValueTypeGeoPoint      ValueType = "GEOGRAPHY(POINT)"
	ValueTypeGeoLineString ValueType = "GEOGRAPHY(LINESTRING)"
	ValueTypeGeoPolygon    ValueType = "GEOGRAPHY(POLYGON)"

	ValueTypeDefault = ValueTypeString
)

var (
	supportedPropValueTypes = map[ValueType]struct{}{
		ValueTypeBool:          {},
		ValueTypeInt:           {},
		ValueTypeString:        {},
		ValueTypeFloat:         {},
		ValueTypeDouble:        {},
		ValueTypeDate:          {},
		ValueTypeDateTime:      {},
		ValueTypeTimestamp:     {},
		ValueTypeGeo:           {},
		ValueTypeGeoPoint:      {},
		ValueTypeGeoLineString: {},
		ValueTypeGeoPolygon:    {},
	}

	supportedNodeIDValueTypes = map[ValueType]struct{}{
		ValueTypeInt:    {},
		ValueTypeString: {},
	}
)

type ValueType string

func IsSupportedPropValueType(t ValueType) bool {
	_, ok := supportedPropValueTypes[ValueType(strings.ToUpper(t.String()))]
	return ok
}

func IsSupportedNodeIDValueType(t ValueType) bool {
	_, ok := supportedNodeIDValueTypes[ValueType(strings.ToUpper(t.String()))]
	return ok
}

func (t ValueType) Equal(vt ValueType) bool {
	if !IsSupportedPropValueType(t) || !IsSupportedPropValueType(vt) {
		return false
	}
	return strings.EqualFold(t.String(), vt.String())
}

func (t ValueType) String() string {
	return string(t)
}
