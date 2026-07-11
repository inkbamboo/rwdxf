package core

import (
	"fmt"
	"strconv"
)

// DataType 表示 DXF Tag 中值的数据类型接口。
// 支持三种具体类型：String、Integer、Float。
type DataType interface {
	DxfElement
	ToString() string  // 返回值的字符串表示
	Value() interface{} // 返回原始的 Go 值
}

// String 表示 DXF tag 中的字符串类型值。
type String struct {
	Val string
}

func (s *String) ToString() string      { return s.Val }
func (s *String) Value() interface{}    { return s.Val }
func (s *String) Equals(other DxfElement) bool {
	if o, ok := other.(*String); ok {
		return s.Val == o.Val
	}
	return false
}

// NewString 从字符串创建 String 类型的 DataType（用于 tag 解析工厂）。
func NewString(val string) (DataType, error) { return &String{Val: val}, nil }

// NewStringValue 直接创建 String 值（不返回 error，用于代码中构造 tag）。
func NewStringValue(val string) *String { return &String{Val: val} }

// Integer 表示 DXF tag 中的整数类型值。
type Integer struct {
	Val int
}

func (i *Integer) ToString() string      { return strconv.Itoa(i.Val) }
func (i *Integer) Value() interface{}    { return i.Val }
func (i *Integer) Equals(other DxfElement) bool {
	if o, ok := other.(*Integer); ok {
		return i.Val == o.Val
	}
	return false
}

// NewInteger 从字符串解析整数，用于 tag 解析工厂。
func NewInteger(val string) (DataType, error) {
	v, err := strconv.Atoi(val)
	if err != nil {
		return nil, err
	}
	return &Integer{Val: v}, nil
}

// NewIntegerValue 直接创建 Integer 值。
func NewIntegerValue(val int) *Integer { return &Integer{Val: val} }

// Float 表示 DXF tag 中的浮点数类型值。
type Float struct {
	Val float64
}

func (f *Float) ToString() string      { return fmt.Sprintf("%f", f.Val) }
func (f *Float) Value() interface{}    { return f.Val }
func (f *Float) Equals(other DxfElement) bool {
	if o, ok := other.(*Float); ok {
		return FloatEquals(f.Val, o.Val)
	}
	return false
}

// NewFloat 从字符串解析浮点数，用于 tag 解析工厂。
func NewFloat(val string) (DataType, error) {
	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return nil, err
	}
	return &Float{Val: v}, nil
}

// NewFloatValue 直接创建 Float 值。
func NewFloatValue(val float64) *Float { return &Float{Val: val} }

// AsString 尝试将 DataType 断言为 *String 并返回其值。
func AsString(d DataType) (string, bool) {
	if s, ok := d.(*String); ok {
		return s.Val, true
	}
	return "", false
}

// AsInt 尝试将 DataType 断言为 *Integer 并返回其值。
func AsInt(d DataType) (int, bool) {
	if i, ok := d.(*Integer); ok {
		return i.Val, true
	}
	return 0, false
}

// AsFloat 尝试将 DataType 断言为 *Float 并返回其值。
func AsFloat(d DataType) (float64, bool) {
	if f, ok := d.(*Float); ok {
		return f.Val, true
	}
	return 0, false
}
