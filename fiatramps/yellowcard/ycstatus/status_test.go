package ycstatus

import "testing"

func TestRejectsUnknownStatus(t *testing.T) {
	if err := ValidateStatusWrite(Pending, "bogus", "bogus"); err == nil {
		t.Fatal("expected unknown status to be rejected")
	}
}

func TestRejectsTransitionOutOfComplete(t *testing.T) {
	if err := ValidateStatusWrite(Complete, Failed, Failed); err == nil {
		t.Fatal("expected transition out of complete to be rejected")
	}
}

func TestRejectsTransitionOutOfFailed(t *testing.T) {
	if err := ValidateStatusWrite(Failed, Complete, Complete); err == nil {
		t.Fatal("expected transition out of failed to be rejected")
	}
}

func TestRejectsCallerProviderMismatch(t *testing.T) {
	if err := ValidateStatusWrite(Pending, Complete, Processing); err == nil {
		t.Fatal("expected mismatch with provider status to be rejected")
	}
}

func TestAllowsValidTransitionMatchingProvider(t *testing.T) {
	if err := ValidateStatusWrite(Processing, Complete, Complete); err != nil {
		t.Fatalf("valid transition rejected: %v", err)
	}
}

func TestSameStatusIsNoOp(t *testing.T) {
	if err := ValidateStatusWrite(Complete, Complete, Complete); err != nil {
		t.Fatalf("same-status write should be allowed: %v", err)
	}
}
