package evaluation

import "testing"

func TestScore(t *testing.T) {
	t.Parallel()
	if !Score(Case{Exact: "READY"}, " READY\n") {
		t.Fatal("exact score failed")
	}
	if !Score(Case{Contains: "4"}, "2+2=4") {
		t.Fatal("contains score failed")
	}
	if Score(Case{Exact: "A"}, "B") {
		t.Fatal("mismatch passed")
	}
}
