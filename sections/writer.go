package sections

import (
	"sort"
	"strconv"

	"github.com/inkbamboo/rwdxf/core"
	"github.com/inkbamboo/rwdxf/entities"
)

func (s *HeaderSection) DxfTags() core.TagSlice {
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("SECTION")),
		core.NewTag(2, core.NewStringValue("HEADER")),
	}

	fixedOrder := []string{"$ACADVER", "$DWGCODEPAGE", "$EXTMIN", "$EXTMAX"}
	written := make(map[string]bool)

	for _, key := range fixedOrder {
		if v, ok := s.Values[key]; ok {
			tags = append(tags, core.NewTag(9, core.NewStringValue(key)))
			tags = append(tags, v...)
			written[key] = true
		}
	}

	for key, valTags := range s.Values {
		if !written[key] {
			tags = append(tags, core.NewTag(9, core.NewStringValue(key)))
			tags = append(tags, valTags...)
		}
	}

	tags = append(tags, core.NewTag(0, core.NewStringValue("ENDSEC")))
	return tags
}

var R12Tables = false

var tableSeq uint64 = 10

func nextTableID() string {
	h := strconv.FormatUint(tableSeq, 10)
	tableSeq++
	return h
}

func (t *TablesSection) DxfTags() core.TagSlice {
	tableSeq = 10
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("SECTION")),
		core.NewTag(2, core.NewStringValue("TABLES")),
	}
	if R12Tables {

		tags = append(tags, lineTypeTableTagsR12(t.LineTypes)...)
		tags = append(tags, layerTableTagsR12(t.Layers)...)
		tags = append(tags, styleTableTagsR12(t.Styles)...)
		tags = append(tags, simpleTableTagsR12("VIEW", 0)...)
		tags = append(tags, simpleTableTagsR12("UCS", 0)...)
		tags = append(tags, simpleTableTagsR12("VPORT", 0)...)
	} else {

		tags = append(tags, appidTableTags()...)
		tags = append(tags, blockRecordTableTags(t.BlockRecords)...)
		tags = append(tags, dimstyleTableTags()...)
		tags = append(tags, layerTableTags(t.Layers)...)
		tags = append(tags, lineTypeTableTags(t.LineTypes)...)
		tags = append(tags, styleTableTags(t.Styles)...)
		tags = append(tags, simpleTableTags("UCS", 0)...)
		tags = append(tags, simpleTableTags("VIEW", 0)...)
		tags = append(tags, simpleTableTags("VPORT", 0)...)
	}
	tags = append(tags, core.NewTag(0, core.NewStringValue("ENDSEC")))
	return tags
}

func simpleTableTags(tableType string, count int) core.TagSlice {
	return core.TagSlice{
		core.NewTag(0, core.NewStringValue("TABLE")),
		core.NewTag(2, core.NewStringValue(tableType)),
		core.NewTag(5, core.NewStringValue(nextTableID())),
		core.NewTag(330, core.NewStringValue("0")),
		core.NewTag(100, core.NewStringValue("AcDbSymbolTable")),
		core.NewTag(70, core.NewIntegerValue(count)),
		core.NewTag(0, core.NewStringValue("ENDTAB")),
	}
}

func layerTableTags(table Table) core.TagSlice {
	tableHdr := nextTableID()
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("TABLE")),
		core.NewTag(2, core.NewStringValue("LAYER")),
		core.NewTag(5, core.NewStringValue(tableHdr)),
		core.NewTag(330, core.NewStringValue("0")),
		core.NewTag(100, core.NewStringValue("AcDbSymbolTable")),
		core.NewTag(70, core.NewIntegerValue(len(table))),
	}

	var names []string
	for name := range table {
		if name != "0" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	if l, ok := table["0"]; ok {
		if layer, ok := l.(*Layer); ok {
			tags = append(tags, layerTagsWithHandle(layer, tableHdr)...)
		}
	}
	for _, name := range names {
		if l, ok := table[name]; ok {
			if layer, ok := l.(*Layer); ok {
				tags = append(tags, layerTagsWithHandle(layer, tableHdr)...)
			}
		}
	}
	tags = append(tags, core.NewTag(0, core.NewStringValue("ENDTAB")))
	return tags
}

func layerTagsWithHandle(l *Layer, tableHdr string) core.TagSlice {
	entry := l.DxfTags()
	return append(entry[:1], append(core.TagSlice{
		core.NewTag(5, core.NewStringValue(nextTableID())),
		core.NewTag(330, core.NewStringValue(tableHdr)),
	}, entry[1:]...)...)
}

func (l *Layer) DxfTags() core.TagSlice {
	color := l.Color
	if !l.On {
		color = -color
	}
	return core.TagSlice{
		core.NewTag(0, core.NewStringValue("LAYER")),
		core.NewTag(100, core.NewStringValue("AcDbSymbolTableRecord")),
		core.NewTag(100, core.NewStringValue("AcDbLayerTableRecord")),
		core.NewTag(2, core.NewStringValue(l.Name)),
		core.NewTag(70, core.NewIntegerValue(l.Flags())),
		core.NewTag(62, core.NewIntegerValue(color)),
		core.NewTag(6, core.NewStringValue(l.LineType)),
		core.NewTag(390, core.NewStringValue("D")),
	}
}

func simpleTableTagsR12(tableType string, count int) core.TagSlice {
	return core.TagSlice{
		core.NewTag(0, core.NewStringValue("TABLE")),
		core.NewTag(2, core.NewStringValue(tableType)),
		core.NewTag(70, core.NewIntegerValue(count)),
		core.NewTag(0, core.NewStringValue("ENDTAB")),
	}
}

func appidTableTags() core.TagSlice {
	return core.TagSlice{
		core.NewTag(0, core.NewStringValue("TABLE")),
		core.NewTag(2, core.NewStringValue("APPID")),
		core.NewTag(5, core.NewStringValue(nextTableID())),
		core.NewTag(330, core.NewStringValue("0")),
		core.NewTag(100, core.NewStringValue("AcDbSymbolTable")),
		core.NewTag(70, core.NewIntegerValue(1)),
		core.NewTag(0, core.NewStringValue("APPID")),
		core.NewTag(5, core.NewStringValue(nextTableID())),
		core.NewTag(100, core.NewStringValue("AcDbSymbolTableRecord")),
		core.NewTag(100, core.NewStringValue("AcDbRegAppTableRecord")),
		core.NewTag(2, core.NewStringValue("ACAD")),
		core.NewTag(70, core.NewIntegerValue(0)),
		core.NewTag(0, core.NewStringValue("ENDTAB")),
	}
}

func blockRecordTableTags(blockRecords map[string]string) core.TagSlice {
	totalCount := 2 + len(blockRecords)
	tableHdr := nextTableID()
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("TABLE")),
		core.NewTag(2, core.NewStringValue("BLOCK_RECORD")),
		core.NewTag(5, core.NewStringValue(tableHdr)),
		core.NewTag(330, core.NewStringValue("0")),
		core.NewTag(100, core.NewStringValue("AcDbSymbolTable")),
		core.NewTag(70, core.NewIntegerValue(totalCount)),
		core.NewTag(0, core.NewStringValue("BLOCK_RECORD")),
		core.NewTag(5, core.NewStringValue(nextTableID())),
		core.NewTag(100, core.NewStringValue("AcDbSymbolTableRecord")),
		core.NewTag(100, core.NewStringValue("AcDbBlockTableRecord")),
		core.NewTag(2, core.NewStringValue("*Model_Space")),
		core.NewTag(340, core.NewStringValue("A")),
		core.NewTag(0, core.NewStringValue("BLOCK_RECORD")),
		core.NewTag(5, core.NewStringValue(nextTableID())),
		core.NewTag(100, core.NewStringValue("AcDbSymbolTableRecord")),
		core.NewTag(100, core.NewStringValue("AcDbBlockTableRecord")),
		core.NewTag(2, core.NewStringValue("*Paper_Space")),
		core.NewTag(340, core.NewStringValue("B")),
	}

	if len(blockRecords) > 0 {
		names := make([]string, 0, len(blockRecords))
		for name := range blockRecords {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			tags = append(tags,
				core.NewTag(0, core.NewStringValue("BLOCK_RECORD")),
				core.NewTag(5, core.NewStringValue(nextTableID())),
				core.NewTag(100, core.NewStringValue("AcDbSymbolTableRecord")),
				core.NewTag(100, core.NewStringValue("AcDbBlockTableRecord")),
				core.NewTag(2, core.NewStringValue(name)),
				core.NewTag(340, core.NewStringValue(blockRecords[name])),
			)
		}
	}
	tags = append(tags, core.NewTag(0, core.NewStringValue("ENDTAB")))
	return tags
}

func dimstyleTableTags() core.TagSlice {
	hdr := nextTableID()
	e1 := nextTableID()
	return core.TagSlice{
		core.NewTag(0, core.NewStringValue("TABLE")),
		core.NewTag(2, core.NewStringValue("DIMSTYLE")),
		core.NewTag(5, core.NewStringValue(hdr)),
		core.NewTag(330, core.NewStringValue("0")),
		core.NewTag(100, core.NewStringValue("AcDbSymbolTable")),
		core.NewTag(70, core.NewIntegerValue(1)),
		core.NewTag(100, core.NewStringValue("AcDbDimStyleTable")),
		core.NewTag(71, core.NewIntegerValue(1)),
		core.NewTag(0, core.NewStringValue("DIMSTYLE")),
		core.NewTag(5, core.NewStringValue(e1)),
		core.NewTag(105, core.NewStringValue(e1)),
		core.NewTag(100, core.NewStringValue("AcDbSymbolTableRecord")),
		core.NewTag(100, core.NewStringValue("AcDbDimStyleTableRecord")),
		core.NewTag(2, core.NewStringValue("STANDARD")),
		core.NewTag(70, core.NewIntegerValue(0)),
		core.NewTag(0, core.NewStringValue("ENDTAB")),
	}
}

func lineTypeTableTags(table Table) core.TagSlice {
	tableHdr := nextTableID()
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("TABLE")),
		core.NewTag(2, core.NewStringValue("LTYPE")),
		core.NewTag(5, core.NewStringValue(tableHdr)),
		core.NewTag(330, core.NewStringValue("0")),
		core.NewTag(100, core.NewStringValue("AcDbSymbolTable")),
		core.NewTag(70, core.NewIntegerValue(len(table))),
	}
	var names []string
	for n := range table {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		lt, ok := table[name]
		if !ok {
			continue
		}
		if ltObj, ok := lt.(*LineType); ok {
			entry := ltObj.DxfTags()
			entry = append(entry[:1], append(core.TagSlice{
				core.NewTag(5, core.NewStringValue(nextTableID())),
				core.NewTag(330, core.NewStringValue(tableHdr)),
			}, entry[1:]...)...)
			tags = append(tags, entry...)
		}
	}
	tags = append(tags, core.NewTag(0, core.NewStringValue("ENDTAB")))
	return tags
}

func styleTableTags(table Table) core.TagSlice {
	tableHdr := nextTableID()
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("TABLE")),
		core.NewTag(2, core.NewStringValue("STYLE")),
		core.NewTag(5, core.NewStringValue(tableHdr)),
		core.NewTag(330, core.NewStringValue("0")),
		core.NewTag(100, core.NewStringValue("AcDbSymbolTable")),
		core.NewTag(70, core.NewIntegerValue(len(table))),
	}
	var names []string
	for n := range table {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		st, ok := table[name]
		if !ok {
			continue
		}
		if stObj, ok := st.(*Style); ok {
			entry := stObj.DxfTags()
			entry = append(entry[:1], append(core.TagSlice{
				core.NewTag(5, core.NewStringValue(nextTableID())),
				core.NewTag(330, core.NewStringValue(tableHdr)),
			}, entry[1:]...)...)
			tags = append(tags, entry...)
		}
	}
	tags = append(tags, core.NewTag(0, core.NewStringValue("ENDTAB")))
	return tags
}

func layerTableTagsR12(table Table) core.TagSlice {
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("TABLE")),
		core.NewTag(2, core.NewStringValue("LAYER")),
		core.NewTag(70, core.NewIntegerValue(len(table))),
	}
	for _, elem := range table {
		if l, ok := elem.(*Layer); ok {
			tags = append(tags, layerDxfTagsR12(l)...)
		}
	}
	tags = append(tags, core.NewTag(0, core.NewStringValue("ENDTAB")))
	return tags
}

func lineTypeTableTagsR12(table Table) core.TagSlice {
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("TABLE")),
		core.NewTag(2, core.NewStringValue("LTYPE")),
		core.NewTag(70, core.NewIntegerValue(len(table))),
	}
	for _, elem := range table {
		if lt, ok := elem.(*LineType); ok {
			tags = append(tags, lineTypeDxfTagsR12(lt)...)
		}
	}
	tags = append(tags, core.NewTag(0, core.NewStringValue("ENDTAB")))
	return tags
}

func styleTableTagsR12(table Table) core.TagSlice {
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("TABLE")),
		core.NewTag(2, core.NewStringValue("STYLE")),
		core.NewTag(70, core.NewIntegerValue(len(table))),
	}
	for _, elem := range table {
		if st, ok := elem.(*Style); ok {
			tags = append(tags, styleDxfTagsR12(st)...)
		}
	}
	tags = append(tags, core.NewTag(0, core.NewStringValue("ENDTAB")))
	return tags
}

func layerDxfTagsR12(l *Layer) core.TagSlice {
	color := l.Color
	if !l.On {
		color = -color
	}
	return core.TagSlice{
		core.NewTag(0, core.NewStringValue("LAYER")),
		core.NewTag(2, core.NewStringValue(l.Name)),
		core.NewTag(70, core.NewIntegerValue(l.Flags())),
		core.NewTag(62, core.NewIntegerValue(color)),
		core.NewTag(6, core.NewStringValue(l.LineType)),
	}
}

func lineTypeDxfTagsR12(l *LineType) core.TagSlice {
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("LTYPE")),
		core.NewTag(2, core.NewStringValue(l.Name)),
		core.NewTag(70, core.NewIntegerValue(0)),
		core.NewTag(3, core.NewStringValue(l.Description)),
		core.NewTag(72, core.NewIntegerValue(65)),
		core.NewTag(73, core.NewIntegerValue(len(l.Pattern))),
		core.NewTag(40, core.NewFloatValue(l.Length)),
	}
	for _, elem := range l.Pattern {
		tags = append(tags, core.NewTag(49, core.NewFloatValue(elem.Length)))
	}
	return tags
}

func styleDxfTagsR12(st *Style) core.TagSlice {
	flags70 := 0
	if st.IsShape {
		flags70 |= shapeBit
	}
	if st.IsVerticalText {
		flags70 |= verticalTextBit
	}
	flags71 := 0
	if st.IsBackwards {
		flags71 |= backwardsBit
	}
	if st.IsUpsideDown {
		flags71 |= upsideDownBit
	}
	font := st.Font
	if font == "" {
		font = "txt"
	}
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("STYLE")),
		core.NewTag(2, core.NewStringValue(st.Name)),
		core.NewTag(70, core.NewIntegerValue(flags70)),
		core.NewTag(40, core.NewFloatValue(st.Height)),
		core.NewTag(41, core.NewFloatValue(st.Width)),
		core.NewTag(50, core.NewFloatValue(st.Oblique)),
		core.NewTag(71, core.NewIntegerValue(flags71)),
		core.NewTag(42, core.NewFloatValue(0)),
		core.NewTag(3, core.NewStringValue(font)),
	}
	if st.BigFont != "" {
		tags = append(tags, core.NewTag(4, core.NewStringValue(st.BigFont)))
	}
	return tags
}

func (l *LineType) DxfTags() core.TagSlice {

	count := len(l.Pattern)
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("LTYPE")),
		core.NewTag(100, core.NewStringValue("AcDbSymbolTableRecord")),
		core.NewTag(100, core.NewStringValue("AcDbLinetypeTableRecord")),
		core.NewTag(2, core.NewStringValue(l.Name)),
		core.NewTag(70, core.NewIntegerValue(0)),
		core.NewTag(3, core.NewStringValue(l.Description)),
		core.NewTag(72, core.NewIntegerValue(65)),
		core.NewTag(73, core.NewIntegerValue(count)),
		core.NewTag(40, core.NewFloatValue(l.Length)),
	}
	for _, elem := range l.Pattern {
		if elem == nil {
			continue
		}
		tags = append(tags, core.NewTag(49, core.NewFloatValue(elem.Length)))

		flags := 0
		if elem.AbsoluteRotation {
			flags |= absRotationBit
		}
		if elem.IsTextString {
			flags |= textStringBit
		}
		if elem.IsShape {
			flags |= elementShapeBit
		}
		tags = append(tags, core.NewTag(74, core.NewIntegerValue(flags)))
		if elem.IsShape {
			tags = append(tags, core.NewTag(75, core.NewIntegerValue(elem.ShapeNumber)))
		}
		if elem.Scale != 0 && !core.FloatEquals(elem.Scale, 1.0) {
			tags = append(tags, core.NewTag(46, core.NewFloatValue(elem.Scale)))
		}
		if !core.FloatEquals(elem.RotationAngle, 0) {
			if elem.AbsoluteRotation {
				tags = append(tags, core.NewTag(50, core.NewFloatValue(elem.RotationAngle)))
			} else {
				tags = append(tags, core.NewTag(50, core.NewFloatValue(elem.RotationAngle)))
			}
		}
		if !core.FloatEquals(elem.XOffset, 0) {
			tags = append(tags, core.NewTag(44, core.NewFloatValue(elem.XOffset)))
		}
		if !core.FloatEquals(elem.YOffset, 0) {
			tags = append(tags, core.NewTag(45, core.NewFloatValue(elem.YOffset)))
		}
		if elem.Text != "" {
			tags = append(tags, core.NewTag(9, core.NewStringValue(elem.Text)))
		}
	}
	return tags
}

func (e *LineElement) DxfTags() core.TagSlice {
	return core.TagSlice{
		core.NewTag(49, core.NewFloatValue(e.Length)),
	}
}

func (st *Style) DxfTags() core.TagSlice {
	flags70 := 0
	if st.IsShape {
		flags70 |= shapeBit
	}
	if st.IsVerticalText {
		flags70 |= verticalTextBit
	}
	flags71 := 0
	if st.IsBackwards {
		flags71 |= backwardsBit
	}
	if st.IsUpsideDown {
		flags71 |= upsideDownBit
	}
	font := st.Font
	if font == "" {
		font = "txt"
	}
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("STYLE")),
		core.NewTag(100, core.NewStringValue("AcDbSymbolTableRecord")),
		core.NewTag(100, core.NewStringValue("AcDbTextStyleTableRecord")),
		core.NewTag(2, core.NewStringValue(st.Name)),
		core.NewTag(70, core.NewIntegerValue(flags70)),
		core.NewTag(40, core.NewFloatValue(st.Height)),
		core.NewTag(41, core.NewFloatValue(st.Width)),
		core.NewTag(50, core.NewFloatValue(st.Oblique)),
		core.NewTag(71, core.NewIntegerValue(flags71)),
		core.NewTag(42, core.NewFloatValue(0)),
		core.NewTag(3, core.NewStringValue(font)),
	}
	if st.BigFont != "" {
		tags = append(tags, core.NewTag(4, core.NewStringValue(st.BigFont)))
	}
	return tags
}

func (b BlocksSection) DxfTags() core.TagSlice {
	resetBlockHandleSeq()
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("SECTION")),
		core.NewTag(2, core.NewStringValue("BLOCKS")),
	}

	var userNames []string
	for name := range b {
		if name != "*Model_Space" && name != "*Paper_Space" {
			userNames = append(userNames, name)
		}
	}
	sort.Strings(userNames)
	if ms, ok := b["*Model_Space"]; ok {
		tags = append(tags, ms.DxfTags()...)
	}
	if ps, ok := b["*Paper_Space"]; ok {
		tags = append(tags, ps.DxfTags()...)
	}
	for _, name := range userNames {
		tags = append(tags, b[name].DxfTags()...)
	}
	tags = append(tags, core.NewTag(0, core.NewStringValue("ENDSEC")))
	return tags
}

func (b *Block) DxfTags() core.TagSlice {
	if R12Tables {
		return b.dxfTagsR12()
	}
	return b.dxfTagsAC1032()
}

var blockHandleSeq uint64 = 100
var blockHandleBase uint64 = 100

func resetBlockHandleSeq() {
	blockHandleSeq = blockHandleBase
}

func nextBlockHandle() string {
	h := strconv.FormatUint(blockHandleSeq, 10)
	blockHandleSeq++
	return h
}

func SetBlockHandleSeq(seq uint64) {
	blockHandleBase = seq
	blockHandleSeq = seq
}

func (b *Block) dxfTagsAC1032() core.TagSlice {
	h := b.Handle
	if h == "" {
		h = nextBlockHandle()
	}
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("BLOCK")),
		core.NewTag(5, core.NewStringValue(h)),
		core.NewTag(100, core.NewStringValue("AcDbEntity")),
		core.NewTag(8, core.NewStringValue(b.LayerName)),
		core.NewTag(100, core.NewStringValue("AcDbBlockBegin")),
		core.NewTag(2, core.NewStringValue(b.Name)),
		core.NewTag(70, core.NewIntegerValue(0)),
		core.NewTag(10, core.NewFloatValue(b.BasePoint.X)),
		core.NewTag(20, core.NewFloatValue(b.BasePoint.Y)),
		core.NewTag(30, core.NewFloatValue(b.BasePoint.Z)),
		core.NewTag(3, core.NewStringValue(b.SecondName)),
	}
	if b.XrefPathName != "" {
		tags = append(tags, core.NewTag(1, core.NewStringValue(b.XrefPathName)))
	}
	if b.Description != "" {
		tags = append(tags, core.NewTag(4, core.NewStringValue(b.Description)))
	}
	for _, e := range b.Entities {
		entities.SetEntityOwner(e, h)
		tags = append(tags, e.DxfTags()...)
	}

	endH := nextBlockHandle()
	tags = append(tags,
		core.NewTag(0, core.NewStringValue("ENDBLK")),
		core.NewTag(5, core.NewStringValue(endH)),
		core.NewTag(100, core.NewStringValue("AcDbEntity")),
		core.NewTag(8, core.NewStringValue(b.LayerName)),
		core.NewTag(100, core.NewStringValue("AcDbBlockEnd")),
	)
	return tags
}

func (b *Block) dxfTagsR12() core.TagSlice {
	layer := b.LayerName
	if layer == "" {
		layer = "0"
	}
	name := b.Name
	secondName := b.SecondName
	if secondName == "" {
		secondName = name
	}
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("BLOCK")),
		core.NewTag(8, core.NewStringValue(layer)),
		core.NewTag(2, core.NewStringValue(name)),
		core.NewTag(70, core.NewIntegerValue(0)),
		core.NewTag(10, core.NewFloatValue(b.BasePoint.X)),
		core.NewTag(20, core.NewFloatValue(b.BasePoint.Y)),
		core.NewTag(30, core.NewFloatValue(b.BasePoint.Z)),
		core.NewTag(3, core.NewStringValue(secondName)),
		core.NewTag(1, core.NewStringValue(b.XrefPathName)),
	}
	if b.Description != "" {
		tags = append(tags, core.NewTag(4, core.NewStringValue(b.Description)))
	}
	for _, e := range b.Entities {
		if !e.IsR12Compatible() {
			continue
		}
		tags = append(tags, e.DxfTags()...)
	}

	tags = append(tags,
		core.NewTag(0, core.NewStringValue("ENDBLK")),
		core.NewTag(8, core.NewStringValue(layer)),
	)
	return tags
}

func (e *EntitiesSection) DxfTags() core.TagSlice {
	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("SECTION")),
		core.NewTag(2, core.NewStringValue("ENTITIES")),
	}
	for _, entity := range e.Entities {

		if entities.R12Mode && !entity.IsR12Compatible() {
			continue
		}
		tags = append(tags, entity.DxfTags()...)
	}
	tags = append(tags, core.NewTag(0, core.NewStringValue("ENDSEC")))
	return tags
}

func WriteEntityTags(es *EntitiesSection) string {
	return entities.WriteTags(es.DxfTags())
}

// NewDefaultHeader 创建带默认 $ACADVER 和 $DWGCODEPAGE 的 Header。
func NewDefaultHeader() *HeaderSection {
	h := &HeaderSection{
		Values: make(map[string]core.TagSlice),
	}
	h.Values[tagACADVER] = core.TagSlice{
		core.NewTag(1, core.NewStringValue("AC1009")),
	}
	h.Values[tagDWGCODEPAGE] = core.TagSlice{
		core.NewTag(3, core.NewStringValue("ANSI_1252")),
	}
	return h
}

// NewDefaultTables 创建带默认 "0" 图层、STANDARD 样式和标准线型的 Tables。
func NewDefaultTables() *TablesSection {
	t := &TablesSection{
		Layers:    make(Table),
		Styles:    make(Table),
		LineTypes: make(Table),
	}

	zeroLayer := &Layer{
		Name:     "0",
		Color:    7,
		LineType: "CONTINUOUS",
		On:       true,
	}
	t.Layers["0"] = zeroLayer

	standardStyle := &Style{
		Name:    "STANDARD",
		Height:  0.0,
		Width:   1.0,
		Font:    "txt",
		BigFont: "",
	}
	t.Styles["STANDARD"] = standardStyle

	t.LineTypes["ByLayer"] = &LineType{
		Name:        "ByLayer",
		Description: "",
		Length:      0.0,
	}
	t.LineTypes["ByBlock"] = &LineType{
		Name:        "ByBlock",
		Description: "",
		Length:      0.0,
	}
	continuous := &LineType{
		Name:        "CONTINUOUS",
		Description: "Solid line",
		Length:      0.0,
	}
	t.LineTypes["CONTINUOUS"] = continuous

	return t
}
