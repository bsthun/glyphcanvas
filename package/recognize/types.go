package recognize

type CharacterFeature struct {
	Unicode        string             `yaml:"unicode"`
	GridSignature  string             `yaml:"grid_signature"`
	DirectionHist  [8]float64         `yaml:"direction_histogram"`
	ZoningFeatures [16]float64        `yaml:"zoning_features"`
	ChainCode      string             `yaml:"chain_code"`
	HuMoments      [7]float64         `yaml:"hu_moments"`
	AspectRatio    float64            `yaml:"aspect_ratio"`
	Density        float64            `yaml:"density"`
	CenterOfMass   [2]float64         `yaml:"center_of_mass"`
	EndPoints      int                `yaml:"end_points"`
	Junctions      int                `yaml:"junctions"`
	RegionCount    int                `yaml:"region_count"`
	RegionFeatures []RegionFeatureSet `yaml:"region_features"`
	TopologyHash   string             `yaml:"topology_hash"`
}

type RegionFeatureSet struct {
	ArcType       string     `yaml:"arc_type"`
	Circularity   float64    `yaml:"circularity"`
	Linearity     float64    `yaml:"linearity"`
	CurveStrength float64    `yaml:"curve_strength"`
	HuMoments     [7]float64 `yaml:"hu_moments"`
	ChainCodeHash string     `yaml:"chain_code_hash"`
	RelativeSize  float64    `yaml:"relative_size"`
	RelativePos   [2]float64 `yaml:"relative_position"`
}

type FeatureDatabase struct {
	Characters map[string]*CharacterFeature `yaml:"characters"`
}

type RecognitionCandidate struct {
	Unicode    string
	Confidence float64
	Distance   float64
}
