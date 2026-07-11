package sections

import "github.com/inkbamboo/rwdxf/core"

// Table 是一个通用的 DXF 表，以名称索引 DxfElement。
type Table map[string]core.DxfElement

func (t Table) Equals(other core.DxfElement) bool {
	if otherTable, ok := other.(Table); ok {
		if len(t) != len(otherTable) {
			return false
		}
		for key, element := range t {
			if otherElement, ok := otherTable[key]; ok {
				if !element.Equals(otherElement) {
					return false
				}
			} else {
				return false
			}
		}
		return true
	}
	return false
}

type TablesSection struct {
	Layers       Table
	Styles       Table
	LineTypes    Table
	BlockRecords map[string]string
}

func (t TablesSection) Equals(other core.DxfElement) bool {
	if o, ok := other.(*TablesSection); ok {
		return t.Layers.Equals(o.Layers) &&
			t.Styles.Equals(o.Styles) &&
			t.LineTypes.Equals(o.LineTypes)
	}
	return false
}

// NewTablesSection 从 TagSlice 解析 Tables 段。
func NewTablesSection(tags core.TagSlice) (*TablesSection, error) {
	tables := new(TablesSection)

	tableParsers := map[string]func(slice core.TagSlice) error{
		"LAYER": func(slice core.TagSlice) error {
			layerTables, err := NewLayerTable(slice)
			tables.Layers = layerTables
			return err
		},
		"STYLE": func(slice core.TagSlice) error {
			styleTables, err := NewStyleTable(slice)
			tables.Styles = styleTables
			return err
		},
		"LTYPE": func(slice core.TagSlice) error {
			lineTypeTables, err := NewLineTypeTable(slice)
			tables.LineTypes = lineTypeTables
			return err
		},
	}

	tags = tags[2:]
	stopTag := core.NewTag(0, core.NewStringValue("ENDSEC"))
	endOfChunk := core.NewTag(0, core.NewStringValue("ENDTAB"))

	for _, tableTags := range SplitTagChunks(tags, stopTag, endOfChunk) {
		entryTagsList, err := TableEntryTags(tableTags)
		if err != nil {
			return nil, err
		}
		for _, entryTags := range entryTagsList {
			tableType := entryTags[0].Value.ToString()
			if tableFactory, ok := tableParsers[tableType]; ok {
				if err := tableFactory(tableTags); err != nil {
					return nil, err
				}
			}
		}
	}
	return tables, nil
}
