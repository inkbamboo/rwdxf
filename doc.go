// Package rwdxf 是一个纯 Go 语言编写的 DXF（Drawing Exchange Format）文件读写库。
//
// 支持 R12 (AC1009) 和 AC1032 (AutoCAD 2018) 两种 DXF 版本，
// 提供完整的 DXF 文件解析、创建、修改和写入功能，
// 同时内置 3D 几何体线框生成辅助模块。
//
// 通过匿名导入汇聚所有子包，确保各包的 init() 注册生效：
//
//	import _ "github.com/inkbamboo/rwdxf"
package rwdxf

import (
	// 核心基础设施：Tag 系统、数据类型、解析器框架
	_ "github.com/inkbamboo/rwdxf/core"

	// 文档级别读写入口
	_ "github.com/inkbamboo/rwdxf/document"

	// 21 种 DXF 实体类型定义
	_ "github.com/inkbamboo/rwdxf/entities"

	// DXF 各段定义与序列化
	_ "github.com/inkbamboo/rwdxf/sections"
)
