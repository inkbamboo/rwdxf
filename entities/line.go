package entities

import (
	"math"

	"github.com/inkbamboo/rwdxf/core"
)

// Line 表示 DXF LINE 实体，由起点和终点定义的一条直线段。
type Line struct {
	RegularEntity
	BaseEntity
	Thickness          float64    // 厚度
	Start              core.Point // 起点
	End                core.Point // 终点
	ExtrusionDirection core.Point // 拉伸方向
}

func (l Line) DxfType() core.DxfTypeName { return core.DxfTypeLine }

func (l Line) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Line); ok {
		return l.BaseEntity.Equals(o.BaseEntity) &&
			core.FloatEquals(l.Thickness, o.Thickness) &&
			l.Start.Equals(o.Start) &&
			l.End.Equals(o.End) &&
			l.ExtrusionDirection.Equals(o.ExtrusionDirection)
	}
	return false
}

func (l *Line) StartPoint() core.Point { return l.Start }

func (l *Line) EndPoint() core.Point { return l.End }

func (l *Line) Length() float64 {
	dx := l.End.X - l.Start.X
	dy := l.End.Y - l.Start.Y
	dz := l.End.Z - l.Start.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

func (l *Line) Delta() core.Point {
	return core.Point{
		X: l.End.X - l.Start.X,
		Y: l.End.Y - l.Start.Y,
		Z: l.End.Z - l.Start.Z,
	}
}

func (l *Line) GetPoints() []core.Point {
	return []core.Point{l.Start, l.End}
}

func (l *Line) MidPoint() core.Point {
	return core.Point{
		X: (l.Start.X + l.End.X) / 2,
		Y: (l.Start.Y + l.End.Y) / 2,
		Z: (l.Start.Z + l.End.Z) / 2,
	}
}

func (l *Line) Translate(dx, dy, dz float64) {
	l.Start.X += dx
	l.Start.Y += dy
	l.Start.Z += dz
	l.End.X += dx
	l.End.Y += dy
	l.End.Z += dz
}

func (l *Line) Scale(s float64) {
	l.Start.X *= s
	l.Start.Y *= s
	l.Start.Z *= s
	l.End.X *= s
	l.End.Y *= s
	l.End.Z *= s
}

// NewLine 从 TagSlice 解析并创建 Line 实体。
func NewLine(tags core.TagSlice) (*Line, error) {
	line := new(Line)
	line.ExtrusionDirection = core.Point{X: 0, Y: 0, Z: 1}
	line.InitBaseEntityParser()
	line.Update(map[int]core.TypeParser{
		39:  core.NewFloatTypeParserToVar(&line.Thickness),
		10:  core.NewFloatTypeParserToVar(&line.Start.X),
		20:  core.NewFloatTypeParserToVar(&line.Start.Y),
		30:  core.NewFloatTypeParserToVar(&line.Start.Z),
		11:  core.NewFloatTypeParserToVar(&line.End.X),
		21:  core.NewFloatTypeParserToVar(&line.End.Y),
		31:  core.NewFloatTypeParserToVar(&line.End.Z),
		210: core.NewFloatTypeParserToVar(&line.ExtrusionDirection.X),
		220: core.NewFloatTypeParserToVar(&line.ExtrusionDirection.Y),
		230: core.NewFloatTypeParserToVar(&line.ExtrusionDirection.Z),
	})
	line.Parse(tags)
	line.XData = CollectXDataFromTags(tags)
	return line, nil
}
func (l *Line) DxfTags() core.TagSlice {
	baseTags := baseEntityTags(&l.BaseEntity, "LINE")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbLine")))
	}
	tags := append(baseTags,
		core.NewTag(10, core.NewFloatValue(l.Start.X)),
		core.NewTag(20, core.NewFloatValue(l.Start.Y)),
		core.NewTag(30, core.NewFloatValue(l.Start.Z)),
		core.NewTag(11, core.NewFloatValue(l.End.X)),
		core.NewTag(21, core.NewFloatValue(l.End.Y)),
		core.NewTag(31, core.NewFloatValue(l.End.Z)),
	)
	if !core.FloatEquals(l.Thickness, 0) {
		tags = append(tags, core.NewTag(39, core.NewFloatValue(l.Thickness)))
	}
	if !isDefaultExtrusion(l.ExtrusionDirection) {
		tags = append(tags, pointToTags210(l.ExtrusionDirection)...)
	}
	return AppendXData(tags, &l.BaseEntity)
}

// NewLineEntity 直接创建一个 Line 实体。
func NewLineEntity(start, end core.Point, layer string) *Line {
	return &Line{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		Start:              start,
		End:                end,
		ExtrusionDirection: core.Point{X: 0, Y: 0, Z: 1},
	}
}

func (l Line) Clone() Entity { n := NewLineEntity(l.Start, l.End, l.LayerName); n.BaseEntity=l.BaseEntity.CloneBase(); n.Thickness=l.Thickness; n.ExtrusionDirection=l.ExtrusionDirection; return n }
