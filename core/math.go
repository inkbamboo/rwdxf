package core

import "math"

const epsilon = 1e-9

// FloatEquals 使用 1e-9 的容差比较两个浮点数是否相等。
func FloatEquals(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}
