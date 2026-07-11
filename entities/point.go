package entities

import (
	"github.com/inkbamboo/rwdxf/core"
)

// PointEntity 表示 DXF POINT 实体。
type PointEntity struct {
	RegularEntity
	BaseEntity
	Thickness          float64
	Position           core.Point
	Angle              float64
	ExtrusionDirection core.Point
}

func (p PointEntity) DxfType() core.DxfTypeName { return core.DxfTypePoint }

func (p PointEntity) Equals(other core.DxfElement) bool {
	if o, ok := other.(*PointEntity); ok {
		return p.BaseEntity.Equals(o.BaseEntity) &&
			core.FloatEquals(p.Thickness, o.Thickness) &&
			p.Position.Equals(o.Position) &&
			core.FloatEquals(p.Angle, o.Angle) &&
			p.ExtrusionDirection.Equals(o.ExtrusionDirection)
	}
	return false
}

func (p *PointEntity) Translate(dx, dy, dz float64) {
	p.Position.X += dx
	p.Position.Y += dy
	p.Position.Z += dz
}

// NewPoint 从 TagSlice 解析并创建 Point 实体。
func NewPoint(tags core.TagSlice) (*PointEntity, error) {
	pt := new(PointEntity)
	pt.ExtrusionDirection = core.Point{X: 0, Y: 0, Z: 1}
	pt.InitBaseEntityParser()
	pt.Update(map[int]core.TypeParser{
		39:  core.NewFloatTypeParserToVar(&pt.Thickness),
		10:  core.NewFloatTypeParserToVar(&pt.Position.X),
		20:  core.NewFloatTypeParserToVar(&pt.Position.Y),
		30:  core.NewFloatTypeParserToVar(&pt.Position.Z),
		50:  core.NewFloatTypeParserToVar(&pt.Angle),
		210: core.NewFloatTypeParserToVar(&pt.ExtrusionDirection.X),
		220: core.NewFloatTypeParserToVar(&pt.ExtrusionDirection.Y),
		230: core.NewFloatTypeParserToVar(&pt.ExtrusionDirection.Z),
	})
	pt.Parse(tags)
	pt.XData = CollectXDataFromTags(tags)
	return pt, nil
}
func (p *PointEntity) DxfTags() core.TagSlice {
	baseTags := baseEntityTags(&p.BaseEntity, "POINT")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbPoint")))
	}
	tags := append(baseTags,
		core.NewTag(10, core.NewFloatValue(p.Position.X)),
		core.NewTag(20, core.NewFloatValue(p.Position.Y)),
		core.NewTag(30, core.NewFloatValue(p.Position.Z)),
	)
	if !core.FloatEquals(p.Thickness, 0) {
		tags = append(tags, core.NewTag(39, core.NewFloatValue(p.Thickness)))
	}
	if !core.FloatEquals(p.Angle, 0) {
		tags = append(tags, core.NewTag(50, core.NewFloatValue(p.Angle)))
	}
	extr := p.ExtrusionDirection
	if core.FloatEquals(extr.X, 0) && core.FloatEquals(extr.Y, 0) && core.FloatEquals(extr.Z, 0) {
		extr = core.Point{X: 0, Y: 0, Z: 1}
	}
	if !isDefaultExtrusion(extr) {
		tags = append(tags, pointToTags210(extr)...)
	}
	return AppendXData(tags, &p.BaseEntity)
}

// NewPointEntity 直接创建一个 Point 实体。
func NewPointEntity(pos core.Point, layer string) *PointEntity {
	return &PointEntity{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		Position:           pos,
		ExtrusionDirection: core.Point{X: 0, Y: 0, Z: 1},
	}
}

func (p PointEntity) Clone() Entity {
	n := NewPointEntity(p.Position, p.LayerName)
	n.BaseEntity = p.BaseEntity.CloneBase()
	n.Thickness = p.Thickness
	n.Angle = p.Angle
	n.ExtrusionDirection = p.ExtrusionDirection
	return n
}
