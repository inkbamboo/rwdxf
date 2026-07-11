package entities

import (
	"github.com/inkbamboo/rwdxf/core"
)

// Text 表示 DXF TEXT 实体（单行文字）。
type Text struct {
	RegularEntity
	BaseEntity
	Thickness          float64
	FirstAlignment     core.Point
	Height             float64
	Value              string
	Rotation           float64
	XScale             float64
	Oblique            float64
	StyleName          string
	MirroredX          bool
	MirroredY          bool
	HorizontalJust     int
	SecondAlignment    core.Point
	VerticalJust       int
	ExtrusionDirection core.Point
}

func (t Text) DxfType() core.DxfTypeName { return core.DxfTypeText }

func (t Text) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Text); ok {
		return t.BaseEntity.Equals(o.BaseEntity) &&
			t.FirstAlignment.Equals(o.FirstAlignment) &&
			core.FloatEquals(t.Height, o.Height) &&
			t.Value == o.Value
	}
	return false
}

// NewText 从 TagSlice 解析并创建 Text 实体。
func NewText(tags core.TagSlice) (*Text, error) {
	text := new(Text)
	text.ExtrusionDirection = core.Point{X: 0, Y: 0, Z: 1}
	text.XScale = 1.0
	text.InitBaseEntityParser()
	text.Update(map[int]core.TypeParser{
		39: core.NewFloatTypeParserToVar(&text.Thickness),
		10: core.NewFloatTypeParserToVar(&text.FirstAlignment.X),
		20: core.NewFloatTypeParserToVar(&text.FirstAlignment.Y),
		30: core.NewFloatTypeParserToVar(&text.FirstAlignment.Z),
		40: core.NewFloatTypeParserToVar(&text.Height),
		1:  core.NewStringTypeParserToVar(&text.Value),
		50: core.NewFloatTypeParserToVar(&text.Rotation),
		41: core.NewFloatTypeParserToVar(&text.XScale),
		51: core.NewFloatTypeParserToVar(&text.Oblique),
		7:  core.NewStringTypeParserToVar(&text.StyleName),
		71: core.NewIntTypeParser(func(v int) {
			text.MirroredX = v&2 != 0
			text.MirroredY = v&4 != 0
		}),
		72:  core.NewIntTypeParserToVar(&text.HorizontalJust),
		11:  core.NewFloatTypeParserToVar(&text.SecondAlignment.X),
		21:  core.NewFloatTypeParserToVar(&text.SecondAlignment.Y),
		31:  core.NewFloatTypeParserToVar(&text.SecondAlignment.Z),
		73:  core.NewIntTypeParserToVar(&text.VerticalJust),
		210: core.NewFloatTypeParserToVar(&text.ExtrusionDirection.X),
		220: core.NewFloatTypeParserToVar(&text.ExtrusionDirection.Y),
		230: core.NewFloatTypeParserToVar(&text.ExtrusionDirection.Z),
	})
	text.Parse(tags)
	text.XData = CollectXDataFromTags(tags)
	return text, nil
}
func (t *Text) DxfTags() core.TagSlice {
	if !R12Mode {
		return t.dxfTagsAsMText()
	}
	return t.dxfTagsAsText()
}

func (t *Text) dxfTagsAsText() core.TagSlice {
	baseTags := baseEntityTags(&t.BaseEntity, "TEXT")
	tags := append(baseTags,
		core.NewTag(10, core.NewFloatValue(t.FirstAlignment.X)),
		core.NewTag(20, core.NewFloatValue(t.FirstAlignment.Y)),
		core.NewTag(30, core.NewFloatValue(t.FirstAlignment.Z)),
		core.NewTag(40, core.NewFloatValue(t.Height)),
		core.NewTag(1, core.NewStringValue(t.Value)),
	)
	if !core.FloatEquals(t.Rotation, 0) {
		tags = append(tags, core.NewTag(50, core.NewFloatValue(t.Rotation)))
	}
	if !core.FloatEquals(t.XScale, 1.0) {
		tags = append(tags, core.NewTag(41, core.NewFloatValue(t.XScale)))
	}
	if t.StyleName != "" {
		tags = append(tags, core.NewTag(7, core.NewStringValue(t.StyleName)))
	}
	if !core.FloatEquals(t.Thickness, 0) {
		tags = append(tags, core.NewTag(39, core.NewFloatValue(t.Thickness)))
	}
	if !core.FloatEquals(t.Oblique, 0) {
		tags = append(tags, core.NewTag(51, core.NewFloatValue(t.Oblique)))
	}
	genFlags := 0
	if t.MirroredX {
		genFlags |= 2
	}
	if t.MirroredY {
		genFlags |= 4
	}
	if genFlags != 0 {
		tags = append(tags, core.NewTag(71, core.NewIntegerValue(genFlags)))
	}
	if t.HorizontalJust != 0 {
		tags = append(tags, core.NewTag(72, core.NewIntegerValue(t.HorizontalJust)))
	}
	if t.HorizontalJust != 0 || t.VerticalJust != 0 {
		if !core.FloatEquals(t.SecondAlignment.X, 0) || !core.FloatEquals(t.SecondAlignment.Y, 0) || !core.FloatEquals(t.SecondAlignment.Z, 0) {
			tags = append(tags,
				core.NewTag(11, core.NewFloatValue(t.SecondAlignment.X)),
				core.NewTag(21, core.NewFloatValue(t.SecondAlignment.Y)),
				core.NewTag(31, core.NewFloatValue(t.SecondAlignment.Z)),
			)
		}
	}
	if t.VerticalJust != 0 {
		tags = append(tags, core.NewTag(73, core.NewIntegerValue(t.VerticalJust)))
	}
	extr := t.ExtrusionDirection
	if core.FloatEquals(extr.X, 0) && core.FloatEquals(extr.Y, 0) && core.FloatEquals(extr.Z, 0) {
		extr = core.Point{X: 0, Y: 0, Z: 1}
	}
	if !isDefaultExtrusion(extr) {
		tags = append(tags, pointToTags210(extr)...)
	}
	return AppendXData(tags, &t.BaseEntity)
}

func (t *Text) dxfTagsAsMText() core.TagSlice {
	baseTags := baseEntityTags(&t.BaseEntity, "MTEXT")
	baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbMText")))
	tags := append(baseTags,
		core.NewTag(10, core.NewFloatValue(t.FirstAlignment.X)),
		core.NewTag(20, core.NewFloatValue(t.FirstAlignment.Y)),
		core.NewTag(30, core.NewFloatValue(t.FirstAlignment.Z)),
		core.NewTag(40, core.NewFloatValue(t.Height)),
		core.NewTag(71, core.NewIntegerValue(1)),
		core.NewTag(72, core.NewIntegerValue(1)),
		core.NewTag(1, core.NewStringValue(t.Value)),
	)
	if t.StyleName != "" {
		tags = append(tags, core.NewTag(7, core.NewStringValue(t.StyleName)))
	}
	if !core.FloatEquals(t.Rotation, 0) {
		tags = append(tags, core.NewTag(50, core.NewFloatValue(t.Rotation)))
	}
	extr := t.ExtrusionDirection
	if core.FloatEquals(extr.X, 0) && core.FloatEquals(extr.Y, 0) && core.FloatEquals(extr.Z, 0) {
		extr = core.Point{X: 0, Y: 0, Z: 1}
	}
	if !isDefaultExtrusion(extr) {
		tags = append(tags, pointToTags210(extr)...)
	}
	return AppendXData(tags, &t.BaseEntity)
}

// NewTextEntity 直接创建一个 Text 实体。
func NewTextEntity(value string, pos core.Point, height float64, layer string) *Text {
	return &Text{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		FirstAlignment: pos,
		Height:         height,
		Value:          value,
		XScale:         1.0,
	}
}

func (t Text) Clone() Entity {
	n := NewTextEntity(t.Value, t.FirstAlignment, t.Height, t.LayerName)
	n.BaseEntity = t.BaseEntity.CloneBase()
	n.Rotation = t.Rotation
	n.XScale = t.XScale
	n.Oblique = t.Oblique
	n.StyleName = t.StyleName
	n.Thickness = t.Thickness
	n.HorizontalJust = t.HorizontalJust
	n.VerticalJust = t.VerticalJust
	n.SecondAlignment = t.SecondAlignment
	n.ExtrusionDirection = t.ExtrusionDirection
	n.MirroredX = t.MirroredX
	n.MirroredY = t.MirroredY
	return n
}
