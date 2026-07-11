package sections

import (
	"github.com/inkbamboo/rwdxf/core"
	"github.com/inkbamboo/rwdxf/entities"
)

// Block 表示一个 DXF 块定义，包含块名、基点、实体列表等。
type Block struct {
	core.DxfParseable
	Name         string
	Handle       string
	LayerName    string
	SecondName   string
	BasePoint    core.Point
	XrefPathName string
	Description  string
	Entities     entities.EntitySlice
}

func (b Block) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Block); ok {
		return b.Name == o.Name &&
			b.Handle == o.Handle &&
			b.LayerName == o.LayerName &&
			b.BasePoint.Equals(o.BasePoint) &&
			b.Entities.Equals(o.Entities)
	}
	return false
}

func NewBlock(tags core.TagSlice) (*Block, error) {
	block := new(Block)
	block.Init(map[int]core.TypeParser{
		1:  core.NewStringTypeParserToVar(&block.XrefPathName),
		2:  core.NewStringTypeParserToVar(&block.Name),
		3:  core.NewStringTypeParserToVar(&block.SecondName),
		4:  core.NewStringTypeParserToVar(&block.Description),
		5:  core.NewStringTypeParserToVar(&block.Handle),
		8:  core.NewStringTypeParserToVar(&block.LayerName),
		10: core.NewFloatTypeParserToVar(&block.BasePoint.X),
		20: core.NewFloatTypeParserToVar(&block.BasePoint.Y),
		30: core.NewFloatTypeParserToVar(&block.BasePoint.Z),
	})
	return block, block.Parse(tags)
}

// BlocksSection 表示 DXF BLOCKS 段，以块名索引 Block。
type BlocksSection map[string]*Block

func (b BlocksSection) Equals(other BlocksSection) bool {
	if len(b) != len(other) {
		return false
	}
	for i, block := range b {
		if !block.Equals(other[i]) {
			return false
		}
	}
	return true
}

// NewBlocksSection 从 TagSlice 解析 Blocks 段。
func NewBlocksSection(tags core.TagSlice) (BlocksSection, error) {
	blocks := make(BlocksSection)

	if len(tags) > 3 {
		groups := make([]core.TagSlice, 0)
		tagGroups := core.TagGroups(tags[2:len(tags)-1], 0)
		for _, group := range tagGroups {
			if group[0].Value.ToString() == "ENDBLK" {
				block, err := NewBlock(groups[0])
				if err != nil {
					return nil, err
				}
				allEntities, err := NewEntityList(groups[1:])
				if err != nil {
					return nil, err
				}
				block.Entities = allEntities
				blocks[block.Name] = block
				groups = make([]core.TagSlice, 0)
			} else {
				groups = append(groups, group)
			}
		}
	}
	return blocks, nil
}
