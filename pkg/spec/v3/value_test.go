package specv3

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Value", func() {
	DescribeTable("IsSupportedPropValueType",
		func(t ValueType, expectIsSupported bool) {
			Expect(IsSupportedPropValueType(t)).To(Equal(expectIsSupported))
		},
		EntryDescription("%[1]s -> %[2]t"),
		Entry(nil, ValueTypeBool, true),
		Entry(nil, ValueTypeInt, true),
		Entry(nil, ValueTypeString, true),
		Entry(nil, ValueTypeFloat, true),
		Entry(nil, ValueTypeDouble, true),
		Entry(nil, ValueTypeDate, true),
		Entry(nil, ValueTypeDateTime, true),
		Entry(nil, ValueTypeTimestamp, true),
		Entry(nil, ValueTypeGeo, true),
		Entry(nil, ValueTypeGeoPoint, true),
		Entry(nil, ValueTypeGeoLineString, true),
		Entry(nil, ValueTypeGeoPolygon, true),
		Entry(nil, ValueType("int"), true),
		Entry(nil, ValueType("inT"), true),
		Entry(nil, ValueType("iNt"), true),
		Entry(nil, ValueType("InT"), true),
		Entry(nil, ValueType("unsupported"), false),
	)

	DescribeTable("IsSupportedNodeIDValueType",
		func(t ValueType, expectIsSupported bool) {
			Expect(IsSupportedNodeIDValueType(t)).To(Equal(expectIsSupported))
		},
		EntryDescription("%[1]s -> %[2]t"),
		Entry(nil, ValueTypeBool, false),
		Entry(nil, ValueTypeInt, true),
		Entry(nil, ValueTypeString, true),
		Entry(nil, ValueTypeFloat, false),
		Entry(nil, ValueTypeDouble, false),
		Entry(nil, ValueTypeDate, false),
		Entry(nil, ValueTypeDateTime, false),
		Entry(nil, ValueTypeTimestamp, false),
		Entry(nil, ValueTypeGeo, false),
		Entry(nil, ValueTypeGeoPoint, false),
		Entry(nil, ValueTypeGeoLineString, false),
		Entry(nil, ValueTypeGeoPolygon, false),
		Entry(nil, ValueType("int"), true),
		Entry(nil, ValueType("inT"), true),
		Entry(nil, ValueType("iNt"), true),
		Entry(nil, ValueType("InT"), true),
		Entry(nil, ValueType("unsupported"), false),
	)

	DescribeTable(".Equal",
		func(t, vt ValueType, expectIsSupported bool) {
			Expect(t.Equal(vt)).To(Equal(expectIsSupported))
		},
		EntryDescription("%[1]s == %[2]s ? %[3]t"),
		Entry(nil, ValueTypeInt, ValueType("int"), true),
		Entry(nil, ValueTypeInt, ValueType("inT"), true),
		Entry(nil, ValueTypeInt, ValueType("iNt"), true),
		Entry(nil, ValueTypeInt, ValueType("InT"), true),
		Entry(nil, ValueTypeInt, ValueType("unsupported"), false),
		Entry(nil, ValueType("unsupported"), ValueTypeInt, false),
		Entry(nil, ValueType("unsupported"), ValueType("unsupported"), false),
	)
})
