package service

import (
	"reflect"
	"testing"

	"github.com/Duke1616/ecmdb/internal/role/internal/domain"
)

func TestFilterAccessibleModelUIDs(t *testing.T) {
	testCases := []struct {
		name      string
		roles     []domain.Role
		modelUIDs []string
		want      []string
	}{
		{
			name:      "role without exclusions can access every model",
			roles:     []domain.Role{{Code: "viewer"}},
			modelUIDs: []string{"host", "database"},
			want:      []string{"host", "database"},
		},
		{
			name:      "single role excludes selected models",
			roles:     []domain.Role{{Code: "viewer", DeniedModelUIDs: []string{"database"}}},
			modelUIDs: []string{"host", "database", "network"},
			want:      []string{"host", "network"},
		},
		{
			name: "multiple roles combine access with union semantics",
			roles: []domain.Role{
				{Code: "role-a", DeniedModelUIDs: []string{"host"}},
				{Code: "role-b", DeniedModelUIDs: []string{"database"}},
			},
			modelUIDs: []string{"host", "database", "network"},
			want:      []string{"host", "database", "network"},
		},
		{
			name: "model is hidden only when every role excludes it",
			roles: []domain.Role{
				{Code: "role-a", DeniedModelUIDs: []string{"host"}},
				{Code: "role-b", DeniedModelUIDs: []string{"host"}},
			},
			modelUIDs: []string{"host", "database"},
			want:      []string{"database"},
		},
		{
			name:      "empty roles grant no model access",
			roles:     nil,
			modelUIDs: []string{"host"},
			want:      []string{},
		},
		{
			name:      "candidate model identifiers are normalized and deduplicated",
			roles:     []domain.Role{{Code: "viewer"}},
			modelUIDs: []string{" host ", "host", "", "database"},
			want:      []string{"host", "database"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterAccessibleModelUIDs(tc.roles, tc.modelUIDs)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("filterAccessibleModelUIDs() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNormalizeModelUIDs(t *testing.T) {
	got := normalizeModelUIDs([]string{" host ", "host", "", "database"})
	want := []string{"host", "database"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeModelUIDs() = %v, want %v", got, want)
	}
}
