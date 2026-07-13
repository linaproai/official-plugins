// This file tests pure helpers in the mail service package.

package mail

import "testing"

func TestNormalizePage(t *testing.T) {
	pageNum, pageSize := normalizePage(0, 0)
	if pageNum != 1 || pageSize != 20 {
		t.Fatalf("defaults: pageNum=%d pageSize=%d", pageNum, pageSize)
	}
	pageNum, pageSize = normalizePage(2, 500)
	if pageNum != 2 || pageSize != 200 {
		t.Fatalf("cap: pageNum=%d pageSize=%d", pageNum, pageSize)
	}
}

func TestUniquePositiveIDs(t *testing.T) {
	got := uniquePositiveIDs([]int64{0, -1, 2, 2, 3})
	if len(got) != 2 || got[0] != 2 || got[1] != 3 {
		t.Fatalf("unexpected ids: %#v", got)
	}
}

func TestValidateConnectionInput(t *testing.T) {
	if err := validateConnectionInput("", "smtp", "smtp.example.com", 587, "", ""); err == nil {
		t.Fatal("expected name required")
	}
	if err := validateConnectionInput("main", "bad", "smtp.example.com", 587, "", ""); err == nil {
		t.Fatal("expected kind invalid")
	}
	if err := validateConnectionInput("main", "smtp", "smtp.example.com", 587, "", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
