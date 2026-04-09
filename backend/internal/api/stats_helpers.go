package api

import "sort"

var tacetSingleCombos = map[int][][2]int{
	8: {{4, 4}, {3, 4}},
	7: {{4, 4}, {4, 3}, {3, 4}, {3, 3}},
	6: {{4, 4}, {4, 3}, {3, 4}, {3, 3}},
	5: {{3, 6}, {3, 5}, {2, 6}, {2, 5}},
}

func splitTacetCombination(solaLevel, goldTubes, purpleTubes, claimCount int) [][2]int {
	if claimCount <= 1 {
		return [][2]int{{goldTubes, purpleTubes}}
	}

	// For double-claim rows, split the merged total into two plausible single-claim combinations.
	// This keeps historical statistics comparable with old data where single claims were recorded directly.
	combos := tacetSingleCombos[solaLevel]
	var matching [][2][2]int
	for _, left := range combos {
		for _, right := range combos {
			if left[0]+right[0] == goldTubes && left[1]+right[1] == purpleTubes {
				pair := [2][2]int{left, right}
				if lessCombo(pair[0], pair[1]) {
					pair[0], pair[1] = pair[1], pair[0]
				}
				matching = append(matching, pair)
			}
		}
	}

	if len(matching) == 0 {
		// If no known split exists for this level, fall back to aggregated values to avoid dropping data.
		return [][2]int{{goldTubes, purpleTubes}}
	}

	sort.Slice(matching, func(i, j int) bool {
		if matching[i][0][0] != matching[j][0][0] {
			return matching[i][0][0] > matching[j][0][0]
		}
		if matching[i][0][1] != matching[j][0][1] {
			return matching[i][0][1] > matching[j][0][1]
		}
		if matching[i][1][0] != matching[j][1][0] {
			return matching[i][1][0] > matching[j][1][0]
		}
		return matching[i][1][1] > matching[j][1][1]
	})

	return [][2]int{matching[0][0], matching[0][1]}
}

func lessCombo(left, right [2]int) bool {
	if left[0] != right[0] {
		return left[0] < right[0]
	}
	return left[1] < right[1]
}

func normalizeResonanceDropForStats(
	solaLevel, claimCount, gold, purple, blue, green int,
) (int, int, int, int) {
	if solaLevel == 8 {
		if claimCount <= 1 {
			blue = 8
		} else if claimCount == 2 {
			blue = 16
		}
	}

	return gold, purple, blue, green
}
