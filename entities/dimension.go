package entities

import (
	"math"
	"strconv"

	"github.com/inkbamboo/rwdxf/core"
)

// Dimension 表示 DXF DIMENSION 实体（尺寸标注）。
type Dimension struct {
	RegularEntity
	BaseEntity

	Version              int
	Geometry             string
	DimStyleName         string
	DefPoint             core.Point
	TextMidpoint         core.Point
	Insert               core.Point
	DimType              int

	AttachmentPoint      int
	LineSpacingStyle     int
	LineSpacingFactor    float64
	ActualMeasurement    float64
	Unknown1             int
	FlipArrow1           bool
	FlipArrow2           bool
	Text                 string
	ObliqueAngle         float64
	TextRotation         float64
	HorizontalDirection  float64
	ExtrusionDirection   core.Point

	DefPoint2  core.Point
	DefPoint3  core.Point
	Angle      float64
	DefPoint4  core.Point
	LeaderLen  float64
	DefPoint5  core.Point
}

const (
	DimLinear    = 0
	DimAligned   = 1
	DimAngular   = 2
	DimDiameter  = 3
	DimRadius    = 4
	DimAngular3P = 5
	DimOrdinate  = 6
)

func (d Dimension) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Dimension); ok {
		return d.BaseEntity.Equals(o.BaseEntity) &&
			d.DimStyleName == o.DimStyleName &&
			d.DefPoint.Equals(o.DefPoint) &&
			d.DimType == o.DimType &&
			d.DefPoint2.Equals(o.DefPoint2) &&
			d.DefPoint3.Equals(o.DefPoint3) &&
			core.FloatEquals(d.Angle, o.Angle) &&
			core.FloatEquals(d.LeaderLen, o.LeaderLen) &&
			d.ExtrusionDirection.Equals(o.ExtrusionDirection)
	}
	return false
}

func (d Dimension) DxfType() core.DxfTypeName { return core.DxfTypeDimension }

func (d Dimension) IsR12Compatible() bool { return true }

var R12DimBlocks = make(map[string]core.TagSlice)

func ResetR12DimBlocks() {
	R12DimBlocks = make(map[string]core.TagSlice)
}

func GetR12DimBlocks() map[string]core.TagSlice {
	result := make(map[string]core.TagSlice, len(R12DimBlocks))
	for k, v := range R12DimBlocks {
		result[k] = v
	}
	return result
}

func (d *Dimension) DimTypeBase() int { return d.DimType & 15 }

func (d *Dimension) HasUserTextPosition() bool {
	return d.DimType&128 != 0
}

func (d *Dimension) IsOrdinateXType() bool {
	return d.DimType&64 != 0
}

func NewDimension(tags core.TagSlice) (*Dimension, error) {
	dim := new(Dimension)
	dim.DimStyleName = "Standard"
	dim.AttachmentPoint = 5
	dim.LineSpacingStyle = 1
	dim.ExtrusionDirection = core.Point{X: 0, Y: 0, Z: 1}
	dim.InitBaseEntityParser()

	dim.Update(map[int]core.TypeParser{

		280: core.NewIntTypeParserToVar(&dim.Version),
		2:   core.NewStringTypeParserToVar(&dim.Geometry),
		3:   core.NewStringTypeParserToVar(&dim.DimStyleName),
		10:  core.NewFloatTypeParserToVar(&dim.DefPoint.X),
		20:  core.NewFloatTypeParserToVar(&dim.DefPoint.Y),
		30:  core.NewFloatTypeParserToVar(&dim.DefPoint.Z),
		11:  core.NewFloatTypeParserToVar(&dim.TextMidpoint.X),
		21:  core.NewFloatTypeParserToVar(&dim.TextMidpoint.Y),
		31:  core.NewFloatTypeParserToVar(&dim.TextMidpoint.Z),
		12:  core.NewFloatTypeParserToVar(&dim.Insert.X),
		22:  core.NewFloatTypeParserToVar(&dim.Insert.Y),
		32:  core.NewFloatTypeParserToVar(&dim.Insert.Z),
		70:  core.NewIntTypeParserToVar(&dim.DimType),
		71:  core.NewIntTypeParserToVar(&dim.AttachmentPoint),
		72:  core.NewIntTypeParserToVar(&dim.LineSpacingStyle),
		41:  core.NewFloatTypeParserToVar(&dim.LineSpacingFactor),
		42:  core.NewFloatTypeParserToVar(&dim.ActualMeasurement),
		73:  core.NewIntTypeParserToVar(&dim.Unknown1),
		74:  core.NewIntTypeParser(func(v int) { dim.FlipArrow1 = v != 0 }),
		75:  core.NewIntTypeParser(func(v int) { dim.FlipArrow2 = v != 0 }),
		1:   core.NewStringTypeParserToVar(&dim.Text),
		52:  core.NewFloatTypeParserToVar(&dim.ObliqueAngle),
		53:  core.NewFloatTypeParserToVar(&dim.TextRotation),
		51:  core.NewFloatTypeParserToVar(&dim.HorizontalDirection),
		210: core.NewFloatTypeParserToVar(&dim.ExtrusionDirection.X),
		220: core.NewFloatTypeParserToVar(&dim.ExtrusionDirection.Y),
		230: core.NewFloatTypeParserToVar(&dim.ExtrusionDirection.Z),

		13: core.NewFloatTypeParserToVar(&dim.DefPoint2.X),
		23: core.NewFloatTypeParserToVar(&dim.DefPoint2.Y),
		33: core.NewFloatTypeParserToVar(&dim.DefPoint2.Z),
		14: core.NewFloatTypeParserToVar(&dim.DefPoint3.X),
		24: core.NewFloatTypeParserToVar(&dim.DefPoint3.Y),
		34: core.NewFloatTypeParserToVar(&dim.DefPoint3.Z),
		50: core.NewFloatTypeParserToVar(&dim.Angle),
		15: core.NewFloatTypeParserToVar(&dim.DefPoint4.X),
		25: core.NewFloatTypeParserToVar(&dim.DefPoint4.Y),
		35: core.NewFloatTypeParserToVar(&dim.DefPoint4.Z),
		40: core.NewFloatTypeParserToVar(&dim.LeaderLen),
		16: core.NewFloatTypeParserToVar(&dim.DefPoint5.X),
		26: core.NewFloatTypeParserToVar(&dim.DefPoint5.Y),
		36: core.NewFloatTypeParserToVar(&dim.DefPoint5.Z),
	})
	dim.Parse(tags)
	dim.XData = CollectXDataFromTags(tags)
	return dim, nil
}

func (d *Dimension) DxfTags() core.TagSlice {
	baseTags := baseEntityTags(&d.BaseEntity, "DIMENSION")

	if !R12Mode {

		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbDimension")))
		dimType := d.DimType
		dimTypeBase := dimType & 15

		tags := append(baseTags,
			core.NewTag(280, core.NewIntegerValue(d.Version)),
		)

		if d.Geometry != "" {
			tags = append(tags, core.NewTag(2, core.NewStringValue(d.Geometry)))
		}
		tags = append(tags,
			core.NewTag(3, core.NewStringValue(d.DimStyleName)),
			core.NewTag(10, core.NewFloatValue(d.DefPoint.X)),
			core.NewTag(20, core.NewFloatValue(d.DefPoint.Y)),
			core.NewTag(30, core.NewFloatValue(d.DefPoint.Z)),
		)

		if !core.FloatEquals(d.TextMidpoint.X, 0) || !core.FloatEquals(d.TextMidpoint.Y, 0) || !core.FloatEquals(d.TextMidpoint.Z, 0) {
			tags = append(tags,
				core.NewTag(11, core.NewFloatValue(d.TextMidpoint.X)),
				core.NewTag(21, core.NewFloatValue(d.TextMidpoint.Y)),
				core.NewTag(31, core.NewFloatValue(d.TextMidpoint.Z)),
			)
		}

		if !core.FloatEquals(d.Insert.X, 0) || !core.FloatEquals(d.Insert.Y, 0) || !core.FloatEquals(d.Insert.Z, 0) {
			tags = append(tags,
				core.NewTag(12, core.NewFloatValue(d.Insert.X)),
				core.NewTag(22, core.NewFloatValue(d.Insert.Y)),
				core.NewTag(32, core.NewFloatValue(d.Insert.Z)),
			)
		}

		tags = append(tags,
			core.NewTag(70, core.NewIntegerValue(dimType)),
			core.NewTag(71, core.NewIntegerValue(d.AttachmentPoint)),
		)

		if d.LineSpacingStyle != 1 {
			tags = append(tags, core.NewTag(72, core.NewIntegerValue(d.LineSpacingStyle)))
		}
		if !core.FloatEquals(d.LineSpacingFactor, 0) {
			tags = append(tags, core.NewTag(41, core.NewFloatValue(d.LineSpacingFactor)))
		}
		if !core.FloatEquals(d.ActualMeasurement, 0) {
			tags = append(tags, core.NewTag(42, core.NewFloatValue(d.ActualMeasurement)))
		}
		if d.Unknown1 != 0 {
			tags = append(tags, core.NewTag(73, core.NewIntegerValue(d.Unknown1)))
		}
		if d.FlipArrow1 {
			tags = append(tags, core.NewTag(74, core.NewIntegerValue(1)))
		}
		if d.FlipArrow2 {
			tags = append(tags, core.NewTag(75, core.NewIntegerValue(1)))
		}

		text := d.Text
		if text == "" {
			text = "<>"
		}
		tags = append(tags, core.NewTag(1, core.NewStringValue(text)))
		if !core.FloatEquals(d.ObliqueAngle, 0) {
			tags = append(tags, core.NewTag(52, core.NewFloatValue(d.ObliqueAngle)))
		}
		if !core.FloatEquals(d.TextRotation, 0) {
			tags = append(tags, core.NewTag(53, core.NewFloatValue(d.TextRotation)))
		}
		if !core.FloatEquals(d.HorizontalDirection, 0) {
			tags = append(tags, core.NewTag(51, core.NewFloatValue(d.HorizontalDirection)))
		}
		if !isDefaultExtrusion(d.ExtrusionDirection) {
			tags = append(tags, pointToTags210(d.ExtrusionDirection)...)
		}

		switch {
		case dimTypeBase == DimLinear:
			tags = append(tags, core.NewTag(100, core.NewStringValue("AcDbAlignedDimension")))
			tags = append(tags, dimPointFieldTags(13, d.DefPoint2)...)
			tags = append(tags, dimPointFieldTags(14, d.DefPoint3)...)
			if !core.FloatEquals(d.Angle, 0) {
				tags = append(tags, core.NewTag(50, core.NewFloatValue(d.Angle)))
			}
			tags = append(tags, core.NewTag(100, core.NewStringValue("AcDbRotatedDimension")))
		case dimTypeBase == DimAligned:

			tags = append(tags, core.NewTag(100, core.NewStringValue("AcDbAlignedDimension")))
			tags = append(tags, dimPointFieldTags(13, d.DefPoint2)...)
			tags = append(tags, dimPointFieldTags(14, d.DefPoint3)...)
			if !core.FloatEquals(d.Angle, 0) {
				tags = append(tags, core.NewTag(50, core.NewFloatValue(d.Angle)))
			}
		case dimTypeBase == DimAngular:
			tags = append(tags, core.NewTag(100, core.NewStringValue("AcDb2LineAngularDimension")))
			tags = append(tags, dimPointFieldTags(13, d.DefPoint2)...)
			tags = append(tags, dimPointFieldTags(14, d.DefPoint3)...)
			tags = append(tags, dimPointFieldTags(15, d.DefPoint4)...)
			tags = append(tags, dimPointFieldTags(16, d.DefPoint5)...)
		case dimTypeBase == DimDiameter:
			tags = append(tags, core.NewTag(100, core.NewStringValue("AcDbDiametricDimension")))
			tags = append(tags, dimPointFieldTags(15, d.DefPoint4)...)
			tags = append(tags, core.NewTag(40, core.NewFloatValue(d.LeaderLen)))
		case dimTypeBase == DimRadius:
			tags = append(tags, core.NewTag(100, core.NewStringValue("AcDbRadialDimension")))
			tags = append(tags, dimPointFieldTags(15, d.DefPoint4)...)
			tags = append(tags, core.NewTag(40, core.NewFloatValue(d.LeaderLen)))
		case dimTypeBase == DimAngular3P:
			tags = append(tags, core.NewTag(100, core.NewStringValue("AcDb3PointAngularDimension")))
			tags = append(tags, dimPointFieldTags(13, d.DefPoint2)...)
			tags = append(tags, dimPointFieldTags(14, d.DefPoint3)...)
			tags = append(tags, dimPointFieldTags(15, d.DefPoint4)...)
			tags = append(tags, dimPointFieldTags(16, d.DefPoint5)...)
		case dimTypeBase == DimOrdinate:
			tags = append(tags, core.NewTag(100, core.NewStringValue("AcDbOrdinateDimension")))
			tags = append(tags, dimPointFieldTags(13, d.DefPoint2)...)
			tags = append(tags, dimPointFieldTags(14, d.DefPoint3)...)
		}
		return AppendXData(tags, &d.BaseEntity)
	}

	tags := baseTags
	tags = append(tags, d.r12DimEntityTags()...)
	return AppendXData(tags, &d.BaseEntity)
}

func (d *Dimension) r12DimEntityTags() core.TagSlice {

	textBlockName, textTags := d.generateR12TextBlock()

	geomBlockName, geomTags := d.generateR12GeomBlock(textBlockName)

	if _, exists := R12DimBlocks[geomBlockName]; !exists {
		R12DimBlocks[geomBlockName] = geomTags
	}
	if textBlockName != "" {
		if _, exists := R12DimBlocks[textBlockName]; !exists {
			R12DimBlocks[textBlockName] = textTags
		}
	}

	var tags core.TagSlice
	tags = append(tags, core.NewTag(2, core.NewStringValue(geomBlockName)))
	tags = append(tags,
		core.NewTag(10, core.NewFloatValue(d.DefPoint.X)),
		core.NewTag(20, core.NewFloatValue(d.DefPoint.Y)),
		core.NewTag(30, core.NewFloatValue(d.DefPoint.Z)),
	)
	if !core.FloatEquals(d.TextMidpoint.X, 0) || !core.FloatEquals(d.TextMidpoint.Y, 0) || !core.FloatEquals(d.TextMidpoint.Z, 0) {
		tags = append(tags,
			core.NewTag(11, core.NewFloatValue(d.TextMidpoint.X)),
			core.NewTag(21, core.NewFloatValue(d.TextMidpoint.Y)),
			core.NewTag(31, core.NewFloatValue(d.TextMidpoint.Z)),
		)
	}

	dimType := d.DimType & 0x7f
	tags = append(tags, core.NewTag(70, core.NewIntegerValue(dimType)))

	tags = append(tags,
		core.NewTag(13, core.NewFloatValue(d.DefPoint2.X)),
		core.NewTag(23, core.NewFloatValue(d.DefPoint2.Y)),
		core.NewTag(33, core.NewFloatValue(d.DefPoint2.Z)),
		core.NewTag(14, core.NewFloatValue(d.DefPoint3.X)),
		core.NewTag(24, core.NewFloatValue(d.DefPoint3.Y)),
		core.NewTag(34, core.NewFloatValue(d.DefPoint3.Z)),
	)

	return AppendXData(tags, &d.BaseEntity)
}

func (d *Dimension) generateR12GeomBlock(textBlockName string) (string, core.TagSlice) {
	seq := NextR12BlockSeq()
	blockName := "*D" + strconv.Itoa(seq)

	handle := nextR12HatchHandle()
	var tags core.TagSlice
	tags = append(tags,
		core.NewTag(0, core.NewStringValue("BLOCK")),
		core.NewTag(5, core.NewStringValue(handle)),
		core.NewTag(8, core.NewStringValue("0")),
		core.NewTag(2, core.NewStringValue(blockName)),
		core.NewTag(70, core.NewIntegerValue(1)),
		core.NewTag(10, core.NewFloatValue(0.0)),
		core.NewTag(20, core.NewFloatValue(0.0)),
		core.NewTag(30, core.NewFloatValue(0.0)),
		core.NewTag(3, core.NewStringValue(blockName)),
		core.NewTag(1, core.NewStringValue("")),
	)

	dimTypeBase := d.DimType & 15

	switch dimTypeBase {
	case DimLinear, DimAligned:
		tags = append(tags, d.r12LinearGeomEntities(textBlockName)...)
	case DimDiameter, DimRadius:
		tags = append(tags, d.r12RadialGeomEntities(dimTypeBase)...)
	default:
		tags = append(tags, d.r12LinearGeomEntities(textBlockName)...)
	}

	tags = append(tags,
		core.NewTag(0, core.NewStringValue("ENDBLK")),
		core.NewTag(5, core.NewStringValue(nextR12HatchHandle())),
		core.NewTag(8, core.NewStringValue("0")),
	)
	return blockName, tags
}

func (d *Dimension) r12LinearGeomEntities(textBlockName string) core.TagSlice {
	var tags core.TagSlice

	p1 := d.DefPoint2
	p2 := d.DefPoint3
	dp := d.DefPoint

	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	dimLen := math.Sqrt(dx*dx + dy*dy)
	if dimLen < 1e-9 {
		return AppendXData(tags, &d.BaseEntity)
	}
	ux := dx / dimLen
	uy := dy / dimLen

	perpX := -uy
	perpY := ux

	arrowSize := 0.18

	t1 := ux*(p1.X-dp.X) + uy*(p1.Y-dp.Y)
	d1X := dp.X + t1*ux
	d1Y := dp.Y + t1*uy

	t2 := ux*(p2.X-dp.X) + uy*(p2.Y-dp.Y)
	d2X := dp.X + t2*ux
	d2Y := dp.Y + t2*uy

	tags = append(tags, makeLineTagR12(p1.X, p1.Y, d1X, d1Y, "0")...)

	tags = append(tags, makeLineTagR12(p2.X, p2.Y, d2X, d2Y, "0")...)

	tags = append(tags, makeLineTagR12(d1X, d1Y, d2X, d2Y, "0")...)

	arrow1X := d1X + ux*arrowSize
	arrow1Y := d1Y + uy*arrowSize
	tags = append(tags, makeArrowSolid(d1X, d1Y, arrow1X, arrow1Y, arrowSize, perpX, perpY)...)

	arrow2X := d2X - ux*arrowSize
	arrow2Y := d2Y - uy*arrowSize
	tags = append(tags, makeArrowSolid(d2X, d2Y, arrow2X, arrow2Y, arrowSize, perpX, perpY)...)

	tags = append(tags, makePointTagR12(p1.X, p1.Y, "DEFPOINTS")...)
	tags = append(tags, makePointTagR12(p2.X, p2.Y, "DEFPOINTS")...)
	tags = append(tags, makePointTagR12(dp.X, dp.Y, "DEFPOINTS")...)

	tags = append(tags,
		core.NewTag(0, core.NewStringValue("INSERT")),
		core.NewTag(5, core.NewStringValue(nextR12HatchHandle())),
		core.NewTag(8, core.NewStringValue("0")),
		core.NewTag(2, core.NewStringValue(textBlockName)),
		core.NewTag(10, core.NewFloatValue(0.0)),
		core.NewTag(20, core.NewFloatValue(0.0)),
		core.NewTag(30, core.NewFloatValue(0.0)),
	)

	return AppendXData(tags, &d.BaseEntity)
}

func (d *Dimension) r12RadialGeomEntities(dimTypeBase int) core.TagSlice {
	var tags core.TagSlice
	center := d.DefPoint
	edge := d.DefPoint4

	tags = append(tags, makeLineTagR12(center.X, center.Y, edge.X, edge.Y, "0")...)

	dx := center.X - edge.X
	dy := center.Y - edge.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist > 1e-9 {
		ux := dx / dist
		uy := dy / dist
		perpX := -uy
		perpY := ux
		arrowSize := 0.18
		arrowTipX := edge.X + ux*arrowSize
		arrowTipY := edge.Y + uy*arrowSize
		tags = append(tags, makeArrowSolid(edge.X, edge.Y, arrowTipX, arrowTipY, arrowSize, perpX, perpY)...)
	}

	tags = append(tags, makePointTagR12(center.X, center.Y, "DEFPOINTS")...)
	tags = append(tags, makePointTagR12(edge.X, edge.Y, "DEFPOINTS")...)
	return AppendXData(tags, &d.BaseEntity)
}

func (d *Dimension) generateR12TextBlock() (string, core.TagSlice) {
	text := d.Text
	if text == "" {
		text = "<>"
	}

	textHeight := 0.18

	tx := d.TextMidpoint.X
	ty := d.TextMidpoint.Y
	if core.FloatEquals(tx, 0) && core.FloatEquals(ty, 0) {
		tx = d.DefPoint.X
		ty = d.DefPoint.Y
	}

	seq := NextR12BlockSeq()
	blockName := "*U" + strconv.Itoa(seq)

	handle := nextR12HatchHandle()
	var tags core.TagSlice
	tags = append(tags,
		core.NewTag(0, core.NewStringValue("BLOCK")),
		core.NewTag(5, core.NewStringValue(handle)),
		core.NewTag(8, core.NewStringValue("0")),
		core.NewTag(2, core.NewStringValue(blockName)),
		core.NewTag(70, core.NewIntegerValue(1)),
		core.NewTag(10, core.NewFloatValue(0.0)),
		core.NewTag(20, core.NewFloatValue(0.0)),
		core.NewTag(30, core.NewFloatValue(0.0)),
		core.NewTag(3, core.NewStringValue(blockName)),
		core.NewTag(1, core.NewStringValue("")),

		core.NewTag(0, core.NewStringValue("TEXT")),
		core.NewTag(5, core.NewStringValue(nextR12HatchHandle())),
		core.NewTag(8, core.NewStringValue("0")),
		core.NewTag(10, core.NewFloatValue(tx)),
		core.NewTag(20, core.NewFloatValue(ty)),
		core.NewTag(30, core.NewFloatValue(0.0)),
		core.NewTag(40, core.NewFloatValue(textHeight)),
		core.NewTag(1, core.NewStringValue(text)),
		core.NewTag(0, core.NewStringValue("ENDBLK")),
		core.NewTag(5, core.NewStringValue(nextR12HatchHandle())),
		core.NewTag(8, core.NewStringValue("0")),
	)
	return blockName, tags
}

func makeLineTagR12(x1, y1, x2, y2 float64, layer string) core.TagSlice {
	handle := nextR12HatchHandle()
	return core.TagSlice{
		core.NewTag(0, core.NewStringValue("LINE")),
		core.NewTag(5, core.NewStringValue(handle)),
		core.NewTag(8, core.NewStringValue(layer)),
		core.NewTag(62, core.NewIntegerValue(0)),
		core.NewTag(10, core.NewFloatValue(x1)),
		core.NewTag(20, core.NewFloatValue(y1)),
		core.NewTag(30, core.NewFloatValue(0.0)),
		core.NewTag(11, core.NewFloatValue(x2)),
		core.NewTag(21, core.NewFloatValue(y2)),
		core.NewTag(31, core.NewFloatValue(0.0)),
	}
}

func makeArrowSolid(tipX, tipY, baseX, baseY, size, perpX, perpY float64) core.TagSlice {
	halfW := size * 0.4
	handle := nextR12HatchHandle()
	return core.TagSlice{
		core.NewTag(0, core.NewStringValue("SOLID")),
		core.NewTag(5, core.NewStringValue(handle)),
		core.NewTag(8, core.NewStringValue("0")),
		core.NewTag(62, core.NewIntegerValue(0)),
		core.NewTag(10, core.NewFloatValue(tipX)),
		core.NewTag(20, core.NewFloatValue(tipY)),
		core.NewTag(30, core.NewFloatValue(0.0)),
		core.NewTag(11, core.NewFloatValue(baseX+perpX*halfW)),
		core.NewTag(21, core.NewFloatValue(baseY+perpY*halfW)),
		core.NewTag(31, core.NewFloatValue(0.0)),
		core.NewTag(12, core.NewFloatValue(baseX-perpX*halfW)),
		core.NewTag(22, core.NewFloatValue(baseY-perpY*halfW)),
		core.NewTag(32, core.NewFloatValue(0.0)),
		core.NewTag(13, core.NewFloatValue(baseX-perpX*halfW)),
		core.NewTag(23, core.NewFloatValue(baseY-perpY*halfW)),
		core.NewTag(33, core.NewFloatValue(0.0)),
	}
}

func makePointTagR12(x, y float64, layer string) core.TagSlice {
	handle := nextR12HatchHandle()
	return core.TagSlice{
		core.NewTag(0, core.NewStringValue("POINT")),
		core.NewTag(5, core.NewStringValue(handle)),
		core.NewTag(8, core.NewStringValue(layer)),
		core.NewTag(62, core.NewIntegerValue(0)),
		core.NewTag(10, core.NewFloatValue(x)),
		core.NewTag(20, core.NewFloatValue(y)),
		core.NewTag(30, core.NewFloatValue(0.0)),
	}
}

func dimPointFieldTags(code int, pt core.Point) core.TagSlice {
	return core.TagSlice{
		core.NewTag(code, core.NewFloatValue(pt.X)),
		core.NewTag(code + 10, core.NewFloatValue(pt.Y)),
		core.NewTag(code + 20, core.NewFloatValue(pt.Z)),
	}
}

func NewLinearDimEntity(base, p1, p2, location core.Point, layer string) *Dimension {
	return &Dimension{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		DimStyleName:       "Standard",
		DimType:            32,
		DefPoint:           base,
		TextMidpoint:       location,
		DefPoint2:          p1,
		DefPoint3:          p2,
		AttachmentPoint:    5,
		LineSpacingStyle:   1,
		ExtrusionDirection: core.Point{X: 0, Y: 0, Z: 1},
	}
}

func NewAlignedDimEntity(p1, p2, location core.Point, layer string) *Dimension {
	return &Dimension{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		DimStyleName:       "Standard",
		DimType:            32 + DimAligned,
		DefPoint:           location,
		TextMidpoint:       location,
		DefPoint2:          p1,
		DefPoint3:          p2,
		AttachmentPoint:    5,
		LineSpacingStyle:   1,
		ExtrusionDirection: core.Point{X: 0, Y: 0, Z: 1},
	}
}

func (d Dimension) Clone() Entity {
	n := NewLinearDimEntity(d.DefPoint, d.DefPoint2, d.DefPoint3, d.TextMidpoint, d.LayerName)
	n.BaseEntity = d.BaseEntity.CloneBase()
	n.DimType = d.DimType
	n.ActualMeasurement = d.ActualMeasurement
	n.Version = d.Version
	n.Geometry = d.Geometry
	n.DimStyleName = d.DimStyleName
	n.Insert = d.Insert
	n.AttachmentPoint = d.AttachmentPoint
	n.LineSpacingStyle = d.LineSpacingStyle
	n.LineSpacingFactor = d.LineSpacingFactor
	n.Unknown1 = d.Unknown1
	n.FlipArrow1 = d.FlipArrow1
	n.FlipArrow2 = d.FlipArrow2
	n.Text = d.Text
	n.ObliqueAngle = d.ObliqueAngle
	n.TextRotation = d.TextRotation
	n.HorizontalDirection = d.HorizontalDirection
	n.ExtrusionDirection = d.ExtrusionDirection
	n.DefPoint4 = d.DefPoint4
	n.DefPoint5 = d.DefPoint5
	n.Angle = d.Angle
	n.LeaderLen = d.LeaderLen
	return n
}
