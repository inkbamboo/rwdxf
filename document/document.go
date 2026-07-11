// Package document 提供 DXF 文档级别的高层读写 API。
//
// 核心功能：
//   - DocumentFromStream：从 io.Reader 解析 DXF 文件（自动 GBK 编解码检测）
//   - NewDocument：创建空白 DXF 文档
//   - Write/WriteRaw：写入 DXF 文件（支持 R12/AC1032 双版本）
//   - 实体增删、图层管理、版本切换
//
// 视口缩放通过 Modelspace 结构体提供 Extents/Window/Center/Objects 等缩放方式。
package document

import (
	"bufio"
	"bytes"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/inkbamboo/rwdxf/core"
	"github.com/inkbamboo/rwdxf/entities"
	"github.com/inkbamboo/rwdxf/sections"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// DXFVersion 表示 DXF 文件版本。
type DXFVersion int

const (
	// VersionAC1032 表示 AutoCAD 2018 格式的 DXF。
	VersionAC1032 DXFVersion = iota
	// VersionR12 表示经典 R12 格式的 DXF。
	VersionR12
)

// Document 表示一个完整的 DXF 文档，聚合了所有段（Header、Tables、Entities、Blocks、Objects 等）。
type Document struct {
	Header     *sections.HeaderSection
	Classes    *sections.ClassesSection
	Tables     *sections.TablesSection
	Entities   *sections.EntitiesSection
	Blocks     sections.BlocksSection
	Objects    *sections.ObjectsSection
	DXFVersion DXFVersion
}

func (d Document) Equals(other *Document) bool {
	return d.Header.Equals(other.Header) &&
		d.Tables.Equals(other.Tables) &&
		d.Entities.Equals(other.Entities) &&
		d.Blocks.Equals(other.Blocks) &&
		d.Objects.Equals(other.Objects)
}

// DocumentFromStream 从 io.Reader 解析 DXF 文件。
// 自动检测 $DWGCODEPAGE 是否为 ANSI_936 来决定是否使用 GBK 解码。
func DocumentFromStream(stream io.Reader) (*Document, error) {
	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data)
	if isGBKEncoded(data) {
		reader = bytes.NewReader(data)
		decoder := transform.NewReader(reader, simplifiedchinese.GBK.NewDecoder())
		return parseDocument(decoder)
	}
	return parseDocument(bytes.NewReader(data))
}

func detectVersion(d *Document) {
	if d.Header == nil {
		return
	}
	v, ok := d.Header.Values["$ACADVER"]
	if !ok || len(v) == 0 {
		d.DXFVersion = VersionAC1032
		return
	}
	ver := v[0].Value.ToString()
	if ver == "AC1009" {
		d.DXFVersion = VersionR12
	} else {
		d.DXFVersion = VersionAC1032
	}
}

func isGBKEncoded(data []byte) bool {

	for i := 0; i < len(data)-20; i++ {
		if data[i] == '$' && string(data[i:i+13]) == "$DWGCODEPAGE" {

			j := i + 13
			for j < len(data) && (data[j] == '\r' || data[j] == '\n' || data[j] == ' ') {
				j++
			}

			k := j
			for k < len(data) && data[k] != '\r' && data[k] != '\n' {
				k++
			}

			k++
			if k < len(data) && data[k] == '\n' {
				k++
			}
			for k < len(data) && (data[k] == '\r' || data[k] == '\n' || data[k] == ' ') {
				k++
			}

			end := k
			for end < len(data) && data[end] != '\r' && data[end] != '\n' {
				end++
			}
			val := strings.TrimSpace(string(data[k:end]))
			return val == "ANSI_936" || val == "ANSI_936\r"
		}
	}
	return false
}

func parseDocument(stream io.Reader) (*Document, error) {
	d := new(Document)
	d.Header = new(sections.HeaderSection)
	d.Tables = new(sections.TablesSection)
	d.Entities = new(sections.EntitiesSection)
	d.Blocks = make(sections.BlocksSection)

	sectionParsers := map[string]func(slice core.TagSlice) error{
		"HEADER": func(slice core.TagSlice) error {
			d.Header = sections.NewHeaderSection(slice)
			return nil
		},
		"TABLES": func(slice core.TagSlice) error {
			section, err := sections.NewTablesSection(slice)
			d.Tables = section
			return err
		},
		"ENTITIES": func(slice core.TagSlice) error {
			section, err := sections.NewEntitiesSection(slice)
			d.Entities = section
			return err
		},
		"BLOCKS": func(slice core.TagSlice) error {
			section, err := sections.NewBlocksSection(slice)
			d.Blocks = section
			return err
		},
		"CLASSES": func(slice core.TagSlice) error {
			d.Classes = &sections.ClassesSection{
				Raw: entities.WriteTags(slice[2 : len(slice)-1]),
			}
			return nil
		},
		"OBJECTS": func(slice core.TagSlice) error {
			section, err := sections.NewObjectsSection(slice)
			d.Objects = section

			d.Objects.Raw = entities.WriteTags(slice[2 : len(slice)-1])
			return err
		},
	}

	next := core.Tagger(stream)
	tags := core.TagSlice(core.AllTags(next))

	stopTag := core.NewTag(0, core.NewStringValue("EOF"))
	endOfChunk := core.NewTag(0, core.NewStringValue("ENDSEC"))
	for _, sectionTags := range sections.SplitTagChunks(tags, stopTag, endOfChunk) {
		sectionType := sectionTags[1].Value.ToString()
		if parserFunc, ok := sectionParsers[sectionType]; ok {
			if err := parserFunc(sectionTags); err != nil {
				return nil, err
			}
		}
	}
	detectVersion(d)
	return d, nil
}

// NewDocument 创建一个带默认 HEADER 和 TABLES 的空白 DXF 文档。
func NewDocument() *Document {
	d := &Document{
		Header:   sections.NewDefaultHeader(),
		Tables:   sections.NewDefaultTables(),
		Entities: &sections.EntitiesSection{},
		Blocks:   make(sections.BlocksSection),
		Objects:  &sections.ObjectsSection{},
	}
	d.SetVersion(VersionAC1032)
	return d
}

// Modelspace 返回文档的模型空间视图，用于视口缩放操作。
func (d *Document) Modelspace() *Modelspace {
	return &Modelspace{
		Entities:   d.Entities.Entities,
		headerVars: d.Header.Values,
	}
}

// Write 将文档以 GBK 编码写入 io.Writer。
// 自动重置实体 Handle、R12 缓存等全局状态。
func (d *Document) Write(w io.Writer) error {
	entities.ResetEntityHandles()
	entities.ResetR12HatchBlocks()
	entities.ResetR12DimBlocks()
	entities.ResetR12ExtraBlocks()
	entities.ResetR12BlockSeq()
	entities.UseGBK = true
	defer func() { entities.UseGBK = false }()
	bw := bufio.NewWriterSize(w, 256*1024)
	err := d.writeVersioned(bw)
	if err != nil {
		return err
	}
	return bw.Flush()
}

// WriteRaw 将文档以原始编码（不转 GBK）写入 io.Writer。
func (d *Document) WriteRaw(w io.Writer) error {
	entities.ResetEntityHandles()
	entities.ResetR12HatchBlocks()
	entities.ResetR12DimBlocks()
	entities.ResetR12ExtraBlocks()
	entities.ResetR12BlockSeq()
	bw := bufio.NewWriterSize(w, 256*1024)
	err := d.writeVersioned(bw)
	if err != nil {
		return err
	}
	return bw.Flush()
}

func (d *Document) writeVersioned(w io.Writer) error {

	if ver, ok := d.Header.Values["$ACADVER"]; ok && len(ver) > 0 {
		if ver[0].Value.ToString() == "AC1009" {
			entities.R12Mode = true
			sections.R12Tables = true
			return d.writeR12(w)
		}
	}
	switch d.DXFVersion {
	case VersionR12:
		entities.R12Mode = true
		sections.R12Tables = true
		return d.writeR12(w)
	default:
		entities.R12Mode = false
		sections.R12Tables = false
		return d.writeAC1032(w)
	}
}

func (d *Document) writeR12(w io.Writer) error {

	entities.ResetR12HatchBlocks()
	entities.ResetR12DimBlocks()
	entities.ResetR12ExtraBlocks()
	entities.ResetR12BlockSeq()

	delete(d.Header.Values, "$HANDSEED")
	d.Classes = nil

	if err := entities.WriteTagsTo(w, d.Header.DxfTags()); err != nil {
		return err
	}
	if err := entities.WriteTagsTo(w, d.Tables.DxfTags()); err != nil {
		return err
	}

	entityTags := d.Entities.DxfTags()

	blocksTags := d.buildR12BlocksTags()
	if len(blocksTags) > 0 {
		if err := entities.WriteTagsTo(w, blocksTags); err != nil {
			return err
		}
	}

	if err := entities.WriteTagsTo(w, entityTags); err != nil {
		return err
	}

	_, err := io.WriteString(w, "0\r\nEOF\r\n")
	return err
}

func (d *Document) buildR12BlocksTags() core.TagSlice {

	userBlocks := make(sections.BlocksSection)
	for name, blk := range d.Blocks {
		if name == "*Model_Space" || name == "*Paper_Space" || name == "$MODEL_SPACE" || name == "$PAPER_SPACE" {
			continue
		}
		userBlocks[name] = blk
	}

	tags := core.TagSlice{
		core.NewTag(0, core.NewStringValue("SECTION")),
		core.NewTag(2, core.NewStringValue("BLOCKS")),
	}
	existingTags := userBlocks.DxfTags()
	if len(existingTags) > 2 {
		tags = append(tags, existingTags[2:len(existingTags)-1]...)
	}

	hatchBlocks := entities.GetR12HatchBlocks()
	dimBlocks := entities.GetR12DimBlocks()
	extraBlocks := entities.R12ExtraBlocks

	writeBlockMap := func(blocks map[string]core.TagSlice) {
		var names []string
		for name := range blocks {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			tags = append(tags, blocks[name]...)
		}
	}
	writeBlockMap(hatchBlocks)
	writeBlockMap(dimBlocks)
	writeBlockMap(extraBlocks)

	if len(tags) <= 2 {
		return nil
	}
	tags = append(tags, core.NewTag(0, core.NewStringValue("ENDSEC")))
	return tags
}

func (d *Document) writeAC1032(w io.Writer) error {

	totalEntities := 0
	for _, e := range d.Entities.Entities {
		totalEntities++
		if poly, ok := e.(*entities.Polyline); ok {
			totalEntities += len(poly.Vertices) + 1
		}
		if ins, ok := e.(*entities.Insert); ok {
			totalEntities += len(ins.Attribs)
		}
	}
	for _, block := range d.Blocks {
		totalEntities += len(block.Entities)
	}

	handSeed := strconv.FormatUint(100000+uint64(totalEntities)+1, 16)
	d.Header.Values["$HANDSEED"] = core.TagSlice{
		core.NewTag(5, core.NewStringValue(handSeed)),
	}
	if err := entities.WriteTagsTo(w, d.Header.DxfTags()); err != nil {
		return err
	}

	classesRaw := d.classesRaw()
	if _, err := io.WriteString(w, "0\r\nSECTION\r\n2\r\nCLASSES\r\n"+classesRaw+"0\r\nENDSEC\r\n"); err != nil {
		return err
	}

	blocks := d.Blocks
	if _, ok := blocks["*Model_Space"]; !ok {
		blocks = make(sections.BlocksSection)
		for k, v := range d.Blocks {
			blocks[k] = v
		}
		blocks["*Model_Space"] = &sections.Block{
			Name: "*Model_Space", Handle: "A", SecondName: "*Model_Space",
			LayerName: "0", BasePoint: core.Point{X: 0, Y: 0, Z: 0},
		}
		blocks["*Paper_Space"] = &sections.Block{
			Name: "*Paper_Space", Handle: "B", SecondName: "*Paper_Space",
			LayerName: "0", BasePoint: core.Point{X: 0, Y: 0, Z: 0},
		}
	}

	handleSeq := uint64(100)
	blockRecords := make(map[string]string)

	var userBlockNames []string
	for name := range blocks {
		if name != "*Model_Space" && name != "*Paper_Space" {
			userBlockNames = append(userBlockNames, name)
		}
	}
	sort.Strings(userBlockNames)
	for _, name := range userBlockNames {
		block := blocks[name]
		if block.Handle == "" {
			block.Handle = strconv.FormatUint(handleSeq, 10)
			handleSeq++
		}
		blockRecords[name] = block.Handle
	}

	if d.Tables != nil && len(blockRecords) > 0 {
		d.Tables.BlockRecords = blockRecords
	}

	sections.SetBlockHandleSeq(handleSeq)

	if d.Tables != nil {
		if err := entities.WriteTagsTo(w, d.Tables.DxfTags()); err != nil {
			return err
		}
	}
	entities.ModelSpaceOwner = "13"
	if err := entities.WriteTagsTo(w, blocks.DxfTags()); err != nil {
		return err
	}
	if err := entities.WriteTagsTo(w, d.Entities.DxfTags()); err != nil {
		return err
	}

	if d.DXFVersion != VersionR12 {
		if err := entities.WriteTagsTo(w, sections.MinObjectsSectionTags()); err != nil {
			return err
		}
	}
	if d.Objects != nil && (d.Objects.Raw != "" || len(d.Objects.Objects) > 0) {
		if d.Objects.Raw != "" {
			if _, err := io.WriteString(w, "0\r\nSECTION\r\n2\r\nOBJECTS\r\n"+d.Objects.Raw+"0\r\nENDSEC\r\n"); err != nil {
				return err
			}
		} else {
			if err := entities.WriteTagsTo(w, d.Objects.DxfTags()); err != nil {
				return err
			}
		}
	}
	_, err := io.WriteString(w, "0\r\nEOF\r\n")
	return err
}

// AddEntity 向文档添加一个实体。
func (d *Document) AddEntity(e entities.Entity) {
	d.Entities.Entities = append(d.Entities.Entities, e)
}

// AddEntities 批量向文档添加多个实体。
func (d *Document) AddEntities(es ...entities.Entity) {
	d.Entities.Entities = append(d.Entities.Entities, es...)
}

// AllEntities 返回文档中的所有实体。
func (d *Document) AllEntities() entities.EntitySlice {
	if d.Entities == nil {
		return nil
	}
	return d.Entities.Entities
}

// Layers 返回文档中所有 Layer 的切片。
func (d *Document) Layers() []*sections.Layer {
	if d.Tables == nil || d.Tables.Layers == nil {
		return nil
	}
	var layers []*sections.Layer
	for _, v := range d.Tables.Layers {
		if l, ok := v.(*sections.Layer); ok {
			layers = append(layers, l)
		}
	}
	return layers
}

// Layer 按名称查找并返回指定图层。
func (d *Document) Layer(name string) *sections.Layer {
	if d.Tables == nil || d.Tables.Layers == nil {
		return nil
	}

	if l, ok := d.Tables.Layers[name]; ok {
		if layer, ok := l.(*sections.Layer); ok {
			return layer
		}
	}
	return nil
}

// AddLayer 向文档添加一个新图层。
func (d *Document) AddLayer(layer *sections.Layer) {
	if d.Tables == nil {
		d.Tables = sections.NewDefaultTables()
	}
	if d.Tables.Layers == nil {
		d.Tables.Layers = make(sections.Table)
	}
	d.Tables.Layers[layer.Name] = layer
}

func (d *Document) classesRaw() string {
	if d.Classes != nil && d.Classes.Raw != "" {
		return d.Classes.Raw
	}
	return sections.MinClassesSectionTags()
}

// SetHeaderVariable 设置 HEADER 段的变量值。
func (d *Document) SetHeaderVariable(name string, tags core.TagSlice) {
	if d.Header == nil {
		d.Header = sections.NewDefaultHeader()
	}
	d.Header.Values[name] = tags
}

// SetVersion 设置 DXF 版本并更新 $ACADVER 和 $DWGCODEPAGE Header 变量。
func (d *Document) SetVersion(v DXFVersion) {
	d.DXFVersion = v
	codepageTag := core.TagSlice{
		core.NewTag(3, core.NewStringValue("ANSI_936")),
	}
	switch v {
	case VersionR12:
		d.Header.Values["$ACADVER"] = core.TagSlice{
			core.NewTag(1, core.NewStringValue("AC1009")),
		}
		d.Header.Values["$DWGCODEPAGE"] = codepageTag
	default:
		d.Header.Values["$ACADVER"] = core.TagSlice{
			core.NewTag(1, core.NewStringValue("AC1032")),
		}
		d.Header.Values["$DWGCODEPAGE"] = codepageTag
	}
}

// GetVersion 返回版本字符串（"R12" 或 "2018"）。
func (d *Document) GetVersion() string {
	switch d.DXFVersion {
	case VersionR12:
		return "R12"
	default:
		return "2018"
	}
}
