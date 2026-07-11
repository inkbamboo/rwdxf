package entities

import "github.com/inkbamboo/rwdxf/core"

// SeqEnd 表示 DXF SEQEND 实体（序列结束标记），用于标记 POLYLINE 等实体的子实体序列结束。
type SeqEnd struct {
	RegularEntity
	BaseEntity
}

func (s SeqEnd) Equals(other core.DxfElement) bool {
	if o, ok := other.(*SeqEnd); ok {
		return s.BaseEntity.Equals(o.BaseEntity)
	}
	return false
}

func (s SeqEnd) IsSeqEnd() bool { return true }

func (s SeqEnd) DxfType() core.DxfTypeName { return core.DxfTypeSeqEnd }

// NewSeqEnd 从 TagSlice 解析并创建 SeqEnd 实体。
func NewSeqEnd(tags core.TagSlice) (*SeqEnd, error) {
	s := new(SeqEnd)
	s.InitBaseEntityParser()
	s.Parse(tags)
	s.XData = CollectXDataFromTags(tags)
	return s, nil
}
func (s *SeqEnd) DxfTags() core.TagSlice {
	tags := baseEntityTags(&s.BaseEntity, "SEQEND")
	if !R12Mode {
		tags = append(tags, core.NewTag(100, core.NewStringValue("AcDbSequenceEnd")))
	}
	return AppendXData(tags, &s.BaseEntity)
}

func (s SeqEnd) Clone() Entity { n:=&SeqEnd{}; n.BaseEntity=s.BaseEntity.CloneBase(); return n }
