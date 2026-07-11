package entities

import (
	"github.com/inkbamboo/rwdxf/core"
)

// Trace 表示 DXF TRACE 实体（宽线）。
type Trace struct {
	RegularEntity
	BaseEntity
	Vtx0               core.Point
	Vtx1               core.Point
	Vtx2               core.Point
	Vtx3               core.Point
	Thickness          float64
	ExtrusionDirection core.Point
}

func (t Trace) DxfType() core.DxfTypeName { return core.DxfTypeTrace }

func (t Trace) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Trace); ok {
		return t.BaseEntity.Equals(o.BaseEntity) &&
			t.Vtx0.Equals(o.Vtx0) &&
			t.Vtx1.Equals(o.Vtx1) &&
			t.Vtx2.Equals(o.Vtx2) &&
			t.Vtx3.Equals(o.Vtx3) &&
			core.FloatEquals(t.Thickness, o.Thickness) &&
			t.ExtrusionDirection.Equals(o.ExtrusionDirection)
	}
	return false
}

func NewTrace(tags core.TagSlice) (*Trace, error) {
	tr := new(Trace)
	tr.ExtrusionDirection = core.Point{X: 0, Y: 0, Z: 1}
	tr.InitBaseEntityParser()
	tr.Update(map[int]core.TypeParser{
		10:  core.NewFloatTypeParserToVar(&tr.Vtx0.X),
		20:  core.NewFloatTypeParserToVar(&tr.Vtx0.Y),
		30:  core.NewFloatTypeParserToVar(&tr.Vtx0.Z),
		11:  core.NewFloatTypeParserToVar(&tr.Vtx1.X),
		21:  core.NewFloatTypeParserToVar(&tr.Vtx1.Y),
		31:  core.NewFloatTypeParserToVar(&tr.Vtx1.Z),
		12:  core.NewFloatTypeParserToVar(&tr.Vtx2.X),
		22:  core.NewFloatTypeParserToVar(&tr.Vtx2.Y),
		32:  core.NewFloatTypeParserToVar(&tr.Vtx2.Z),
		13:  core.NewFloatTypeParserToVar(&tr.Vtx3.X),
		23:  core.NewFloatTypeParserToVar(&tr.Vtx3.Y),
		33:  core.NewFloatTypeParserToVar(&tr.Vtx3.Z),
		39:  core.NewFloatTypeParserToVar(&tr.Thickness),
		210: core.NewFloatTypeParserToVar(&tr.ExtrusionDirection.X),
		220: core.NewFloatTypeParserToVar(&tr.ExtrusionDirection.Y),
		230: core.NewFloatTypeParserToVar(&tr.ExtrusionDirection.Z),
	})
	tr.Parse(tags)
	tr.XData = CollectXDataFromTags(tags)
	return tr, nil
}

func (t *Trace) DxfTags() core.TagSlice {
	baseTags := baseEntityTags(&t.BaseEntity, "TRACE")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbTrace")))
	}
	vtx3 := t.Vtx3
	if core.FloatEquals(vtx3.X, 0) && core.FloatEquals(vtx3.Y, 0) && core.FloatEquals(vtx3.Z, 0) {
		vtx3 = t.Vtx2
	}
	tags := append(baseTags,
		core.NewTag(10, core.NewFloatValue(t.Vtx0.X)),
		core.NewTag(20, core.NewFloatValue(t.Vtx0.Y)),
		core.NewTag(30, core.NewFloatValue(t.Vtx0.Z)),
		core.NewTag(11, core.NewFloatValue(t.Vtx1.X)),
		core.NewTag(21, core.NewFloatValue(t.Vtx1.Y)),
		core.NewTag(31, core.NewFloatValue(t.Vtx1.Z)),
		core.NewTag(12, core.NewFloatValue(t.Vtx2.X)),
		core.NewTag(22, core.NewFloatValue(t.Vtx2.Y)),
		core.NewTag(32, core.NewFloatValue(t.Vtx2.Z)),
		core.NewTag(13, core.NewFloatValue(vtx3.X)),
		core.NewTag(23, core.NewFloatValue(vtx3.Y)),
		core.NewTag(33, core.NewFloatValue(vtx3.Z)),
	)
	if !core.FloatEquals(t.Thickness, 0) {
		tags = append(tags, core.NewTag(39, core.NewFloatValue(t.Thickness)))
	}
	if !isDefaultExtrusion(t.ExtrusionDirection) {
		tags = append(tags, pointToTags210(t.ExtrusionDirection)...)
	}
	return AppendXData(tags, &t.BaseEntity)
}

func NewTraceEntity(points []core.Point, layer string) *Trace {
	tr := &Trace{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		ExtrusionDirection: core.Point{X: 0, Y: 0, Z: 1},
	}
	if len(points) > 0 {
		tr.Vtx0 = points[0]
	}
	if len(points) > 1 {
		tr.Vtx1 = points[1]
	}
	if len(points) > 2 {
		tr.Vtx2 = points[2]
	}
	if len(points) > 3 {
		tr.Vtx3 = points[3]
	} else {
		tr.Vtx3 = tr.Vtx2
	}
	return tr
}

func (t Trace) Clone() Entity {
	n := NewTraceEntity([]core.Point{t.Vtx0, t.Vtx1, t.Vtx2, t.Vtx3}, t.LayerName)
	n.BaseEntity = t.BaseEntity.CloneBase()
	n.Thickness = t.Thickness
	n.ExtrusionDirection = t.ExtrusionDirection
	return n
}
