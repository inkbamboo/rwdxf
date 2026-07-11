// Package render 提供 3D 几何体线框生成和向量运算辅助。
//
// 包括：
//   - 基础几何体：立方体 (CubeLines)、圆柱体 (CylinderLines)、球体 (SphereLines)
//   - 拉伸体：ExtrudeLinear（线性拉伸）、Sweep（沿路径扫掠）
//   - 剖面生成：圆形 (CircleProfile)、方形 (SquareProfile)、多边形 (NgonProfile)、星形 (StarProfile)
//   - 向量运算：加减、归一化、叉积、点积、旋转矩阵
package render

import (
	"math"

	"github.com/inkbamboo/rwdxf/core"
	"github.com/inkbamboo/rwdxf/entities"
)

// SweepProfile 沿路径扫掠一个剖面，返回每个路径点上的旋转后剖面。
func SweepProfile(profile []core.Point, path []core.Point) [][]core.Point {
	if len(path) < 2 || len(profile) == 0 {
		return nil
	}
	n := len(path)
	profiles := make([][]core.Point, n)

	for i, pt := range path {
		var forward core.Point
		if i < n-1 {
			forward = normalize(sub(path[i+1], pt))
		} else {
			forward = normalize(sub(pt, path[i-1]))
		}
		var prev core.Point
		if i > 0 {
			prev = normalize(sub(pt, path[i-1]))
		} else {
			prev = forward
		}

		profiles[i] = make([]core.Point, len(profile))
		rot := rotationToZ(prev, forward, pt)
		for j, p := range profile {
			profiles[i][j] = transform(rot, p)
		}
	}
	return profiles
}

// CubeLines 生成一个立方体的 12 条边线框。
func CubeLines(origin core.Point, size float64, layer string) []entities.Entity {
	x, y, z := origin.X, origin.Y, origin.Z
	s := size
	v := [8]core.Point{
		{X: x, Y: y, Z: z}, {X: x + s, Y: y, Z: z},
		{X: x + s, Y: y + s, Z: z}, {X: x, Y: y + s, Z: z},
		{X: x, Y: y, Z: z + s}, {X: x + s, Y: y, Z: z + s},
		{X: x + s, Y: y + s, Z: z + s}, {X: x, Y: y + s, Z: z + s},
	}
	edges := [12][2]int{
		{0, 1}, {1, 2}, {2, 3}, {3, 0},
		{4, 5}, {5, 6}, {6, 7}, {7, 4},
		{0, 4}, {1, 5}, {2, 6}, {3, 7},
	}
	result := make([]entities.Entity, 12)
	for i, e := range edges {
		result[i] = entities.NewLineEntity(v[e[0]], v[e[1]], layer)
	}
	return result
}

// CylinderLines 生成一个圆柱体的线框（两个圆底 + 侧边竖线）。
func CylinderLines(origin core.Point, radius, height float64, segments int, layer string) []entities.Entity {
	topZ := origin.Z + height
	bottomC := entities.NewCircleEntity(origin, radius, layer)
	topC := entities.NewCircleEntity(core.Point{X: origin.X, Y: origin.Y, Z: topZ}, radius, layer)

	result := []entities.Entity{bottomC, topC}
	step := 2 * math.Pi / float64(segments)
	for i := 0; i < segments; i++ {
		a := float64(i) * step
		cx := origin.X + radius*math.Cos(a)
		cy := origin.Y + radius*math.Sin(a)
		result = append(result,
			entities.NewLineEntity(core.Point{X: cx, Y: cy, Z: origin.Z}, core.Point{X: cx, Y: cy, Z: topZ}, layer),
		)
	}
	return result
}

// SphereLines 生成一个球体的代表性线框（XY/XZ/YZ 三个大圆）。
func SphereLines(center core.Point, radius float64, segments int, layer string) []entities.Entity {
	cXY := entities.NewCircleEntity(center, radius, layer)
	cXZ := entities.NewCircleEntity(center, radius, layer)
	cYZ := entities.NewCircleEntity(center, radius, layer)
	return []entities.Entity{cXY, cXZ, cYZ}
}

// SweepResult 包含扫掠体的剖面线和连接边线。
type SweepResult struct {
	Profiles []entities.Entity // 剖面线
	Edges    []entities.Entity // 连接边线
}

// Sweep 沿路径扫掠剖面，返回剖面线和连接边线。
func Sweep(profile []core.Point, path []core.Point, profileLayer, edgeLayer string) SweepResult {
	profiles := SweepProfile(profile, path)
	if len(profiles) < 2 {
		return SweepResult{}
	}
	n := len(profiles[0])

	var result SweepResult

	for _, sec := range profiles {
		for j := 0; j < n; j++ {
			next := (j + 1) % n
			result.Profiles = append(result.Profiles,
				entities.NewLineEntity(sec[j], sec[next], profileLayer))
		}
	}

	for i := 0; i < len(profiles)-1; i++ {
		for j := 0; j < n; j++ {
			result.Edges = append(result.Edges,
				entities.NewLineEntity(profiles[i][j], profiles[i+1][j], edgeLayer))
		}
	}

	return result
}

// ExtrudeLinear 沿指定方向线性拉伸一个剖面。
func ExtrudeLinear(profile []core.Point, direction core.Point, distance float64, profileLayer, edgeLayer string) SweepResult {
	dir := normalize(direction)
	end := scale(dir, distance)
	path := []core.Point{{X: 0, Y: 0, Z: 0}, end}
	return Sweep(profile, path, profileLayer, edgeLayer)
}
