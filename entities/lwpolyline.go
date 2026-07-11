package entities

import (
	"github.com/inkbamboo/rwdxf/core"
)

const closedBit = 0x1
const plinegenBit = 0x80

// LWPolyline 表示 DXF LWPOLYLINE 实体（轻量多段线）。
type LWPolyline struct {
	RegularEntity
	BaseEntity
	Closed             bool
	Plinegen           bool
	ConstantWidth      float64
	Elevation          float64
	Thickness          float64
	Points             LWPolyLinePointSlice
	ExtrusionDirection core.Point
}

func (p LWPolyline) Equals(other core.DxfElement) bool {
	if o, ok := other.(*LWPolyline); ok {
		return p.BaseEntity.Equals(o.BaseEntity) &&
			p.Closed == o.Closed &&
			p.Plinegen == o.Plinegen &&
			core.FloatEquals(p.ConstantWidth, o.ConstantWidth) &&
			core.FloatEquals(p.Elevation, o.Elevation) &&
			core.FloatEquals(p.Thickness, o.Thickness) &&
			p.Points.Equals(o.Points) &&
			p.ExtrusionDirection.Equals(o.ExtrusionDirection)
	}
	return false
}

func (p LWPolyline) DxfType() core.DxfTypeName { return core.DxfTypeLWPolyline }

func (p LWPolyline) IsR12Compatible() bool { return true }

func (p *LWPolyline) NumPoints() int { return len(p.Points) }

func (p *LWPolyline) HasBulge() bool {
	for _, pt := range p.Points {
		if !core.FloatEquals(pt.Bulge, 0) {
			return true
		}
	}
	return false
}

func (p *LWPolyline) GetPoints() []core.Point {
	pts := make([]core.Point, len(p.Points))
	for i, v := range p.Points {
		pts[i] = v.Point
	}
	return pts
}

func (p *LWPolyline) HasWidth() bool {
	for _, pt := range p.Points {
		if !core.FloatEquals(pt.StartingWidth, 0) || !core.FloatEquals(pt.EndWidth, 0) {
			return true
		}
	}
	return false
}

// LWPolyLinePoint 表示轻量多段线的一个顶点。
type LWPolyLinePoint struct {
	Point         core.Point
	ID            int
	StartingWidth float64
	EndWidth      float64
	Bulge         float64
}

func (p LWPolyLinePoint) Equals(other LWPolyLinePoint) bool {
	return p.Point.Equals(other.Point) &&
		p.ID == other.ID &&
		core.FloatEquals(p.StartingWidth, other.StartingWidth) &&
		core.FloatEquals(p.EndWidth, other.EndWidth) &&
		core.FloatEquals(p.Bulge, other.Bulge)
}

// LWPolyLinePointSlice 是 LWPolyLinePoint 的切片。
type LWPolyLinePointSlice []LWPolyLinePoint

func (p LWPolyLinePointSlice) Equals(other LWPolyLinePointSlice) bool {
	if len(p) != len(other) {
		return false
	}
	for i, pt := range p {
		if !pt.Equals(other[i]) {
			return false
		}
	}
	return true
}

// NewLWPolyline 从 TagSlice 解析并创建 LWPolyline 实体。
func NewLWPolyline(tags core.TagSlice) (*LWPolyline, error) {
	poly := new(LWPolyline)
	poly.ExtrusionDirection = core.Point{X: 0, Y: 0, Z: 1}
	poly.InitBaseEntityParser()
	pointIndex := -1
	poly.Update(map[int]core.TypeParser{
		70: core.NewIntTypeParser(func(flags int) {
			poly.Closed = flags&closedBit != 0
			poly.Plinegen = flags&plinegenBit != 0
		}),
		90: core.NewIntTypeParser(func(value int) {
			poly.Points = make(LWPolyLinePointSlice, value)
		}),
		38: core.NewFloatTypeParserToVar(&poly.Elevation),
		39: core.NewFloatTypeParserToVar(&poly.Thickness),
		43: core.NewFloatTypeParserToVar(&poly.ConstantWidth),
		10: core.NewFloatTypeParser(func(x float64) {
			pointIndex++
			poly.Points[pointIndex].Point.X = x
		}),
		20: core.NewFloatTypeParser(func(y float64) {
			poly.Points[pointIndex].Point.Y = y
		}),
		91: core.NewIntTypeParser(func(value int) {
			poly.Points[pointIndex].ID = value
		}),
		40: core.NewFloatTypeParser(func(value float64) {
			poly.Points[pointIndex].StartingWidth = value
		}),
		41: core.NewFloatTypeParser(func(value float64) {
			poly.Points[pointIndex].EndWidth = value
		}),
		42: core.NewFloatTypeParser(func(value float64) {
			poly.Points[pointIndex].Bulge = value
		}),
		210: core.NewFloatTypeParserToVar(&poly.ExtrusionDirection.X),
		220: core.NewFloatTypeParserToVar(&poly.ExtrusionDirection.Y),
		230: core.NewFloatTypeParserToVar(&poly.ExtrusionDirection.Z),
	})
	poly.Parse(tags)
	poly.XData = CollectXDataFromTags(tags)
	return poly, nil
}
func (p *LWPolyline) DxfTags() core.TagSlice {
	if R12Mode {
		return p.dxfTagsR12()
	}
	flags := 0
	if p.Closed {
		flags |= 1
	}
	if p.Plinegen {
		flags |= 128
	}
	baseTags := baseEntityTags(&p.BaseEntity, "LWPOLYLINE")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbPolyline")))
	}
	tags := append(baseTags,
		core.NewTag(90, core.NewIntegerValue(len(p.Points))),
		core.NewTag(70, core.NewIntegerValue(flags)),
	)
	if !core.FloatEquals(p.Elevation, 0) {
		tags = append(tags, core.NewTag(38, core.NewFloatValue(p.Elevation)))
	}
	if !core.FloatEquals(p.Thickness, 0) {
		tags = append(tags, core.NewTag(39, core.NewFloatValue(p.Thickness)))
	}
	if !core.FloatEquals(p.ConstantWidth, 0) {
		tags = append(tags, core.NewTag(43, core.NewFloatValue(p.ConstantWidth)))
	}
	for _, pt := range p.Points {
		tags = append(tags,
			core.NewTag(10, core.NewFloatValue(pt.Point.X)),
			core.NewTag(20, core.NewFloatValue(pt.Point.Y)),
		)
		if !core.FloatEquals(pt.Bulge, 0) {
			tags = append(tags, core.NewTag(42, core.NewFloatValue(pt.Bulge)))
		}
		if !core.FloatEquals(pt.StartingWidth, 0) {
			tags = append(tags, core.NewTag(40, core.NewFloatValue(pt.StartingWidth)))
		}
		if !core.FloatEquals(pt.EndWidth, 0) {
			tags = append(tags, core.NewTag(41, core.NewFloatValue(pt.EndWidth)))
		}
	}
	if !isDefaultExtrusion(p.ExtrusionDirection) {
		tags = append(tags, pointToTags210(p.ExtrusionDirection)...)
	}
	return AppendXData(tags, &p.BaseEntity)
}

func (p *LWPolyline) dxfTagsR12() core.TagSlice {
	layerName := p.LayerName
	if layerName == "" {
		layerName = "0"
	}

	polyFlag := 0
	if p.Closed {
		polyFlag = 1
	}

	var tags core.TagSlice
	tags = append(tags,
		core.NewTag(0, core.NewStringValue("POLYLINE")),
		core.NewTag(8, core.NewStringValue(layerName)),
		core.NewTag(66, core.NewIntegerValue(1)),
		core.NewTag(10, core.NewFloatValue(0.0)),
		core.NewTag(20, core.NewFloatValue(0.0)),
		core.NewTag(30, core.NewFloatValue(0.0)),
		core.NewTag(70, core.NewIntegerValue(polyFlag)),
	)

	if !core.FloatEquals(p.Elevation, 0) {
		tags = append(tags, core.NewTag(30, core.NewFloatValue(p.Elevation)))
	}
	if !core.FloatEquals(p.Thickness, 0) {
		tags = append(tags, core.NewTag(39, core.NewFloatValue(p.Thickness)))
	}

	for _, pt := range p.Points {
		vtxFlags := 0
		if !core.FloatEquals(pt.Bulge, 0) {
			vtxFlags = 0
		}
		tags = append(tags,
			core.NewTag(0, core.NewStringValue("VERTEX")),
			core.NewTag(8, core.NewStringValue(layerName)),
			core.NewTag(10, core.NewFloatValue(pt.Point.X)),
			core.NewTag(20, core.NewFloatValue(pt.Point.Y)),
			core.NewTag(30, core.NewFloatValue(0.0)),
			core.NewTag(70, core.NewIntegerValue(vtxFlags)),
		)
		if !core.FloatEquals(pt.Bulge, 0) {
			tags = append(tags, core.NewTag(42, core.NewFloatValue(pt.Bulge)))
		}
		if !core.FloatEquals(pt.StartingWidth, 0) {
			tags = append(tags, core.NewTag(40, core.NewFloatValue(pt.StartingWidth)))
		}
		if !core.FloatEquals(pt.EndWidth, 0) {
			tags = append(tags, core.NewTag(41, core.NewFloatValue(pt.EndWidth)))
		}
	}

	tags = append(tags,
		core.NewTag(0, core.NewStringValue("SEQEND")),
		core.NewTag(8, core.NewStringValue(layerName)),
	)

	return tags
}

// NewLWPolylineEntity 直接创建一个 LWPolyline 实体。
func NewLWPolylineEntity(points []core.Point, closed bool, layer string) *LWPolyline {
	lwPoints := make(LWPolyLinePointSlice, len(points))
	for i, pt := range points {
		lwPoints[i] = LWPolyLinePoint{Point: pt}
	}
	return &LWPolyline{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		Closed:             closed,
		Points:             lwPoints,
		ExtrusionDirection: core.Point{X: 0, Y: 0, Z: 1},
	}
}

func (p LWPolyline) Clone() Entity {
	pts := make([]core.Point, len(p.Points))
	for i, pt := range p.Points {
		pts[i] = pt.Point
	}
	n := NewLWPolylineEntity(pts, p.Closed, p.LayerName)
	n.BaseEntity = p.BaseEntity.CloneBase()
	n.Plinegen = p.Plinegen
	n.ConstantWidth = p.ConstantWidth
	n.Elevation = p.Elevation
	n.Thickness = p.Thickness
	n.ExtrusionDirection = p.ExtrusionDirection
	return n
}
