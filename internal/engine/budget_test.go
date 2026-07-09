package engine

import "testing"

func TestLedgerEnforcesLimits(t *testing.T) {
	t.Parallel()
	ledger := NewLedger(2, 1)
	if err := ledger.Reserve(true); err != nil {
		t.Fatal(err)
	}
	if err := ledger.Reserve(true); err == nil {
		t.Fatal("second advisor call was accepted")
	}
	if err := ledger.Reserve(false); err != nil {
		t.Fatal(err)
	}
	if err := ledger.Reserve(false); err == nil {
		t.Fatal("third call was accepted")
	}
}
