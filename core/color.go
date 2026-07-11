package core

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ACI 颜色索引常量。
const (
	ByBlock  = 0   // 随块
	ByLayer  = 256 // 随层
	ByObject = 257 // 随对象

	// 标准 ACI 颜色（1-9）
	Red     = 1
	Yellow  = 2
	Green   = 3
	Cyan    = 4
	Blue    = 5
	Magenta = 6
	Black   = 7
	White   = 7
	Gray    = 8
	LtGray  = 9
)

// TrueColor 表示 RGB 真彩色，每个通道范围为 0-255。
type TrueColor struct {
	R, G, B uint8
}

// ToInt 将真彩色编码为 24 位整数值。
func (c TrueColor) ToInt() int {
	return int(c.R)<<16 | int(c.G)<<8 | int(c.B)
}

// ToHex 返回 "#RRGGBB" 格式的十六进制颜色字符串。
func (c TrueColor) ToHex() string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

// ToFloats 将颜色分量转换为 [0, 1] 范围的浮点数。
func (c TrueColor) ToFloats() (r, g, b float64) {
	return float64(c.R) / 255, float64(c.G) / 255, float64(c.B) / 255
}

// Luminance 计算颜色的相对亮度（感知亮度）。
func (c TrueColor) Luminance() float64 {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	return math.Round(math.Sqrt(0.299*r*r+0.587*g*g+0.114*b*b)*1000) / 1000
}

// Equals 判断两个 TrueColor 是否完全相同。
func (c TrueColor) Equals(other TrueColor) bool {
	return c.R == other.R && c.G == other.G && c.B == other.B
}

// NewTrueColor 使用 RGB 分量创建 TrueColor。
func NewTrueColor(r, g, b uint8) TrueColor {
	return TrueColor{R: r, G: g, B: b}
}

// NewTrueColorFromInt 从 24 位整数值解析 TrueColor。
func NewTrueColorFromInt(v int) TrueColor {
	return TrueColor{
		R: uint8((v >> 16) & 0xFF),
		G: uint8((v >> 8) & 0xFF),
		B: uint8(v & 0xFF),
	}
}

// NewTrueColorFromHex 从 "#RRGGBB" 格式的十六进制字符串创建 TrueColor。
func NewTrueColorFromHex(hex string) (TrueColor, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return TrueColor{}, fmt.Errorf("invalid hex color: %q", hex)
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return TrueColor{R: uint8(r), G: uint8(g), B: uint8(b)}, nil
}

// NewTrueColorFromFloats 从 [0, 1] 范围的浮点数创建 TrueColor。
func NewTrueColorFromFloats(r, g, b float64) TrueColor {
	clamp := func(v float64) uint8 {
		return uint8(max(0, math.Min(255, math.Round(v*255))))
	}
	return TrueColor{R: clamp(r), G: clamp(g), B: clamp(b)}
}

// Int2RGB 将 24 位整数值解码为 R、G、B 三个分量。
func Int2RGB(value int) (r, g, b uint8) {
	return uint8((value >> 16) & 0xFF),
		uint8((value >> 8) & 0xFF),
		uint8(value & 0xFF)
}

// RGB2Int 将 R、G、B 分量编码为 24 位整数值。
func RGB2Int(r, g, b uint8) int {
	return (int(r) << 16) | (int(g) << 8) | int(b)
}

// Luminance 计算指定 RGB 分量的相对亮度。
func Luminance(r, g, b uint8) float64 {
	rf := float64(r) / 255
	gf := float64(g) / 255
	bf := float64(b) / 255
	return math.Round(math.Sqrt(0.299*rf*rf+0.587*gf*gf+0.114*bf*bf)*1000) / 1000
}

// ACI2RGB 将 ACI（AutoCAD Color Index）索引转换为 TrueColor。
// 返回的第二个值指示转换是否成功。
func ACI2RGB(index int) (TrueColor, bool) {
	if index < 1 || index >= len(dxfDefaultColors) {
		return TrueColor{}, false
	}
	return NewTrueColorFromInt(dxfDefaultColors[index]), true
}

// ACI2RGBInt 将 ACI 索引转换为 24 位整数值。
func ACI2RGBInt(index int) (int, bool) {
	if index < 1 || index >= len(dxfDefaultColors) {
		return 0, false
	}
	return dxfDefaultColors[index], true
}

// DXF 原始颜色编码中的颜色类型常量。
const (
	colorTypeByLayer  = 0xC0 // 随层
	colorTypeByBlock  = 0xC1 // 随块
	colorTypeRGB      = 0xC2 // RGB 真彩色
	colorTypeACI      = 0xC3 // ACI 索引色
	colorTypeWindowBG = 0xC8 // 窗口背景色
)

// DXF 原始颜色编码的预计算值。
const (
	ByLayerRawValue  = -1073741824 // 随层
	ByBlockRawValue  = -1056964608 // 随块
	WindowBGRawValue = -939524096  // 窗口背景色
)

// DecodeRawColor 解码 DXF 的原始颜色值。
// 返回颜色类型和对应的颜色值（ACI 索引或 RGB 整数值）。
func DecodeRawColor(value int) (colorType int, colorValue int) {
	flags := (value >> 24) & 0xFF
	switch flags {
	case colorTypeByLayer:
		return colorTypeByLayer, ByLayer
	case colorTypeByBlock:
		return colorTypeByBlock, ByBlock
	case colorTypeACI:
		return colorTypeACI, value & 0xFF
	case colorTypeRGB:
		return colorTypeRGB, value & 0xFFFFFF
	case colorTypeWindowBG:
		return colorTypeWindowBG, 0
	default:
		return 0, 0
	}
}

// EncodeRawColor 将 ACI 颜色和 TrueColor 编码为 DXF 原始颜色值。
func EncodeRawColor(aci int, tc TrueColor, hasTrueColor bool) int {
	if hasTrueColor {
		return -(^((colorTypeRGB<<24)+tc.ToInt()) + 1)
	}
	switch aci {
	case ByBlock:
		return ByBlockRawValue
	case ByLayer:
		return ByLayerRawValue
	default:
		if aci > 0 && aci < 256 {
			return -(^((colorTypeACI << 24) | aci) + 1)
		}
		return ByLayerRawValue
	}
}

// 透明度常量。
const (
	TransparencyByBlock = 0x01000000 // 随块
	Opaque              = 0x020000FF // 不透明
)

// Float2Transparency 将 [0, 1] 范围的浮点数转换为 DXF 透明度值。
// 0 表示不透明，1 表示完全透明。
func Float2Transparency(value float64) int {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	return int((1.0-value)*255) | 0x02000000
}

// Transparency2Float 将 DXF 透明度值转换为 [0, 1] 范围的浮点数。
func Transparency2Float(value int) float64 {
	return 1.0 - float64(value&0xFF)/255.0
}

// TrueColorSlice 是一组 TrueColor 的切片，提供等值比较方法。
type TrueColorSlice []TrueColor

func (s TrueColorSlice) Equals(other TrueColorSlice) bool {
	if len(s) != len(other) {
		return false
	}
	for i, c := range s {
		if !c.Equals(other[i]) {
			return false
		}
	}
	return true
}

// dxfDefaultColors 是 ACI 1-255 的默认 RGB 颜色表。
var dxfDefaultColors = []int{
	0x000000,
	0xFF0000, 0xFFFF00, 0x00FF00, 0x00FFFF, 0x0000FF, 0xFF00FF, 0xFFFFFF,
	0x808080, 0xC0C0C0,
	0xFF0000, 0xFF7F7F, 0xA50000, 0xA55252, 0x7F0000, 0x7F3F3F, 0x4C0000, 0x4C2626, 0x260000, 0x261313,
	0xFF3F00, 0xFF9F7F, 0xA52900, 0xA56752, 0x7F1F00, 0x7F4F3F, 0x4C1300, 0x4C2F26, 0x260900, 0x261713,
	0xFF7F00, 0xFFBF7F, 0xA55200, 0xA57C52, 0x7F3F00, 0x7F5F3F, 0x4C2600, 0x4C3926, 0x261300, 0x261C13,
	0xFFBF00, 0xFFDF7F, 0xA57C00, 0xA59152, 0x7F5F00, 0x7F6F3F, 0x4C3900, 0x4C4226, 0x261C00, 0x262113,
	0xFFFF00, 0xFFFF7F, 0xA5A500, 0xA5A552, 0x7F7F00, 0x7F7F3F, 0x4C4C00, 0x4C4C26, 0x262600, 0x262613,
	0xBFFF00, 0xDFFF7F, 0x7CA500, 0x91A552, 0x5F7F00, 0x6F7F3F, 0x394C00, 0x424C26, 0x1C2600, 0x212613,
	0x7FFF00, 0xBFFF7F, 0x52A500, 0x7CA552, 0x3F7F00, 0x5F7F3F, 0x264C00, 0x394C26, 0x132600, 0x1C2613,
	0x3FFF00, 0x9FFF7F, 0x29A500, 0x67A552, 0x1F7F00, 0x4F7F3F, 0x134C00, 0x2F4C26, 0x092600, 0x172613,
	0x00FF00, 0x7FFF7F, 0x00A500, 0x52A552, 0x007F00, 0x3F7F3F, 0x004C00, 0x264C26, 0x002600, 0x132613,
	0x00FF3F, 0x7FFF9F, 0x00A529, 0x52A567, 0x007F1F, 0x3F7F4F, 0x004C13, 0x264C2F, 0x002609, 0x135817,
	0x00FF7F, 0x7FFFBF, 0x00A552, 0x52A57C, 0x007F3F, 0x3F7F5F, 0x004C26, 0x264C39, 0x002613, 0x13581C,
	0x00FFBF, 0x7FFFDF, 0x00A57C, 0x52A591, 0x007F5F, 0x3F7F6F, 0x004C39, 0x264C42, 0x00261C, 0x135858,
	0x00FFFF, 0x7FFFFF, 0x00A5A5, 0x52A5A5, 0x007F7F, 0x3F7F7F, 0x004C4C, 0x264C4C, 0x002626, 0x135858,
	0x00BFFF, 0x7FDFFF, 0x007CA5, 0x5291A5, 0x005F7F, 0x3F6F7F, 0x00394C, 0x26427E, 0x001C26, 0x135858,
	0x007FFF, 0x7FBFFF, 0x0052A5, 0x527CA5, 0x003F7F, 0x3F5F7F, 0x00264C, 0x26397E, 0x001326, 0x131C58,
	0x003FFF, 0x7F9FFF, 0x0029A5, 0x5267A5, 0x001F7F, 0x3F4F7F, 0x00134C, 0x262F7E, 0x000926, 0x131758,
	0x0000FF, 0x7F7FFF, 0x0000A5, 0x5252A5, 0x00007F, 0x3F3F7F, 0x00004C, 0x26267E, 0x000026, 0x131358,
	0x3F00FF, 0x9F7FFF, 0x2900A5, 0x6752A5, 0x1F007F, 0x4F3F7F, 0x13004C, 0x2F267E, 0x090026, 0x171358,
	0x7F00FF, 0xBF7FFF, 0x5200A5, 0x7C52A5, 0x3F007F, 0x5F3F7F, 0x26004C, 0x39267E, 0x130026, 0x1C1358,
	0xBF00FF, 0xDF7FFF, 0x7C00A5, 0x9152A5, 0x5F007F, 0x6F3F7F, 0x39004C, 0x42264C, 0x1C0026, 0x581358,
	0xFF00FF, 0xFF7FFF, 0xA500A5, 0xA552A5, 0x7F007F, 0x7F3F7F, 0x4C004C, 0x4C264C, 0x260026, 0x581358,
	0xFF00BF, 0xFF7FDF, 0xA5007C, 0xA55291, 0x7F005F, 0x7F3F6F, 0x4C0039, 0x4C2642, 0x26001C, 0x581358,
	0xFF007F, 0xFF7FBF, 0xA50052, 0xA5527C, 0x7F003F, 0x7F3F5F, 0x4C0026, 0x4C2639, 0x260013, 0x58131C,
	0xFF003F, 0xFF7F9F, 0xA50029, 0xA55267, 0x7F001F, 0x7F3F4F, 0x4C0013, 0x4C262F, 0x260009, 0x581317,
	0x000000, 0x656565, 0x666666, 0x999999, 0xCCCCCC, 0xFFFFFF,
}

// IsValidACI 判断给定的 ACI 索引是否有效（1-255）。
func IsValidACI(aci int) bool {
	return aci >= 1 && aci < len(dxfDefaultColors)
}

// IsValidLayerACI 判断给定的 ACI 是否可作为图层颜色（1-255）。
func IsValidLayerACI(aci int) bool {
	return aci > 0 && aci < 256
}

// MaxACICount 返回 ACI 颜色表的最大索引值。
func MaxACICount() int { return len(dxfDefaultColors) - 1 }
