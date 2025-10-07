package regionHelper

import "math"

func RegionComputeHuInvariants(moments map[string]float64) []float64 {
	hu := make([]float64, 7)

	m00 := moments["m00"]
	if m00 == 0 {
		return hu
	}

	norm := math.Pow(m00, 2.5)

	eta20 := moments["mu20"] / norm
	eta02 := moments["mu02"] / norm
	eta11 := moments["mu11"] / norm
	eta30 := moments["mu30"] / math.Pow(m00, 3.5)
	eta21 := moments["mu21"] / math.Pow(m00, 3.5)
	eta12 := moments["mu12"] / math.Pow(m00, 3.5)
	eta03 := moments["mu03"] / math.Pow(m00, 3.5)

	hu[0] = eta20 + eta02
	hu[1] = math.Pow(eta20-eta02, 2) + 4*math.Pow(eta11, 2)
	hu[2] = math.Pow(eta30-3*eta12, 2) + math.Pow(3*eta21-eta03, 2)
	hu[3] = math.Pow(eta30+eta12, 2) + math.Pow(eta21+eta03, 2)
	hu[4] = (eta30-3*eta12)*(eta30+eta12)*(math.Pow(eta30+eta12, 2)-3*math.Pow(eta21+eta03, 2)) +
		(3*eta21-eta03)*(eta21+eta03)*(3*math.Pow(eta30+eta12, 2)-math.Pow(eta21+eta03, 2))
	hu[5] = (eta20-eta02)*(math.Pow(eta30+eta12, 2)-math.Pow(eta21+eta03, 2)) +
		4*eta11*(eta30+eta12)*(eta21+eta03)
	hu[6] = (3*eta21-eta03)*(eta30+eta12)*(math.Pow(eta30+eta12, 2)-3*math.Pow(eta21+eta03, 2)) -
		(eta30-3*eta12)*(eta21+eta03)*(3*math.Pow(eta30+eta12, 2)-math.Pow(eta21+eta03, 2))

	return hu
}
