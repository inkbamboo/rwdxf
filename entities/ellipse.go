package entities

import (
	"math"

	"github.com/inkbamboo/rwdxf/core"
)

// Ellipse 表示 DXF ELLIPSE 实体。
type Ellipse struct {
	RegularEntity
	BaseEntity
	Center             core.Point
	MajorAxisEndPoint  core.Point
	AxisRatio          float64
	StartParameter     float64
	EndParameter       float64
	ExtrusionDirection core.Point
}

func (e Ellipse) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Ellipse); ok {
		return e.BaseEntity.Equals(o.BaseEntity) &&
			e.Center.Equals(o.Center) &&
			e.MajorAxisEndPoint.Equals(o.MajorAxisEndPoint) &&
			core.FloatEquals(e.AxisRatio, o.AxisRatio) &&
			core.FloatEquals(e.StartParameter, o.StartParameter) &&
			core.FloatEquals(e.EndParameter, o.EndParameter)
	}
	return false
}

func (e Ellipse) DxfType() core.DxfTypeName { return core.DxfTypeEllipse }

func (e Ellipse) IsR12Compatible() bool { return true }

func (e *Ellipse) MajorAxis() core.Point { return e.MajorAxisEndPoint }

func (e *Ellipse) MajorRadius() float64 {
	p := e.MajorAxisEndPoint
	return math.Sqrt(p.X*p.X + p.Y*p.Y + p.Z*p.Z)
}

func (e *Ellipse) MinorRadius() float64 {
	return e.MajorRadius() * e.AxisRatio
}

func (e *Ellipse) StartPoint() core.Point { return e.ParamPoint(e.StartParameter) }

func (e *Ellipse) EndPoint() core.Point { return e.ParamPoint(e.EndParameter) }

func (e *Ellipse) ParamPoint(t float64) core.Point {
	rx := e.MajorRadius()
	ry := e.MinorRadius()

	return core.Point{
		X: e.Center.X + rx*math.Cos(t),
		Y: e.Center.Y + ry*math.Sin(t),
		Z: e.Center.Z,
	}
}

func (e *Ellipse) IsFullEllipse() bool {
	return core.FloatEquals(e.StartParameter, 0) &&
		core.FloatEquals(e.EndParameter, 2*math.Pi)
}

func (e *Ellipse) Flattening(segments int) []core.Point {
	if segments <= 0 {
		segments = 64
	}
	startParam, endParam := e.StartParameter, e.EndParameter
	if endParam < startParam {
		endParam += 2 * math.Pi
	}
	pts := make([]core.Point, segments+1)
	for i := 0; i <= segments; i++ {
		t := startParam + (endParam-startParam)*float64(i)/float64(segments)
		pts[i] = e.ParamPoint(t)
	}
	return pts
}

func (e *Ellipse) Translate(dx, dy, dz float64) {
	e.Center.X += dx
	e.Center.Y += dy
	e.Center.Z += dz
}

func NewEllipse(tags core.TagSlice) (*Ellipse, error) {
	el := new(Ellipse)
	el.ExtrusionDirection = core.Point{X: 0, Y: 0, Z: 1}
	el.AxisRatio = 1.0
	el.InitBaseEntityParser()
	el.Update(map[int]core.TypeParser{
		10:  core.NewFloatTypeParserToVar(&el.Center.X),
		20:  core.NewFloatTypeParserToVar(&el.Center.Y),
		30:  core.NewFloatTypeParserToVar(&el.Center.Z),
		11:  core.NewFloatTypeParserToVar(&el.MajorAxisEndPoint.X),
		21:  core.NewFloatTypeParserToVar(&el.MajorAxisEndPoint.Y),
		31:  core.NewFloatTypeParserToVar(&el.MajorAxisEndPoint.Z),
		40:  core.NewFloatTypeParserToVar(&el.AxisRatio),
		41:  core.NewFloatTypeParserToVar(&el.StartParameter),
		42:  core.NewFloatTypeParserToVar(&el.EndParameter),
		210: core.NewFloatTypeParserToVar(&el.ExtrusionDirection.X),
		220: core.NewFloatTypeParserToVar(&el.ExtrusionDirection.Y),
		230: core.NewFloatTypeParserToVar(&el.ExtrusionDirection.Z),
	})
	el.Parse(tags)
	el.XData = CollectXDataFromTags(tags)
	return el, nil
}
func (e *Ellipse) DxfTags() core.TagSlice {
	if R12Mode {
		return e.dxfTagsR12()
	}
	baseTags := baseEntityTags(&e.BaseEntity, "ELLIPSE")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbEllipse")))
	}
	tags := append(baseTags,
		core.NewTag(10, core.NewFloatValue(e.Center.X)),
		core.NewTag(20, core.NewFloatValue(e.Center.Y)),
		core.NewTag(30, core.NewFloatValue(e.Center.Z)),
		core.NewTag(11, core.NewFloatValue(e.MajorAxisEndPoint.X)),
		core.NewTag(21, core.NewFloatValue(e.MajorAxisEndPoint.Y)),
		core.NewTag(31, core.NewFloatValue(e.MajorAxisEndPoint.Z)),
	)
	if !core.FloatEquals(e.AxisRatio, 1.0) {
		tags = append(tags, core.NewTag(40, core.NewFloatValue(e.AxisRatio)))
	}

	tags = append(tags, core.NewTag(41, core.NewFloatValue(e.StartParameter)))
	if !core.FloatEquals(e.EndParameter, 0) {
		tags = append(tags, core.NewTag(42, core.NewFloatValue(e.EndParameter)))
	}

	extr := e.ExtrusionDirection
	if core.FloatEquals(extr.X, 0) && core.FloatEquals(extr.Y, 0) && core.FloatEquals(extr.Z, 0) {
		extr = core.Point{X: 0, Y: 0, Z: 1}
	}
	if !isDefaultExtrusion(extr) {
		tags = append(tags, pointToTags210(extr)...)
	}
	return AppendXData(tags, &e.BaseEntity)
}

func (e *Ellipse) dxfTagsR12() core.TagSlice {
	layerName := e.LayerName
	if layerName == "" {
		layerName = "0"
	}
	var tags core.TagSlice

	closed := core.FloatEquals(e.EndParameter-e.StartParameter, 2*math.Pi)
	polyFlag := 0
	if closed {
		polyFlag = 1
	}

	tags = append(tags,
		core.NewTag(0, core.NewStringValue("POLYLINE")),
		core.NewTag(8, core.NewStringValue(layerName)),
		core.NewTag(66, core.NewIntegerValue(1)),
		core.NewTag(10, core.NewFloatValue(0.0)),
		core.NewTag(20, core.NewFloatValue(0.0)),
		core.NewTag(30, core.NewFloatValue(0.0)),
		core.NewTag(70, core.NewIntegerValue(polyFlag)),
	)

	numPts := 72
	if closed {
		numPts = 72
	}
	if !closed && numPts < 2 {
		numPts = 2
	}
	step := (e.EndParameter - e.StartParameter) / float64(numPts-1)
	for i := 0; i < numPts; i++ {
		t := e.StartParameter + step*float64(i)
		pt := e.ParamPoint(t)
		tags = append(tags,
			core.NewTag(0, core.NewStringValue("VERTEX")),
			core.NewTag(8, core.NewStringValue(layerName)),
			core.NewTag(10, core.NewFloatValue(pt.X)),
			core.NewTag(20, core.NewFloatValue(pt.Y)),
			core.NewTag(30, core.NewFloatValue(0.0)),
			core.NewTag(70, core.NewIntegerValue(0)),
		)
	}

	if closed && numPts > 0 {
		pt0 := e.ParamPoint(e.StartParameter)
		ptLast := e.ParamPoint(e.EndParameter)
		if !core.FloatEquals(pt0.X, ptLast.X) || !core.FloatEquals(pt0.Y, ptLast.Y) {
			tags = append(tags,
				core.NewTag(0, core.NewStringValue("VERTEX")),
				core.NewTag(8, core.NewStringValue(layerName)),
				core.NewTag(10, core.NewFloatValue(pt0.X)),
				core.NewTag(20, core.NewFloatValue(pt0.Y)),
				core.NewTag(30, core.NewFloatValue(0.0)),
				core.NewTag(70, core.NewIntegerValue(0)),
			)
		}
	}

	tags = append(tags,
		core.NewTag(0, core.NewStringValue("SEQEND")),
		core.NewTag(8, core.NewStringValue(layerName)),
	)
	return AppendXData(tags, &e.BaseEntity)
}

func NewEllipseEntity(center, majorAxis core.Point, axisRatio, startParam, endParam float64, layer string) *Ellipse {
	return &Ellipse{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		Center:            center,
		MajorAxisEndPoint: majorAxis,
		AxisRatio:         axisRatio,
		StartParameter:    startParam,
		EndParameter:      endParam,
	}
}

func (e Ellipse) Clone() Entity {
	n := NewEllipseEntity(e.Center, e.MajorAxisEndPoint, e.AxisRatio, e.StartParameter, e.EndParameter, e.LayerName)
	n.BaseEntity = e.BaseEntity.CloneBase()
	n.ExtrusionDirection = e.ExtrusionDirection
	return n
}
