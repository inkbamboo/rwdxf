package entities

import (
	"github.com/inkbamboo/rwdxf/core"
)

// Face3D 表示 DXF 3DFACE 实体（三维面）。
type Face3D struct {
	RegularEntity
	BaseEntity
	Vtx0           core.Point
	Vtx1           core.Point
	Vtx2           core.Point
	Vtx3           core.Point
	InvisibleEdges int
}

func (f Face3D) DxfType() core.DxfTypeName { return core.DxfTypeFace3D }

func (f Face3D) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Face3D); ok {
		return f.BaseEntity.Equals(o.BaseEntity) &&
			f.Vtx0.Equals(o.Vtx0) &&
			f.Vtx1.Equals(o.Vtx1) &&
			f.Vtx2.Equals(o.Vtx2) &&
			f.Vtx3.Equals(o.Vtx3) &&
			f.InvisibleEdges == o.InvisibleEdges
	}
	return false
}

func (f *Face3D) IsInvisibleEdge(num int) bool {
	return f.InvisibleEdges&(1<<num) != 0
}

func (f *Face3D) SetEdgeVisibility(num int, visible bool) {
	if visible {
		f.InvisibleEdges &= ^(1 << num)
	} else {
		f.InvisibleEdges |= (1 << num)
	}
}

func NewFace3D(tags core.TagSlice) (*Face3D, error) {
	f := new(Face3D)
	f.InitBaseEntityParser()
	f.Update(map[int]core.TypeParser{
		10: core.NewFloatTypeParserToVar(&f.Vtx0.X),
		20: core.NewFloatTypeParserToVar(&f.Vtx0.Y),
		30: core.NewFloatTypeParserToVar(&f.Vtx0.Z),
		11: core.NewFloatTypeParserToVar(&f.Vtx1.X),
		21: core.NewFloatTypeParserToVar(&f.Vtx1.Y),
		31: core.NewFloatTypeParserToVar(&f.Vtx1.Z),
		12: core.NewFloatTypeParserToVar(&f.Vtx2.X),
		22: core.NewFloatTypeParserToVar(&f.Vtx2.Y),
		32: core.NewFloatTypeParserToVar(&f.Vtx2.Z),
		13: core.NewFloatTypeParserToVar(&f.Vtx3.X),
		23: core.NewFloatTypeParserToVar(&f.Vtx3.Y),
		33: core.NewFloatTypeParserToVar(&f.Vtx3.Z),
		70: core.NewIntTypeParserToVar(&f.InvisibleEdges),
	})
	f.Parse(tags)
	f.XData = CollectXDataFromTags(tags)
	return f, nil
}

func (f *Face3D) DxfTags() core.TagSlice {
	baseTags := baseEntityTags(&f.BaseEntity, "3DFACE")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbFace")))
	}

	vtx3 := f.Vtx3
	if core.FloatEquals(vtx3.X, 0) && core.FloatEquals(vtx3.Y, 0) && core.FloatEquals(vtx3.Z, 0) {
		vtx3 = f.Vtx2
	}
	tags := append(baseTags,
		core.NewTag(10, core.NewFloatValue(f.Vtx0.X)),
		core.NewTag(20, core.NewFloatValue(f.Vtx0.Y)),
		core.NewTag(30, core.NewFloatValue(f.Vtx0.Z)),
		core.NewTag(11, core.NewFloatValue(f.Vtx1.X)),
		core.NewTag(21, core.NewFloatValue(f.Vtx1.Y)),
		core.NewTag(31, core.NewFloatValue(f.Vtx1.Z)),
		core.NewTag(12, core.NewFloatValue(f.Vtx2.X)),
		core.NewTag(22, core.NewFloatValue(f.Vtx2.Y)),
		core.NewTag(32, core.NewFloatValue(f.Vtx2.Z)),
		core.NewTag(13, core.NewFloatValue(vtx3.X)),
		core.NewTag(23, core.NewFloatValue(vtx3.Y)),
		core.NewTag(33, core.NewFloatValue(vtx3.Z)),
	)
	if f.InvisibleEdges != 0 {
		tags = append(tags, core.NewTag(70, core.NewIntegerValue(f.InvisibleEdges)))
	}
	return AppendXData(tags, &f.BaseEntity)
}

func NewFace3DEntity(points []core.Point, layer string) *Face3D {
	f := &Face3D{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
	}
	if len(points) > 0 {
		f.Vtx0 = points[0]
	}
	if len(points) > 1 {
		f.Vtx1 = points[1]
	}
	if len(points) > 2 {
		f.Vtx2 = points[2]
	}
	if len(points) > 3 {
		f.Vtx3 = points[3]
	} else {
		f.Vtx3 = f.Vtx2
	}
	return f
}

func (f Face3D) Clone() Entity {
	n := NewFace3DEntity([]core.Point{f.Vtx0, f.Vtx1, f.Vtx2, f.Vtx3}, f.LayerName)
	n.BaseEntity = f.BaseEntity.CloneBase()
	n.InvisibleEdges = f.InvisibleEdges
	return n
}
