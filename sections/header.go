package sections

import "github.com/inkbamboo/rwdxf/core"

const tagACADVER = "$ACADVER"
const tagDWGCODEPAGE = "$DWGCODEPAGE"

// HeaderSection 表示 DXF HEADER 段，存储 $ACADVER、$DWGCODEPAGE 等 Header 变量。
type HeaderSection struct {
	Values map[string]core.TagSlice
}

func (s HeaderSection) Equals(other core.DxfElement) bool {
	if o, ok := other.(*HeaderSection); ok {
		if len(s.Values) != len(o.Values) {
			return false
		}
		for key, slice := range s.Values {
			if otherSlice, ok := o.Values[key]; ok {
				if !slice.Equals(otherSlice) {
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

// NewHeaderSection 从 TagSlice 解析 Header 段。
// 若缺少 $ACADVER 和 $DWGCODEPAGE，自动设置默认值。
func NewHeaderSection(tags core.TagSlice) *HeaderSection {
	header := new(HeaderSection)
	header.Values = make(map[string]core.TagSlice)

	if len(tags) > 3 {
		groups := core.TagGroups(tags[2:len(tags)-1], 9)
		for _, group := range groups {
			headerKey := group[0].Value.ToString()
			var groupTags core.TagSlice
			if keyTags, ok := header.Values[headerKey]; ok {
				groupTags = keyTags
			} else {
				groupTags = make(core.TagSlice, 0)
			}
			groupTags = append(groupTags, group[1:]...)
			header.Values[headerKey] = groupTags
		}
	}

	if _, ok := header.Values[tagACADVER]; !ok {
		header.Values[tagACADVER] = core.TagSlice{
			core.NewTag(1, core.NewStringValue("AC1009")),
		}
	}
	if _, ok := header.Values[tagDWGCODEPAGE]; !ok {
		header.Values[tagDWGCODEPAGE] = core.TagSlice{
			core.NewTag(3, core.NewStringValue("ANSI_1252")),
		}
	}
	return header
}

// Get 按变量名获取 Header 变量值。
func (s *HeaderSection) Get(key string) core.TagSlice {
	if keyTags, ok := s.Values[key]; ok {
		return keyTags
	}
	return core.TagSlice{}
}
