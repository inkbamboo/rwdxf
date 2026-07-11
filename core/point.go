package core

// Point 表示 3D 空间中的一个点，包含 X、Y、Z 三个坐标分量。
type Point struct {
	X float64
	Y float64
	Z float64
}

// Equals 判断两个点的坐标是否相等（使用 FloatEquals 容差比较）。
func (p Point) Equals(other Point) bool {
	return FloatEquals(p.X, other.X) &&
		FloatEquals(p.Y, other.Y) &&
		FloatEquals(p.Z, other.Z)
}

// PointSlice 表示一组点的切片，提供等值比较方法。
type PointSlice []Point

// Equals 判断两个点切片是否逐个相等。
func (p PointSlice) Equals(other PointSlice) bool {
	if len(p) != len(other) {
		return false
	}
	for i, pt := range p {
		if !pt.Equals(other[i]) {
			return false
		}
	}
	return true
}
