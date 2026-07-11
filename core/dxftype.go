package core

// DxfTypeName 表示 DXF 实体类型名称，用于标识和区分不同的 DXF 实体。
type DxfTypeName string

// 支持的 DXF 实体类型名称常量。
const (
	DxfTypeArc        DxfTypeName = "ARC"        // 圆弧
	DxfTypeCircle     DxfTypeName = "CIRCLE"     // 圆
	DxfTypeDimension  DxfTypeName = "DIMENSION"  // 尺寸标注
	DxfTypeEllipse    DxfTypeName = "ELLIPSE"    // 椭圆
	DxfTypeFace3D     DxfTypeName = "3DFACE"     // 三维面
	DxfTypeHatch      DxfTypeName = "HATCH"      // 填充图案
	DxfTypeInsert     DxfTypeName = "INSERT"     // 块引用
	DxfTypeLeader     DxfTypeName = "LEADER"     // 引线标注
	DxfTypeLine       DxfTypeName = "LINE"       // 直线
	DxfTypeLWPolyline DxfTypeName = "LWPOLYLINE" // 轻量多段线
	DxfTypeMLine      DxfTypeName = "MLINE"      // 多线
	DxfTypeMText      DxfTypeName = "MTEXT"      // 多行文字
	DxfTypePoint      DxfTypeName = "POINT"      // 点
	DxfTypePolyline   DxfTypeName = "POLYLINE"   // 多段线
	DxfTypeRay        DxfTypeName = "RAY"        // 射线
	DxfTypeSeqEnd     DxfTypeName = "SEQEND"     // 序列结束标记
	DxfTypeSolid      DxfTypeName = "SOLID"      // 实心填充
	DxfTypeSpline     DxfTypeName = "SPLINE"     // 样条曲线
	DxfTypeText       DxfTypeName = "TEXT"       // 单行文字
	DxfTypeTrace      DxfTypeName = "TRACE"      // 宽线
	DxfTypeVertex     DxfTypeName = "VERTEX"     // 顶点
	DxfTypeXLine      DxfTypeName = "XLINE"      // 构造线
)
