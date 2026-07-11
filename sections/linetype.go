package sections

import "github.com/inkbamboo/rwdxf/core"

// 线型元素位标志。
const absRotationBit = 0x1  // 绝对旋转
const textStringBit = 0x2   // 文本字符串
const elementShapeBit = 0x4 // 形状元素

// LineElement 表示线型定义中的一个图案元素（线段、文本或形状）。
type LineElement struct {
	Length           float64
	AbsoluteRotation bool
	IsTextString     bool
	IsShape          bool
	ShapeNumber      int
	Scale            float64
	RotationAngle    float64
	XOffset          float64
	YOffset          float64
	Text             string
}

func (e LineElement) Equals(other LineElement) bool {
	return core.FloatEquals(e.Length, other.Length) &&
		e.AbsoluteRotation == other.AbsoluteRotation &&
		e.IsTextString == other.IsTextString &&
		e.IsShape == other.IsShape &&
		e.ShapeNumber == other.ShapeNumber &&
		core.FloatEquals(e.Scale, other.Scale) &&
		core.FloatEquals(e.RotationAngle, other.RotationAngle) &&
		core.FloatEquals(e.XOffset, other.XOffset) &&
		core.FloatEquals(e.YOffset, other.YOffset) &&
		e.Text == other.Text
}

// LineType 表示 DXF 线型定义，包含名称、描述和图案元素列表。
type LineType struct {
	core.DxfParseable
	Name        string
	Description string
	Length      float64
	Pattern     []*LineElement
}

func (l LineType) Equals(other core.DxfElement) bool {
	if o, ok := other.(*LineType); ok {
		if l.Name != o.Name ||
			l.Description != o.Description ||
			!core.FloatEquals(l.Length, o.Length) ||
			len(l.Pattern) != len(o.Pattern) {
			return false
		}
		for i, p1 := range l.Pattern {
			if !p1.Equals(*o.Pattern[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func (l *LineType) Clone() *LineType {
	c := &LineType{
		Name:        l.Name,
		Description: l.Description,
		Length:      l.Length,
		Pattern:     make([]*LineElement, len(l.Pattern)),
	}
	for i, e := range l.Pattern {
		if e != nil {
			el := *e
			c.Pattern[i] = &el
		}
	}
	return c
}

// NewLineType 从 TagSlice 解析线型。
func NewLineType(tags core.TagSlice) (*LineType, error) {
	ltype := new(LineType)
	ltype.Pattern = make([]*LineElement, 0)
	flags74 := 0
	var lineElement *LineElement

	ltype.Init(map[int]core.TypeParser{
		2:  core.NewStringTypeParserToVar(&ltype.Name),
		3:  core.NewStringTypeParserToVar(&ltype.Description),
		40: core.NewFloatTypeParserToVar(&ltype.Length),
		49: core.NewFloatTypeParser(func(length float64) {
			if lineElement != nil {
				ltype.Pattern = append(ltype.Pattern, lineElement)
			}
			lineElement = new(LineElement)
			lineElement.Scale = 1.0
			lineElement.Length = length
		}),
		74: core.NewIntTypeParser(func(flags int) {
			flags74 = flags
			if flags74 > 0 {
				lineElement.AbsoluteRotation = flags74&absRotationBit > 0
				lineElement.IsTextString = flags74&textStringBit > 0
				lineElement.IsShape = flags74&elementShapeBit > 0
			}
		}),
		75: core.NewIntTypeParser(func(flags int) {
			if lineElement != nil && lineElement.IsShape {
				lineElement.ShapeNumber = flags
			}
		}),
		46: core.NewFloatTypeParser(func(scale float64) {
			if lineElement != nil {
				lineElement.Scale = scale
			}
		}),
		50: core.NewFloatTypeParser(func(angle float64) {
			if lineElement != nil {
				lineElement.RotationAngle = angle
			}
		}),
		44: core.NewFloatTypeParser(func(x float64) {
			if lineElement != nil {
				lineElement.XOffset = x
			}
		}),
		45: core.NewFloatTypeParser(func(y float64) {
			if lineElement != nil {
				lineElement.YOffset = y
			}
		}),
		9: core.NewStringTypeParser(func(text string) {
			if lineElement != nil {
				lineElement.Text = text
			}
		}),
	})

	err := ltype.Parse(tags)

	if lineElement != nil {
		ltype.Pattern = append(ltype.Pattern, lineElement)
	}
	return ltype, err
}

// NewLineTypeTable 从 TagSlice 解析线型表。
func NewLineTypeTable(tags core.TagSlice) (Table, error) {
	table := make(Table)
	tableSlices, err := TableEntryTags(tags)
	if err != nil {
		return table, err
	}
	for _, slice := range tableSlices {
		ltype, err := NewLineType(slice)
		if err != nil {
			return nil, err
		}
		table[ltype.Name] = ltype
	}
	return table, nil
}
