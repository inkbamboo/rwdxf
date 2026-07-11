package entities

import (
	"github.com/inkbamboo/rwdxf/core"
)

// XLine 表示 DXF XLINE 实体（构造线，向两端无限延伸）。
type XLine struct {
	RegularEntity
	BaseEntity
	Start       core.Point
	UnitVector  core.Point
}

func (x XLine) Equals(other core.DxfElement) bool {
	if o, ok := other.(*XLine); ok {
		return x.BaseEntity.Equals(o.BaseEntity) &&
			x.Start.Equals(o.Start) &&
			x.UnitVector.Equals(o.UnitVector)
	}
	return false
}

func (x XLine) DxfType() core.DxfTypeName { return core.DxfTypeXLine }

func (x XLine) IsR12Compatible() bool { return true }

func NewXLine(tags core.TagSlice) (*XLine, error) {
	xl := new(XLine)
	xl.UnitVector = core.Point{X: 0, Y: 0, Z: 1}
	xl.InitBaseEntityParser()
	xl.Update(map[int]core.TypeParser{
		10: core.NewFloatTypeParserToVar(&xl.Start.X),
		20: core.NewFloatTypeParserToVar(&xl.Start.Y),
		30: core.NewFloatTypeParserToVar(&xl.Start.Z),
		11: core.NewFloatTypeParserToVar(&xl.UnitVector.X),
		21: core.NewFloatTypeParserToVar(&xl.UnitVector.Y),
		31: core.NewFloatTypeParserToVar(&xl.UnitVector.Z),
	})
	xl.Parse(tags)
	xl.XData = CollectXDataFromTags(tags)
	return xl, nil
}

func (x *XLine) DxfTags() core.TagSlice {
	if R12Mode {
		return x.dxfTagsR12()
	}
	baseTags := baseEntityTags(&x.BaseEntity, "XLINE")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbXline")))
	}
	tags := append(baseTags,
		core.NewTag(10, core.NewFloatValue(x.Start.X)),
		core.NewTag(20, core.NewFloatValue(x.Start.Y)),
		core.NewTag(30, core.NewFloatValue(x.Start.Z)),
		core.NewTag(11, core.NewFloatValue(x.UnitVector.X)),
		core.NewTag(21, core.NewFloatValue(x.UnitVector.Y)),
		core.NewTag(31, core.NewFloatValue(x.UnitVector.Z)),
	)
	return AppendXData(tags, &x.BaseEntity)
}

func (x *XLine) dxfTagsR12() core.TagSlice {
	layerName := x.LayerName
	if layerName == "" {
		layerName = "0"
	}

	extent := 1e6
	ux, uy := x.UnitVector.X, x.UnitVector.Y
	sx, sy := x.Start.X, x.Start.Y
	return core.TagSlice{
		core.NewTag(0, core.NewStringValue("LINE")),
		core.NewTag(8, core.NewStringValue(layerName)),
		core.NewTag(10, core.NewFloatValue(sx - ux*extent)),
		core.NewTag(20, core.NewFloatValue(sy - uy*extent)),
		core.NewTag(30, core.NewFloatValue(0.0)),
		core.NewTag(11, core.NewFloatValue(sx + ux*extent)),
		core.NewTag(21, core.NewFloatValue(sy + uy*extent)),
		core.NewTag(31, core.NewFloatValue(0.0)),
	}
}

func NewXLineEntity(start, unitVector core.Point, layer string) *XLine {
	return &XLine{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		Start:      start,
		UnitVector: unitVector,
	}
}

func (x XLine) Clone() Entity { n := NewXLineEntity(x.Start, x.UnitVector, x.LayerName); n.BaseEntity = x.BaseEntity.CloneBase(); return n }
