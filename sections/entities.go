package sections

import (
	"github.com/inkbamboo/rwdxf/core"
	"github.com/inkbamboo/rwdxf/entities"
)

// EntitiesSection 表示 DXF ENTITIES 段，包含文档中的所有实体。
type EntitiesSection struct {
	Entities entities.EntitySlice
}

func (e EntitiesSection) Equals(other core.DxfElement) bool {
	if o, ok := other.(*EntitiesSection); ok {
		return e.Entities.Equals(o.Entities)
	}
	return false
}

// NewEntitiesSection 从 TagSlice 解析 Entities 段。
func NewEntitiesSection(tags core.TagSlice) (*EntitiesSection, error) {
	section := new(EntitiesSection)
	if len(tags) == 3 {
		return section, nil
	}
	ents, err := NewEntityList(core.TagGroups(tags[2:len(tags)-1], 0))
	if err != nil {
		return nil, err
	}
	section.Entities = ents
	return section, nil
}

type entityAccumulator struct {
	parent   entities.Entity
	entities entities.EntitySlice
}

func (e *entityAccumulator) Stop() {
	e.parent.AddNestedEntities(e.entities)
}

func isNestedChild(parent, child entities.Entity) bool {
	switch parent.(type) {
	case *entities.Polyline:
		_, ok := child.(*entities.Vertex)
		return ok || child.IsSeqEnd()
	case *entities.Insert:
		_, ok := child.(*entities.Text)
		return ok
	}
	return false
}

func newEntityAccumulator(parent entities.Entity) *entityAccumulator {
	return &entityAccumulator{
		parent:   parent,
		entities: make(entities.EntitySlice, 0),
	}
}

type entityFactoryFunc func(tags core.TagSlice) (entities.Entity, error)

var entityFactory map[string]entityFactoryFunc

func init() {
	entityFactory = map[string]entityFactoryFunc{
		"LINE": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewLine(tags)
		},
		"POINT": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewPoint(tags)
		},
		"CIRCLE": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewCircle(tags)
		},
		"ARC": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewArc(tags)
		},
		"TEXT": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewText(tags)
		},
		"INSERT": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewInsert(tags)
		},
		"SEQEND": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewSeqEnd(tags)
		},
		"POLYLINE": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewPolyline(tags)
		},
		"VERTEX": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewVertex(tags)
		},
		"LWPOLYLINE": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewLWPolyline(tags)
		},
		"ELLIPSE": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewEllipse(tags)
		},
		"SPLINE": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewSpline(tags)
		},
		"MTEXT": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewMText(tags)
		},
		"HATCH": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewHatch(tags)
		},
		"RAY": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewRay(tags)
		},
		"XLINE": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewXLine(tags)
		},
		"SOLID": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewSolid(tags)
		},
		"TRACE": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewTrace(tags)
		},
		"3DFACE": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewFace3D(tags)
		},
		"LEADER": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewLeader(tags)
		},
		"MLINE": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewMLine(tags)
		},
		"DIMENSION": func(tags core.TagSlice) (entities.Entity, error) {
			return entities.NewDimension(tags)
		},
	}
}

// NewEntityList 从 Tag 切片列表解析实体列表。
// 使用 entityAccumulator 模式处理嵌套实体（如 POLYLINE→VERTEX+SEQEND）。
func NewEntityList(tags []core.TagSlice) (entities.EntitySlice, error) {
	entityList := make(entities.EntitySlice, 0)
	var accumulator *entityAccumulator

	for _, group := range tags {
		entityType := group[0].Value.ToString()
		if factory, ok := entityFactory[entityType]; ok {
			entity, err := factory(group)
			if err != nil {
				return nil, err
			}
			if accumulator != nil {
				if entity.IsSeqEnd() {
					accumulator.Stop()
					entityList = append(entityList, accumulator.parent)
					accumulator = nil
				} else if isNestedChild(accumulator.parent, entity) {
					accumulator.entities = append(accumulator.entities, entity)
				} else {

					accumulator.Stop()
					entityList = append(entityList, accumulator.parent)
					accumulator = nil
					entityList = append(entityList, entity)
				}
			} else if entity.HasNestedEntities() {
				accumulator = newEntityAccumulator(entity)
			} else {
				entityList = append(entityList, entity)
			}
		}
	}

	if accumulator != nil {
		accumulator.Stop()
		entityList = append(entityList, accumulator.parent)
	}
	return entityList, nil
}
