package entities

import (
	"math"
	"strconv"

	"github.com/inkbamboo/rwdxf/core"
)

// Hatch 表示 DXF HATCH 实体（填充图案）。
type Hatch struct {
	RegularEntity
	BaseEntity
	ElevationPoint     core.Point
	ExtrusionDirection core.Point
	PatternName        string
	SolidFill          bool
	Associative        bool
	HatchStyle         int
	PatternType        int
	PatternAngle       float64
	PatternScale       float64
	PatternDouble      bool

	BoundaryPaths []core.TagSlice

	SeedPoints core.PointSlice

	R12BlockName      string
	R12BoundaryPoints []core.PointSlice
}

func (h Hatch) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Hatch); ok {
		return h.BaseEntity.Equals(o.BaseEntity) &&
			h.PatternName == o.PatternName &&
			h.SolidFill == o.SolidFill
	}
	return false
}

func (h Hatch) DxfType() core.DxfTypeName { return core.DxfTypeHatch }

func (h Hatch) IsR12Compatible() bool { return true }

var r12HatchBlocks = make(map[string]core.TagSlice)

func ResetR12HatchBlocks() {
	r12HatchBlocks = make(map[string]core.TagSlice)
	r12HatchHandleSeq = 500
}

func GetR12HatchBlocks() map[string]core.TagSlice {
	result := make(map[string]core.TagSlice, len(r12HatchBlocks))
	for k, v := range r12HatchBlocks {
		result[k] = v
	}
	return result
}
var r12HatchHandleSeq = 500

func NewHatch(tags core.TagSlice) (*Hatch, error) {
	hatch := new(Hatch)
	hatch.ExtrusionDirection = core.Point{X: 0, Y: 0, Z: 1}
	hatch.PatternScale = 1.0

	hatch.PatternType = 1
	hatch.InitBaseEntityParser()
	hatch.Update(map[int]core.TypeParser{
		10:  core.NewFloatTypeParserToVar(&hatch.ElevationPoint.X),
		20:  core.NewFloatTypeParserToVar(&hatch.ElevationPoint.Y),
		30:  core.NewFloatTypeParserToVar(&hatch.ElevationPoint.Z),
		210: core.NewFloatTypeParserToVar(&hatch.ExtrusionDirection.X),
		220: core.NewFloatTypeParserToVar(&hatch.ExtrusionDirection.Y),
		230: core.NewFloatTypeParserToVar(&hatch.ExtrusionDirection.Z),
		2:   core.NewStringTypeParserToVar(&hatch.PatternName),
		70: core.NewIntTypeParser(func(v int) {
			hatch.SolidFill = v == 1
		}),
		71: core.NewIntTypeParser(func(v int) {
			hatch.Associative = v == 1
		}),
		75: core.NewIntTypeParserToVar(&hatch.HatchStyle),
		76: core.NewIntTypeParserToVar(&hatch.PatternType),
		52: core.NewFloatTypeParserToVar(&hatch.PatternAngle),
		41: core.NewFloatTypeParserToVar(&hatch.PatternScale),
		77: core.NewIntTypeParser(func(v int) {
			hatch.PatternDouble = v == 1
		}),
	})
	hatch.Parse(tags)
	hatch.XData = CollectXDataFromTags(tags)

	hatch.parseBoundaryPaths(tags)
	return hatch, nil
}

func (h *Hatch) parseBoundaryPaths(tags core.TagSlice) {
	regularTags := tags.RegularTags()

	boundaryCount := 0
	for _, t := range regularTags {
		if t.Code == 91 {
			if v, ok := core.AsInt(t.Value); ok {
				boundaryCount = v
			}
			break
		}
	}
	if boundaryCount <= 0 {
		return
	}
	inPath := false
	var currentPath core.TagSlice
	for _, t := range regularTags {
		if t.Code == 75 {
			break
		}
		if t.Code == 92 {
			if inPath && len(currentPath) > 0 {
				h.BoundaryPaths = append(h.BoundaryPaths, currentPath)
			}
			if len(h.BoundaryPaths) >= boundaryCount {
				break
			}
			currentPath = core.TagSlice{t}
			inPath = true
		} else if inPath {
			currentPath = append(currentPath, t)
			if t.Code == 97 {

				h.BoundaryPaths = append(h.BoundaryPaths, currentPath[:len(currentPath)-1])
				currentPath = nil
				inPath = false
				if len(h.BoundaryPaths) >= boundaryCount {
					break
				}
			}
		}
	}
}

func (h *Hatch) DxfTags() core.TagSlice {
	if R12Mode {
		return h.dxfTagsR12()
	}
	return h.dxfTagsAC1032()
}

func (h *Hatch) dxfTagsAC1032() core.TagSlice {
	solidFlag := 0
	if h.SolidFill {
		solidFlag = 1
	}
	assocFlag := 0
	if h.Associative {
		assocFlag = 1
	}
	doubleFlag := 0
	if h.PatternDouble {
		doubleFlag = 1
	}
	baseTags := baseEntityTags(&h.BaseEntity, "HATCH")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbHatch")))
	}
	tags := append(baseTags,
		core.NewTag(10, core.NewFloatValue(h.ElevationPoint.X)),
		core.NewTag(20, core.NewFloatValue(h.ElevationPoint.Y)),
		core.NewTag(30, core.NewFloatValue(h.ElevationPoint.Z)),
		core.NewTag(210, core.NewFloatValue(h.ExtrusionDirection.X)),
		core.NewTag(220, core.NewFloatValue(h.ExtrusionDirection.Y)),
		core.NewTag(230, core.NewFloatValue(h.ExtrusionDirection.Z)),
		core.NewTag(2, core.NewStringValue(h.PatternName)),
		core.NewTag(70, core.NewIntegerValue(solidFlag)),
		core.NewTag(71, core.NewIntegerValue(assocFlag)),
		core.NewTag(91, core.NewIntegerValue(len(h.BoundaryPaths))),
	)

	for _, bp := range h.BoundaryPaths {
		tags = append(tags, bp...)

		tags = append(tags, core.NewTag(97, core.NewIntegerValue(0)))
	}

	tags = append(tags,
		core.NewTag(75, core.NewIntegerValue(h.HatchStyle)),
		core.NewTag(76, core.NewIntegerValue(h.PatternType)),
	)
	if !h.SolidFill {

		angleRad := h.PatternAngle * math.Pi / 180.0
		baseSpacing := 0.125
		offsetX := -baseSpacing * math.Sin(angleRad) * h.PatternScale
		offsetY := baseSpacing * math.Cos(angleRad) * h.PatternScale
		tags = append(tags,
			core.NewTag(52, core.NewFloatValue(h.PatternAngle)),
			core.NewTag(41, core.NewFloatValue(h.PatternScale)),
			core.NewTag(77, core.NewIntegerValue(doubleFlag)),

			core.NewTag(78, core.NewIntegerValue(1)),
			core.NewTag(53, core.NewFloatValue(h.PatternAngle)),
			core.NewTag(43, core.NewFloatValue(0.0)),
			core.NewTag(44, core.NewFloatValue(0.0)),
			core.NewTag(45, core.NewFloatValue(offsetX)),
			core.NewTag(46, core.NewFloatValue(offsetY)),
			core.NewTag(79, core.NewIntegerValue(0)),
		)
	}

	if len(h.SeedPoints) > 0 {
		tags = append(tags, core.NewTag(98, core.NewIntegerValue(len(h.SeedPoints))))
		for _, pt := range h.SeedPoints {
			tags = append(tags,
				core.NewTag(10, core.NewFloatValue(pt.X)),
				core.NewTag(20, core.NewFloatValue(pt.Y)),
			)
		}
	} else {
		tags = append(tags, core.NewTag(98, core.NewIntegerValue(0)))
	}

	return AppendXData(tags, &h.BaseEntity)
}

func (h *Hatch) dxfTagsR12() core.TagSlice {
	var tags core.TagSlice

	layerName := h.LayerName
	if layerName == "" {
		layerName = "0"
	}

	h.R12BlockName = "*H" + strconv.Itoa(NextR12BlockSeq())

	insertHandle := nextR12HatchHandle()

	h.R12BoundaryPoints = h.parseBoundaryPoints()
	boundaryHandles := make([]string, 0, len(h.R12BoundaryPoints))

	for _, pts := range h.R12BoundaryPoints {
		if len(pts) < 2 {
			continue
		}
		handle := nextR12HatchHandle()
		boundaryHandles = append(boundaryHandles, handle)

		closed := pts[0].Equals(pts[len(pts)-1])
		tags = append(tags, core.NewTag(0, core.NewStringValue("POLYLINE")))
		tags = append(tags, core.NewTag(5, core.NewStringValue(handle)))
		tags = append(tags, core.NewTag(8, core.NewStringValue(layerName)))
		tags = append(tags, core.NewTag(66, core.NewIntegerValue(1)))
		if closed {
			tags = append(tags, core.NewTag(70, core.NewIntegerValue(1)))
		}
		tags = append(tags, core.NewTag(10, core.NewFloatValue(0.0)))
		tags = append(tags, core.NewTag(20, core.NewFloatValue(0.0)))
		tags = append(tags, core.NewTag(30, core.NewFloatValue(0.0)))

		vertexCount := len(pts)
		if closed {
			vertexCount--
		}
		for j := 0; j < vertexCount; j++ {
			tags = append(tags, core.NewTag(0, core.NewStringValue("VERTEX")))
			tags = append(tags, core.NewTag(5, core.NewStringValue(nextR12HatchHandle())))
			tags = append(tags, core.NewTag(8, core.NewStringValue(layerName)))
			tags = append(tags, core.NewTag(10, core.NewFloatValue(pts[j].X)))
			tags = append(tags, core.NewTag(20, core.NewFloatValue(pts[j].Y)))
			tags = append(tags, core.NewTag(30, core.NewFloatValue(pts[j].Z)))
		}

		tags = append(tags, core.NewTag(0, core.NewStringValue("SEQEND")))
		tags = append(tags, core.NewTag(5, core.NewStringValue(nextR12HatchHandle())))
		tags = append(tags, core.NewTag(8, core.NewStringValue(layerName)))
	}

	tags = append(tags, core.NewTag(0, core.NewStringValue("INSERT")))
	tags = append(tags, core.NewTag(5, core.NewStringValue(insertHandle)))
	tags = append(tags, core.NewTag(8, core.NewStringValue(layerName)))
	tags = append(tags, core.NewTag(2, core.NewStringValue(h.R12BlockName)))
	tags = append(tags, core.NewTag(10, core.NewFloatValue(0.0)))
	tags = append(tags, core.NewTag(20, core.NewFloatValue(0.0)))
	tags = append(tags, core.NewTag(30, core.NewFloatValue(0.0)))

	tags = append(tags, h.r12XDataTags(boundaryHandles, insertHandle)...)

	h.registerR12HatchBlock(layerName)

	return AppendXData(tags, &h.BaseEntity)
}

func (h *Hatch) parseBoundaryPoints() []core.PointSlice {
	var result []core.PointSlice
	for _, bp := range h.BoundaryPaths {
		var pts core.PointSlice
		var startX, startY float64
		var endX, endY float64
		edgeIndex := 0

		collectStart := func(x, y float64) {
			startX, startY = x, y
		}
		collectEnd := func(x, y float64) {
			endX, endY = x, y
		}
		flushEdge := func() {
			if edgeIndex == 0 {

				pts = append(pts, core.Point{X: startX, Y: startY, Z: 0})
			}
			pts = append(pts, core.Point{X: endX, Y: endY, Z: 0})
			edgeIndex++
		}

		for _, t := range bp {
			switch t.Code {
			case 72:
				if v, ok := core.AsInt(t.Value); ok && v == 1 {

				}
			case 10:
				if v, ok := core.AsFloat(t.Value); ok {
					collectStart(v, startY)
				}
			case 20:
				if v, ok := core.AsFloat(t.Value); ok {
					collectStart(startX, v)
				}
			case 11:
				if v, ok := core.AsFloat(t.Value); ok {
					collectEnd(v, endY)
				}
			case 21:
				if v, ok := core.AsFloat(t.Value); ok {
					collectEnd(endX, v)
					flushEdge()
				}
			}
		}
		if len(pts) > 0 {

			if !pts[0].Equals(pts[len(pts)-1]) {
				pts = append(pts, pts[0])
			}
			result = append(result, pts)
		}
	}
	return result
}

func (h *Hatch) r12XDataTags(boundaryHandles []string, insertHandle string) core.TagSlice {
	var xd core.TagSlice

	solidFlag := 0
	if h.SolidFill {
		solidFlag = 1
	}
	assocFlag := 0
	if h.Associative {
		assocFlag = 1
	}

	patternName := h.PatternName
	if patternName == "" {
		patternName = "SOLID"
	}

	scale := h.PatternScale
	if core.FloatEquals(scale, 0.0) {
		scale = 1.0
	}

	xd = append(xd, core.NewTag(1001, core.NewStringValue("ACAD")))
	xd = append(xd, core.NewTag(1000, core.NewStringValue("HATCH")))
	xd = append(xd, core.NewTag(1002, core.NewStringValue("{")))

	xd = append(xd, core.NewTag(1070, core.NewIntegerValue(19)))

	xd = append(xd, core.NewTag(1000, core.NewStringValue(patternName)))

	xd = append(xd, core.NewTag(1040, core.NewFloatValue(scale)))

	xd = append(xd, core.NewTag(1040, core.NewFloatValue(h.PatternAngle)))

	xd = append(xd, core.NewTag(1000, core.NewStringValue("ASC_BOUNDS")))
	xd = append(xd, core.NewTag(1070, core.NewIntegerValue(len(boundaryHandles))))
	for _, bh := range boundaryHandles {

		xd = append(xd, core.NewTag(1070, core.NewIntegerValue(1)))
		xd = append(xd, core.NewTag(1070, core.NewIntegerValue(1)))
		xd = append(xd, core.NewTag(1005, core.NewStringValue(bh)))
	}

	xd = append(xd, core.NewTag(1000, core.NewStringValue("ASC_SEEDPOINT")))

	seedX, seedY := h.computeSeedPoint()
	xd = append(xd, core.NewTag(1011, core.NewFloatValue(seedX)))
	xd = append(xd, core.NewTag(1021, core.NewFloatValue(seedY)))
	xd = append(xd, core.NewTag(1031, core.NewFloatValue(0.0)))

	xd = append(xd, core.NewTag(1021, core.NewFloatValue(seedY+h.PatternScale*0.1)))
	xd = append(xd, core.NewTag(1031, core.NewFloatValue(0.0)))

	xd = append(xd, core.NewTag(1000, core.NewStringValue("R14_HATCH_DATA")))
	xd = append(xd, core.NewTag(1000, core.NewStringValue(insertHandle)))

	xd = append(xd, core.NewTag(1011, core.NewFloatValue(1.0)))
	xd = append(xd, core.NewTag(1021, core.NewFloatValue(0.0)))
	xd = append(xd, core.NewTag(1031, core.NewFloatValue(0.0)))
	xd = append(xd, core.NewTag(1021, core.NewFloatValue(0.0)))
	xd = append(xd, core.NewTag(1031, core.NewFloatValue(0.0)))

	xd = append(xd, core.NewTag(1011, core.NewFloatValue(0.0)))
	xd = append(xd, core.NewTag(1021, core.NewFloatValue(0.0)))
	xd = append(xd, core.NewTag(1031, core.NewFloatValue(0.0)))
	xd = append(xd, core.NewTag(1021, core.NewFloatValue(1.0)))
	xd = append(xd, core.NewTag(1031, core.NewFloatValue(0.0)))

	xd = append(xd, core.NewTag(1011, core.NewFloatValue(0.0)))
	xd = append(xd, core.NewTag(1021, core.NewFloatValue(0.0)))
	xd = append(xd, core.NewTag(1031, core.NewFloatValue(0.0)))
	xd = append(xd, core.NewTag(1021, core.NewFloatValue(0.0)))
	xd = append(xd, core.NewTag(1031, core.NewFloatValue(1.0)))

	xd = append(xd, core.NewTag(1040, core.NewFloatValue(0.0)))

	xd = append(xd, core.NewTag(1010, core.NewFloatValue(0.0)))
	xd = append(xd, core.NewTag(1020, core.NewFloatValue(0.0)))
	xd = append(xd, core.NewTag(1030, core.NewFloatValue(1.0)))

	xd = append(xd, core.NewTag(1000, core.NewStringValue(patternName)))

	xd = append(xd, core.NewTag(1070, core.NewIntegerValue(solidFlag)))
	xd = append(xd, core.NewTag(1070, core.NewIntegerValue(assocFlag)))

	totalLoops := len(h.BoundaryPaths)
	totalEdges := 0
	for _, pts := range h.R12BoundaryPoints {
		if len(pts) >= 2 {
			closed := pts[0].Equals(pts[len(pts)-1])
			if closed {
				totalEdges += len(pts) - 1
			} else {
				totalEdges += len(pts)
			}
		}
	}
	xd = append(xd, core.NewTag(1071, core.NewIntegerValue(totalLoops)))
	xd = append(xd, core.NewTag(1071, core.NewIntegerValue(totalEdges)))

	xd = append(xd, core.NewTag(1070, core.NewIntegerValue(h.HatchStyle)))
	xd = append(xd, core.NewTag(1070, core.NewIntegerValue(h.PatternType)))

	totalEdgeCount := 0
	for i, pts := range h.R12BoundaryPoints {
		if len(pts) < 2 {
			continue
		}
		closed := pts[0].Equals(pts[len(pts)-1])
		edgeCount := len(pts)
		if closed {
			edgeCount = len(pts) - 1
		}
		xd = append(xd, core.NewTag(1071, core.NewIntegerValue(edgeCount)))

		for j := 0; j < edgeCount; j++ {
			xd = append(xd, core.NewTag(1040, core.NewFloatValue(pts[j].X)))
			xd = append(xd, core.NewTag(1040, core.NewFloatValue(pts[j].Y)))
		}
		totalEdgeCount += edgeCount

		xd = append(xd, core.NewTag(1071, core.NewIntegerValue(1)))
		if i < len(boundaryHandles) {
			xd = append(xd, core.NewTag(1005, core.NewStringValue(boundaryHandles[i])))
		}
	}

	angleRad := h.PatternAngle * math.Pi / 180.0
	baseSpacing := 0.125
	offsetX := -baseSpacing * math.Sin(angleRad) * scale
	offsetY := baseSpacing * math.Cos(angleRad) * scale

	xd = append(xd, core.NewTag(1070, core.NewIntegerValue(0)))
	xd = append(xd, core.NewTag(1070, core.NewIntegerValue(1)))
	xd = append(xd, core.NewTag(1040, core.NewFloatValue(h.PatternAngle)))
	xd = append(xd, core.NewTag(1040, core.NewFloatValue(0.0)))
	xd = append(xd, core.NewTag(1040, core.NewFloatValue(0.0)))
	xd = append(xd, core.NewTag(1040, core.NewFloatValue(offsetX)))
	xd = append(xd, core.NewTag(1040, core.NewFloatValue(offsetY)))
	xd = append(xd, core.NewTag(1070, core.NewIntegerValue(0)))
	xd = append(xd, core.NewTag(1040, core.NewFloatValue(1.0)))

	xd = append(xd, core.NewTag(1071, core.NewIntegerValue(1)))
	xd = append(xd, core.NewTag(1040, core.NewFloatValue(seedX)))
	xd = append(xd, core.NewTag(1040, core.NewFloatValue(seedY)))
	xd = append(xd, core.NewTag(1040, core.NewFloatValue(0.0)))

	xd = append(xd, core.NewTag(1002, core.NewStringValue("}")))

	return xd
}

func (h *Hatch) computeSeedPoint() (float64, float64) {
	for _, pts := range h.R12BoundaryPoints {
		if len(pts) >= 3 {
			var sumX, sumY float64
			count := len(pts)

			if pts[0].Equals(pts[count-1]) {
				count--
			}
			if count > 0 {
				for i := 0; i < count; i++ {
					sumX += pts[i].X
					sumY += pts[i].Y
				}
				return sumX / float64(count), sumY / float64(count)
			}
		}
	}
	return 0.0, 0.0
}

func (h *Hatch) registerR12HatchBlock(layerName string) {
	if _, exists := r12HatchBlocks[h.R12BlockName]; exists {
		return
	}
	handle := nextR12HatchHandle()
	blockTags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("BLOCK")),
		core.NewTag(5, core.NewStringValue(handle)),
		core.NewTag(8, core.NewStringValue(layerName)),
		core.NewTag(2, core.NewStringValue(h.R12BlockName)),
		core.NewTag(70, core.NewIntegerValue(1)),
		core.NewTag(10, core.NewFloatValue(0.0)),
		core.NewTag(20, core.NewFloatValue(0.0)),
		core.NewTag(30, core.NewFloatValue(0.0)),
		core.NewTag(3, core.NewStringValue(h.R12BlockName)),
		core.NewTag(1, core.NewStringValue("")),
	}

	if h.SolidFill {
		blockTags = append(blockTags, h.generateSolidLines(layerName)...)
	} else {
		blockTags = append(blockTags, h.generatePatternLines(layerName)...)
	}

	blockTags = append(blockTags,
		core.NewTag(0, core.NewStringValue("ENDBLK")),
		core.NewTag(5, core.NewStringValue(nextR12HatchHandle())),
		core.NewTag(8, core.NewStringValue(layerName)),
	)
	r12HatchBlocks[h.R12BlockName] = blockTags
}

func (h *Hatch) generateSolidLines(layerName string) core.TagSlice {
	var tags core.TagSlice
	spacing := 0.5
	if !core.FloatEquals(h.PatternScale, 0) {
		spacing = 0.5 * h.PatternScale
	}

	for _, pts := range h.R12BoundaryPoints {
		hLines := h.scanLineFill(pts, 0.0, spacing)
		for _, seg := range hLines {
			tags = append(tags, makeLineTag(seg[0], seg[1], seg[2], seg[3], layerName)...)
		}
	}
	return AppendXData(tags, &h.BaseEntity)
}

func (h *Hatch) generatePatternLines(layerName string) core.TagSlice {
	var tags core.TagSlice
	angle := h.PatternAngle
	scale := h.PatternScale
	if core.FloatEquals(scale, 0) {
		scale = 1.0
	}

	spacing := 0.125 * scale

	for _, pts := range h.R12BoundaryPoints {
		hLines := h.scanLineFill(pts, angle, spacing)
		for _, seg := range hLines {
			tags = append(tags, makeLineTag(seg[0], seg[1], seg[2], seg[3], layerName)...)
		}
	}
	return AppendXData(tags, &h.BaseEntity)
}

func (h *Hatch) scanLineFill(polygon core.PointSlice, angle, spacing float64) [][4]float64 {
	if len(polygon) < 3 || spacing <= 0 {
		return nil
	}

	n := len(polygon)
	if !polygon[0].Equals(polygon[n-1]) {
		polygon = append(polygon, polygon[0])
		n++
	}

	rad := angle * math.Pi / 180.0
	cosN := math.Cos(-rad)
	sinN := math.Sin(-rad)

	rotated := make(core.PointSlice, n)
	for i := 0; i < n; i++ {
		rotated[i] = core.Point{
			X: polygon[i].X*cosN - polygon[i].Y*sinN,
			Y: polygon[i].X*sinN + polygon[i].Y*cosN,
			Z: 0,
		}
	}

	minY, maxY := rotated[0].Y, rotated[0].Y
	for _, p := range rotated {
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	var segments [][4]float64
	eps := 1e-9

	cosP := math.Cos(rad)
	sinP := math.Sin(rad)

	for y := minY + spacing*0.5; y <= maxY; y += spacing {
		var intersections []float64

		for i := 0; i < n-1; i++ {
			y1, y2 := rotated[i].Y, rotated[i+1].Y
			if (y1 < y && y2 >= y) || (y2 < y && y1 >= y) {
				if math.Abs(y2-y1) < eps {
					continue
				}
				t := (y - y1) / (y2 - y1)
				x := rotated[i].X + t*(rotated[i+1].X-rotated[i].X)
				intersections = append(intersections, x)
			}
		}

		sortFloat64(intersections)

		for i := 0; i+1 < len(intersections); i += 2 {
			x1 := intersections[i]
			x2 := intersections[i+1]

			if math.Abs(x2-x1) < eps {
				continue
			}

			ox1 := x1*cosP - y*sinP
			oy1 := x1*sinP + y*cosP
			ox2 := x2*cosP - y*sinP
			oy2 := x2*sinP + y*cosP

			segments = append(segments, [4]float64{ox1, oy1, ox2, oy2})
		}
	}

	return segments
}

func sortFloat64(a []float64) {
	for i := 0; i < len(a); i++ {
		for j := i + 1; j < len(a); j++ {
			if a[i] > a[j] {
				a[i], a[j] = a[j], a[i]
			}
		}
	}
}

func makeLineTag(x1, y1, x2, y2 float64, layerName string) core.TagSlice {
	handle := nextR12HatchHandle()
	return core.TagSlice{
		core.NewTag(0, core.NewStringValue("LINE")),
		core.NewTag(5, core.NewStringValue(handle)),
		core.NewTag(8, core.NewStringValue(layerName)),
		core.NewTag(10, core.NewFloatValue(x1)),
		core.NewTag(20, core.NewFloatValue(y1)),
		core.NewTag(30, core.NewFloatValue(0.0)),
		core.NewTag(11, core.NewFloatValue(x2)),
		core.NewTag(21, core.NewFloatValue(y2)),
		core.NewTag(31, core.NewFloatValue(0.0)),
	}
}

func nextR12HatchHandle() string {
	h := r12HatchHandleSeq
	r12HatchHandleSeq++
	return strconv.Itoa(h)
}

func NewHatchEntity(solid bool, patternName string, layer string) *Hatch {
	return &Hatch{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		PatternName:        patternName,
		SolidFill:          solid,
		Associative:        true,
		PatternType:        1,
		PatternScale:       1.0,
		ExtrusionDirection: core.Point{X: 0, Y: 0, Z: 1},
	}
}

func (h Hatch) Clone() Entity {
	n := NewHatchEntity(h.SolidFill, h.PatternName, h.LayerName)
	n.BaseEntity = h.BaseEntity.CloneBase()
	n.PatternAngle = h.PatternAngle
	n.PatternScale = h.PatternScale
	n.PatternType = h.PatternType
	n.Associative = h.Associative
	n.HatchStyle = h.HatchStyle
	n.PatternDouble = h.PatternDouble
	n.ElevationPoint = h.ElevationPoint
	n.ExtrusionDirection = h.ExtrusionDirection
	n.SeedPoints = make(core.PointSlice, len(h.SeedPoints))
	copy(n.SeedPoints, h.SeedPoints)
	for _, bp := range h.BoundaryPaths {
		cbp := make(core.TagSlice, len(bp))
		copy(cbp, bp)
		n.BoundaryPaths = append(n.BoundaryPaths, cbp)
	}
	return n
}
