package sections

import (
	"github.com/inkbamboo/rwdxf/core"
	"github.com/inkbamboo/rwdxf/entities"
)

// ObjectsSection 表示 DXF OBJECTS 段。
type ObjectsSection struct {
	Objects []core.TagSlice
	Raw     string
}

// MinObjectsSectionTags 返回最小的 OBJECTS 段内容（包含必要的字典和多线样式）。
func MinObjectsSectionTags() core.TagSlice {
	return core.TagSlice{
		core.NewTag(0, core.NewStringValue("SECTION")),
		core.NewTag(2, core.NewStringValue("OBJECTS")),

		core.NewTag(0, core.NewStringValue("DICTIONARY")),
		core.NewTag(5, core.NewStringValue("C")),
		core.NewTag(330, core.NewStringValue("0")),
		core.NewTag(100, core.NewStringValue("AcDbDictionary")),
		core.NewTag(281, core.NewIntegerValue(1)),
		core.NewTag(3, core.NewStringValue("ACAD_PLOTSTYLENAME")),
		core.NewTag(350, core.NewStringValue("D")),

		core.NewTag(0, core.NewStringValue("ACDBDICTIONARYWDFLT")),
		core.NewTag(5, core.NewStringValue("D")),
		core.NewTag(330, core.NewStringValue("C")),
		core.NewTag(100, core.NewStringValue("AcDbDictionary")),
		core.NewTag(100, core.NewStringValue("AcDbDictionaryWithDefault")),
		core.NewTag(281, core.NewIntegerValue(1)),
		core.NewTag(3, core.NewStringValue("Normal")),

		core.NewTag(0, core.NewStringValue("MLINESTYLE")),
		core.NewTag(5, core.NewStringValue("1")),
		core.NewTag(102, core.NewStringValue("{ACAD_REACTORS")),
		core.NewTag(330, core.NewStringValue("0")),
		core.NewTag(102, core.NewStringValue("}")),
		core.NewTag(330, core.NewStringValue("0")),
		core.NewTag(100, core.NewStringValue("AcDbMlineStyle")),
		core.NewTag(2, core.NewStringValue("Standard")),
		core.NewTag(70, core.NewIntegerValue(0)),
		core.NewTag(3, core.NewStringValue("")),
		core.NewTag(62, core.NewIntegerValue(256)),
		core.NewTag(51, core.NewFloatValue(90.0)),
		core.NewTag(52, core.NewFloatValue(90.0)),
		core.NewTag(71, core.NewIntegerValue(2)),
		core.NewTag(49, core.NewFloatValue(0.5)),
		core.NewTag(62, core.NewIntegerValue(256)),
		core.NewTag(6, core.NewStringValue("BYLAYER")),
		core.NewTag(49, core.NewFloatValue(-0.5)),
		core.NewTag(62, core.NewIntegerValue(256)),
		core.NewTag(6, core.NewStringValue("BYLAYER")),
		core.NewTag(0, core.NewStringValue("ENDSEC")),
	}
}

func (o ObjectsSection) Equals(other core.DxfElement) bool {
	if o2, ok := other.(*ObjectsSection); ok {
		if len(o.Objects) != len(o2.Objects) {
			return false
		}
		for i, obj := range o.Objects {
			if !obj.Equals(o2.Objects[i]) {
				return false
			}
		}
		return true
	}
	return false
}

// NewObjectsSection 从 TagSlice 解析 Objects 段。
func NewObjectsSection(tags core.TagSlice) (*ObjectsSection, error) {
	section := new(ObjectsSection)
	if len(tags) <= 3 {
		return section, nil
	}

	innerTags := tags[2 : len(tags)-1]
	objects := core.TagGroups(innerTags, 0)
	section.Objects = objects
	return section, nil
}

func (o *ObjectsSection) DxfTags() core.TagSlice {
	if len(o.Objects) == 0 {
		return core.TagSlice{
			core.NewTag(0, core.NewStringValue("SECTION")),
			core.NewTag(2, core.NewStringValue("OBJECTS")),
			core.NewTag(0, core.NewStringValue("ENDSEC")),
		}
	}
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("SECTION")),
		core.NewTag(2, core.NewStringValue("OBJECTS")),
	}
	for _, obj := range o.Objects {
		tags = append(tags, obj...)
	}
	tags = append(tags, core.NewTag(0, core.NewStringValue("ENDSEC")))
	return tags
}

func WriteObjectsTags(o *ObjectsSection) string {
	return entities.WriteTags(o.DxfTags())
}
