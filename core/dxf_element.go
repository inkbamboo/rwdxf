// Package core 提供 rwdxf 库的基础设施类型和解析框架。
//
// 核心组件包括：
//   - Tag 系统：DXF 文件的"组码-值"对表示与扫描器
//   - DataType 接口：统一 String/Integer/Float 三种数据类型的处理
//   - DxfParseable 解析器框架：基于策略模式的声明式 Tag 解析
//   - Point：3D 点结构
//   - TrueColor / ACI 颜色表：完整的 DXF 颜色系统
//   - 线型常量与浮点比较工具
package core

// DxfElement 是 DXF 元素的统一接口，要求实现 Equals 方法进行相等性比较。
// Tag、TagSlice、各实体类型和段类型均实现此接口。
type DxfElement interface {
	Equals(other DxfElement) bool
}
