package entities

import (
	"github.com/inkbamboo/rwdxf/core"
)

// Vertex 表示 DXF VERTEX 实体（POLYLINE 的顶点）。
type Vertex struct {
	RegularEntity
	BaseEntity
	Location            core.Point
	StartingWidth       float64
	EndWidth            float64
	Bulge               float64
	Flags               int
	CurveFitTangentDir  float64
	PolyfaceMeshVertex0 int
	PolyfaceMeshVertex1 int
	PolyfaceMeshVertex2 int
	PolyfaceMeshVertex3 int
}

func (v Vertex) DxfType() core.DxfTypeName { return core.DxfTypeVertex }

func (v Vertex) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Vertex); ok {
		return v.BaseEntity.Equals(o.BaseEntity) &&
			v.Location.Equals(o.Location) &&
			core.FloatEquals(v.Bulge, o.Bulge) &&
			v.Flags == o.Flags
	}
	return false
}

// NewVertex 从 TagSlice 解析并创建 Vertex 实体。
func NewVertex(tags core.TagSlice) (*Vertex, error) {
	v := new(Vertex)
	v.InitBaseEntityParser()
	v.Update(map[int]core.TypeParser{
		10: core.NewFloatTypeParserToVar(&v.Location.X),
		20: core.NewFloatTypeParserToVar(&v.Location.Y),
		30: core.NewFloatTypeParserToVar(&v.Location.Z),
		40: core.NewFloatTypeParserToVar(&v.StartingWidth),
		41: core.NewFloatTypeParserToVar(&v.EndWidth),
		42: core.NewFloatTypeParserToVar(&v.Bulge),
		70: core.NewIntTypeParserToVar(&v.Flags),
		50: core.NewFloatTypeParserToVar(&v.CurveFitTangentDir),
		71: core.NewIntTypeParserToVar(&v.PolyfaceMeshVertex0),
		72: core.NewIntTypeParserToVar(&v.PolyfaceMeshVertex1),
		73: core.NewIntTypeParserToVar(&v.PolyfaceMeshVertex2),
		74: core.NewIntTypeParserToVar(&v.PolyfaceMeshVertex3),
	})
	v.Parse(tags)
	v.XData = CollectXDataFromTags(tags)
	return v, nil
}
func (v *Vertex) DxfTags() core.TagSlice {
	tags := baseEntityTags(&v.BaseEntity, "VERTEX")
	if !R12Mode {
		tags = append(tags,
			core.NewTag(100, core.NewStringValue("AcDbVertex")),
			core.NewTag(100, core.NewStringValue("AcDb2dVertex")),
		)
	}
	tags = append(tags,
		core.NewTag(10, core.NewFloatValue(v.Location.X)),
		core.NewTag(20, core.NewFloatValue(v.Location.Y)),
		core.NewTag(30, core.NewFloatValue(v.Location.Z)),
		core.NewTag(42, core.NewFloatValue(v.Bulge)),
		core.NewTag(70, core.NewIntegerValue(v.Flags)),
	)
	return AppendXData(tags, &v.BaseEntity)
}

func (v Vertex) Clone() Entity {
	n := &Vertex{
		Location: v.Location, Bulge: v.Bulge, Flags: v.Flags,
		StartingWidth: v.StartingWidth, EndWidth: v.EndWidth,
		CurveFitTangentDir: v.CurveFitTangentDir,
		PolyfaceMeshVertex0: v.PolyfaceMeshVertex0,
		PolyfaceMeshVertex1: v.PolyfaceMeshVertex1,
		PolyfaceMeshVertex2: v.PolyfaceMeshVertex2,
		PolyfaceMeshVertex3: v.PolyfaceMeshVertex3,
	}
	n.BaseEntity = v.BaseEntity.CloneBase()
	return n
}
