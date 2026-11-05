package payments

import "testing"

func TestRecomputeLocalAmount(t *testing.T) {
	if got := RecomputeLocalAmount(10, 1500); got != 15000 {
		t.Fatalf("got %v", got)
	}
	if LocalAmountMatchesVerifiedDeposit(1, 10, 1500) {
		t.Fatal("mismatched amount should not match")
	}
	if !LocalAmountMatchesVerifiedDeposit(15000, 10, 1500) {
		t.Fatal("expected match")
	}
}
