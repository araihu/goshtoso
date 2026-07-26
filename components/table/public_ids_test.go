package table_test

import (
	"testing"

	"github.com/araihu/goshtoso/components/table"
)

func TestConfigPublishesFragmentTargetIDs(t *testing.T) {
	cfg := table.Config{ID: "resources"}

	if got := cfg.TbodyID(); got != "resources-tbody" {
		t.Fatalf("TbodyID() = %q, want resources-tbody", got)
	}
	if got := cfg.TheadID(); got != "resources-thead" {
		t.Fatalf("TheadID() = %q, want resources-thead", got)
	}
	if got := cfg.PaginationID(); got != "resources-pagination" {
		t.Fatalf("PaginationID() = %q, want resources-pagination", got)
	}
}
