package ops

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorSentinelClassification(t *testing.T) {
	wrapped := wrapOps("not_found", "store.get", ErrOpsNotFound)
	if !errors.Is(wrapped, ErrOpsNotFound) {
		t.Fatal("wrapped error should be recognizable by errors.Is")
	}
	if opsCode(wrapped) != "not_found" {
		t.Fatalf("code=%s want not_found", opsCode(wrapped))
	}
}

func TestErrorUnwrapChain(t *testing.T) {
	inner := wrapOps("conflict", "store.update", ErrOpsConflict)
	outer := fmt.Errorf("outer: %w", inner)
	if !errors.Is(outer, ErrOpsConflict) {
		t.Fatal("nested wrap chain broken")
	}
	if opsCode(outer) != "conflict" {
		t.Fatalf("code=%s want conflict", opsCode(outer))
	}
}

func TestErrorDirectValue(t *testing.T) {
	if opsCode(ErrOpsNotFound) != "not_found" {
		t.Fatalf("direct sentinel code=%s", opsCode(ErrOpsNotFound))
	}
	if opsCode(ErrOpsPolicy) != "policy" {
		t.Fatalf("policy code=%s", opsCode(ErrOpsPolicy))
	}
}

func TestErrorUnknownCode(t *testing.T) {
	if opsCode(errors.New("random")) != "internal" {
		t.Fatal("unknown error should map to internal")
	}
}
