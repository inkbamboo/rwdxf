package core

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// Tag 表示 DXF 文件中的一个组码-值对。
// Code 是 DXF 组码，Value 是该组码对应的数据值。
type Tag struct {
	Code  int
	Value DataType
}

func (tag Tag) ToString() string {
	return "{ Code: " + strconv.Itoa(tag.Code) + "; Value: " + tag.Value.ToString() + " }"
}

func (tag Tag) Equals(other DxfElement) bool {
	if otherTag, ok := other.(*Tag); ok {
		return tag.Code == otherTag.Code && tag.Value.Equals(otherTag.Value)
	}
	return false
}

// NewTag 创建一个新的 Tag 实例。
func NewTag(code int, value DataType) *Tag {
	return &Tag{Code: code, Value: value}
}

// NoneTag 是一个哨兵 Tag，用于标记 Tag 流的结束。
var NoneTag = Tag{999999, NewStringValue("NONE")}

// 特殊组码常量。
const appDataMarker = 102  // 应用程序数据标记
const subclassMarker = 100 // 子类标记

// NextTagFunction 是一个闭包类型，每次调用返回下一个 Tag。
type NextTagFunction func() (*Tag, error)

// Tagger 从 io.Reader 创建一个 Tag 扫描器闭包。
// 每次调用返回的闭包会读取两行（组码 + 值），
// 根据组码范围自动推断数据类型（0-9→字符串，10-59→浮点，60-99→整数等）。
func Tagger(stream io.Reader) NextTagFunction {
	scanner := bufio.NewScanner(stream)

	readLine := func() (string, error) {
		if scanner.Scan() {
			return scanner.Text(), nil
		}
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", nil
	}

	return func() (*Tag, error) {
		code, err := readLine()
		if err != nil {
			return &NoneTag, err
		}
		value, err := readLine()
		if err != nil {
			return &NoneTag, err
		}

	charsToTrim := " \r\n"
	if len(code) > 0 {
		intCode, err := strconv.Atoi(strings.Trim(code, charsToTrim))
		if err != nil {
			return &NoneTag, err
		}
		var valueType DataType
		if factory, ok := groupCodeTypes[intCode]; ok {
			valueType, _ = factory(strings.Trim(value, charsToTrim))
		} else {

			valueType, _ = NewString(strings.Trim(value, charsToTrim))
		}
		return NewTag(intCode, valueType), nil
	}
	return &NoneTag, nil
	}
}

// AllTags 从 NextTagFunction 中读取所有 Tag 直到遇到 NoneTag。
func AllTags(next NextTagFunction) []*Tag {
	tags := make([]*Tag, 0)
	tag, _ := next()
	for *tag != NoneTag {
		tags = append(tags, tag)
		tag, _ = next()
	}
	return tags
}

// TagSlice 是一组 Tag 的切片，提供过滤、查找、分组等方法。
type TagSlice []*Tag

func (slice TagSlice) Equals(other DxfElement) bool {
	if otherSlice, ok := other.(TagSlice); ok {
		if len(slice) != len(otherSlice) {
			return false
		}
		for i, tag := range slice {
			if !tag.Equals(otherSlice[i]) {
				return false
			}
		}
		return true
	}
	return false
}

// TagIndex 在 [start, end) 范围内查找指定组码的第一个 Tag 索引，未找到返回 -1。
func (slice TagSlice) TagIndex(tagCode int, start, end int) int {
	for i := start; i < end; i++ {
		if slice[i].Code == tagCode {
			return i
		}
	}
	return -1
}

// AllWithCode 返回所有组码等于指定值的 Tag。
func (slice TagSlice) AllWithCode(tagCode int) []*Tag {
	tags := make([]*Tag, 0)
	for _, tag := range slice {
		if tag.Code == tagCode {
			tags = append(tags, tag)
		}
	}
	return tags
}

// RegularTags 返回过滤掉 XData（组码 >=1000）和应用数据标记段的常规 Tag。
func (slice TagSlice) RegularTags() []*Tag {
	tags := make([]*Tag, 0)
	inAppDataRange := false
	for _, tag := range slice {
		if tag.Code >= 1000 {
			continue
		}
		if tag.Code == appDataMarker {
			if inAppDataRange {
				inAppDataRange = false
			} else {
				inAppDataRange = true
			}
			continue
		}
		if !inAppDataRange {
			tags = append(tags, tag)
		}
	}
	return tags
}

// XDataTags 返回所有扩展数据 Tag（组码 > 999）。
func (slice TagSlice) XDataTags() []*Tag {
	tags := make([]*Tag, 0)
	for _, tag := range slice {
		if tag.Code > 999 {
			tags = append(tags, tag)
		}
	}
	return tags
}

// TagGroups 按指定的 splitCode 将 TagSlice 拆分为多个 TagSlice 组。
// 每个组以 splitCode 开头，后续 Tag 归入同一组直到下一个 splitCode 出现。
func TagGroups(tags TagSlice, splitCode int) []TagSlice {
	groups := make([]TagSlice, 0)
	group := make(TagSlice, 0)
	for _, tag := range tags {
		if tag.Code == splitCode {
			if len(group) > 0 {
				groups = append(groups, group)
				group = make(TagSlice, 0)
			}
			group = append(group, tag)
		} else if len(group) > 0 {
			group = append(group, tag)
		}
	}
	if len(group) > 0 && group[0].Code == splitCode {
		groups = append(groups, group)
	}
	return groups
}

type dataTypeFactory func(string) (DataType, error)

var groupCodeTypes map[int]dataTypeFactory

func init() {
	groupCodeTypes = make(map[int]dataTypeFactory)
	for code := 0; code < 10; code++ {
		groupCodeTypes[code] = NewString
	}
	for code := 10; code < 60; code++ {
		groupCodeTypes[code] = NewFloat
	}
	for code := 60; code < 100; code++ {
		groupCodeTypes[code] = NewInteger
	}
	for code := 100; code < 106; code++ {
		groupCodeTypes[code] = NewString
	}
	for code := 110; code < 150; code++ {
		groupCodeTypes[code] = NewFloat
	}
	for code := 170; code < 180; code++ {
		groupCodeTypes[code] = NewInteger
	}
	for code := 210; code < 240; code++ {
		groupCodeTypes[code] = NewFloat
	}
	for code := 270; code < 300; code++ {
		groupCodeTypes[code] = NewInteger
	}
	for code := 300; code < 370; code++ {
		groupCodeTypes[code] = NewString
	}
	for code := 370; code < 390; code++ {
		groupCodeTypes[code] = NewInteger
	}
	for code := 390; code < 400; code++ {
		groupCodeTypes[code] = NewString
	}
	for code := 400; code < 410; code++ {
		groupCodeTypes[code] = NewInteger
	}
	for code := 410; code < 420; code++ {
		groupCodeTypes[code] = NewString
	}
	for code := 420; code < 430; code++ {
		groupCodeTypes[code] = NewInteger
	}
	for code := 430; code < 440; code++ {
		groupCodeTypes[code] = NewString
	}
	for code := 440; code < 460; code++ {
		groupCodeTypes[code] = NewInteger
	}
	for code := 460; code < 470; code++ {
		groupCodeTypes[code] = NewFloat
	}
	for code := 470; code < 482; code++ {
		groupCodeTypes[code] = NewString
	}
	for code := 999; code < 1010; code++ {
		groupCodeTypes[code] = NewString
	}
	for code := 1010; code < 1060; code++ {
		groupCodeTypes[code] = NewFloat
	}
	for code := 1060; code < 1072; code++ {
		groupCodeTypes[code] = NewInteger
	}
}
