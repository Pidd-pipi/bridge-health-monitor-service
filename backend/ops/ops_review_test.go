package ops

import (
	"context"
	"testing"
)

func TestEventReviewFlow(t *testing.T) {
	svc := serviceWithSeed(t)
	rec, err := svc.Review(context.Background(), "evt-1001", 3, "质检员")
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	if rec.Status != OpsStatusReviewing {
		t.Fatalf("review status=%s want reviewing", rec.Status)
	}
	closed, err := svc.CloseReviewed(context.Background(), "evt-1001", rec.Revision, "质检员")
	if err != nil {
		t.Fatalf("close reviewed: %v", err)
	}
	if closed.Status != OpsStatusClosed {
		t.Fatalf("close status=%s want closed", closed.Status)
	}
}

func TestEventReviewWriteBackCorrectStatus(t *testing.T) {
	svc := serviceWithSeed(t)
	_, err := svc.Review(context.Background(), "evt-1001", 3, "质检员")
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	got, _ := svc.Get(context.Background(), "evt-1001")
	if got.Status != OpsStatusReviewing {
		t.Fatalf("stored status=%s want reviewing (write-back wrong)", got.Status)
	}
}

func TestEventReviewWrongRevisionConflict(t *testing.T) {
	svc := serviceWithSeed(t)
	if _, err := svc.Review(context.Background(), "evt-1001", 1, "质检员"); err != ErrOpsConflict {
		t.Fatalf("stale review: err=%v want conflict", err)
	}
}

func TestEventCloseReviewedRequiresReviewing(t *testing.T) {
	svc := serviceWithSeed(t)
	if _, err := svc.CloseReviewed(context.Background(), "evt-1002", 1, "质检员"); err == nil {
		t.Fatal("closing a queued record without review should fail")
	}
}
