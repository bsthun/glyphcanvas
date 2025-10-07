package region

type ArcType int

const (
	ArcTypeCircle ArcType = iota
	ArcTypeStrengthLine
	ArcTypeCurveLine
	ArcTypeTriangle
	ArcTypeRectangle
)

type ArcFillType int

const (
	ArcFillTypeFill ArcFillType = iota
	ArcFillTypeStroke
)

type Arc struct {
	Type               ArcType
	Fill               ArcFillType
	CircleEllipseRatio float32
	LineDegree         float32
	ArcLineTheta       float32
}
