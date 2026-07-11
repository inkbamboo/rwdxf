package entities

import (
	"github.com/inkbamboo/rwdxf/core"
)

// Insert 表示 DXF INSERT 实体（块引用）。
type Insert struct {
	BaseEntity
	Name               string
	Position           core.Point
	ScaleX             float64
	ScaleY             float64
	ScaleZ             float64
	Rotation           float64
	ColumnCount        int
	RowCount           int
	ColumnSpacing      float64
	RowSpacing         float64
	ExtrusionDirection core.Point
	Attribs            EntitySlice
}

func (i *Insert) IsSeqEnd() bool                  { return false }
func (i *Insert) HasNestedEntities() bool         { return true }
func (i *Insert) AddNestedEntities(e EntitySlice) { i.Attribs = e }

func (i Insert) DxfType() core.DxfTypeName { return core.DxfTypeInsert }

func (i *Insert) IsR12Compatible() bool { return true }

func (i Insert) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Insert); ok {
		return i.BaseEntity.Equals(o.BaseEntity) &&
			i.Name == o.Name &&
			i.Position.Equals(o.Position) &&
			core.FloatEquals(i.ScaleX, o.ScaleX) &&
			core.FloatEquals(i.ScaleY, o.ScaleY) &&
			core.FloatEquals(i.ScaleZ, o.ScaleZ) &&
			core.FloatEquals(i.Rotation, o.Rotation)
	}
	return false
}

func (i *Insert) Translate(dx, dy, dz float64) {
	i.Position.X += dx
	i.Position.Y += dy
	i.Position.Z += dz
}

func (i *Insert) AttributeCount() int { return len(i.Attribs) }

// NewInsert 从 TagSlice 解析并创建 Insert 实体。
func NewInsert(tags core.TagSlice) (*Insert, error) {
	ins := new(Insert)
	ins.ExtrusionDirection = core.Point{X: 0, Y: 0, Z: 1}
	ins.ScaleX = 1.0
	ins.ScaleY = 1.0
	ins.ScaleZ = 1.0
	ins.InitBaseEntityParser()
	ins.Update(map[int]core.TypeParser{
		2:   core.NewStringTypeParserToVar(&ins.Name),
		10:  core.NewFloatTypeParserToVar(&ins.Position.X),
		20:  core.NewFloatTypeParserToVar(&ins.Position.Y),
		30:  core.NewFloatTypeParserToVar(&ins.Position.Z),
		41:  core.NewFloatTypeParserToVar(&ins.ScaleX),
		42:  core.NewFloatTypeParserToVar(&ins.ScaleY),
		43:  core.NewFloatTypeParserToVar(&ins.ScaleZ),
		50:  core.NewFloatTypeParserToVar(&ins.Rotation),
		70:  core.NewIntTypeParserToVar(&ins.ColumnCount),
		71:  core.NewIntTypeParserToVar(&ins.RowCount),
		44:  core.NewFloatTypeParserToVar(&ins.ColumnSpacing),
		45:  core.NewFloatTypeParserToVar(&ins.RowSpacing),
		210: core.NewFloatTypeParserToVar(&ins.ExtrusionDirection.X),
		220: core.NewFloatTypeParserToVar(&ins.ExtrusionDirection.Y),
		230: core.NewFloatTypeParserToVar(&ins.ExtrusionDirection.Z),
	})
	ins.Parse(tags)
	ins.XData = CollectXDataFromTags(tags)
	return ins, nil
}
func (i *Insert) DxfTags() core.TagSlice {
	baseTags := baseEntityTags(&i.BaseEntity, "INSERT")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbBlockReference")))
	}
	tags := append(baseTags,
		core.NewTag(2, core.NewStringValue(i.Name)),
		core.NewTag(10, core.NewFloatValue(i.Position.X)),
		core.NewTag(20, core.NewFloatValue(i.Position.Y)),
		core.NewTag(30, core.NewFloatValue(i.Position.Z)),
	)
	if !core.FloatEquals(i.ScaleX, 1.0) {
		tags = append(tags, core.NewTag(41, core.NewFloatValue(i.ScaleX)))
	}
	if !core.FloatEquals(i.ScaleY, 1.0) {
		tags = append(tags, core.NewTag(42, core.NewFloatValue(i.ScaleY)))
	}
	if !core.FloatEquals(i.ScaleZ, 1.0) {
		tags = append(tags, core.NewTag(43, core.NewFloatValue(i.ScaleZ)))
	}
	if !core.FloatEquals(i.Rotation, 0) {
		tags = append(tags, core.NewTag(50, core.NewFloatValue(i.Rotation)))
	}
	if i.ColumnCount > 0 {
		tags = append(tags, core.NewTag(70, core.NewIntegerValue(i.ColumnCount)))
	}
	if i.RowCount > 0 {
		tags = append(tags, core.NewTag(71, core.NewIntegerValue(i.RowCount)))
	}
	if !core.FloatEquals(i.ColumnSpacing, 0) {
		tags = append(tags, core.NewTag(44, core.NewFloatValue(i.ColumnSpacing)))
	}
	if !core.FloatEquals(i.RowSpacing, 0) {
		tags = append(tags, core.NewTag(45, core.NewFloatValue(i.RowSpacing)))
	}
	if !isDefaultExtrusion(i.ExtrusionDirection) {
		tags = append(tags, pointToTags210(i.ExtrusionDirection)...)
	}

	for _, attr := range i.Attribs {
		if a, ok := attr.(*Text); ok {
			tags = append(tags, a.DxfTags()...)
		}
	}
	return AppendXData(tags, &i.BaseEntity)
}

// NewInsertEntity 直接创建一个 Insert 实体。
func NewInsertEntity(blockName string, pos core.Point, layer string) *Insert {
	return &Insert{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		Name:               blockName,
		Position:           pos,
		ScaleX:             1.0,
		ScaleY:             1.0,
		ScaleZ:             1.0,
		ExtrusionDirection: core.Point{X: 0, Y: 0, Z: 1},
	}
}

func (i Insert) Clone() Entity { n:=NewInsertEntity(i.Name,i.Position,i.LayerName); n.BaseEntity=i.BaseEntity.CloneBase(); n.ScaleX,n.ScaleY,n.ScaleZ=i.ScaleX,i.ScaleY,i.ScaleZ; n.Rotation=i.Rotation; n.ExtrusionDirection=i.ExtrusionDirection; for _,a:=range i.Attribs{n.Attribs=append(n.Attribs,a.Clone())}; return n }
