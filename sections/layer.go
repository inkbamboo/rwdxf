package sections

import "github.com/inkbamboo/rwdxf/core"

// 图层状态位掩码。
const (
	FrozenMask = 0x01
	ThawMask   = 0xFE
	LockMask   = 0x04
	UnlockMask = 0xFB
)

// 线宽特殊常量。
const (
	LineweightByLayer = -1 // 随层
	LineweightByBlock = -2 // 随块
	LineweightDefault = -3 // 默认线宽
)

// Layer 表示 DXF 图层定义，包含颜色、线型、冻结/锁定状态等属性。
type Layer struct {
	core.DxfParseable
	Name            string
	Color           int
	TrueColor       int
	LineType        string
	LineWeight      int
	Plot            bool
	PlotStyleHandle string
	MaterialHandle  string
	Locked          bool
	Frozen          bool
	On              bool
}

func (l Layer) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Layer); ok {
		return l.Name == o.Name &&
			l.Color == o.Color &&
			l.TrueColor == o.TrueColor &&
			l.LineType == o.LineType &&
			l.LineWeight == o.LineWeight &&
			l.Plot == o.Plot &&
			l.PlotStyleHandle == o.PlotStyleHandle &&
			l.MaterialHandle == o.MaterialHandle &&
			l.Locked == o.Locked &&
			l.Frozen == o.Frozen &&
			l.On == o.On
	}
	return false
}

func (l *Layer) IsFrozen() bool {
	return l.Frozen
}

func (l *Layer) Freeze() {
	l.Frozen = true
}

func (l *Layer) Thaw() {
	l.Frozen = false
}

func (l *Layer) IsLocked() bool {
	return l.Locked
}

func (l *Layer) Lock() {
	l.Locked = true
}

func (l *Layer) Unlock() {
	l.Locked = false
}

func (l *Layer) IsOff() bool {
	return !l.On
}

func (l *Layer) IsOn() bool {
	return l.On
}

func (l *Layer) TurnOn() {
	l.On = true
}

func (l *Layer) TurnOff() {
	l.On = false
}

func (l *Layer) GetColor() int {
	return l.Color
}

func (l *Layer) SetColor(color int) {
	l.Color = color
}

func IsValidLayerColor(aci int) bool {
	return aci > 0 && aci < 256
}

func FixLayerColor(aci int) int {
	if IsValidLayerColor(aci) {
		return aci
	}
	return 7
}

func IsValidLayerLineWeight(lw int) bool {
	return lw >= 0 && lw <= 211
}

func (l *Layer) SetRGB(r, g, b uint8) {
	l.TrueColor = int(r)<<16 | int(g)<<8 | int(b)
}

func (l *Layer) RGB() (r, g, b uint8, ok bool) {
	if l.TrueColor == 0 {
		return 0, 0, 0, false
	}
	return uint8((l.TrueColor >> 16) & 0xFF),
		uint8((l.TrueColor >> 8) & 0xFF),
		uint8(l.TrueColor & 0xFF),
		true
}

func (l *Layer) Flags() int {
	flags := 0
	if l.Frozen {
		flags |= FrozenMask
	}
	if l.Locked {
		flags |= LockMask
	}
	return flags
}

func (l *Layer) SetFlags(flags int) {
	l.Frozen = flags&FrozenMask != 0
	l.Locked = flags&LockMask != 0
}

func (l *Layer) Clone() *Layer {
	return &Layer{
		Name:            l.Name,
		Color:           l.Color,
		TrueColor:       l.TrueColor,
		LineType:        l.LineType,
		LineWeight:      l.LineWeight,
		Plot:            l.Plot,
		PlotStyleHandle: l.PlotStyleHandle,
		MaterialHandle:  l.MaterialHandle,
		Locked:          l.Locked,
		Frozen:          l.Frozen,
		On:              l.On,
	}
}

// NewLayer 从 TagSlice 解析 Layer。
func NewLayer(tags core.TagSlice) (*Layer, error) {
	layer := new(Layer)
	layer.On = true
	layer.Color = 7
	layer.Plot = true

	layer.Init(map[int]core.TypeParser{
		2: core.NewStringTypeParserToVar(&layer.Name),
		70: core.NewIntTypeParser(func(flags int) {
			layer.Frozen = flags&FrozenMask != 0
			layer.Locked = flags&LockMask != 0
		}),
		62: core.NewIntTypeParser(func(color int) {
			if color < 0 {
				layer.On = false
				layer.Color = -color
			} else {
				layer.Color = color
			}
		}),
		6:   core.NewStringTypeParserToVar(&layer.LineType),
		290: core.NewIntTypeParser(func(v int) { layer.Plot = v != 0 }),
		370: core.NewIntTypeParserToVar(&layer.LineWeight),
		390: core.NewStringTypeParserToVar(&layer.PlotStyleHandle),
		420: core.NewIntTypeParserToVar(&layer.TrueColor),
		347: core.NewStringTypeParserToVar(&layer.MaterialHandle),
	})
	return layer, layer.Parse(tags)
}

// NewLayerTable 从 TagSlice 解析图层表。
func NewLayerTable(tags core.TagSlice) (Table, error) {
	table := make(Table)
	tableSlices, err := TableEntryTags(tags)
	if err != nil {
		return table, err
	}
	for _, slice := range tableSlices {
		layer, err := NewLayer(slice)
		if err != nil {
			return nil, err
		}
		table[layer.Name] = layer
	}
	return table, nil
}
