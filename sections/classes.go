package sections

// ClassesSection 表示 DXF CLASSES 段，以原始字符串形式保存。
type ClassesSection struct {
	Raw string
}

// MinClassesSectionTags 返回最小的 CLASSES 段内容（包含基本类定义）。
func MinClassesSectionTags() string {

	return "0\r\nCLASS\r\n1\r\nACDBDICTIONARYWDFLT\r\n2\r\nAcDbDictionaryWithDefault\r\n3\r\nObjectDBX Classes\r\n90\r\n0\r\n91\r\n1\r\n280\r\n0\r\n281\r\n0\r\n" +
		"0\r\nCLASS\r\n1\r\nDICTIONARYVAR\r\n2\r\nAcDbDictionaryVar\r\n3\r\nObjectDBX Classes\r\n90\r\n0\r\n91\r\n0\r\n280\r\n0\r\n281\r\n0\r\n" +
		"0\r\nCLASS\r\n1\r\nLAYER_INDEX\r\n2\r\nAcDbLayerIndex\r\n3\r\nObjectDBX Classes\r\n90\r\n0\r\n91\r\n0\r\n280\r\n0\r\n281\r\n0\r\n" +
		"0\r\nCLASS\r\n1\r\nSPATIAL_INDEX\r\n2\r\nAcDbSpatialIndex\r\n3\r\nObjectDBX Classes\r\n90\r\n0\r\n91\r\n0\r\n280\r\n0\r\n281\r\n0\r\n"
}
