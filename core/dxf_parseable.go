package core

import "fmt"

// TypeParser 定义 Tag 值解析器接口。
// 每个 TypeParser 实现从 DataType 中提取值并设入目标结构体字段。
type TypeParser interface {
	Parse(d DataType) error
}

// SetStringFunc 是设置字符串字段的回调函数类型。
type SetStringFunc func(string)

// SetIntFunc 是设置整数字段的回调函数类型。
type SetIntFunc func(int)

// SetFloatFunc 是设置浮点字段的回调函数类型。
type SetFloatFunc func(float64)

// StringTypeParser 将 Tag 值解析为字符串并调用 setter 回调。
type StringTypeParser struct {
	setter SetStringFunc
}

func (p StringTypeParser) Parse(d DataType) error {
	if value, ok := AsString(d); ok {
		p.setter(value)
		return nil
	}
	return fmt.Errorf("error parsing %#v as String", d)
}

// NewStringTypeParser 使用自定义 setter 创建字符串解析器。
func NewStringTypeParser(setter SetStringFunc) *StringTypeParser {
	return &StringTypeParser{setter: setter}
}

// NewStringTypeParserToVar 创建将值直接设置到目标变量的字符串解析器。
func NewStringTypeParserToVar(variable *string) *StringTypeParser {
	return &StringTypeParser{setter: func(value string) { *variable = value }}
}

// IntTypeParser 将 Tag 值解析为整数并调用 setter 回调。
type IntTypeParser struct {
	setter SetIntFunc
}

func (p IntTypeParser) Parse(d DataType) error {
	if value, ok := AsInt(d); ok {
		p.setter(value)
		return nil
	}
	return fmt.Errorf("error parsing %#v as Integer", d)
}

// NewIntTypeParser 使用自定义 setter 创建整数解析器。
func NewIntTypeParser(setter SetIntFunc) *IntTypeParser {
	return &IntTypeParser{setter: setter}
}

// NewIntTypeParserToVar 创建将值直接设置到目标变量的整数解析器。
func NewIntTypeParserToVar(variable *int) *IntTypeParser {
	return &IntTypeParser{setter: func(value int) { *variable = value }}
}

// FloatTypeParser 将 Tag 值解析为浮点数并调用 setter 回调。
type FloatTypeParser struct {
	setter SetFloatFunc
}

func (p FloatTypeParser) Parse(d DataType) error {
	if value, ok := AsFloat(d); ok {
		p.setter(value)
		return nil
	}
	return fmt.Errorf("error parsing %#v as Float", d)
}

// NewFloatTypeParser 使用自定义 setter 创建浮点数解析器。
func NewFloatTypeParser(setter SetFloatFunc) *FloatTypeParser {
	return &FloatTypeParser{setter: setter}
}

// NewFloatTypeParserToVar 创建将值直接设置到目标变量的浮点数解析器。
func NewFloatTypeParserToVar(variable *float64) *FloatTypeParser {
	return &FloatTypeParser{setter: func(value float64) { *variable = value }}
}

// DxfParseable 提供声明式的 Tag 解析框架。
//
// 通过注册"组码 → TypeParser"的映射关系，
// 调用 Parse 方法时自动将每个 Tag 的值设入对应的结构体字段。
// 这使得扩展新实体类型只需注册额外的组码映射即可。
type DxfParseable struct {
	tagParsers map[int]TypeParser
}

// Init 使用给定的解析器映射初始化（会覆盖原有映射）。
func (e *DxfParseable) Init(parsers map[int]TypeParser) {
	e.tagParsers = parsers
}

// Update 合并新的解析器映射到现有映射中（不会覆盖已有映射）。
func (e *DxfParseable) Update(parsers map[int]TypeParser) {
	if len(e.tagParsers) == 0 {
		e.tagParsers = parsers
	} else {
		for k, v := range parsers {
			e.tagParsers[k] = v
		}
	}
}

// Parse 遍历 tags 中的常规 Tag（排除 XData），
// 对每个 Tag 匹配注册的解析器并调用其 Parse 方法。
func (e *DxfParseable) Parse(tags TagSlice) error {
	for _, tag := range tags.RegularTags() {
		if parser, ok := e.tagParsers[tag.Code]; ok {
			if err := parser.Parse(tag.Value); err != nil {
				return err
			}
		}
	}
	return nil
}
