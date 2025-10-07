package character

import "fmt"

type CharacterConfig struct {
	// Anchor Detection Configuration
	AnchorDetectionThreshold float64 `json:"anchorDetectionThreshold"` // Threshold for anchor point significance
	MinAnchorDistance        float64 `json:"minAnchorDistance"`        // Minimum distance between anchor points
	CurvatureThreshold       float64 `json:"curvatureThreshold"`       // Curvature threshold for anchor detection

	// Medial Axis Configuration
	MedialAxisEpsilon        float64 `json:"medialAxisEpsilon"`        // Precision for medial axis computation
	MedialAxisSimplification float64 `json:"medialAxisSimplification"` // Simplification factor for medial axis
	SkeletonPruningThreshold float64 `json:"skeletonPruningThreshold"` // Threshold for pruning short skeleton branches

	// Region Decomposition Configuration
	MinRegionSize        uint16  `json:"minRegionSize"`        // Minimum size for a valid region
	RegionMergeThreshold float64 `json:"regionMergeThreshold"` // Threshold for merging adjacent regions
	ConnectivityType     int     `json:"connectivityType"`     // 4-connectivity (0) or 8-connectivity (1)

	// Character Analysis Configuration
	EnableStrokeAnalysis    bool `json:"enableStrokeAnalysis"`    // Enable stroke-based analysis
	EnableTopologyAnalysis  bool `json:"enableTopologyAnalysis"`  // Enable topology preservation
	EnableJunctionDetection bool `json:"enableJunctionDetection"` // Enable junction point detection

	// Geometric Analysis Configuration
	CircularityThreshold    float64 `json:"circularityThreshold"`    // Threshold for circular region classification
	LinearityThreshold      float64 `json:"linearityThreshold"`      // Threshold for linear region classification
	RectangularityThreshold float64 `json:"rectangularityThreshold"` // Threshold for rectangular region classification

	// Performance Configuration
	EnableParallelProcessing bool `json:"enableParallelProcessing"` // Enable parallel processing where applicable
	MaxRegions               int  `json:"maxRegions"`               // Maximum number of regions to analyze
	ComputationTimeout       int  `json:"computationTimeout"`       // Timeout in milliseconds for analysis
}

func DefaultCharacterConfig() *CharacterConfig {
	return &CharacterConfig{
		// Anchor Detection
		AnchorDetectionThreshold: 0.7,
		MinAnchorDistance:        3.0,
		CurvatureThreshold:       0.5,

		// Medial Axis
		MedialAxisEpsilon:        0.1,
		MedialAxisSimplification: 0.2,
		SkeletonPruningThreshold: 5.0,

		// Region Decomposition
		MinRegionSize:        4,
		RegionMergeThreshold: 0.8,
		ConnectivityType:     1, // 8-connectivity

		// Character Analysis
		EnableStrokeAnalysis:    true,
		EnableTopologyAnalysis:  true,
		EnableJunctionDetection: true,

		// Geometric Analysis
		CircularityThreshold:    0.85,
		LinearityThreshold:      0.9,
		RectangularityThreshold: 0.8,

		// Performance
		EnableParallelProcessing: true,
		MaxRegions:               100,
		ComputationTimeout:       5000, // 5 seconds
	}
}

func (config *CharacterConfig) Validate() error {
	if config.AnchorDetectionThreshold < 0 || config.AnchorDetectionThreshold > 1 {
		return fmt.Errorf("anchorDetectionThreshold must be between 0 and 1")
	}
	if config.MinAnchorDistance < 0 {
		return fmt.Errorf("minAnchorDistance must be non-negative")
	}
	if config.MedialAxisEpsilon <= 0 {
		return fmt.Errorf("medialAxisEpsilon must be positive")
	}
	if config.MinRegionSize == 0 {
		return fmt.Errorf("minRegionSize must be positive")
	}
	if config.ConnectivityType != 0 && config.ConnectivityType != 1 {
		return fmt.Errorf("connectivityType must be 0 (4-connectivity) or 1 (8-connectivity)")
	}
	if config.MaxRegions <= 0 {
		return fmt.Errorf("maxRegions must be positive")
	}
	if config.ComputationTimeout <= 0 {
		return fmt.Errorf("computationTimeout must be positive")
	}
	return nil
}
