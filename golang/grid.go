package main

func preparePointDistances(positions []float64) []float64 {
	maxPos := positions[len(positions)-1]
	ld := int(maxPos+0.5) + 1

	closest := make([]int, ld, ld)
	distances := make([]float64, ld, ld)
	for i := 1; i < ld; i++ {
		closest[i] = int(maxPos + 0.5)
		distances[i] = maxPos
	}

	for _, position := range positions {
		idx := int(position + 0.5)
		closest[idx] = 0
		distances[idx] = position
		for i := 1; i < idx+1; i++ {
			if i < closest[idx-i] {
				closest[idx-i] = i
				distances[idx-i] = position
			} else {
				break
			}
		}

		for i := 1; i < ld-idx; i++ {
			if i < closest[idx+i] {
				closest[idx+i] = i
				distances[idx+i] = position
			} else {
				break
			}
		}
	}

	return distances
}
