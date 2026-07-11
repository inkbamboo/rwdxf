package entities

import (
	"github.com/inkbamboo/rwdxf/core"
)

// Ray 表示 DXF RAY 实体（射线，从一个点无限延伸）。
type Ray struct {
	RegularEntity
	BaseEntity
	Start      core.Point
	UnitVector core.Point
}

func (r Ray) Equals(other core.DxfElement) bool {
	if o, ok := other.(*Ray); ok {
		return r.BaseEntity.Equals(o.BaseEntity) &&
			r.Start.Equals(o.Start) &&
			r.UnitVector.Equals(o.UnitVector)
	}
	return false
}

func (r Ray) DxfType() core.DxfTypeName { return core.DxfTypeRay }

func (r Ray) IsR12Compatible() bool { return true }

func NewRay(tags core.TagSlice) (*Ray, error) {
	ray := new(Ray)
	ray.UnitVector = core.Point{X: 0, Y: 0, Z: 1}
	ray.InitBaseEntityParser()
	ray.Update(map[int]core.TypeParser{
		10: core.NewFloatTypeParserToVar(&ray.Start.X),
		20: core.NewFloatTypeParserToVar(&ray.Start.Y),
		30: core.NewFloatTypeParserToVar(&ray.Start.Z),
		11: core.NewFloatTypeParserToVar(&ray.UnitVector.X),
		21: core.NewFloatTypeParserToVar(&ray.UnitVector.Y),
		31: core.NewFloatTypeParserToVar(&ray.UnitVector.Z),
	})
	ray.Parse(tags)
	ray.XData = CollectXDataFromTags(tags)
	return ray, nil
}

func (r *Ray) DxfTags() core.TagSlice {
	if R12Mode {
		return r.dxfTagsR12()
	}
	baseTags := baseEntityTags(&r.BaseEntity, "RAY")
	if !R12Mode {
		baseTags = append(baseTags, core.NewTag(100, core.NewStringValue("AcDbRay")))
	}
	tags := append(baseTags,
		core.NewTag(10, core.NewFloatValue(r.Start.X)),
		core.NewTag(20, core.NewFloatValue(r.Start.Y)),
		core.NewTag(30, core.NewFloatValue(r.Start.Z)),
		core.NewTag(11, core.NewFloatValue(r.UnitVector.X)),
		core.NewTag(21, core.NewFloatValue(r.UnitVector.Y)),
		core.NewTag(31, core.NewFloatValue(r.UnitVector.Z)),
	)
	return AppendXData(tags, &r.BaseEntity)
}

func (r *Ray) dxfTagsR12() core.TagSlice {
	layerName := r.LayerName
	if layerName == "" {
		layerName = "0"
	}
	extent := 1e6
	ux, uy := r.UnitVector.X, r.UnitVector.Y
	sx, sy := r.Start.X, r.Start.Y
	return core.TagSlice{
		core.NewTag(0, core.NewStringValue("LINE")),
		core.NewTag(8, core.NewStringValue(layerName)),
		core.NewTag(10, core.NewFloatValue(sx)),
		core.NewTag(20, core.NewFloatValue(sy)),
		core.NewTag(30, core.NewFloatValue(0.0)),
		core.NewTag(11, core.NewFloatValue(sx + ux*extent)),
		core.NewTag(21, core.NewFloatValue(sy + uy*extent)),
		core.NewTag(31, core.NewFloatValue(0.0)),
	}
}

func NewRayEntity(start, unitVector core.Point, layer string) *Ray {
	return &Ray{
		BaseEntity: BaseEntity{
			LayerName:    layer,
			On:           true,
			Visible:      true,
			LineTypeName: "BYLAYER",
		},
		Start:      start,
		UnitVector: unitVector,
	}
}

func (r Ray) Clone() Entity {
	n := NewRayEntity(r.Start, r.UnitVector, r.LayerName)
	n.BaseEntity = r.BaseEntity.CloneBase()
	return n
}
