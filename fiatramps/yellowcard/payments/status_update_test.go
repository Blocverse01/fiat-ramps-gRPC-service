package payments

import (
	"fmt"
	"testing"

	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/ycstatus"
)

func TestApplyPaymentStatusUpdateRejectsUnknownAndTerminal(t *testing.T) {
	current := StoredPaymentRequestRecord{SequenceId: "azza-p-1", YellowCardStatus: ycstatus.Complete}
	load := func(sequenceId string) (StoredPaymentRequestRecord, error) {
		if sequenceId != current.SequenceId {
			return StoredPaymentRequestRecord{}, fmt.Errorf("not found")
		}
		return current, nil
	}
	patched := ""
	patch := func(sequenceId, status string) error {
		patched = status
		return nil
	}

	err := applyPaymentStatusUpdate("azza-p-1", "not-a-status", func(string) (string, error) {
		return "not-a-status", nil
	}, load, patch)
	if err == nil {
		t.Fatal("expected unknown status to be rejected")
	}

	err = applyPaymentStatusUpdate("azza-p-1", ycstatus.Failed, func(string) (string, error) {
		return ycstatus.Failed, nil
	}, load, patch)
	if err == nil {
		t.Fatal("expected transition out of complete to be rejected")
	}

	current.YellowCardStatus = ycstatus.Failed
	err = applyPaymentStatusUpdate("azza-p-1", ycstatus.Complete, func(string) (string, error) {
		return ycstatus.Complete, nil
	}, load, patch)
	if err == nil {
		t.Fatal("expected transition out of failed to be rejected")
	}

	current.YellowCardStatus = ycstatus.Processing
	if err := applyPaymentStatusUpdate("azza-p-1", ycstatus.Complete, func(string) (string, error) {
		return ycstatus.Complete, nil
	}, load, patch); err != nil {
		t.Fatalf("valid transition rejected: %v", err)
	}
	if patched != ycstatus.Complete {
		t.Fatalf("patch status = %q", patched)
	}
}
