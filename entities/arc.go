package entities

import (
	"math"

	"github.com/inkbamboo/rwdxf/core"
)

// Arc 表示 DXF ARC 实体。
type Arc struct {
	RegularEntity
	BaseEntity
	Thickness          float64    // 厚度
	Center             core.Point // 圆心
	Radius             float64    // 半径
	StartAngle         float64    // 起始角度（弧度）
	EndAngle           float64    // 终止角度（弧度）
	ExtrusionDirection core.Point // 拉伸方向
}

func (a Arc) DxfType() core.DxfTypeName { return core.DxfTypeArc }

func (a Arc) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Arc); ok {
		return a.BaseEntity.Equals(o.BaseEntity) &&
			core.FloatEquals(a.Thickness, o.Thickness) &&
			a.Center.Equals(o.Center) &&
			core.FloatEquals(a.Radius, o.Radius) &&
			core.FloatEquals(a.StartAngle, o.StartAngle) &&
			core.FloatEquals(a.EndAngle, o.EndAngle) &&
			a.ExtrusionDirection.Equals(o.ExtrusionDirection)
	}
	return false
}

func (a *Arc) StartPoint() core.Point {
	return core.Point{
		X: a.Center.X + a.Radius*math.Cos(a.StartAngle),
		Y: a.Center.Y + a.Radius*math.Sin(a.StartAngle),
		Z: a.Center.Z,
	}
}

func (a *Arc) EndPoint() core.Point {
	return core.Point{
		X: a.Center.X + a.Radius*math.Cos(a.EndAngle),
		Y: a.Center.Y + a.Radius*math.Sin(a.EndAngle),
		Z: a.Center.Z,
	}
}

func (a *Arc) TotalAngle() float64 {
	if a.EndAngle > a.StartAngle {
		return a.EndAngle - a.StartAngle
	}
	return a.EndAngle - a.StartAngle + 2*math.Pi
}

func (a *Arc) Flattening(segments int) []core.Point {
	if segments <= 0 {
		segments = 32
	}
	cx, cy, r := a.Center.X, a.Center.Y, a.Radius
	startAngle, endAngle := a.StartAngle, a.EndAngle
	if endAngle < startAngle {
		endAngle += 2 * math.Pi
	}
	pts := make([]core.Point, segments+1)
	for i := 0; i <= segments; i++ {
		angle := startAngle + (endAngle-startAngle)*float64(i)/float64(segments)
		pts[i] = core.Point{
			X: cx + r*math.Cos(angle),
			Y: cy + r*math.Sin(angle),
			Z: a.Center.Z,
		}
	}
	return pts
}

func (a *Arc) IsFullCircle() bool {
	start := math.Mod(a.StartAngle, 2*math.Pi)
	end := math.Mod(a.EndAngle, 2*math.Pi)
	return core.FloatEquals(start, end)
}

func (a *Arc) Translate(dx, dy, dz float64) {
	a.Center.X += dx
	a.Center.Y += dy
	a.Center.Z += dz
}

// NewArc 从 TagSlice 解析并创建 Arc 实体。
func NewArc(tags core.TagSlice) (*Arc, error) {
	arc := new(Arc)
	arc.ExtrusionDirection = core.Point{X: 0, Y: 0, Z: 1}
	arc.InitBaseEntityParser()
	arc.Update(map[int]core.TypeParser{
		39:  core.NewFloatTypeParserToVar(&arc.Thickness),
		10:  core.NewFloatTypeParserToVar(&arc.Center.X),
		20:  core.NewFloatTypeParserToVar(&arc.Center.Y),
		30:  core.NewFloatTypeParserToVar(&arc.Center.Z),
		40:  core.NewFloatTypeParserToVar(&arc.Radius),
		50:  core.NewFloatTypeParser(func(v float64) { arc.StartAngle = v * math.Pi / 180 }),
		51:  core.NewFloatTypeParser(func(v float64) { arc.EndAngle = v * math.Pi / 180 }),
		210: core.NewFloatTypeParserToVar(&arc.ExtrusionDirection.X),
		220: core.NewFloatTypeParserToVar(&arc.ExtrusionDirection.Y),
		230: core.NewFloatTypeParserToVar(&arc.ExtrusionDirection.Z),
	})
	arc.Parse(tags)
	arc.XData = CollectXDataFromTags(tags)
	return arc, nil
}
func (a *Arc) DxfTags() core.TagSlice {

	baseTags := baseEntityTags(&a.BaseEntity, "ARC")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbCircle")))
	}

	tags := append(baseTags,
		core.NewTag(10, core.NewFloatValue(a.Center.X)),
		core.NewTag(20, core.NewFloatValue(a.Center.Y)),
		core.NewTag(30, core.NewFloatValue(a.Center.Z)),
		core.NewTag(40, core.NewFloatValue(a.Radius)),
	)
	if !R12Mode {
		tags = append(tags, core.NewTag(100, core.NewStringValue("AcDbArc")))
	}

	tags = append(tags,
		core.NewTag(50, core.NewFloatValue(a.StartAngle*180/math.Pi)),
		core.NewTag(51, core.NewFloatValue(a.EndAngle*180/math.Pi)),
	)
	if !core.FloatEquals(a.Thickness, 0) {
		tags = append(tags, core.NewTag(39, core.NewFloatValue(a.Thickness)))
	}
	if !isDefaultExtrusion(a.ExtrusionDirection) {
		tags = append(tags, pointToTags210(a.ExtrusionDirection)...)
	}
	return AppendXData(tags, &a.BaseEntity)
}

// NewArcEntity 直接创建一个 Arc 实体。
func NewArcEntity(center core.Point, radius, startAngle, endAngle float64, layer string) *Arc {
	return &Arc{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		Center:             center,
		Radius:             radius,
		StartAngle:         startAngle,
		EndAngle:           endAngle,
		ExtrusionDirection: core.Point{X: 0, Y: 0, Z: 1},
	}
}

func (a Arc) Clone() Entity {
	n := NewArcEntity(a.Center, a.Radius, a.StartAngle, a.EndAngle, a.LayerName)
	n.BaseEntity = a.BaseEntity.CloneBase()
	n.Thickness = a.Thickness
	n.ExtrusionDirection = a.ExtrusionDirection
	return n
}
