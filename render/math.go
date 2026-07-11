package render

import (
	"math"

	"github.com/inkbamboo/rwdxf/core"
)

func sub(a, b core.Point) core.Point {
	return core.Point{X: a.X - b.X, Y: a.Y - b.Y, Z: a.Z - b.Z}
}

func add(a, b core.Point) core.Point {
	return core.Point{X: a.X + b.X, Y: a.Y + b.Y, Z: a.Z + b.Z}
}

func scale(p core.Point, s float64) core.Point {
	return core.Point{X: p.X * s, Y: p.Y * s, Z: p.Z * s}
}

func length(p core.Point) float64 {
	return math.Sqrt(p.X*p.X + p.Y*p.Y + p.Z*p.Z)
}

func normalize(p core.Point) core.Point {
	l := length(p)
	if l < 1e-12 {
		return core.Point{X: 0, Y: 0, Z: 1}
	}
	return core.Point{X: p.X / l, Y: p.Y / l, Z: p.Z / l}
}

func cross(a, b core.Point) core.Point {
	return core.Point{
		X: a.Y*b.Z - a.Z*b.Y,
		Y: a.Z*b.X - a.X*b.Z,
		Z: a.X*b.Y - a.Y*b.X,
	}
}

func dot(a, b core.Point) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func rotationToZ(prevDir, forwardDir, position core.Point) [3][4]float64 {
	dir := normalize(core.Point{
		X: prevDir.X + forwardDir.X,
		Y: prevDir.Y + forwardDir.Y,
		Z: prevDir.Z + forwardDir.Z,
	})
	if dir.Z < 0 {
		dir = scale(dir, -1)
	}

	z := dir
	worldY := core.Point{X: 0, Y: 1, Z: 0}
	if math.Abs(z.Y) > 0.999 {
		worldY = core.Point{X: 1, Y: 0, Z: 0}
	}
	x := normalize(cross(worldY, z))
	y := cross(z, x)

	return [3][4]float64{
		{x.X, y.X, z.X, position.X},
		{x.Y, y.Y, z.Y, position.Y},
		{x.Z, y.Z, z.Z, position.Z},
	}
}

func transform(m [3][4]float64, p core.Point) core.Point {
	return core.Point{
		X: m[0][0]*p.X + m[0][1]*p.Y + m[0][2]*p.Z + m[0][3],
		Y: m[1][0]*p.X + m[1][1]*p.Y + m[1][2]*p.Z + m[1][3],
		Z: m[2][0]*p.X + m[2][1]*p.Y + m[2][2]*p.Z + m[2][3],
	}
}

// CircleProfile 生成圆形剖面点集。
func CircleProfile(radius float64, segments int) []core.Point {
	pts := make([]core.Point, segments)
	step := 2 * math.Pi / float64(segments)
	for i := 0; i < segments; i++ {
		a := float64(i) * step
		pts[i] = core.Point{X: radius * math.Cos(a), Y: radius * math.Sin(a), Z: 0}
	}
	return pts
}

// SquareProfile 生成正方形剖面点集。
func SquareProfile(size float64) []core.Point {
	h := size / 2
	return []core.Point{
		{X: -h, Y: -h}, {X: h, Y: -h}, {X: h, Y: h}, {X: -h, Y: h},
	}
}

// NgonProfile 生成正多边形剖面点集。
func NgonProfile(sides int, radius float64) []core.Point {
	if sides < 3 {
		sides = 3
	}
	pts := make([]core.Point, sides)
	step := 2 * math.Pi / float64(sides)
	for i := 0; i < sides; i++ {
		a := float64(i)*step - math.Pi/2
		pts[i] = core.Point{X: radius * math.Cos(a), Y: radius * math.Sin(a), Z: 0}
	}
	return pts
}

// StarProfile 生成星形剖面点集（交替使用外半径和内半径）。
func StarProfile(points int, outerRadius, innerRadius float64) []core.Point {
	n := points * 2
	pts := make([]core.Point, n)
	step := 2 * math.Pi / float64(n)
	for i := 0; i < n; i++ {
		a := float64(i)*step - math.Pi/2
		r := outerRadius
		if i%2 == 1 {
			r = innerRadius
		}
		pts[i] = core.Point{X: r * math.Cos(a), Y: r * math.Sin(a), Z: 0}
	}
	return pts
}
