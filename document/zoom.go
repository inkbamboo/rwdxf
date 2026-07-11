package document

import (
	"math"

	"github.com/inkbamboo/rwdxf/core"
	"github.com/inkbamboo/rwdxf/entities"
)

// Modelspace 表示文档的模型空间视图，用于设置视口范围。
// 通过 Header 变量 $EXTMIN 和 $EXTMAX 控制显示范围。
type Modelspace struct {
	Entities   entities.EntitySlice
	headerVars map[string]core.TagSlice
}

// Extents 根据模型空间中所有实体的包围盒设置视口。
func Extents(msp *Modelspace) {
	ext := bbox(msp.Entities)
	if ext.hasData {
		center, size := ext.centerAndSize()
		msp.setViewport(guessHeight(size), center)
	}
}

// Objects 根据指定实体的包围盒设置视口。
func Objects(msp *Modelspace, ents []entities.Entity) {
	ext := bbox(ents)
	if ext.hasData {
		center, size := ext.centerAndSize()
		msp.setViewport(guessHeight(size), center)
	}
}

// Window 根据指定的两个角点设置视口范围。
func Window(msp *Modelspace, p1, p2 core.Point) {
	bb := boundingBox{
		minX:    math.Min(p1.X, p2.X),
		minY:    math.Min(p1.Y, p2.Y),
		maxX:    math.Max(p1.X, p2.X),
		maxY:    math.Max(p1.Y, p2.Y),
		hasData: true,
	}
	center, size := bb.centerAndSize()
	msp.setViewport(guessHeight(size), center)
}

// Center 将视口中心设置为指定点，大小由 size 参数控制。
func Center(msp *Modelspace, point, size core.Point) {
	msp.setViewport(guessHeight(size), point)
}

func (m *Modelspace) setViewport(height float64, center core.Point) {
	halfW := height * 1.6 / 2.0
	halfH := height / 2.0

	setHeaderVar(m.headerVars, "$EXTMIN", core.Point{X: center.X - halfW, Y: center.Y - halfH, Z: 0})
	setHeaderVar(m.headerVars, "$EXTMAX", core.Point{X: center.X + halfW, Y: center.Y + halfH, Z: 0})
}

func guessHeight(size core.Point) float64 {
	return math.Max(size.X/2.0, size.Y)
}

type boundingBox struct {
	minX, minY, maxX, maxY float64
	hasData                bool
}

func (b boundingBox) centerAndSize() (center, size core.Point) {
	return core.Point{X: b.minX + (b.maxX-b.minX)/2, Y: b.minY + (b.maxY-b.minY)/2},
		core.Point{X: b.maxX - b.minX, Y: b.maxY - b.minY}
}

func bbox(ents []entities.Entity) boundingBox {
	if len(ents) == 0 {
		return boundingBox{}
	}
	first := true
	ext := boundingBox{}
	for _, e := range ents {
		for _, pt := range entityPoints(e) {
			if first {
				ext.minX, ext.minY = pt.X, pt.Y
				ext.maxX, ext.maxY = pt.X, pt.Y
				ext.hasData = true
				first = false
				continue
			}
			if pt.X < ext.minX {
				ext.minX = pt.X
			}
			if pt.X > ext.maxX {
				ext.maxX = pt.X
			}
			if pt.Y < ext.minY {
				ext.minY = pt.Y
			}
			if pt.Y > ext.maxY {
				ext.maxY = pt.Y
			}
		}
	}
	return ext
}

func entityPoints(e entities.Entity) []core.Point {
	switch ent := e.(type) {
	case *entities.Line:
		return []core.Point{ent.Start, ent.End}
	case *entities.Circle:
		return []core.Point{
			{X: ent.Center.X - ent.Radius, Y: ent.Center.Y - ent.Radius},
			{X: ent.Center.X + ent.Radius, Y: ent.Center.Y + ent.Radius},
		}
	case *entities.Arc:
		return []core.Point{
			{X: ent.Center.X - ent.Radius, Y: ent.Center.Y - ent.Radius},
			{X: ent.Center.X + ent.Radius, Y: ent.Center.Y + ent.Radius},
		}
	case *entities.Ellipse:
		return []core.Point{
			{X: ent.Center.X - ent.MajorAxisEndPoint.X, Y: ent.Center.Y - ent.MajorAxisEndPoint.Y},
			{X: ent.Center.X + ent.MajorAxisEndPoint.X, Y: ent.Center.Y + ent.MajorAxisEndPoint.Y},
		}
	case *entities.Text:
		return []core.Point{ent.FirstAlignment}
	case *entities.MText:
		return []core.Point{ent.InsertionPoint}
	case *entities.PointEntity:
		return []core.Point{ent.Position}
	case *entities.LWPolyline:
		pts := make([]core.Point, len(ent.Points))
		for i, p := range ent.Points {
			pts[i] = p.Point
		}
		return pts
	case *entities.Polyline:
		pts := make([]core.Point, 0, len(ent.Vertices))
		for _, v := range ent.Vertices {
			if vertex, ok := v.(*entities.Vertex); ok {
				pts = append(pts, vertex.Location)
			}
		}
		return pts
	case *entities.Insert:
		return []core.Point{ent.Position}
	case *entities.Spline:
		pts := make([]core.Point, 0, len(ent.ControlPoints)+len(ent.FitPoints))
		pts = append(pts, ent.ControlPoints...)
		pts = append(pts, ent.FitPoints...)
		return pts
	}
	return nil
}

func setHeaderVar(vars map[string]core.TagSlice, name string, pt core.Point) {
	vars[name] = core.TagSlice{
		core.NewTag(10, core.NewFloatValue(pt.X)),
		core.NewTag(20, core.NewFloatValue(pt.Y)),
		core.NewTag(30, core.NewFloatValue(pt.Z)),
	}
}

