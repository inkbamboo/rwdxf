package entities

import (
	"strconv"

	"github.com/inkbamboo/rwdxf/core"
)

// MText 表示 DXF MTEXT 实体（多行文字）。
type MText struct {
	RegularEntity
	BaseEntity
	InsertionPoint      core.Point
	Height              float64
	Value               string
	AttachmentPoint     int
	DrawingDirection    int
	StyleName           string
	Rotation            float64
	RectWidth           float64
	LineSpacingFactor   float64
	LineSpacingStyle    int
	BackgroundFillFlag  int
	BackgroundFillColor int
	BackgroundFillRGB   core.TrueColor
	ExtrusionDirection  core.Point
}

func (m MText) Equals(other core.DxfElement) bool {
	if o, ok := other.(*MText); ok {
		return m.BaseEntity.Equals(o.BaseEntity) &&
			m.InsertionPoint.Equals(o.InsertionPoint) &&
			core.FloatEquals(m.Height, o.Height) &&
			m.Value == o.Value &&
			m.AttachmentPoint == o.AttachmentPoint
	}
	return false
}

func (m MText) DxfType() core.DxfTypeName { return core.DxfTypeMText }

func (m MText) IsR12Compatible() bool { return true }

// NewMText 从 TagSlice 解析并创建 MText 实体。
func NewMText(tags core.TagSlice) (*MText, error) {
	mtext := new(MText)
	mtext.ExtrusionDirection = core.Point{X: 0, Y: 0, Z: 1}
	mtext.AttachmentPoint = 1
	mtext.DrawingDirection = 1
	mtext.LineSpacingStyle = 1

	mtext.InitBaseEntityParser()
	mtext.Update(map[int]core.TypeParser{
		10: core.NewFloatTypeParserToVar(&mtext.InsertionPoint.X),
		20: core.NewFloatTypeParserToVar(&mtext.InsertionPoint.Y),
		30: core.NewFloatTypeParserToVar(&mtext.InsertionPoint.Z),
		40: core.NewFloatTypeParserToVar(&mtext.Height),
		1:  core.NewStringTypeParserToVar(&mtext.Value),
		71: core.NewIntTypeParserToVar(&mtext.AttachmentPoint),
		72: core.NewIntTypeParserToVar(&mtext.DrawingDirection),
		7:  core.NewStringTypeParserToVar(&mtext.StyleName),
		50: core.NewFloatTypeParserToVar(&mtext.Rotation),
		41: core.NewFloatTypeParserToVar(&mtext.RectWidth),
		44: core.NewFloatTypeParserToVar(&mtext.LineSpacingFactor),
		73: core.NewIntTypeParserToVar(&mtext.LineSpacingStyle),
		90: core.NewIntTypeParserToVar(&mtext.BackgroundFillFlag),
		63: core.NewIntTypeParserToVar(&mtext.BackgroundFillColor),
		420: core.NewIntTypeParser(func(color int) {
			mtext.BackgroundFillRGB.R = uint8((color >> 16) & 0xFF)
			mtext.BackgroundFillRGB.G = uint8((color >> 8) & 0xFF)
			mtext.BackgroundFillRGB.B = uint8(color & 0xFF)
		}),
		210: core.NewFloatTypeParserToVar(&mtext.ExtrusionDirection.X),
		220: core.NewFloatTypeParserToVar(&mtext.ExtrusionDirection.Y),
		230: core.NewFloatTypeParserToVar(&mtext.ExtrusionDirection.Z),
	})
	mtext.Parse(tags)
	mtext.XData = CollectXDataFromTags(tags)
	return mtext, nil
}
func (m *MText) DxfTags() core.TagSlice {
	if R12Mode {
		return m.dxfTagsR12()
	}
	baseTags := baseEntityTags(&m.BaseEntity, "MTEXT")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbMText")))
	}
	tags := append(baseTags,
		core.NewTag(10, core.NewFloatValue(m.InsertionPoint.X)),
		core.NewTag(20, core.NewFloatValue(m.InsertionPoint.Y)),
		core.NewTag(30, core.NewFloatValue(m.InsertionPoint.Z)),
		core.NewTag(40, core.NewFloatValue(m.Height)),
		core.NewTag(71, core.NewIntegerValue(m.AttachmentPoint)),
		core.NewTag(72, core.NewIntegerValue(m.DrawingDirection)),
		core.NewTag(1, core.NewStringValue(escapeMText(m.Value))),
	)
	if m.StyleName != "" {
		tags = append(tags, core.NewTag(7, core.NewStringValue(m.StyleName)))
	}
	if !core.FloatEquals(m.Rotation, 0) {
		tags = append(tags, core.NewTag(50, core.NewFloatValue(m.Rotation)))
	}
	if !core.FloatEquals(m.RectWidth, 0) {
		tags = append(tags, core.NewTag(41, core.NewFloatValue(m.RectWidth)))
	}
	if !core.FloatEquals(m.LineSpacingFactor, 0) {
		tags = append(tags, core.NewTag(44, core.NewFloatValue(m.LineSpacingFactor)))
	}
	if m.LineSpacingStyle != 1 {
		tags = append(tags, core.NewTag(73, core.NewIntegerValue(m.LineSpacingStyle)))
	}
	if m.BackgroundFillFlag != 0 {
		tags = append(tags,
			core.NewTag(90, core.NewIntegerValue(m.BackgroundFillFlag)),
		)
		if m.BackgroundFillRGB.R != 0 || m.BackgroundFillRGB.G != 0 || m.BackgroundFillRGB.B != 0 {
			v := int(m.BackgroundFillRGB.R)<<16 | int(m.BackgroundFillRGB.G)<<8 | int(m.BackgroundFillRGB.B)
			tags = append(tags, core.NewTag(420, core.NewIntegerValue(v)))
		}
	}
	if !isDefaultExtrusion(m.ExtrusionDirection) {
		tags = append(tags, pointToTags210(m.ExtrusionDirection)...)
	}
	return AppendXData(tags, &m.BaseEntity)
}

func (m *MText) dxfTagsR12() core.TagSlice {
	layerName := m.LayerName
	if layerName == "" {
		layerName = "0"
	}
	blockName := "*U" + strconv.Itoa(NextR12BlockSeq())

	height := m.Height
	if core.FloatEquals(height, 0) {
		height = 0.2
	}
	lineSpacing := height * 1.667
	if !core.FloatEquals(m.LineSpacingFactor, 0) {
		lineSpacing = height * m.LineSpacingFactor
	}

	h1 := nextR12HatchHandle()

	blockTags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("BLOCK")),
		core.NewTag(5, core.NewStringValue(h1)),
		core.NewTag(8, core.NewStringValue(layerName)),
		core.NewTag(2, core.NewStringValue(blockName)),
		core.NewTag(70, core.NewIntegerValue(1)),
		core.NewTag(10, core.NewFloatValue(0.0)),
		core.NewTag(20, core.NewFloatValue(0.0)),
		core.NewTag(30, core.NewFloatValue(0.0)),
		core.NewTag(3, core.NewStringValue(blockName)),
		core.NewTag(1, core.NewStringValue("")),
	}

	lines := splitLines(m.Value)
	for i, line := range lines {
		y := m.InsertionPoint.Y - float64(i)*lineSpacing
		h := nextR12HatchHandle()
		blockTags = append(blockTags,
			core.NewTag(0, core.NewStringValue("TEXT")),
			core.NewTag(5, core.NewStringValue(h)),
			core.NewTag(8, core.NewStringValue(layerName)),
			core.NewTag(10, core.NewFloatValue(m.InsertionPoint.X)),
			core.NewTag(20, core.NewFloatValue(y)),
			core.NewTag(30, core.NewFloatValue(m.InsertionPoint.Z)),
			core.NewTag(40, core.NewFloatValue(height)),
			core.NewTag(1, core.NewStringValue(line)),
		)
	}

	hEnd := nextR12HatchHandle()
	blockTags = append(blockTags,
		core.NewTag(0, core.NewStringValue("ENDBLK")),
		core.NewTag(5, core.NewStringValue(hEnd)),
		core.NewTag(8, core.NewStringValue(layerName)),
	)
	R12ExtraBlocks[blockName] = blockTags

	var tags core.TagSlice
	tags = append(tags,
		core.NewTag(0, core.NewStringValue("INSERT")),
		core.NewTag(8, core.NewStringValue(layerName)),
		core.NewTag(2, core.NewStringValue(blockName)),
		core.NewTag(10, core.NewFloatValue(0.0)),
		core.NewTag(20, core.NewFloatValue(0.0)),
		core.NewTag(30, core.NewFloatValue(0.0)),
	)
	return AppendXData(tags, &m.BaseEntity)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]

			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}

func escapeMText(s string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			result += "\\P"
		} else if s[i] == '\r' {

		} else {
			result += string(s[i])
		}
	}
	return result
}

// NewMTextEntity 直接创建一个 MText 实体。
func NewMTextEntity(value string, pos core.Point, height float64, layer string) *MText {
	return &MText{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		InsertionPoint:     pos,
		Height:             height,
		Value:              value,
		AttachmentPoint:    1,
		DrawingDirection:   1,
		ExtrusionDirection: core.Point{X: 0, Y: 0, Z: 1},
	}
}

func (m MText) Clone() Entity {
	n := NewMTextEntity(m.Value, m.InsertionPoint, m.Height, m.LayerName)
	n.BaseEntity = m.BaseEntity.CloneBase()
	n.AttachmentPoint = m.AttachmentPoint
	n.DrawingDirection = m.DrawingDirection
	n.Rotation = m.Rotation
	n.StyleName = m.StyleName
	n.RectWidth = m.RectWidth
	n.LineSpacingFactor = m.LineSpacingFactor
	n.LineSpacingStyle = m.LineSpacingStyle
	n.BackgroundFillFlag = m.BackgroundFillFlag
	n.BackgroundFillColor = m.BackgroundFillColor
	n.BackgroundFillRGB = m.BackgroundFillRGB
	n.ExtrusionDirection = m.ExtrusionDirection
	return n
}
