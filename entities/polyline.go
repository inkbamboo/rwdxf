package entities

import (
	"github.com/inkbamboo/rwdxf/core"
)

const polyClosedBit = 0x1
const curveFitBit = 0x2
const splineFitBit = 0x4
const poly3DBit = 0x8
const mesh3DBit = 0x10
const meshClosedNBit = 0x20
const polyfaceBit = 0x40

// Polyline 表示 DXF POLYLINE 实体（多段线），可能包含嵌套的 VERTEX 和 SEQEND 子实体。
type Polyline struct {
	BaseEntity
	Thickness                 float64
	Closed                    bool
	CurveFitVerticesAdded     bool
	SplineFitVerticesAdded    bool
	Is3DPolyline              bool
	Is3DPolygonMesh           bool
	Is3DPolygonMeshClosed     bool
	IsPolyfaceMesh            bool
	ContinuousLinetypePattern bool
	Elevation                 float64
	ExtrusionDirection        core.Point
	Vertices                  EntitySlice
}

func (p *Polyline) IsSeqEnd() bool                  { return false }
func (p *Polyline) HasNestedEntities() bool         { return true }
func (p *Polyline) AddNestedEntities(e EntitySlice) { p.Vertices = e }

func (p Polyline) DxfType() core.DxfTypeName { return core.DxfTypePolyline }

func (p *Polyline) IsR12Compatible() bool { return true }

func (p Polyline) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Polyline); ok {
		return p.BaseEntity.Equals(o.BaseEntity) &&
			core.FloatEquals(p.Thickness, o.Thickness) &&
			p.Closed == o.Closed &&
			p.Is3DPolyline == o.Is3DPolyline &&
			p.ExtrusionDirection.Equals(o.ExtrusionDirection) &&
			p.Vertices.Equals(o.Vertices)
	}
	return false
}

func (p *Polyline) NumVertices() int { return len(p.Vertices) }

func (p *Polyline) Is2D() bool { return !p.Is3DPolyline }

func (p *Polyline) Is3DMesh() bool { return p.Is3DPolygonMesh }

func (p *Polyline) IsPolyface() bool { return p.IsPolyfaceMesh }

func (p *Polyline) IsCurveFit() bool { return p.CurveFitVerticesAdded }

func (p *Polyline) IsSplineFit() bool { return p.SplineFitVerticesAdded }

func (p *Polyline) GetPoints() []core.Point {
	var pts []core.Point
	for _, v := range p.Vertices {
		if vert, ok := v.(*Vertex); ok {
			pts = append(pts, vert.Location)
		}
	}
	return pts
}

// NewPolyline 从 TagSlice 解析并创建 Polyline 实体。
func NewPolyline(tags core.TagSlice) (*Polyline, error) {
	poly := new(Polyline)
	poly.ExtrusionDirection = core.Point{X: 0, Y: 0, Z: 1}
	poly.InitBaseEntityParser()
	poly.Update(map[int]core.TypeParser{
		39: core.NewFloatTypeParserToVar(&poly.Thickness),
		70: core.NewIntTypeParser(func(flags int) {
			poly.Closed = flags&polyClosedBit != 0
			poly.CurveFitVerticesAdded = flags&curveFitBit != 0
			poly.SplineFitVerticesAdded = flags&splineFitBit != 0
			poly.Is3DPolyline = flags&poly3DBit != 0
			poly.Is3DPolygonMesh = flags&mesh3DBit != 0
			poly.Is3DPolygonMeshClosed = flags&meshClosedNBit != 0
			poly.IsPolyfaceMesh = flags&polyfaceBit != 0
		}),
		30:  core.NewFloatTypeParserToVar(&poly.Elevation),
		210: core.NewFloatTypeParserToVar(&poly.ExtrusionDirection.X),
		220: core.NewFloatTypeParserToVar(&poly.ExtrusionDirection.Y),
		230: core.NewFloatTypeParserToVar(&poly.ExtrusionDirection.Z),
	})
	poly.Parse(tags)
	poly.XData = CollectXDataFromTags(tags)
	return poly, nil
}
func (p *Polyline) DxfTags() core.TagSlice {
	flags := 0
	if p.Closed {
		flags |= 1
	}
	baseTags := baseEntityTags(&p.BaseEntity, "POLYLINE")
	if !R12Mode {
		subclass := "AcDb2dPolyline"
		if p.Is3DPolyline {
			subclass = "AcDb3dPolyline"
		}
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue(subclass)))
	}
	tags := append(baseTags,
		core.NewTag(66, core.NewIntegerValue(1)),
		core.NewTag(70, core.NewIntegerValue(flags)),
	)
	if !core.FloatEquals(p.Elevation, 0) {
		tags = append(tags, core.NewTag(30, core.NewFloatValue(p.Elevation)))
	}
	if !core.FloatEquals(p.Thickness, 0) {
		tags = append(tags, core.NewTag(39, core.NewFloatValue(p.Thickness)))
	}
	if !isDefaultExtrusion(p.ExtrusionDirection) {
		tags = append(tags, pointToTags210(p.ExtrusionDirection)...)
	}

	for _, v := range p.Vertices {
		if vertex, ok := v.(*Vertex); ok {
			tags = append(tags, vertex.DxfTags()...)
		}
	}

	if R12Mode {
		seqLayer := p.LayerName
		if seqLayer == "" {
			seqLayer = "0"
		}
		tags = append(tags,
			core.NewTag(0, core.NewStringValue("SEQEND")),
			core.NewTag(8, core.NewStringValue(seqLayer)),
		)
	}
	return AppendXData(tags, &p.BaseEntity)
}

// NewPolylineEntity 直接创建一个 Polyline 实体。
func NewPolylineEntity(points []core.Point, closed bool, layer string) *Polyline {
	poly := &Polyline{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		Closed:             closed,
		Vertices:           make(EntitySlice, 0, len(points)),
		ExtrusionDirection: core.Point{X: 0, Y: 0, Z: 1},
	}
	for _, pt := range points {
		poly.Vertices = append(poly.Vertices, &Vertex{
			BaseEntity: BaseEntity{LayerName: layer},
			Location:   pt,
		})
	}
	return poly
}

func (p Polyline) Clone() Entity { pts:=make([]core.Point,len(p.Vertices)); for i,v:=range p.Vertices {if vv,ok:=v.(*Vertex);ok{pts[i]=vv.Location}}; n:=NewPolylineEntity(pts,p.Closed,p.LayerName); n.BaseEntity=p.BaseEntity.CloneBase(); n.Is3DPolyline=p.Is3DPolyline; n.ExtrusionDirection=p.ExtrusionDirection; return n }
