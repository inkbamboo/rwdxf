package entities

import (
	"github.com/inkbamboo/rwdxf/core"
)

// Leader 表示 DXF LEADER 实体（引线标注）。
type Leader struct {
	RegularEntity
	BaseEntity
	DimStyleName           string
	HasArrowHead           bool
	PathType               int
	AnnotationType         int
	HooklineDirection      int
	HasHookline            bool
	TextHeight             float64
	TextWidth              float64
	Vertices               core.PointSlice
	BlockColor             int
	AnnotationHandle       string
	NormalVector           core.Point
	HorizontalDirection    core.Point
	LeaderOffsetBlockRef   core.Point
	LeaderOffsetAnnPlacement core.Point
	ExtrusionDirection     core.Point
}

func (l Leader) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Leader); ok {
		return l.BaseEntity.Equals(o.BaseEntity) &&
			l.DimStyleName == o.DimStyleName &&
			l.HasArrowHead == o.HasArrowHead &&
			l.PathType == o.PathType &&
			l.AnnotationType == o.AnnotationType &&
			l.Vertices.Equals(o.Vertices) &&
			l.NormalVector.Equals(o.NormalVector)
	}
	return false
}

func (l Leader) DxfType() core.DxfTypeName { return core.DxfTypeLeader }

func (l Leader) IsR12Compatible() bool { return false }

func NewLeader(tags core.TagSlice) (*Leader, error) {
	leader := new(Leader)
	leader.DimStyleName = "Standard"
	leader.HasArrowHead = true
	leader.AnnotationType = 3
	leader.HasHookline = true
	leader.TextHeight = 1.0
	leader.TextWidth = 1.0
	leader.NormalVector = core.Point{X: 0, Y: 0, Z: 1}
	leader.HorizontalDirection = core.Point{X: 1, Y: 0, Z: 0}
	leader.ExtrusionDirection = core.Point{X: 0, Y: 0, Z: 1}
	leader.InitBaseEntityParser()

	leader.Update(map[int]core.TypeParser{
		3:   core.NewStringTypeParserToVar(&leader.DimStyleName),
		71: core.NewIntTypeParser(func(v int) { leader.HasArrowHead = v != 0 }),
		72: core.NewIntTypeParserToVar(&leader.PathType),
		73: core.NewIntTypeParserToVar(&leader.AnnotationType),
		74: core.NewIntTypeParserToVar(&leader.HooklineDirection),
		75: core.NewIntTypeParser(func(v int) { leader.HasHookline = v != 0 }),
		40: core.NewFloatTypeParserToVar(&leader.TextHeight),
		41: core.NewFloatTypeParserToVar(&leader.TextWidth),
		77: core.NewIntTypeParserToVar(&leader.BlockColor),
		340: core.NewStringTypeParserToVar(&leader.AnnotationHandle),
		210: core.NewFloatTypeParserToVar(&leader.NormalVector.X),
		220: core.NewFloatTypeParserToVar(&leader.NormalVector.Y),
		230: core.NewFloatTypeParserToVar(&leader.NormalVector.Z),
		211: core.NewFloatTypeParserToVar(&leader.HorizontalDirection.X),
		221: core.NewFloatTypeParserToVar(&leader.HorizontalDirection.Y),
		231: core.NewFloatTypeParserToVar(&leader.HorizontalDirection.Z),
		212: core.NewFloatTypeParserToVar(&leader.LeaderOffsetBlockRef.X),
		222: core.NewFloatTypeParserToVar(&leader.LeaderOffsetBlockRef.Y),
		232: core.NewFloatTypeParserToVar(&leader.LeaderOffsetBlockRef.Z),
		213: core.NewFloatTypeParserToVar(&leader.LeaderOffsetAnnPlacement.X),
		223: core.NewFloatTypeParserToVar(&leader.LeaderOffsetAnnPlacement.Y),
		233: core.NewFloatTypeParserToVar(&leader.LeaderOffsetAnnPlacement.Z),
	})

	leader.Parse(tags)
	leader.XData = CollectXDataFromTags(tags)
	leader.parseVertices(tags)
	return leader, nil
}

func (l *Leader) parseVertices(tags core.TagSlice) {
	var current core.Point
	pts := make(core.PointSlice, 0)
	for _, tag := range tags.RegularTags() {
		switch tag.Code {
		case 10:
			if v, ok := core.AsFloat(tag.Value); ok {
				current = core.Point{X: v}
			}
		case 20:
			if v, ok := core.AsFloat(tag.Value); ok {
				current.Y = v
			}
		case 30:
			if v, ok := core.AsFloat(tag.Value); ok {
				current.Z = v
				pts = append(pts, current)
				current = core.Point{}
			}
		}
	}
	l.Vertices = pts
}

func (l *Leader) DxfTags() core.TagSlice {
	baseTags := baseEntityTags(&l.BaseEntity, "LEADER")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbLeader")))
	}

	arrowFlag := 0
	if l.HasArrowHead {
		arrowFlag = 1
	}
	hooklineFlag := 0
	if l.HasHookline {
		hooklineFlag = 1
	}

	tags := append(baseTags,
		core.NewTag(3, core.NewStringValue(l.DimStyleName)),
		core.NewTag(71, core.NewIntegerValue(arrowFlag)),
		core.NewTag(72, core.NewIntegerValue(l.PathType)),
		core.NewTag(73, core.NewIntegerValue(l.AnnotationType)),
		core.NewTag(74, core.NewIntegerValue(l.HooklineDirection)),
		core.NewTag(75, core.NewIntegerValue(hooklineFlag)),
		core.NewTag(40, core.NewFloatValue(l.TextHeight)),
		core.NewTag(41, core.NewFloatValue(l.TextWidth)),
	)

	for _, pt := range l.Vertices {
		tags = append(tags,
			core.NewTag(10, core.NewFloatValue(pt.X)),
			core.NewTag(20, core.NewFloatValue(pt.Y)),
			core.NewTag(30, core.NewFloatValue(pt.Z)),
		)
	}

	if l.BlockColor != 0 {
		tags = append(tags, core.NewTag(77, core.NewIntegerValue(l.BlockColor)))
	}
	if l.AnnotationHandle != "" {
		tags = append(tags, core.NewTag(340, core.NewStringValue(l.AnnotationHandle)))
	}

	norm := l.NormalVector
	if !core.FloatEquals(norm.X, 0) || !core.FloatEquals(norm.Y, 0) || !core.FloatEquals(norm.Z, 1) {
		tags = append(tags,
			core.NewTag(210, core.NewFloatValue(norm.X)),
			core.NewTag(220, core.NewFloatValue(norm.Y)),
			core.NewTag(230, core.NewFloatValue(norm.Z)),
		)
	}
	if !core.FloatEquals(l.HorizontalDirection.X, 1) || !core.FloatEquals(l.HorizontalDirection.Y, 0) || !core.FloatEquals(l.HorizontalDirection.Z, 0) {
		tags = append(tags,
			core.NewTag(211, core.NewFloatValue(l.HorizontalDirection.X)),
			core.NewTag(221, core.NewFloatValue(l.HorizontalDirection.Y)),
			core.NewTag(231, core.NewFloatValue(l.HorizontalDirection.Z)),
		)
	}

	if !core.FloatEquals(l.LeaderOffsetBlockRef.X, 0) || !core.FloatEquals(l.LeaderOffsetBlockRef.Y, 0) || !core.FloatEquals(l.LeaderOffsetBlockRef.Z, 0) {
		tags = append(tags,
			core.NewTag(212, core.NewFloatValue(l.LeaderOffsetBlockRef.X)),
			core.NewTag(222, core.NewFloatValue(l.LeaderOffsetBlockRef.Y)),
			core.NewTag(232, core.NewFloatValue(l.LeaderOffsetBlockRef.Z)),
		)
	}
	if !core.FloatEquals(l.LeaderOffsetAnnPlacement.X, 0) || !core.FloatEquals(l.LeaderOffsetAnnPlacement.Y, 0) || !core.FloatEquals(l.LeaderOffsetAnnPlacement.Z, 0) {
		tags = append(tags,
			core.NewTag(213, core.NewFloatValue(l.LeaderOffsetAnnPlacement.X)),
			core.NewTag(223, core.NewFloatValue(l.LeaderOffsetAnnPlacement.Y)),
			core.NewTag(233, core.NewFloatValue(l.LeaderOffsetAnnPlacement.Z)),
		)
	}

	return AppendXData(tags, &l.BaseEntity)
}

func NewLeaderEntity(vertices core.PointSlice, layer string) *Leader {
	return &Leader{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		DimStyleName:         "Standard",
		HasArrowHead:         true,
		AnnotationType:       3,
		HasHookline:          true,
		TextHeight:           1.0,
		TextWidth:            1.0,
		Vertices:             vertices,
		BlockColor:           0,
		NormalVector:         core.Point{X: 0, Y: 0, Z: 1},
		HorizontalDirection:  core.Point{X: 1, Y: 0, Z: 0},
		ExtrusionDirection:   core.Point{X: 0, Y: 0, Z: 1},
	}
}

func (l Leader) Clone() Entity {
	verts := make(core.PointSlice, len(l.Vertices))
	copy(verts, l.Vertices)
	n := NewLeaderEntity(verts, l.LayerName)
	n.BaseEntity = l.BaseEntity.CloneBase()
	n.DimStyleName = l.DimStyleName
	n.HasArrowHead = l.HasArrowHead
	n.PathType = l.PathType
	n.AnnotationType = l.AnnotationType
	n.HooklineDirection = l.HooklineDirection
	n.HasHookline = l.HasHookline
	n.TextHeight = l.TextHeight
	n.TextWidth = l.TextWidth
	n.BlockColor = l.BlockColor
	n.AnnotationHandle = l.AnnotationHandle
	n.NormalVector = l.NormalVector
	n.HorizontalDirection = l.HorizontalDirection
	n.LeaderOffsetBlockRef = l.LeaderOffsetBlockRef
	n.LeaderOffsetAnnPlacement = l.LeaderOffsetAnnPlacement
	return n
}
