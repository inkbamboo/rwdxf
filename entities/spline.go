package entities

import (
	"github.com/inkbamboo/rwdxf/core"
)

// Spline 表示 DXF SPLINE 实体（样条曲线）。
type Spline struct {
	RegularEntity
	BaseEntity
	NormalVector          core.Point
	ExtrusionDirection    core.Point
	Closed                bool
	Periodic              bool
	Rational              bool
	Planar                bool
	Linear                bool
	Degree                int
	NumberOfKnots         int
	NumberOfControlPoints int
	NumberOfFitPoints     int
	KnotValues            []float64
	ControlPoints         core.PointSlice
	FitPoints             core.PointSlice
	StartTangent          core.Point
	EndTangent            core.Point
	KnotTolerance         float64
	ControlPointTolerance float64
	FitTolerance          float64
}

func (s Spline) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Spline); ok {
		return s.BaseEntity.Equals(o.BaseEntity) &&
			s.Closed == o.Closed &&
			s.Periodic == o.Periodic &&
			s.Rational == o.Rational &&
			s.Degree == o.Degree &&
			s.ControlPoints.Equals(o.ControlPoints)
	}
	return false
}

func (s Spline) DxfType() core.DxfTypeName { return core.DxfTypeSpline }

func (s Spline) IsR12Compatible() bool { return true }

func (s *Spline) IsClosed() bool { return s.Closed }

func (s *Spline) IsPeriodic() bool { return s.Periodic }

func (s *Spline) Order() int { return s.Degree + 1 }

func (s *Spline) NumControlPoints() int { return len(s.ControlPoints) }

func (s *Spline) NumFitPoints() int { return len(s.FitPoints) }

func (s *Spline) NumKnots() int { return len(s.KnotValues) }

func (s *Spline) Flattening(subSteps int) []core.Point {
	if subSteps <= 0 {
		subSteps = 20
	}

	pts := core.PointSlice(s.FitPoints)
	if len(pts) == 0 {
		pts = s.ControlPoints
	}
	if len(pts) < 2 {
		return core.PointSlice(pts)
	}

	var result core.PointSlice
	result = append(result, pts[0])

	for i := 1; i < len(pts); i++ {
		p0 := pts[maxInt(i-2, 0)]
		p1 := pts[i-1]
		p2 := pts[i]
		p3 := pts[minInt(i+1, len(pts)-1)]

		for j := 1; j <= subSteps; j++ {
			t := float64(j) / float64(subSteps)
			result = append(result, catmullRom(p0, p1, p2, p3, t))
		}
	}

	return result
}

func NewSpline(tags core.TagSlice) (*Spline, error) {
	spline := new(Spline)
	spline.InitBaseEntityParser()
	spline.Update(map[int]core.TypeParser{
		210: core.NewFloatTypeParserToVar(&spline.NormalVector.X),
		220: core.NewFloatTypeParserToVar(&spline.NormalVector.Y),
		230: core.NewFloatTypeParserToVar(&spline.NormalVector.Z),
		70: core.NewIntTypeParser(func(flags int) {
			spline.Closed = flags&1 != 0
			spline.Periodic = flags&2 != 0
			spline.Rational = flags&4 != 0
			spline.Planar = flags&8 != 0
			spline.Linear = flags&16 != 0
		}),
		71: core.NewIntTypeParserToVar(&spline.Degree),
		72: core.NewIntTypeParserToVar(&spline.NumberOfKnots),
		73: core.NewIntTypeParserToVar(&spline.NumberOfControlPoints),
		74: core.NewIntTypeParserToVar(&spline.NumberOfFitPoints),
	})
	spline.Parse(tags)
	spline.XData = CollectXDataFromTags(tags)

	spline.KnotValues = make([]float64, 0)
	spline.ControlPoints = make(core.PointSlice, 0)
	spline.FitPoints = make(core.PointSlice, 0)
	cpIdx := 0
	fpIdx := 0
	for _, tag := range tags.RegularTags() {
		switch tag.Code {
		case 40:
			if v, ok := core.AsFloat(tag.Value); ok {
				spline.KnotValues = append(spline.KnotValues, v)
			}
		case 10:
			if v, ok := core.AsFloat(tag.Value); ok {
				if cpIdx >= len(spline.ControlPoints) {
					spline.ControlPoints = append(spline.ControlPoints, core.Point{})
				}
				spline.ControlPoints[cpIdx].X = v
			}
		case 20:
			if cpIdx < len(spline.ControlPoints) {
				if v, ok := core.AsFloat(tag.Value); ok {
					spline.ControlPoints[cpIdx].Y = v
				}
			}
		case 30:
			if cpIdx < len(spline.ControlPoints) {
				if v, ok := core.AsFloat(tag.Value); ok {
					spline.ControlPoints[cpIdx].Z = v
					cpIdx++
				}
			}
		case 11:
			if v, ok := core.AsFloat(tag.Value); ok {
				if fpIdx >= len(spline.FitPoints) {
					spline.FitPoints = append(spline.FitPoints, core.Point{})
				}
				spline.FitPoints[fpIdx].X = v
			}
		case 21:
			if fpIdx < len(spline.FitPoints) {
				if v, ok := core.AsFloat(tag.Value); ok {
					spline.FitPoints[fpIdx].Y = v
				}
			}
		case 31:
			if fpIdx < len(spline.FitPoints) {
				if v, ok := core.AsFloat(tag.Value); ok {
					spline.FitPoints[fpIdx].Z = v
					fpIdx++
				}
			}
		}
	}
	return spline, nil
}
func (s *Spline) DxfTags() core.TagSlice {
	if R12Mode {
		return s.dxfTagsR12()
	}
	flags := 0
	if s.Closed {
		flags |= 1
	}
	if s.Periodic {
		flags |= 2
	}
	if s.Rational {
		flags |= 4
	}
	if s.Planar {
		flags |= 8
	}
	if s.Linear {
		flags |= 16
	}
	baseTags := baseEntityTags(&s.BaseEntity, "SPLINE")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbSpline")))
	}
	tags := append(baseTags,
		core.NewTag(70, core.NewIntegerValue(flags)),
		core.NewTag(71, core.NewIntegerValue(s.Degree)),
		core.NewTag(72, core.NewIntegerValue(len(s.KnotValues))),
		core.NewTag(73, core.NewIntegerValue(len(s.ControlPoints))),
		core.NewTag(74, core.NewIntegerValue(len(s.FitPoints))),
	)

	if !R12Mode {
		tags = append(tags,
			core.NewTag(42, core.NewFloatValue(s.KnotTolerance)),
			core.NewTag(43, core.NewFloatValue(s.ControlPointTolerance)),
		)
		if len(s.FitPoints) > 0 {
			tags = append(tags, core.NewTag(44, core.NewFloatValue(s.FitTolerance)))
		}
	}
	for _, k := range s.KnotValues {
		tags = append(tags, core.NewTag(40, core.NewFloatValue(k)))
	}
	for _, pt := range s.ControlPoints {
		tags = append(tags,
			core.NewTag(10, core.NewFloatValue(pt.X)),
			core.NewTag(20, core.NewFloatValue(pt.Y)),
			core.NewTag(30, core.NewFloatValue(pt.Z)),
		)
	}
	for _, pt := range s.FitPoints {
		tags = append(tags,
			core.NewTag(11, core.NewFloatValue(pt.X)),
			core.NewTag(21, core.NewFloatValue(pt.Y)),
			core.NewTag(31, core.NewFloatValue(pt.Z)),
		)
	}

	extr := s.ExtrusionDirection
	if core.FloatEquals(extr.X, 0) && core.FloatEquals(extr.Y, 0) && core.FloatEquals(extr.Z, 0) {
		extr = core.Point{X: 0, Y: 0, Z: 1}
	}
	tags = append(tags, pointToTags210(extr)...)
	return AppendXData(tags, &s.BaseEntity)
}

func (s *Spline) dxfTagsR12() core.TagSlice {
	layerName := s.LayerName
	if layerName == "" {
		layerName = "0"
	}

	smoothPts := s.Flattening(20)
	if len(smoothPts) < 2 {
		return core.TagSlice{
			core.NewTag(0, core.NewStringValue("SPLINE")),
			core.NewTag(8, core.NewStringValue(layerName)),
		}
	}

	var tags core.TagSlice
	tags = append(tags,
		core.NewTag(0, core.NewStringValue("POLYLINE")),
		core.NewTag(8, core.NewStringValue(layerName)),
		core.NewTag(66, core.NewIntegerValue(1)),
		core.NewTag(10, core.NewFloatValue(0.0)),
		core.NewTag(20, core.NewFloatValue(0.0)),
		core.NewTag(30, core.NewFloatValue(0.0)),
		core.NewTag(70, core.NewIntegerValue(0)),
	)

	for _, pt := range smoothPts {
		tags = append(tags,
			core.NewTag(0, core.NewStringValue("VERTEX")),
			core.NewTag(8, core.NewStringValue(layerName)),
			core.NewTag(10, core.NewFloatValue(pt.X)),
			core.NewTag(20, core.NewFloatValue(pt.Y)),
			core.NewTag(30, core.NewFloatValue(0.0)),
			core.NewTag(70, core.NewIntegerValue(0)),
		)
	}

	tags = append(tags,
		core.NewTag(0, core.NewStringValue("SEQEND")),
		core.NewTag(8, core.NewStringValue(layerName)),
	)
	return AppendXData(tags, &s.BaseEntity)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func catmullRom(p0, p1, p2, p3 core.Point, t float64) core.Point {
	t2 := t * t
	t3 := t2 * t
	return core.Point{
		X: 0.5 * ((2*p1.X) + (-p0.X+p2.X)*t + (2*p0.X-5*p1.X+4*p2.X-p3.X)*t2 + (-p0.X+3*p1.X-3*p2.X+p3.X)*t3),
		Y: 0.5 * ((2*p1.Y) + (-p0.Y+p2.Y)*t + (2*p0.Y-5*p1.Y+4*p2.Y-p3.Y)*t2 + (-p0.Y+3*p1.Y-3*p2.Y+p3.Y)*t3),
		Z: 0.5 * ((2*p1.Z) + (-p0.Z+p2.Z)*t + (2*p0.Z-5*p1.Z+4*p2.Z-p3.Z)*t2 + (-p0.Z+3*p1.Z-3*p2.Z+p3.Z)*t3),
	}
}

func NewSplineEntity(degree int, controlPoints core.PointSlice, knotValues []float64, layer string) *Spline {
	return &Spline{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		Degree:                degree,
		NumberOfKnots:         len(knotValues),
		NumberOfControlPoints: len(controlPoints),
		KnotValues:            knotValues,
		ControlPoints:         controlPoints,
		ExtrusionDirection:    core.Point{X: 0, Y: 0, Z: 1},
		KnotTolerance:         0.000000001,
		ControlPointTolerance: 0.0000000001,
		FitTolerance:          0.0000000001,
	}
}

func (s Spline) Clone() Entity {
	cp := make(core.PointSlice, len(s.ControlPoints))
	copy(cp, s.ControlPoints)
	kv := make([]float64, len(s.KnotValues))
	copy(kv, s.KnotValues)
	fp := make(core.PointSlice, len(s.FitPoints))
	copy(fp, s.FitPoints)
	n := NewSplineEntity(s.Degree, cp, kv, s.LayerName)
	n.BaseEntity = s.BaseEntity.CloneBase()
	n.Closed = s.Closed
	n.Periodic = s.Periodic
	n.Rational = s.Rational
	n.Planar = s.Planar
	n.Linear = s.Linear
	n.FitPoints = fp
	n.StartTangent = s.StartTangent
	n.EndTangent = s.EndTangent
	n.KnotTolerance = s.KnotTolerance
	n.ControlPointTolerance = s.ControlPointTolerance
	n.FitTolerance = s.FitTolerance
	n.ExtrusionDirection = s.ExtrusionDirection
	n.NormalVector = s.NormalVector
	return n
}
