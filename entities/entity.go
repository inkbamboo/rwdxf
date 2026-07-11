// Package entities 定义了所有 DXF 实体类型（共 21 种）。
//
// 核心设计:
//   - Entity 接口：所有实体必须实现的方法集合
//   - BaseEntity：所有实体的公共属性基类（Handle、Layer、Color、LineType 等）
//   - RegularEntity：普通实体 mixin，提供默认的非嵌套实体行为
//   - DxfParseable 策略解析：通过注册"组码→解析器"映射实现声明式 Tag 解析
//
// 全局标志：
//   - R12Mode：控制输出为 R12 格式（AC1009）还是 AC1032 格式
//   - UseGBK：写入时是否启用 GBK 编码（仅对非 ASCII 文本转码）
package entities

import (
	"io"
	"strconv"
	"strings"

	"github.com/inkbamboo/rwdxf/core"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// handleCounter 为实体自动分配 Handle 的全局计数器。
var handleCounter uint64 = 100000

// R12Mode 控制是否以 R12（AC1009）格式输出。
// 为 true 时跳过 handle/subclass marker 等 AC1032 特性。
var R12Mode = false

// UseGBK 控制写入时是否对非 ASCII 文本进行 GBK 编码。
var UseGBK = false

// ModelSpaceOwner 表示模型空间的 Owner handle，在写入 AC1032 时设置为 "13"。
var ModelSpaceOwner = "0"

// CloneBase 创建 BaseEntity 的浅拷贝，清除 Handle、Owner 和 XData，
// 用于实体 Clone 时的基础属性复制。
func (b BaseEntity) CloneBase() BaseEntity {
	c := b
	c.Handle = ""
	c.Owner = ""
	c.XData = nil
	if c.LineTypeName != "" && !core.IsStandardLineType(c.LineTypeName) {
		c.LineTypeName = ""
	}
	return c
}

func nextEntityHandle() string {
	h := strconv.FormatUint(handleCounter, 16)
	handleCounter++
	return h
}

// ResetEntityHandles 重置实体 Handle 计数器。
func ResetEntityHandles() {
	handleCounter = 100000
}

// HandleSeed 返回当前 Handle 计数器的十六进制字符串表示。
func HandleSeed() string {
	return strconv.FormatUint(handleCounter, 16)
}

// Entity 是所有 DXF 实体必须实现的接口。
// 定义了实体的类型标识、序列化、克隆和嵌套管理等核心能力。
type Entity interface {
	core.DxfElement
	IsSeqEnd() bool                 // 是否为序列结束标记
	HasNestedEntities() bool        // 是否有子实体（如 POLYLINE 包含 VERTEX）
	AddNestedEntities(entities EntitySlice) // 添加子实体

	Layer() string   // 获取所在图层名
	SetLayer(name string) // 设置图层名

	DxfTags() core.TagSlice // 序列化为 DXF Tag 列表

	IsR12Compatible() bool // 是否兼容 R12 格式

	Clone() Entity // 深拷贝

	DxfType() core.DxfTypeName // 返回 DXF 实体类型名
}

// EntitySlice 是一组 Entity 的切片，提供等值比较方法。
type EntitySlice []Entity

func (e EntitySlice) Equals(other EntitySlice) bool {
	if len(e) != len(other) {
		return false
	}
	for i, entity := range e {
		if !entity.Equals(other[i]) {
			return false
		}
	}
	return true
}

// Space 表示实体所在空间（模型空间或图纸空间）。
type Space int

const (
	MODEL Space = iota // 模型空间
	PAPER              // 图纸空间
)

// ShadowMode 表示实体的阴影模式。
type ShadowMode int

const (
	CASTS_AND_RECEIVE ShadowMode = iota // 投射并接收阴影
	CASTS                                // 仅投射阴影
	RECEIVES                             // 仅接收阴影
	IGNORES                              // 忽略阴影
)

// RegularEntity 是普通实体（非 SEQEND、非嵌套容器）的 mixin。
// 嵌入此类型即可获得默认的 IsSeqEnd=false、HasNestedEntities=false 等行为。
type RegularEntity struct{}

func (r RegularEntity) IsSeqEnd() bool                  { return false }
func (r RegularEntity) HasNestedEntities() bool         { return false }
func (r RegularEntity) AddNestedEntities(e EntitySlice) {}
func (r RegularEntity) IsR12Compatible() bool           { return true }

// BaseEntity 是所有实体的公共属性基类。
// 包含 Handle、Owner、Layer、Color、LineType 等通用 DXF 属性。
// 通过嵌入 DxfParseable 实现基于组码的策略解析。
type BaseEntity struct {
	core.DxfParseable
	Handle        string          // DXF 句柄（AC1032）
	Owner         string          // 所属 Owner 的 Handle
	Space         Space           // 模型空间/图纸空间
	LayoutTabName string          // 布局标签名
	LayerName     string          // 图层名
	LineTypeName  string          // 线型名
	On            bool            // 颜色是否为打开状态
	Color         int             // ACI 颜色索引
	LineWeight    int             // 线宽
	LineTypeScale float64         // 线型比例
	Visible       bool            // 可见性
	TrueColor     core.TrueColor  // 真彩色
	ColorName     string          // 颜色名
	Transparency  int             // 透明度
	ShadowMode    ShadowMode      // 阴影模式

	XData core.TagSlice // 扩展数据 (eXtended Data)
}

// Equals 判断两个 BaseEntity 的所有基础属性是否相等。
func (b BaseEntity) Equals(other BaseEntity) bool {
	return b.Handle == other.Handle &&
		b.Owner == other.Owner &&
		b.Space == other.Space &&
		b.LayoutTabName == other.LayoutTabName &&
		b.LayerName == other.LayerName &&
		b.LineTypeName == other.LineTypeName &&
		b.On == other.On &&
		b.Color == other.Color &&
		b.LineWeight == other.LineWeight &&
		core.FloatEquals(b.LineTypeScale, other.LineTypeScale) &&
		b.Visible == other.Visible &&
		b.ColorName == other.ColorName &&
		b.Transparency == other.Transparency &&
		b.ShadowMode == other.ShadowMode
}

// Layer 返回实体所在图层名。
func (b BaseEntity) Layer() string { return b.LayerName }

// SetLayer 设置实体所在图层名。
func (b *BaseEntity) SetLayer(name string) { b.LayerName = name }

// SetOwner 设置实体的 Owner Handle。
func (b *BaseEntity) SetOwner(id string) { b.Owner = id }

// SetEntityOwner 根据实体类型设置其 Owner Handle。
// 对于 Polyline 类型，同时将其所有 Vertex 的 Owner 设为 "1"。
func SetEntityOwner(e Entity, id string) {
	switch ent := e.(type) {
	case *Line:
		ent.Owner = id
	case *Circle:
		ent.Owner = id
	case *Arc:
		ent.Owner = id
	case *Text:
		ent.Owner = id
	case *PointEntity:
		ent.Owner = id
	case *LWPolyline:
		ent.Owner = id
	case *Polyline:
		ent.Owner = id
		for _, v := range ent.Vertices {
			if vertex, ok := v.(*Vertex); ok {
				vertex.Owner = "1"
			}
		}
	case *Insert:
		ent.Owner = id
	case *Ellipse:
		ent.Owner = id
	case *Spline:
		ent.Owner = id
	case *MText:
		ent.Owner = id
	case *Hatch:
		ent.Owner = id
	case *Ray:
		ent.Owner = id
	case *XLine:
		ent.Owner = id
	case *Solid:
		ent.Owner = id
	case *Trace:
		ent.Owner = id
	case *Face3D:
		ent.Owner = id
	case *Leader:
		ent.Owner = id
	case *MLine:
		ent.Owner = id
	case *Dimension:
		ent.Owner = id
	}
}

// InitBaseEntityParser 注册所有基础属性的组码解析器。
// 应在每个具体实体的构造函数中首先调用，再通过 Update 追加特定属性解析器。
func (b *BaseEntity) InitBaseEntityParser() {
	b.On = true
	b.Visible = true
	b.Init(map[int]core.TypeParser{
		5:   core.NewStringTypeParserToVar(&b.Handle),
		330: core.NewStringTypeParserToVar(&b.Owner),
		67: core.NewIntTypeParser(func(space int) {
			if space == 1 {
				b.Space = PAPER
			} else {
				b.Space = MODEL
			}
		}),
		410: core.NewStringTypeParserToVar(&b.LayoutTabName),
		8:   core.NewStringTypeParserToVar(&b.LayerName),
		6:   core.NewStringTypeParserToVar(&b.LineTypeName),
		62: core.NewIntTypeParser(func(color int) {
			if color < 0 {
				b.On = false
				b.Color = -color
			} else {
				b.Color = color
			}
		}),
		370: core.NewIntTypeParserToVar(&b.LineWeight),
		48:  core.NewFloatTypeParserToVar(&b.LineTypeScale),
		60: core.NewIntTypeParser(func(v int) {
			b.Visible = v == 0
		}),
		420: core.NewIntTypeParser(func(color int) {
			b.TrueColor.R = uint8((color >> 16) & 0xFF)
			b.TrueColor.G = uint8((color >> 8) & 0xFF)
			b.TrueColor.B = uint8(color & 0xFF)
		}),
		430: core.NewStringTypeParserToVar(&b.ColorName),
		440: core.NewIntTypeParserToVar(&b.Transparency),
		284: core.NewIntTypeParser(func(mode int) {
			b.ShadowMode = ShadowMode(mode)
		}),
	})
}

// baseEntityTags 生成实体的基础 Tag 列表（类型标识、Handle、Owner、图层等）。
func baseEntityTags(b *BaseEntity, entityType string) core.TagSlice {
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue(entityType)),
	}
	if !R12Mode {

		if b.Handle == "" {
			b.Handle = nextEntityHandle()
		}
		owner := b.Owner
		if owner == "" {
			owner = ModelSpaceOwner
		}
		tags = append(tags,
			core.NewTag(5, core.NewStringValue(b.Handle)),
			core.NewTag(330, core.NewStringValue(owner)),
			core.NewTag(100, core.NewStringValue("AcDbEntity")),
		)
	}

	layer := b.LayerName
	if layer == "" {
		layer = "0"
	}
	tags = append(tags, core.NewTag(8, core.NewStringValue(layer)))
	if b.LineTypeName != "" && b.LineTypeName != "BYLAYER" {
		tags = append(tags, core.NewTag(6, core.NewStringValue(b.LineTypeName)))
	}
	if b.Color != 0 {
		color := b.Color
		if !b.On {
			color = -color
		}
		tags = append(tags, core.NewTag(62, core.NewIntegerValue(color)))
	}
	if !core.FloatEquals(b.LineTypeScale, 0) {
		tags = append(tags, core.NewTag(48, core.NewFloatValue(b.LineTypeScale)))
	}

	return tags
}

// AppendXData 将 BaseEntity 的 XData 追加到 Tag 列表中。
// 在 R12 模式下会过滤掉 Hatch 相关的 XData。
func AppendXData(tags core.TagSlice, b *BaseEntity) core.TagSlice {
	if len(b.XData) > 0 {

		if !R12Mode && isHatchXData(b.XData) {
			return tags
		}
		tags = append(tags, b.XData...)
	}
	return tags
}

func isHatchXData(xdata core.TagSlice) bool {
	sawAcad := false
	for _, t := range xdata {
		if t.Code == 1001 {
			if v, ok := core.AsString(t.Value); ok {
				sawAcad = v == "ACAD"
			}
		}
		if sawAcad && t.Code == 1000 {
			if v, ok := core.AsString(t.Value); ok {
				if v == "HATCH" || v == "R14_HATCH_DATA" {
					return true
				}
			}
		}
	}
	return false
}

// getBaseEntity 通过类型断言获取任何实体的 *BaseEntity 指针。
func getBaseEntity(e Entity) *BaseEntity {
	switch ent := e.(type) {
	case *Line:
		return &ent.BaseEntity
	case *Circle:
		return &ent.BaseEntity
	case *Arc:
		return &ent.BaseEntity
	case *Text:
		return &ent.BaseEntity
	case *MText:
		return &ent.BaseEntity
	case *PointEntity:
		return &ent.BaseEntity
	case *Polyline:
		return &ent.BaseEntity
	case *Vertex:
		return &ent.BaseEntity
	case *SeqEnd:
		return &ent.BaseEntity
	case *Insert:
		return &ent.BaseEntity
	case *Ellipse:
		return &ent.BaseEntity
	case *Spline:
		return &ent.BaseEntity
	case *LWPolyline:
		return &ent.BaseEntity
	case *Solid:
		return &ent.BaseEntity
	case *Trace:
		return &ent.BaseEntity
	case *Face3D:
		return &ent.BaseEntity
	case *Hatch:
		return &ent.BaseEntity
	case *Dimension:
		return &ent.BaseEntity
	case *Leader:
		return &ent.BaseEntity
	case *MLine:
		return &ent.BaseEntity
	case *XLine:
		return &ent.BaseEntity
	case *Ray:
		return &ent.BaseEntity
	}
	return nil
}

// AppendEntityXData 将实体的 XData 追加到 Tag 列表中。
func AppendEntityXData(tags core.TagSlice, e Entity) core.TagSlice {
	b := getBaseEntity(e)
	if b != nil && len(b.XData) > 0 {
		tags = append(tags, b.XData...)
	}
	return tags
}

// CollectXDataFromTags 从 Tag 列表中提取所有 XData（组码 >= 1000）。
func CollectXDataFromTags(tags core.TagSlice) core.TagSlice {
	var xdata core.TagSlice
	for _, t := range tags {
		if t.Code >= 1000 {
			xdata = append(xdata, t)
		}
	}
	return xdata
}

// R12ExtraBlocks 存储在 R12 模式下自动生成的额外 Block 定义。
// 用于不支持 R12 的实体（如 MTEXT 多行→单行拆分）转换为 Block 引用。
var R12ExtraBlocks = make(map[string]core.TagSlice)

// ResetR12ExtraBlocks 重置 R12 额外 Block 缓存。
func ResetR12ExtraBlocks() {
	R12ExtraBlocks = make(map[string]core.TagSlice)
}

var nextR12BlockSeqVal = 0

// NextR12BlockSeq 返回下一个 R12 自动 Block 序号。
func NextR12BlockSeq() int {
	nextR12BlockSeqVal++
	return nextR12BlockSeqVal
}

// ResetR12BlockSeq 重置 R12 自动 Block 序号计数器。
func ResetR12BlockSeq() {
	nextR12BlockSeqVal = 0
}

func isDefaultExtrusion(p core.Point) bool {
	return core.FloatEquals(p.X, 0) && core.FloatEquals(p.Y, 0) && core.FloatEquals(p.Z, 1)
}

func pointToTags210(p core.Point) core.TagSlice {
	return core.TagSlice{
		core.NewTag(210, core.NewFloatValue(p.X)),
		core.NewTag(220, core.NewFloatValue(p.Y)),
		core.NewTag(230, core.NewFloatValue(p.Z)),
	}
}

// WriteTags 将 TagSlice 序列化为 DXF 文本字符串。
func WriteTags(tags core.TagSlice) string {
	var b strings.Builder
	writeTagsTo(&b, tags)
	return b.String()
}

// WriteTagsTo 将 TagSlice 序列化后写入 io.Writer。
func WriteTagsTo(w io.Writer, tags core.TagSlice) error {
	var b strings.Builder
	writeTagsTo(&b, tags)
	_, err := io.WriteString(w, b.String())
	return err
}

func writeTagsTo(b *strings.Builder, tags core.TagSlice) {
	for _, tag := range tags {
		b.WriteString(strconv.Itoa(tag.Code))
		b.WriteString("\r\n")
		v := tag.Value.ToString()
		if UseGBK && !isASCII(v) {
			encoded, _, err := transform.String(simplifiedchinese.GBK.NewEncoder(), v)
			if err != nil {
				b.WriteString(v)
			} else {
				b.WriteString(encoded)
			}
		} else {
			b.WriteString(v)
		}
		b.WriteString("\r\n")
	}
}

// WriteEntity 将实体序列化为 DXF 文本字符串。
func WriteEntity(e Entity) string {
	return WriteTags(e.DxfTags())
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}
