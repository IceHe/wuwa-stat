package api

import (
	"reflect"
	"testing"
)

func TestSplitTacetCombinationDoubleClaimLevel8(t *testing.T) {
	got := splitTacetCombination(8, 7, 8, 2)
	want := [][2]int{{4, 4}, {3, 4}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitTacetCombination() = %v, want %v", got, want)
	}
}

func TestSplitTacetCombinationFallsBackWhenUnknown(t *testing.T) {
	got := splitTacetCombination(8, 9, 9, 2)
	want := [][2]int{{9, 9}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitTacetCombination() = %v, want %v", got, want)
	}
}

func TestNormalizeResonanceDropForStatsLevel8SingleClaim(t *testing.T) {
	gold, purple, blue, green := normalizeResonanceDropForStats(8, 1, 0, 2, 0, 7)
	if gold != 0 || purple != 2 || blue != 8 || green != 7 {
		t.Fatalf("normalizeResonanceDropForStats() = (%d, %d, %d, %d), want (0, 2, 8, 7)", gold, purple, blue, green)
	}
}

func TestNormalizeResonanceDropForStatsLevel8DoubleClaim(t *testing.T) {
	gold, purple, blue, green := normalizeResonanceDropForStats(8, 2, 1, 4, 12, 14)
	if gold != 1 || purple != 4 || blue != 16 || green != 14 {
		t.Fatalf("normalizeResonanceDropForStats() = (%d, %d, %d, %d), want (1, 4, 16, 14)", gold, purple, blue, green)
	}
}

func TestNormalizeResonanceDropForStatsOtherLevelsUnchanged(t *testing.T) {
	gold, purple, blue, green := normalizeResonanceDropForStats(7, 1, 1, 2, 6, 7)
	if gold != 1 || purple != 2 || blue != 6 || green != 7 {
		t.Fatalf("normalizeResonanceDropForStats() = (%d, %d, %d, %d), want (1, 2, 6, 7)", gold, purple, blue, green)
	}
}
