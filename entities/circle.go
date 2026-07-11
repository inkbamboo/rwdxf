package entities

import (
	"math"

	"github.com/inkbamboo/rwdxf/core"
)

// Circle 表示 DXF CIRCLE 实体。
type Circle struct {
	RegularEntity
	BaseEntity
	Thickness          float64    // 厚度
	Center             core.Point // 圆心
	Radius             float64    // 半径
	ExtrusionDirection core.Point // 拉伸方向
}

func (c Circle) DxfType() core.DxfTypeName { return core.DxfTypeCircle }

func (c Circle) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Circle); ok {
		return c.BaseEntity.Equals(o.BaseEntity) &&
			core.FloatEquals(c.Thickness, o.Thickness) &&
			c.Center.Equals(o.Center) &&
			core.FloatEquals(c.Radius, o.Radius) &&
			c.ExtrusionDirection.Equals(o.ExtrusionDirection)
	}
	return false
}

func (c *Circle) Diameter() float64 { return 2 * c.Radius }

func (c *Circle) Circumference() float64 { return 2 * math.Pi * c.Radius }

func (c *Circle) Area() float64 { return math.Pi * c.Radius * c.Radius }

func (c *Circle) Translate(dx, dy, dz float64) {
	c.Center.X += dx
	c.Center.Y += dy
	c.Center.Z += dz
}

func (c *Circle) Flattening(segments int) []core.Point {
	if segments <= 0 {
		segments = 64
	}
	cx, cy, r := c.Center.X, c.Center.Y, c.Radius
	pts := make([]core.Point, segments+1)
	for i := 0; i <= segments; i++ {
		angle := 2 * math.Pi * float64(i) / float64(segments)
		pts[i] = core.Point{
			X: cx + r*math.Cos(angle),
			Y: cy + r*math.Sin(angle),
			Z: c.Center.Z,
		}
	}
	return pts
}

func (c *Circle) Scale(s float64) {
	c.Radius *= s
	c.Center.X *= s
	c.Center.Y *= s
	c.Center.Z *= s
}

// NewCircle 从 TagSlice 解析并创建 Circle 实体。
func NewCircle(tags core.TagSlice) (*Circle, error) {
	circle := new(Circle)
	circle.ExtrusionDirection = core.Point{X: 0, Y: 0, Z: 1}
	circle.InitBaseEntityParser()
	circle.Update(map[int]core.TypeParser{
		39:  core.NewFloatTypeParserToVar(&circle.Thickness),
		10:  core.NewFloatTypeParserToVar(&circle.Center.X),
		20:  core.NewFloatTypeParserToVar(&circle.Center.Y),
		30:  core.NewFloatTypeParserToVar(&circle.Center.Z),
		40:  core.NewFloatTypeParserToVar(&circle.Radius),
		210: core.NewFloatTypeParserToVar(&circle.ExtrusionDirection.X),
		220: core.NewFloatTypeParserToVar(&circle.ExtrusionDirection.Y),
		230: core.NewFloatTypeParserToVar(&circle.ExtrusionDirection.Z),
	})
	circle.Parse(tags)
	circle.XData = CollectXDataFromTags(tags)
	return circle, nil
}
func (c *Circle) DxfTags() core.TagSlice {
	baseTags := baseEntityTags(&c.BaseEntity, "CIRCLE")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbCircle")))
	}
	tags := append(baseTags,
		core.NewTag(10, core.NewFloatValue(c.Center.X)),
		core.NewTag(20, core.NewFloatValue(c.Center.Y)),
		core.NewTag(30, core.NewFloatValue(c.Center.Z)),
		core.NewTag(40, core.NewFloatValue(c.Radius)),
	)
	if !core.FloatEquals(c.Thickness, 0) {
		tags = append(tags, core.NewTag(39, core.NewFloatValue(c.Thickness)))
	}
	if !isDefaultExtrusion(c.ExtrusionDirection) {
		tags = append(tags, pointToTags210(c.ExtrusionDirection)...)
	}
	return AppendXData(tags, &c.BaseEntity)
}

// NewCircleEntity 直接创建一个 Circle 实体。
func NewCircleEntity(center core.Point, radius float64, layer string) *Circle {
	return &Circle{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		Center:             center,
		Radius:             radius,
		ExtrusionDirection: core.Point{X: 0, Y: 0, Z: 1},
	}
}

func (c Circle) Clone() Entity {
	n := NewCircleEntity(c.Center, c.Radius, c.LayerName)
	n.BaseEntity = c.BaseEntity.CloneBase()
	n.Thickness = c.Thickness
	n.ExtrusionDirection = c.ExtrusionDirection
	return n
}
