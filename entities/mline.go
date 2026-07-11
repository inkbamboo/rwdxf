package entities

import (
	"math"
	"strconv"

	"github.com/inkbamboo/rwdxf/core"
)

// MLine 表示 DXF MLINE 实体（多线）。
type MLine struct {
	RegularEntity
	BaseEntity
	StyleName          string
	StyleHandle        string
	Scale              float64
	Justification      int
	Flags              int
	NumberOfVertices   int
	NumberOfElements   int
	StartPoint         core.Point
	ExtrusionDirection core.Point
	Vertices           core.PointSlice

	VertexParams []float64
	AreaFillParams []float64
}

func (m MLine) DxfType() core.DxfTypeName { return core.DxfTypeMLine }

func (m MLine) IsR12Compatible() bool { return true }

func (m MLine) Equals(other core.DxfElement) bool {
	if o, ok := other.(*MLine); ok {
		return m.BaseEntity.Equals(o.BaseEntity) &&
			m.StyleName == o.StyleName &&
			core.FloatEquals(m.Scale, o.Scale) &&
			m.Justification == o.Justification &&
			m.Vertices.Equals(o.Vertices) &&
			m.ExtrusionDirection.Equals(o.ExtrusionDirection)
	}
	return false
}

func NewMLine(tags core.TagSlice) (*MLine, error) {
	ml := new(MLine)
	ml.Scale = 1.0
	ml.ExtrusionDirection = core.Point{X: 0, Y: 0, Z: 1}
	ml.InitBaseEntityParser()

	ml.Update(map[int]core.TypeParser{
		2:   core.NewStringTypeParserToVar(&ml.StyleName),
		340: core.NewStringTypeParserToVar(&ml.StyleHandle),
		40:  core.NewFloatTypeParserToVar(&ml.Scale),
		70:  core.NewIntTypeParserToVar(&ml.Justification),
		71:  core.NewIntTypeParserToVar(&ml.Flags),
		72:  core.NewIntTypeParserToVar(&ml.NumberOfVertices),
		73:  core.NewIntTypeParserToVar(&ml.NumberOfElements),
		210: core.NewFloatTypeParserToVar(&ml.ExtrusionDirection.X),
		220: core.NewFloatTypeParserToVar(&ml.ExtrusionDirection.Y),
		230: core.NewFloatTypeParserToVar(&ml.ExtrusionDirection.Z),
	})

	ml.Parse(tags)
	ml.XData = CollectXDataFromTags(tags)
	ml.parseMLineVertices(tags)
	return ml, nil
}

func (m *MLine) parseMLineVertices(tags core.TagSlice) {
	var current core.Point
	pts := make(core.PointSlice, 0)
	firstPoint := true

	for _, tag := range tags.RegularTags() {
		switch tag.Code {
		case 10:
			if v, ok := core.AsFloat(tag.Value); ok {
				if firstPoint {
					m.StartPoint = core.Point{X: v}
				} else {
					current = core.Point{X: v}
				}
			}
		case 20:
			if v, ok := core.AsFloat(tag.Value); ok {
				if firstPoint {
					m.StartPoint.Y = v
					m.StartPoint.Z = 0
					firstPoint = false
					pts = append(pts, m.StartPoint)
				} else {
					current.Y = v
				}
			}
		case 30:
			if v, ok := core.AsFloat(tag.Value); ok {
				if !firstPoint {
					current.Z = v
				}
			}
		case 11:
			if v, ok := core.AsFloat(tag.Value); ok {
				current = core.Point{X: v}
			}
		case 21:
			if v, ok := core.AsFloat(tag.Value); ok {
				current.Y = v
				current.Z = 0
				pts = append(pts, current)
				current = core.Point{}
			}
		}
	}
	m.Vertices = pts
}

func (m *MLine) DxfTags() core.TagSlice {
	if R12Mode {
		return m.dxfTagsR12()
	}
	baseTags := baseEntityTags(&m.BaseEntity, "MLINE")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbMline")))
	}

	tags := append(baseTags,
		core.NewTag(2, core.NewStringValue(m.StyleName)),
	)

	if m.StyleHandle != "" {
		tags = append(tags, core.NewTag(340, core.NewStringValue(m.StyleHandle)))
	} else {
		tags = append(tags, core.NewTag(340, core.NewStringValue("1")))
	}
	tags = append(tags,
		core.NewTag(40, core.NewFloatValue(m.Scale)),
		core.NewTag(70, core.NewIntegerValue(m.Justification)),
		core.NewTag(71, core.NewIntegerValue(m.Flags)),
	)

	numVerts := len(m.Vertices)
	numElements := 2
	tags = append(tags,
		core.NewTag(72, core.NewIntegerValue(numVerts)),
		core.NewTag(73, core.NewIntegerValue(numElements)),
	)

	sp := m.StartPoint
	if sp.X == 0 && sp.Y == 0 && sp.Z == 0 && numVerts > 0 {
		sp = m.Vertices[0]
	}
	tags = append(tags,
		core.NewTag(10, core.NewFloatValue(sp.X)),
		core.NewTag(20, core.NewFloatValue(sp.Y)),
		core.NewTag(30, core.NewFloatValue(sp.Z)),
	)

	tags = append(tags, pointToTags210(m.ExtrusionDirection)...)

	for i := 0; i < numVerts; i++ {
		pt := m.Vertices[i]

		var dirX, dirY float64
		if i < numVerts-1 {
			nx := m.Vertices[i+1]
			dx := nx.X - pt.X
			dy := nx.Y - pt.Y
			mag := math.Sqrt(dx*dx + dy*dy)
			if mag > 0 {
				dirX = dx / mag
				dirY = dy / mag
			}
		} else if i > 0 {
			px := m.Vertices[i-1]
			dx := pt.X - px.X
			dy := pt.Y - px.Y
			mag := math.Sqrt(dx*dx + dy*dy)
			if mag > 0 {
				dirX = dx / mag
				dirY = dy / mag
			}
		}

		var miterX, miterY float64
		if i == 0 || i == numVerts-1 {
			miterX = -dirY
			miterY = dirX
		} else {

			px := m.Vertices[i-1]
			dx1 := pt.X - px.X
			dy1 := pt.Y - px.Y
			mag1 := math.Sqrt(dx1*dx1 + dy1*dy1)
			var perpInX, perpInY float64
			if mag1 > 0 {
				perpInX = -dy1 / mag1
				perpInY = dx1 / mag1
			}
			nx := m.Vertices[i+1]
			dx2 := nx.X - pt.X
			dy2 := nx.Y - pt.Y
			mag2 := math.Sqrt(dx2*dx2 + dy2*dy2)
			var perpOutX, perpOutY float64
			if mag2 > 0 {
				perpOutX = -dy2 / mag2
				perpOutY = dx2 / mag2
			}
			mx := perpInX + perpOutX
			my := perpInY + perpOutY
			mLen := math.Sqrt(mx*mx + my*my)
			if mLen > 0 {
				miterX = mx / mLen
				miterY = my / mLen
			}
		}

		tags = append(tags,
			core.NewTag(11, core.NewFloatValue(pt.X)),
			core.NewTag(21, core.NewFloatValue(pt.Y)),
			core.NewTag(31, core.NewFloatValue(pt.Z)),
			core.NewTag(12, core.NewFloatValue(dirX)),
			core.NewTag(22, core.NewFloatValue(dirY)),
			core.NewTag(32, core.NewFloatValue(0.0)),
			core.NewTag(13, core.NewFloatValue(miterX)),
			core.NewTag(23, core.NewFloatValue(miterY)),
			core.NewTag(33, core.NewFloatValue(0.0)),
			core.NewTag(74, core.NewIntegerValue(numElements)),
		)
		for j := 0; j < numElements; j++ {
			tags = append(tags, core.NewTag(41, core.NewFloatValue(0.0)))
		}
		tags = append(tags, core.NewTag(75, core.NewIntegerValue(0)))

		tags = append(tags, core.NewTag(74, core.NewIntegerValue(numElements)))
		for j := 0; j < numElements; j++ {
			v := 0.0
			if j == 0 {
				v = -1.0
			}
			tags = append(tags, core.NewTag(41, core.NewFloatValue(v)))
		}
		tags = append(tags, core.NewTag(75, core.NewIntegerValue(0)))
	}

	return AppendXData(tags, &m.BaseEntity)
}

func (m *MLine) dxfTagsR12() core.TagSlice {
	layerName := m.LayerName
	if layerName == "" {
		layerName = "0"
	}
	blockName := "*U" + strconv.Itoa(NextR12BlockSeq())
	scale := m.Scale
	if core.FloatEquals(scale, 0) {
		scale = 1.0
	}

	offsets := []float64{0.5, -0.5}

	h1 := nextR12HatchHandle()
	hEnd := nextR12HatchHandle()

	blockTags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("BLOCK")),
		core.NewTag(5, core.NewStringValue(h1)),
		core.NewTag(8, core.NewStringValue(layerName)),
		core.NewTag(2, core.NewStringValue(blockName)),
		core.NewTag(70, core.NewIntegerValue(1)),
		core.NewTag(10, core.NewFloatValue(0.0)),
		core.NewTag(20, core.NewFloatValue(0.0)),
		core.NewTag(30, core.NewFloatValue(0.0)),
		core.NewTag(3, core.NewStringValue(blockName)),
		core.NewTag(1, core.NewStringValue("")),
	}

	type segInfo struct {
		p1, p2 core.Point
		dx, dy float64
		length  float64
		px, py  float64
	}
	segments := make([]segInfo, len(m.Vertices)-1)
	for i := 1; i < len(m.Vertices); i++ {
		p1 := m.Vertices[i-1]
		p2 := m.Vertices[i]
		dx := p2.X - p1.X
		dy := p2.Y - p1.Y
		length := math.Sqrt(dx*dx + dy*dy)
		if length < 1e-9 {
			continue
		}
		segments[i-1] = segInfo{
			p1: p1, p2: p2,
			dx: dx, dy: dy, length: length,
			px: -dy / length, py: dx / length,
		}
	}

	type miterPoint struct {
		point core.Point
		valid bool
	}

	for _, off := range offsets {
		offDist := off * scale

		miterPts := make([]miterPoint, len(m.Vertices))
		for vi := 1; vi < len(m.Vertices)-1; vi++ {
			prevSeg := segments[vi-1]
			nextSeg := segments[vi]

			A := core.Point{X: m.Vertices[vi].X + prevSeg.px*offDist, Y: m.Vertices[vi].Y + prevSeg.py*offDist}
			U := core.Point{X: prevSeg.dx / prevSeg.length, Y: prevSeg.dy / prevSeg.length}

			B := core.Point{X: m.Vertices[vi].X + nextSeg.px*offDist, Y: m.Vertices[vi].Y + nextSeg.py*offDist}
			V := core.Point{X: nextSeg.dx / nextSeg.length, Y: nextSeg.dy / nextSeg.length}

			denom := U.X*V.Y - U.Y*V.X
			if math.Abs(denom) > 1e-9 {
				t := ((B.X-A.X)*V.Y - (B.Y-A.Y)*V.X) / denom
				miterPts[vi] = miterPoint{
					point: core.Point{X: A.X + t*U.X, Y: A.Y + t*U.Y, Z: 0},
					valid: true,
				}
			}
		}

		for i, seg := range segments {
			if seg.length < 1e-9 {
				continue
			}

			var startPt core.Point
			if i == 0 {
				startPt = core.Point{X: seg.p1.X + seg.px*offDist, Y: seg.p1.Y + seg.py*offDist, Z: seg.p1.Z}
			} else {
				mp := miterPts[i]
				if mp.valid {
					startPt = mp.point
				} else {
					startPt = core.Point{X: seg.p1.X + seg.px*offDist, Y: seg.p1.Y + seg.py*offDist, Z: seg.p1.Z}
				}
			}

			var endPt core.Point
			if i == len(segments)-1 {
				endPt = core.Point{X: seg.p2.X + seg.px*offDist, Y: seg.p2.Y + seg.py*offDist, Z: seg.p2.Z}
			} else {
				mp := miterPts[i+1]
				if mp.valid {
					endPt = mp.point
				} else {
					endPt = core.Point{X: seg.p2.X + seg.px*offDist, Y: seg.p2.Y + seg.py*offDist, Z: seg.p2.Z}
				}
			}

			h := nextR12HatchHandle()
			blockTags = append(blockTags,
				core.NewTag(0, core.NewStringValue("LINE")),
				core.NewTag(5, core.NewStringValue(h)),
				core.NewTag(8, core.NewStringValue(layerName)),
				core.NewTag(10, core.NewFloatValue(startPt.X)),
				core.NewTag(20, core.NewFloatValue(startPt.Y)),
				core.NewTag(30, core.NewFloatValue(startPt.Z)),
				core.NewTag(11, core.NewFloatValue(endPt.X)),
				core.NewTag(21, core.NewFloatValue(endPt.Y)),
				core.NewTag(31, core.NewFloatValue(endPt.Z)),
			)
		}
	}

	blockTags = append(blockTags,
		core.NewTag(0, core.NewStringValue("ENDBLK")),
		core.NewTag(5, core.NewStringValue(hEnd)),
		core.NewTag(8, core.NewStringValue(layerName)),
	)
	R12ExtraBlocks[blockName] = blockTags

	var tags core.TagSlice
	tags = append(tags,
		core.NewTag(0, core.NewStringValue("INSERT")),
		core.NewTag(8, core.NewStringValue(layerName)),
		core.NewTag(2, core.NewStringValue(blockName)),
		core.NewTag(10, core.NewFloatValue(0.0)),
		core.NewTag(20, core.NewFloatValue(0.0)),
		core.NewTag(30, core.NewFloatValue(0.0)),
	)
	return AppendXData(tags, &m.BaseEntity)
}

func NewMLineEntity(vertices core.PointSlice, layer string) *MLine {
	ml := &MLine{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		StyleName:          "Standard",
		Scale:              1.0,
		Flags:              1,
		Vertices:           vertices,
		ExtrusionDirection: core.Point{X: 0, Y: 0, Z: 1},
	}
	if len(vertices) > 0 {
		ml.StartPoint = vertices[0]
	}
	return ml
}

func (m MLine) Clone() Entity {
	verts := make(core.PointSlice, len(m.Vertices))
	copy(verts, m.Vertices)
	n := NewMLineEntity(verts, m.LayerName)
	n.BaseEntity = m.BaseEntity.CloneBase()
	n.Scale = m.Scale
	n.Justification = m.Justification
	n.StyleName = m.StyleName
	n.StyleHandle = m.StyleHandle
	n.Flags = m.Flags
	n.StartPoint = m.StartPoint
	n.ExtrusionDirection = m.ExtrusionDirection
	n.VertexParams = make([]float64, len(m.VertexParams))
	copy(n.VertexParams, m.VertexParams)
	n.AreaFillParams = make([]float64, len(m.AreaFillParams))
	copy(n.AreaFillParams, m.AreaFillParams)
	return n
}
