# rwdxf

[![Go Version](https://img.shields.io/badge/Go-1.25.6-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

**rwdxf** 是一个纯 Go 语言编写的 DXF（Drawing Exchange Format）文件读写库，支持 R12 (AC1009) 和 AC1032 (2018) 两种 DXF 版本，并提供 3D 几何体线框生成功能。

## 特性

- 📄 **完整读写支持**：支持 DXF R12 和 AC1032 两种版本的读写
- 🧩 **21 种实体类型**：LINE, CIRCLE, ARC, TEXT, MTEXT, POINT, POLYLINE, LWPOLYLINE, VERTEX, SEQEND, INSERT, ELLIPSE, SPLINE, HATCH, RAY, XLINE, SOLID, TRACE, 3DFACE, LEADER, MLINE, DIMENSION
- 🌐 **编码兼容**：自动检测并支持 GBK 编码的 DXF 文件（中文 CAD 常用）
- 🎨 **颜色系统**：完整的 ACI 颜色表（1-255）、真彩色 (TrueColor) 支持
- 📐 **3D 几何生成**：独立 `render` 包提供立方体、圆柱、球体线框及 Sweep/Extrude 拉伸体生成
- 🔧 **策略解析模式**：声明式的 Tag → 字段映射解析器，易于扩展新实体
- 📋 **图层管理**：完整的图层创建、冻结/锁定、颜色设置等操作
- 🔍 **视口缩放**：支持 Extents、Window、Center、Objects 等多种缩放方式

## 安装

```bash
go get github.com/inkbamboo/rwdxf
```

## 快速开始

### 读取 DXF 文件

```go
package main

import (
    "os"
    "github.com/inkbamboo/rwdxf/document"
)

func main() {
    f, _ := os.Open("example.dxf")
    defer f.Close()

    doc, err := document.DocumentFromStream(f)
    if err != nil {
        panic(err)
    }

    // 遍历所有实体
    for _, entity := range doc.AllEntities() {
        println(entity.DxfType())
    }
}
```

### 创建并写入 DXF 文件

```go
package main

import (
    "os"
    "github.com/inkbamboo/rwdxf/core"
    "github.com/inkbamboo/rwdxf/document"
    "github.com/inkbamboo/rwdxf/entities"
)

func main() {
    doc := document.NewDocument()
    doc.SetVersion(document.VersionAC1032)

    // 添加一条线
    line := entities.NewLineEntity(
        core.Point{X: 0, Y: 0, Z: 0},
        core.Point{X: 10, Y: 10, Z: 0},
        "0",
    )
    doc.AddEntity(line)

    // 添加一个圆
    circle := entities.NewCircleEntity(
        core.Point{X: 50, Y: 50, Z: 0},
        25,
        "0",
    )
    doc.AddEntity(circle)

    // 写入文件
    f, _ := os.Create("output.dxf")
    defer f.Close()
    doc.Write(f)
}
```

### 3D 几何体生成

```go
package main

import (
    "github.com/inkbamboo/rwdxf/core"
    "github.com/inkbamboo/rwdxf/entities"
    "github.com/inkbamboo/rwdxf/render"
)

func main() {
    // 生成一个边长为 100 的立方体线框
    cube := render.CubeLines(core.Point{X: 0, Y: 0, Z: 0}, 100, "0")

    // 生成一个圆柱体线框
    cylinder := render.CylinderLines(
        core.Point{X: 200, Y: 0, Z: 0},
        50, 100, 32, "0",
    )

    // 将所有实体加入文档
    doc := document.NewDocument()
    for _, e := range cube {
        doc.AddEntity(e)
    }
    for _, e := range cylinder {
        doc.AddEntity(e)
    }
}
```

## 项目结构

```
rwdxf/
├── core/           # 核心基础设施：Tag 系统、数据类型、Point、颜色、解析器框架
├── entities/       # 21 种 DXF 实体类型定义（Entity 接口 + 具体实现）
├── sections/       # DXF 段定义：HEADER/TABLES/CLASSES/BLOCKS/ENTITIES/OBJECTS
├── document/       # 文档级别：DXF 文档解析、写入、版本切换、视口缩放
├── render/         # 3D 几何辅助：向量运算、剖面生成、几何体线框构建
└── doc.go          # 根包入口，匿名导入所有子包
```

## 支持的 DXF 版本

| 版本 | ACADVER 值 | 说明 |
|------|-----------|------|
| R12 | AC1009 | 经典 DXF 格式，兼容性最广 |
| AC1032 | AC1032 | AutoCAD 2018 格式 |

非 R12 兼容的实体（如 ELLIPSE、SPLINE 的某些变体）在 R12 模式下会自动转为 Block 表示。

## 核心概念

### Tag 系统

DXF 文件本质上是"组码-值"对的序列。`core.Tag` 是其最小单元：

```go
type Tag struct {
    Code  int       // DXF 组码
    Value DataType  // 值（String/Integer/Float 三种类型）
}
```

### 实体接口

所有实体都实现 `entities.Entity` 接口：

```go
type Entity interface {
    DxfType() core.DxfTypeName    // 实体类型名
    DxfTags() core.TagSlice       // 序列化为 DXF tags
    Clone() Entity                // 深拷贝
    IsR12Compatible() bool        // 是否兼容 R12
    // ...
}
```

### 解析器框架

基于 `core.DxfParseable` 的策略解析模式，通过注册"组码→解析器"映射实现声明式解析：

```go
// 初始化基础属性解析器
baseEntity.InitBaseEntityParser()

// 追加特定属性解析器
baseEntity.Update(map[int]core.TypeParser{
    10: core.NewFloatTypeParserToVar(&start.X),  // X 坐标
    20: core.NewFloatTypeParserToVar(&start.Y),  // Y 坐标
})
```

## 许可证

[MIT License](LICENSE)
