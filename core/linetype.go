package core

// LineTypeName 表示 DXF 线型名称。
type LineTypeName string

// 标准线型名称常量。
const (
	LineTypeByLayer    LineTypeName = "BYLAYER"
	LineTypeByBlock    LineTypeName = "BYBLOCK"
	LineTypeContinuous LineTypeName = "CONTINUOUS"
	LineTypeCenter     LineTypeName = "CENTER"
	LineTypeCenter2    LineTypeName = "CENTER2"
	LineTypeCenterX2   LineTypeName = "CENTERX2"
	LineTypeDashDot    LineTypeName = "DASHDOT"
	LineTypeDashDot2   LineTypeName = "DASHDOT2"
	LineTypeDashDotX2  LineTypeName = "DASHDOTX2"
	LineTypeDashed     LineTypeName = "DASHED"
	LineTypeDashed2    LineTypeName = "DASHED2"
	LineTypeDashedX2   LineTypeName = "DASHEDX2"
	LineTypeDivide     LineTypeName = "DIVIDE"
	LineTypeDivide2    LineTypeName = "DIVIDE2"
	LineTypeDivideX2   LineTypeName = "DIVIDEX2"
	LineTypeDot        LineTypeName = "DOT"
	LineTypeDot2       LineTypeName = "DOT2"
	LineTypeDotX2      LineTypeName = "DOTX2"
	LineTypeHidden     LineTypeName = "HIDDEN"
	LineTypeHidden2    LineTypeName = "HIDDEN2"
	LineTypeHiddenX2   LineTypeName = "HIDDENX2"
	LineTypePhantom    LineTypeName = "PHANTOM"
	LineTypePhantom2   LineTypeName = "PHANTOM2"
	LineTypePhantomX2  LineTypeName = "PHANTOMX2"
	LineTypeBorder     LineTypeName = "BORDER"
	LineTypeBorder2    LineTypeName = "BORDER2"
	LineTypeBorderX2   LineTypeName = "BORDERX2"
)

// standardLineTypes 包含所有标准线型名称的集合。
var standardLineTypes = map[LineTypeName]bool{
	LineTypeByLayer:    true,
	LineTypeByBlock:    true,
	LineTypeContinuous: true,
	LineTypeCenter:     true,
	LineTypeCenter2:    true,
	LineTypeCenterX2:   true,
	LineTypeDashDot:    true,
	LineTypeDashDot2:   true,
	LineTypeDashDotX2:  true,
	LineTypeDashed:     true,
	LineTypeDashed2:    true,
	LineTypeDashedX2:   true,
	LineTypeDivide:     true,
	LineTypeDivide2:    true,
	LineTypeDivideX2:   true,
	LineTypeDot:        true,
	LineTypeDot2:       true,
	LineTypeDotX2:      true,
	LineTypeHidden:     true,
	LineTypeHidden2:    true,
	LineTypeHiddenX2:   true,
	LineTypePhantom:    true,
	LineTypePhantom2:   true,
	LineTypePhantomX2:  true,
	LineTypeBorder:     true,
	LineTypeBorder2:    true,
	LineTypeBorderX2:   true,
}

// IsStandardLineType 判断给定的线型名称是否为 AutoCAD 标准线型。
func IsStandardLineType(name string) bool {
	return standardLineTypes[LineTypeName(name)]
}
