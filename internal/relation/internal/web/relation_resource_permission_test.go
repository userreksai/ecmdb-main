package web

import (
	"reflect"
	"testing"

	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
)

func TestRelationModelUIDs(t *testing.T) {
	got, err := relationModelUIDs("host_default_database")
	if err != nil {
		t.Fatalf("relationModelUIDs() error = %v", err)
	}
	want := []string{"host", "database"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("relationModelUIDs() = %v, want %v", got, want)
	}

	if _, err = relationModelUIDs("invalid_relation"); err == nil {
		t.Fatal("relationModelUIDs() expected invalid relation error")
	}
}

func TestFilterRelationsByAllowed(t *testing.T) {
	relations := []domain.ResourceRelation{
		{ID: 1, SourceModelUID: "host", TargetModelUID: "database"},
		{ID: 2, SourceModelUID: "host", TargetModelUID: "network"},
	}
	allowed := map[string]struct{}{"host": {}, "database": {}}

	got := filterRelationsByAllowed(relations, allowed)
	if len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("filterRelationsByAllowed() = %v, want relation 1", got)
	}
}

func TestFilterAggregatedByAllowed(t *testing.T) {
	items := []domain.ResourceAggregatedAssets{
		{ModelUid: "database", Total: 2},
		{ModelUid: "network", Total: 3},
	}
	allowed := map[string]struct{}{"database": {}}

	got := filterAggregatedByAllowed(items, allowed)
	if len(got) != 1 || got[0].ModelUid != "database" {
		t.Fatalf("filterAggregatedByAllowed() = %v, want database only", got)
	}
}
