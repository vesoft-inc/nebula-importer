package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("string", func() {
	DescribeTable("IsUnsignedInteger/IsInteger",
		func(s string, isUnsignedInteger bool) {
			Expect(IsUnsignedInteger(s)).To(Equal(isUnsignedInteger))
			Expect(IsInteger(s)).To(Equal(isUnsignedInteger))
			if len(s) > 0 && (s[0] != '-' && s[0] != '+') {
				Expect(IsInteger("-" + s)).To(Equal(isUnsignedInteger))
				Expect(IsInteger("+" + s)).To(Equal(isUnsignedInteger))
			}
		},
		EntryDescription("IsInteger(%[1]q) == %[2]t"),
		Entry(nil, "", false),
		Entry(nil, "0", true),
		Entry(nil, "02", true),
		Entry(nil, "12", true),
		Entry(nil, "0x0", true),
		Entry(nil, "0X0", true),
		Entry(nil, "0x", false),
		Entry(nil, "0X", false),
		Entry(nil, "0123456789", true),
		Entry(nil, "1234567890", true),
		Entry(nil, "1234567890a", false),
		Entry(nil, "a1234567890", false),
		Entry(nil, "12345a67890", false),
		Entry(nil, "0X1234567890", true),
		Entry(nil, "0X1234567890abcdef", true),
		Entry(nil, "0X1234567890ABCDEF", true),
		Entry(nil, "0X1A2B3C4D5F6F7890", true),
		Entry(nil, "0X1A2B3C4D5F6F7890g", false),
	)

	DescribeTable("IsDigit",
		func(b byte, isDigit bool) {
			Expect(IsDigit(b)).To(Equal(isDigit))
		},
		EntryDescription("IsDigit('%[1]c') == %[2]t"),
		Entry(nil, byte(0), false),
		Entry(nil, byte('0')-1, false),
		Entry(nil, byte('0'), true),
		Entry(nil, byte('1'), true),
		Entry(nil, byte('2'), true),
		Entry(nil, byte('3'), true),
		Entry(nil, byte('4'), true),
		Entry(nil, byte('5'), true),
		Entry(nil, byte('6'), true),
		Entry(nil, byte('7'), true),
		Entry(nil, byte('8'), true),
		Entry(nil, byte('9'), true),
		Entry(nil, byte('9')+1, false),
		Entry(nil, byte('a'), false),
		Entry(nil, byte('A'), false),
		Entry(nil, byte('\n'), false),
	)

	DescribeTable("IsDigit",
		func(b byte, isDigit bool) {
			Expect(IsHexDigit(b)).To(Equal(isDigit))
		},
		EntryDescription("IsDigit('%[1]c') == %[2]t"),
		Entry(nil, byte(0), false),
		Entry(nil, byte('0')-1, false),
		Entry(nil, byte('0'), true),
		Entry(nil, byte('1'), true),
		Entry(nil, byte('2'), true),
		Entry(nil, byte('3'), true),
		Entry(nil, byte('4'), true),
		Entry(nil, byte('5'), true),
		Entry(nil, byte('6'), true),
		Entry(nil, byte('7'), true),
		Entry(nil, byte('8'), true),
		Entry(nil, byte('9'), true),
		Entry(nil, byte('9')+1, false),
		Entry(nil, byte('a')-1, false),
		Entry(nil, byte('a'), true),
		Entry(nil, byte('b'), true),
		Entry(nil, byte('c'), true),
		Entry(nil, byte('d'), true),
		Entry(nil, byte('e'), true),
		Entry(nil, byte('f'), true),
		Entry(nil, byte('f')+1, false),
		Entry(nil, byte('A')-1, false),
		Entry(nil, byte('A'), true),
		Entry(nil, byte('B'), true),
		Entry(nil, byte('C'), true),
		Entry(nil, byte('D'), true),
		Entry(nil, byte('E'), true),
		Entry(nil, byte('F'), true),
		Entry(nil, byte('F')+1, false),
		Entry(nil, byte('\n'), false),
	)

	DescribeTable("ConvertIdentifier",
		func(s, expect string) {
			Expect(ConvertIdentifier(s)).To(Equal(expect))
		},
		EntryDescription("ConvertIdentifier(%[1]s) = %[2]s"),
		Entry(nil, "", "``"),
		Entry(nil, "`", "`\\``"),
		Entry(nil, "``", "`\\`\\``"),
		Entry(nil, "a`b`c", "`a\\`b\\`c`"),
		Entry(nil, "`a`b`c", "`\\`a\\`b\\`c`"),
		Entry(nil, "a`b`c`", "`a\\`b\\`c\\``"),
		Entry(nil, "`a`b`c`", "`\\`a\\`b\\`c\\``"),
		Entry(nil, "\\", "`\\\\`"),
		Entry(nil, "\\\\", "`\\\\\\\\`"),
		Entry(nil, "a\\b\\c", "`a\\\\b\\\\c`"),
		Entry(nil, "\\a\\b\\c", "`\\\\a\\\\b\\\\c`"),
		Entry(nil, "a\\b\\c\\", "`a\\\\b\\\\c\\\\`"),
		Entry(nil, "\\a\\b\\c\\", "`\\\\a\\\\b\\\\c\\\\`"),
		Entry(nil, "`\\a\\`b`\\c\\`", "`\\`\\\\a\\\\\\`b\\`\\\\c\\\\\\``"),
	)
})
