package regionHelper

import "github.com/bsthun/glyphcanvas/package/region"

func RegionComputeMoments(reg *region.Region) map[string]float64 {
	moments := make(map[string]float64)

	m00, m10, m01, m11, m20, m02, m21, m12, m30, m03 := 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0

	for x := uint16(0); x < reg.GetSizeX(); x++ {
		for y := uint16(0); y < reg.GetSizeY(); y++ {
			if reg.IsDrew(x, y) {
				fx := float64(x)
				fy := float64(y)

				m00 += 1
				m10 += fx
				m01 += fy
				m11 += fx * fy
				m20 += fx * fx
				m02 += fy * fy
				m21 += fx * fx * fy
				m12 += fx * fy * fy
				m30 += fx * fx * fx
				m03 += fy * fy * fy
			}
		}
	}

	moments["m00"] = m00
	moments["m10"] = m10
	moments["m01"] = m01
	moments["m11"] = m11
	moments["m20"] = m20
	moments["m02"] = m02
	moments["m21"] = m21
	moments["m12"] = m12
	moments["m30"] = m30
	moments["m03"] = m03

	if m00 > 0 {
		cx := m10 / m00
		cy := m01 / m00
		moments["cx"] = cx
		moments["cy"] = cy

		mu20 := m20 - cx*m10
		mu02 := m02 - cy*m01
		mu11 := m11 - cx*m01
		mu30 := m30 - 3*cx*m20 + 2*cx*cx*m10
		mu21 := m21 - 2*cx*m11 - cy*m20 + 2*cx*cx*m01
		mu12 := m12 - 2*cy*m11 - cx*m02 + 2*cy*cy*m10
		mu03 := m03 - 3*cy*m02 + 2*cy*cy*m01

		moments["mu20"] = mu20
		moments["mu02"] = mu02
		moments["mu11"] = mu11
		moments["mu30"] = mu30
		moments["mu21"] = mu21
		moments["mu12"] = mu12
		moments["mu03"] = mu03
	}

	return moments
}
