package sections

import "github.com/inkbamboo/rwdxf/core"

// 文字样式位标志。
const verticalTextBit = 0x4 // 垂直文字
const shapeBit = 0x1         // 形状字体
const backwardsBit = 0x2     // 反向
const upsideDownBit = 0x4    // 倒置

// Style 表示 DXF 文字样式（Text Style），用于控制 TEXT/MTEXT 的外观。
type Style struct {
	core.DxfParseable
	Name           string
	Height         float64
	Width          float64
	Oblique        float64
	IsBackwards    bool
	IsUpsideDown   bool
	IsShape        bool
	IsVerticalText bool
	Font           string
	BigFont        string
}

func (s Style) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Style); ok {
		return s.Name == o.Name &&
			core.FloatEquals(s.Height, o.Height) &&
			core.FloatEquals(s.Width, o.Width) &&
			core.FloatEquals(s.Oblique, o.Oblique) &&
			s.IsBackwards == o.IsBackwards &&
			s.IsUpsideDown == o.IsUpsideDown &&
			s.IsShape == o.IsShape &&
			s.IsVerticalText == o.IsVerticalText &&
			s.Font == o.Font &&
			s.BigFont == o.BigFont
	}
	return false
}

// NewStyle 从 TagSlice 解析文字样式。
func NewStyle(tags core.TagSlice) (*Style, error) {
	style := new(Style)
	style.Height = 1.0
	style.Width = 1.0

	style.Init(map[int]core.TypeParser{
		2:  core.NewStringTypeParserToVar(&style.Name),
		3:  core.NewStringTypeParserToVar(&style.Font),
		4:  core.NewStringTypeParserToVar(&style.BigFont),
		40: core.NewFloatTypeParserToVar(&style.Height),
		41: core.NewFloatTypeParserToVar(&style.Width),
		50: core.NewFloatTypeParserToVar(&style.Oblique),
		70: core.NewIntTypeParser(func(flags int) {
			style.IsShape = flags&shapeBit != 0
			style.IsVerticalText = flags&verticalTextBit != 0
		}),
		71: core.NewIntTypeParser(func(flags int) {
			style.IsBackwards = flags&backwardsBit != 0
			style.IsUpsideDown = flags&upsideDownBit != 0
		}),
	})
	return style, style.Parse(tags)
}

// NewStyleTable 从 TagSlice 解析文字样式表。
func NewStyleTable(tags core.TagSlice) (Table, error) {
	table := make(Table)
	tableSlices, err := TableEntryTags(tags)
	if err != nil {
		return table, err
	}
	for _, slice := range tableSlices {
		style, err := NewStyle(slice)
		if err != nil {
			return nil, err
		}
		table[style.Name] = style
	}
	return table, nil
}
