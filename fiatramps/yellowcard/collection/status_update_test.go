package collection

import (
	"testing"

	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/ycstatus"
)

func TestApplyCollectionStatusUpdateRejectsUnknownAndTerminal(t *testing.T) {
	current := RequestRecord{SequenceId: "azza-c-1", YellowCardStatus: ycstatus.Complete}
	load := func(string) (RequestRecord, error) { return current, nil }
	patch := func(string, string) error { return nil }

	if err := applyCollectionStatusUpdate("azza-c-1", "nope", func(string) (string, error) {
		return "nope", nil
	}, load, patch); err == nil {
		t.Fatal("expected unknown status to be rejected")
	}

	if err := applyCollectionStatusUpdate("azza-c-1", ycstatus.Failed, func(string) (string, error) {
		return ycstatus.Failed, nil
	}, load, patch); err == nil {
		t.Fatal("expected transition out of complete to be rejected")
	}

	current.YellowCardStatus = ycstatus.Failed
	if err := applyCollectionStatusUpdate("azza-c-1", ycstatus.Pending, func(string) (string, error) {
		return ycstatus.Pending, nil
	}, load, patch); err == nil {
		t.Fatal("expected transition out of failed to be rejected")
	}
}
