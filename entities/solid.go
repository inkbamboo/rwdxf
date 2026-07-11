package entities

import (
	"github.com/inkbamboo/rwdxf/core"
)

// Solid 表示 DXF SOLID 实体（实心填充区域）。
type Solid struct {
	RegularEntity
	BaseEntity
	Vtx0               core.Point
	Vtx1               core.Point
	Vtx2               core.Point
	Vtx3               core.Point
	Thickness          float64
	ExtrusionDirection core.Point
}

func (s Solid) DxfType() core.DxfTypeName { return core.DxfTypeSolid }

func (s Solid) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Solid); ok {
		return s.BaseEntity.Equals(o.BaseEntity) &&
			s.Vtx0.Equals(o.Vtx0) &&
			s.Vtx1.Equals(o.Vtx1) &&
			s.Vtx2.Equals(o.Vtx2) &&
			s.Vtx3.Equals(o.Vtx3) &&
			core.FloatEquals(s.Thickness, o.Thickness) &&
			s.ExtrusionDirection.Equals(o.ExtrusionDirection)
	}
	return false
}

func NewSolid(tags core.TagSlice) (*Solid, error) {
	s := new(Solid)
	s.ExtrusionDirection = core.Point{X: 0, Y: 0, Z: 1}
	s.InitBaseEntityParser()
	s.Update(map[int]core.TypeParser{
		10:  core.NewFloatTypeParserToVar(&s.Vtx0.X),
		20:  core.NewFloatTypeParserToVar(&s.Vtx0.Y),
		30:  core.NewFloatTypeParserToVar(&s.Vtx0.Z),
		11:  core.NewFloatTypeParserToVar(&s.Vtx1.X),
		21:  core.NewFloatTypeParserToVar(&s.Vtx1.Y),
		31:  core.NewFloatTypeParserToVar(&s.Vtx1.Z),
		12:  core.NewFloatTypeParserToVar(&s.Vtx2.X),
		22:  core.NewFloatTypeParserToVar(&s.Vtx2.Y),
		32:  core.NewFloatTypeParserToVar(&s.Vtx2.Z),
		13:  core.NewFloatTypeParserToVar(&s.Vtx3.X),
		23:  core.NewFloatTypeParserToVar(&s.Vtx3.Y),
		33:  core.NewFloatTypeParserToVar(&s.Vtx3.Z),
		39:  core.NewFloatTypeParserToVar(&s.Thickness),
		210: core.NewFloatTypeParserToVar(&s.ExtrusionDirection.X),
		220: core.NewFloatTypeParserToVar(&s.ExtrusionDirection.Y),
		230: core.NewFloatTypeParserToVar(&s.ExtrusionDirection.Z),
	})
	s.Parse(tags)
	s.XData = CollectXDataFromTags(tags)
	return s, nil
}

func (s *Solid) DxfTags() core.TagSlice {
	baseTags := baseEntityTags(&s.BaseEntity, "SOLID")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbTrace")))
	}

	vtx3 := s.Vtx3
	if core.FloatEquals(vtx3.X, 0) && core.FloatEquals(vtx3.Y, 0) && core.FloatEquals(vtx3.Z, 0) {
		vtx3 = s.Vtx2
	}
	tags := append(baseTags,
		core.NewTag(10, core.NewFloatValue(s.Vtx0.X)),
		core.NewTag(20, core.NewFloatValue(s.Vtx0.Y)),
		core.NewTag(30, core.NewFloatValue(s.Vtx0.Z)),
		core.NewTag(11, core.NewFloatValue(s.Vtx1.X)),
		core.NewTag(21, core.NewFloatValue(s.Vtx1.Y)),
		core.NewTag(31, core.NewFloatValue(s.Vtx1.Z)),
		core.NewTag(12, core.NewFloatValue(s.Vtx2.X)),
		core.NewTag(22, core.NewFloatValue(s.Vtx2.Y)),
		core.NewTag(32, core.NewFloatValue(s.Vtx2.Z)),
		core.NewTag(13, core.NewFloatValue(vtx3.X)),
		core.NewTag(23, core.NewFloatValue(vtx3.Y)),
		core.NewTag(33, core.NewFloatValue(vtx3.Z)),
	)
	if !core.FloatEquals(s.Thickness, 0) {
		tags = append(tags, core.NewTag(39, core.NewFloatValue(s.Thickness)))
	}
	if !isDefaultExtrusion(s.ExtrusionDirection) {
		tags = append(tags, pointToTags210(s.ExtrusionDirection)...)
	}
	return AppendXData(tags, &s.BaseEntity)
}

func NewSolidEntity(points []core.Point, layer string) *Solid {
	s := &Solid{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		ExtrusionDirection: core.Point{X: 0, Y: 0, Z: 1},
	}
	if len(points) > 0 {
		s.Vtx0 = points[0]
	}
	if len(points) > 1 {
		s.Vtx1 = points[1]
	}
	if len(points) > 2 {
		s.Vtx2 = points[2]
	}
	if len(points) > 3 {
		s.Vtx3 = points[3]
	} else {
		s.Vtx3 = s.Vtx2
	}
	return s
}

func (s Solid) Clone() Entity {
	n := NewSolidEntity([]core.Point{s.Vtx0, s.Vtx1, s.Vtx2, s.Vtx3}, s.LayerName)
	n.BaseEntity = s.BaseEntity.CloneBase()
	n.Thickness = s.Thickness
	n.ExtrusionDirection = s.ExtrusionDirection
	return n
}
